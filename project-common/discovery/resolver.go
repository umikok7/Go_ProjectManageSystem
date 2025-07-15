package discovery

import (
	"context"
	"strings"
	"time"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"google.golang.org/grpc/resolver"
)

const (
	schema = "etcd"
)

// Resolver for grpc client
type Resolver struct {
	schema      string // 协议名称
	EtcdAddrs   []string
	DialTimeout int

	closeCh      chan struct{}
	watchCh      clientv3.WatchChan
	cli          *clientv3.Client
	keyPrifix    string             // 服务查询前缀
	srvAddrsList []resolver.Address // 可用服务地址列表

	cc     resolver.ClientConn // gRPC 连接
	logger *zap.Logger
}

// NewResolver create a new resolver.Builder base on etcd
func NewResolver(etcdAddrs []string, logger *zap.Logger) *Resolver {
	return &Resolver{
		schema:      schema,
		EtcdAddrs:   etcdAddrs,
		DialTimeout: 3,
		logger:      logger,
	}
}

// Scheme returns the scheme supported by this resolver.
func (r *Resolver) Scheme() string {
	return r.schema
}

// Build creates a new resolver.Resolver for the given target （构建解析器）
func (r *Resolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	// r.cc = cc

	// // 使用新版本 Target 的方法和字段
	// serviceName := target.Endpoint() // 使用内置的 Endpoint() 方法获取服务名
	// version := target.URL.Scheme     // 从 URL.Scheme 获取版本信息

	// // 如果没有版本信息，使用默认版本
	// if version == "" {
	// 	version = "v1"
	// }

	// // 如果服务名为空，尝试从其他字段获取
	// if serviceName == "" {
	// 	serviceName = target.URL.Host
	// }

	// // 确保有默认的服务名
	// if serviceName == "" {
	// 	serviceName = "default-service"
	// }

	// r.keyPrifix = BuildPrefix(Server{Name: serviceName, Version: version})
	// if _, err := r.start(); err != nil {
	// 	return nil, err
	// }
	// return r, nil

	r.cc = cc
	// 修复服务名解析逻辑
	serviceName := target.Endpoint()
	if serviceName == "" {
		// 从 URL 路径中提取服务名
		path := strings.TrimPrefix(target.URL.Path, "/")
		if path != "" {
			serviceName = path
		} else {
			serviceName = target.URL.Host
		}
		// 如果还是空，使用默认服务名
		if serviceName == "" {
			serviceName = "user" // 默认服务名
		}
	}

	// 使用固定版本
	version := "1.0.0"
	r.keyPrifix = BuildPrefix(Server{Name: serviceName, Version: version})
	if _, err := r.start(); err != nil {
		r.logger.Error("启动 resolver 失败", zap.Error(err))
		return nil, err
	}
	r.logger.Info("Resolver 构建成功", zap.String("serviceName", serviceName),
		zap.String("keyPrefix", r.keyPrifix))

	return r, nil
}

// ResolveNow resolver.Resolver interface
func (r *Resolver) ResolveNow(o resolver.ResolveNowOptions) {}

// Close resolver.Resolver interface
func (r *Resolver) Close() {
	r.closeCh <- struct{}{}
}

// start
func (r *Resolver) start() (chan<- struct{}, error) {
	var err error
	r.cli, err = clientv3.New(clientv3.Config{
		Endpoints:   r.EtcdAddrs,
		DialTimeout: time.Duration(r.DialTimeout) * time.Second,
	})
	if err != nil {
		return nil, err
	}
	resolver.Register(r)

	r.closeCh = make(chan struct{})

	if err = r.sync(); err != nil {
		return nil, err
	}

	go r.watch()

	return r.closeCh, nil
}

// watch update events
func (r *Resolver) watch() {
	ticker := time.NewTicker(time.Minute)
	r.watchCh = r.cli.Watch(context.Background(), r.keyPrifix, clientv3.WithPrefix())

	for {
		select {
		case <-r.closeCh:
			return
		case res, ok := <-r.watchCh:
			if ok {
				r.update(res.Events)
			}
		case <-ticker.C:
			if err := r.sync(); err != nil {
				r.logger.Error("sync failed", zap.Error(err))
			}
		}
	}
}

// update
func (r *Resolver) update(events []*clientv3.Event) {
	for _, ev := range events {
		var info Server
		var err error

		switch ev.Type {
		case mvccpb.PUT:
			info, err = ParseValue(ev.Kv.Value)
			if err != nil {
				continue
			}
			addr := resolver.Address{Addr: info.Addr, Metadata: info.Weight}
			if !Exist(r.srvAddrsList, addr) {
				r.srvAddrsList = append(r.srvAddrsList, addr)
				r.cc.UpdateState(resolver.State{Addresses: r.srvAddrsList})
			}
		case mvccpb.DELETE:
			info, err = SplitPath(string(ev.Kv.Key))
			if err != nil {
				continue
			}
			addr := resolver.Address{Addr: info.Addr}
			if s, ok := Remove(r.srvAddrsList, addr); ok {
				r.srvAddrsList = s
				r.cc.UpdateState(resolver.State{Addresses: r.srvAddrsList})
			}
		}
	}
}

// sync 同步获取所有地址信息
func (r *Resolver) sync() error {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	res, err := r.cli.Get(ctx, r.keyPrifix, clientv3.WithPrefix())
	if err != nil {
		return err
	}
	r.srvAddrsList = []resolver.Address{}

	for _, v := range res.Kvs {
		info, err := ParseValue(v.Value)
		if err != nil {
			continue
		}
		addr := resolver.Address{Addr: info.Addr, Metadata: info.Weight}
		r.srvAddrsList = append(r.srvAddrsList, addr)
	}
	r.cc.UpdateState(resolver.State{Addresses: r.srvAddrsList})
	return nil
}
