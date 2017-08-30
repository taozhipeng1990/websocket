# qm WebSocket

qm WebSocket 是整合Gorilla WebSocket的一个集成，方便快速开发websocket服务

### Installation

    go get github.com/taozhipeng1990/websocket


### Documentation
	
	server := websocket.Server{}//相关于实例化一个Websocket服务端的类
	server.Create("0.0.0.0", 1234, "/echo",5*60)//创建websocket 第4个参数是心跳检测时间间隔
	//注册连接时的函数 isopen为false为阻止客户端连接 可以在这里做登陆验证
	server.OnOpen = func(r *http.Request) (isopen bool, code int) {
		isopen = true
		code = 101
		return
	}
	//接收来自客户端的消息
	server.OnMessage = func(fd int64, data []byte) {
		fmt.Printf("收到来自%d客户端的消息:%s\n", fd, string(data))
		server.Send(fd, []byte("我也给你发消息"))
	}
	//客户端关闭
	server.OnClose = func(fd int64) {
		fmt.Printf("客户端%d关闭了\n", fd)
	}


	//server.Runing.GetInfo()返回服务运行时的一些相关信息

	//启动websocket
	server.Start()