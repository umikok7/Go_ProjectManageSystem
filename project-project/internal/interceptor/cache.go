package interceptor

import (
	"context"
	"encoding/json"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"test.com/project-common/encrypts"
	"test.com/project-grpc/project"
	"test.com/project-project/internal/dao"
	"test.com/project-project/internal/repo"
	"time"
)

type CacheInterceptor struct {
	cache    repo.Cache
	cacheMap map[string]any
}

type CacheRespOption struct {
	path   string
	typ    any
	expire time.Duration
}

func New() *CacheInterceptor {
	cacheMap := make(map[string]any)
	cacheMap["/login.service.v1.ProjectService/FindProjectByMemId"] = &project.MyProjectResponse{}
	return &CacheInterceptor{
		cache:    dao.Rc,
		cacheMap: cacheMap,
	}
}

func (c *CacheInterceptor) CacheInterceptor() func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
	// 创建拦截器
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		respType := c.cacheMap[info.FullMethod]
		if respType == nil {
			// 不在拦截的范围之内则直接调用实际的处理函数
			return handler(ctx, req)
		}
		// 先查询是否有缓存，有则直接返回缓存的内容，没有则先请求再存入缓存
		con, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		marshal, _ := json.Marshal(req)
		cacheKey := encrypts.Md5(string(marshal))

		respJson, _ := c.cache.Get(con, info.FullMethod+"::"+cacheKey)
		if respJson != "" {
			json.Unmarshal([]byte(respJson), &respType)
			zap.L().Info(info.FullMethod + "走了缓存")
			return respType, nil
		}

		resp, err = handler(ctx, req) // 缓存未命中，则调用实际的处理函数
		bytes, _ := json.Marshal(resp)
		c.cache.Put(con, info.FullMethod+"::"+cacheKey, string(bytes), 5*time.Minute)
		zap.L().Info(info.FullMethod + "放入缓存")
		return
	}
}
