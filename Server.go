package websocket

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

type config struct {
	ip      string
	port    int
	routing string
	timeout time.Duration
}

type runing struct {
	starttime time.Time
	run       time.Duration
	sum       int64
	current   int64
	close     int64
}

//GetInfo 获取服务执行的一些信息
func (info runing) GetInfo() string {
	info.run = time.Now().Sub(info.starttime)
	return fmt.Sprintf("启动时间:%s\n运行时间:%s\n总连接数:%d\n当前连接数:%d\n关闭连接数:%d\n", info.starttime, info.run, info.sum, info.current, info.close)
}

//Server websocket服务端
type Server struct {
	Clients   map[int64]*websocket.Conn
	autoFd    int64
	upgrader  websocket.Upgrader
	Runing    runing
	Config    config
	OnOpen    func(r *http.Request) (isopen bool, code int)
	OnMessage func(fd int64, data []byte)
	OnClose   func(fd int64)
}

//Create 创建一个websocket服务端
//ip 0.0.0.0
//port 1234
//routing /
//timeout 秒为单位 心跳检测的时间间隔 N秒内，如果客户端没有发来消息就踢下线
func (serv *Server) Create(ip string, port int, routing string, timeout time.Duration) {
	conf := config{ip, port, routing, timeout}
	serv.Config = conf
	serv.upgrader = websocket.Upgrader{
		ReadBufferSize:  4096,
		WriteBufferSize: 4096,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	serv.Clients = make(map[int64]*websocket.Conn)
}

//Start 启动websocket服务端
func (serv *Server) Start() {
	http.HandleFunc(serv.Config.routing, serv.connect)
	port := strconv.Itoa(serv.Config.port)
	addr := bytes.Buffer{}
	addr.WriteString(":")
	addr.WriteString(port)
	serv.Runing.starttime = time.Now()
	fmt.Printf("websocket启动 监听%s端口\n", port)
	http.ListenAndServe(addr.String(), nil)
}

//Send 发送消息给某fd客户端
func (serv *Server) Send(fd int64, data []byte) error {
	return serv.Clients[fd].WriteMessage(1, data)
}

//connect 内部处理 客户端连接
func (serv *Server) connect(w http.ResponseWriter, r *http.Request) {
	if serv.OnOpen != nil {
		isopen, code := serv.OnOpen(r)
		if isopen == false {
			w.WriteHeader(code)
			return
		}
	}

	client, err := serv.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	fd := serv.createFd()
	serv.Clients[fd] = client
	serv.Runing.sum++
	serv.Runing.current++
	go serv.receive(fd, client)
}

//receive 内部处理 接受客户端发来的消息
func (serv *Server) receive(fd int64, client *websocket.Conn) {
	timeout := serv.Config.timeout * time.Second
	t1 := time.NewTimer(timeout)
	c1 := make(chan []byte)

	go func() {
		for {
			_, data, err := client.ReadMessage()
			if err != nil {
				serv.Close(fd)
				break
			}
			if serv.OnMessage != nil {
				serv.OnMessage(fd, data)
			}
			c1 <- data
		}
	}()

	for {
		select {
		case <-c1:
			t1.Reset(timeout)
		case <-t1.C:
			serv.Close(fd)
			fmt.Println("超时关闭\n")
			break
		}
	}
}

//Close 关闭客户端
func (serv *Server) Close(fd int64) {

	if serv.Clients[fd] == nil {
		return
	}
	serv.Clients[fd].Close()
	serv.Runing.current--
	serv.Runing.close++
	delete(serv.Clients, fd)
	if serv.OnClose != nil {
		serv.OnClose(fd)
	}
}

//createFd 创建fd
func (serv *Server) createFd() int64 {
	serv.autoFd = serv.autoFd + 1
	return serv.autoFd
}
