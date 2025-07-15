# 项目疑惑点总结

## 1. 前后端分离框架的理解

如下图所展示，理解一下前后端是如何协作的
![前后端分离框架.png](%E5%89%8D%E5%90%8E%E7%AB%AF%E5%88%86%E7%A6%BB%E6%A1%86%E6%9E%B6.png)

关键理解点：
- 8045端口：Vue开发服务器监听，提供前端资源，用于浏览器进行访问
- 80端口：project-api文件夹监听，提供后端的api

总结来讲就是说：前端监听8045端口是为了提供开发环境， 但前端的JavaScript代码会主动向
80端口发起HTTP请求来获取后端数据。

## 2. gRPC是如何服务本项目的

首先通过下图理解项目协作的框架


![gRPC服务.png](gRPC%E6%9C%8D%E5%8A%A1.png)

### 2.1 协作架构

服务角色划分
- project-user: gRPC 服务端，提供用户相关的业务逻辑服务
- project-api: gRPC 客户端 + HTTP API 网关，对外提供 REST API，内部调用 gRPC 服务

### 2.2 gRPC协作流程

#### gRPC 服务定义
在 `login_service.proto` 中定义服务接口（通过 protobuf 生成代码），
生成了 `login_service_grpc.pb.go`。

#### 服务端实现 (project-user)
1. **服务注册**: 在 router.go 的 `RegisterGrpc` 函数中：
   ```go
   loginServiceV1.RegisterLoginServiceServer(g, loginServiceV1.New())
   ```

2. **服务启动**: 启动 gRPC 服务器监听 `8080` 端口
   ```go
   lis, err := net.Listen("tcp", c.Addr) // 创建监听
   s.Serve(lis) // 启动gRPC服务
   ```

3. **业务实现** :在 login.service.v1 目录下实现具体的业务逻辑，包括 GetCaptcha 等方法。

#### 客户端调用 (project-api)
1. **客户端初始化**: 在 rpc.go 中初始化 gRPC 客户端：
   ```go
   // 此处的8881端口在配置文件可以看到是用于gRPC服务的端口，注意与8080或者80端口区分
   conn, err := grpc.Dial("127.0.0.1:8881", grpc.WithTransportCredentials(insecure.NewCredentials()))
   LoginServiceClient = loginServiceV1.NewLoginServiceClient(conn)
   ```

2. **API 层调用**: 在 user.go 的 `getCaptcha` 方法中：
   ```go
   _, err := LoginServiceClient.GetCaptcha(c, &loginServiceV1.CaptchaMessage{Mobile: mobile})
   ```

### 3. 完整的请求流程

```
前端HTTP请求 (localhost:8045)
    ↓
project-api HTTP接收 (端口80)
    ↓
project-api/api/user/user.go::getCaptcha (HTTP处理)
    ↓
gRPC调用 LoginServiceClient.GetCaptcha
    ↓
project-user gRPC服务接收 (端口8881)
    ↓
project-user 业务逻辑处理 (pkg/service/login.service.v1/)
    ↓
gRPC响应返回 (端口8881 → 端口80)
    ↓
project-api HTTP响应返回给前端
```

具体步骤：
1. 前端发起请求: 浏览器从 localhost:8045 加载页面后，JavaScript 向 127.0.0.1:80 发送 HTTP POST 请求
2. API网关接收: project-api目录下的user.go 的 getCaptcha 方法接收 HTTP 请求
3. gRPC调用: 通过 LoginServiceClient 调用 GetCaptcha gRPC 方法
4. 服务端处理: project-user 的 gRPC 服务器（端口8881）接收请求
5. 业务逻辑: 在 login.service.v1 中执行验证码生成、存储等业务逻辑
6. 响应返回: gRPC 响应返回给 project-api
7. HTTP响应: project-api 将 gRPC 响应转换为 HTTP JSON 响应返回给前端


关键端口补充说明：

