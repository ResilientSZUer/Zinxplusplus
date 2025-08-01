package ziface

/*
IRouter 路由接口
使用框架者给该连接自定义的处理业务方法。
路由里的 IRequest 则包含该连接的连接信息和请求数据信息。
*/
type IRouter interface {
	PreHandle(request IRequest)

	Handle(request IRequest)

	PostHandle(request IRequest)
}