| 服务 | 端口 | 协议 | 用途 |
|------|------|------|------|
| 前端开发服务器 | 8045 | HTTP | 提供前端页面和静态资源 |
| project-api | 80 | HTTP | REST API 网关 |
| project-user | 8881 | gRPC | 用户业务服务 |

- 这种架构实现了微服务间的解耦，project-api 作为网关层处理 HTTP 请求，
  project-user 专注于用户业务逻辑，通过 gRPC 实现高效的内部通信。

- 协议分离: 前端使用 HTTP，内部服务使用高效的 gRPC

# pprof 压力测试

## 步骤总结
基本命令：`go run ./main.go --mem=mem.pprof`，以及`go tool pprof ./mem.pprof`

* 通过监控平台监测到内存或cpu问题。 (http://127.0.0.1/debug/pprof/下进行查看)
* 通过浏览器方式大致判断是哪些可能的问题。 (-http模式)
* 通过命令行方式抓取几个时间点的profile 
* 使用`web`命令查看函数调用图
* 使用`top` 、`traces`、`list` 命令定位问题
* 如果出现了goroutine泄漏或者内存泄漏等随着时间持续增长的问题，`go tool pprof -base`比较两个不同时间点的状态更方便我们定位问题。

# 压力测试

可进入以下连接进行学习
```
https://github.com/link1st/go-stress-testing?tab=readme-ov-file#12-项目体验
```


# 一些有意思的实现

## struct{}{}
体会go中set数据结构的实现

```go
func ToAuthNodeTreeList(list []*ProjectNode, checkedList []string) []*ProjectNodeAuthTree {
	checkedMap := make(map[string]struct{})
	for _, v := range checkedList {
		checkedMap[v] = struct{}{} // 空结构体类型的零值实例，优势：每个值占用 0 字节内存，这样的操作实现了set数据结构
	}
	var roots []*ProjectNodeAuthTree
	for _, v := range list {
		paths := strings.Split(v.Node, "/")
		if len(paths) == 1 {
			// 检查该节点是否已授权
			checked := false
			if _, ok := checkedMap[v.Node]; ok {
				checked = true
			}
			//根节点
			root := &ProjectNodeAuthTree{
				Id:       v.Id,
				Node:     v.Node,
				Pnode:    "",
				IsLogin:  v.IsLogin,
				IsMenu:   v.IsMenu,
				IsAuth:   v.IsAuth,
				Title:    v.Title,
				Children: []*ProjectNodeAuthTree{},
				Checked:  checked, // 权限状态
				Key:      v.Node,  // 用于前端组件的 key
			}
			roots = append(roots, root)
		}
	}
	for _, v := range roots {
		addAuthNodeChild(list, v, 2, checkedMap)
	}
	return roots
}
```


```go
// 假设 checkedList = ["project", "project/task", "project/task/list"]
checkedMap = map[string]struct{}{
    "project":           {},
    "project/task":      {},
    "project/task/list": {},
}

// 检查权限：O(1) 时间复杂度
_, hasPermission := checkedMap["project/task"]  // true
_, hasPermission := checkedMap["project/member"]  // false
```


struct{}{} 的核心价值：

- 零内存占用：完美适合 Set 数据结构
- 语义清晰：表达"只关心键的存在，不关心值"
- 性能优异：提供 O(1) 查找复杂度
- 内存友好：在大量数据时显著节省内存


# Nacos
docker启动后需要访问：http://your_ip:8848/nacos/index.html 进行同一的服务配置

# Jaeger
docker启动后在：http://your_ip:16686/search 中进行链路追踪

# ELK日志采集系统

- E: ElasticSearch 是一种NoSQL数据库，支持全文索引
- L：负责消费Kafka消息队列中的原始数据，并将消费的数据上报到ElasticSearch进行存储
- K：负责可视化ElasticSearch中存储的数据，并提供查询、聚合、图表、导出等功能