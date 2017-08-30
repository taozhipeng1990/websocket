package websocket

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
)

type config struct {
	ip      string
	port    int
	routing string
}

//Server websocket服务端
type Server struct {
	Clients   map[int64]*websocket.Conn
	autoFd    int64
	upgrader  websocket.Upgrader
	Config    config
	OnOpen    func(r *http.Request) (isopen bool, code int)
	OnMessage func(fd int64, data []byte)
	OnClose   func(fd int64)
}

//Create 创建一个websocket服务端
func (serv *Server) Create(ip string, port int, routing string) {
	conf := config{ip, port, routing}
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
	go serv.receive(fd, client)
}

//receive 内部处理 接受客户端发来的消息
func (serv *Server) receive(fd int64, client *websocket.Conn) {
	for {
		_, data, err := client.ReadMessage()
		if err != nil {
			serv.Close(fd)
			break
		}
		if serv.OnMessage != nil {
			serv.OnMessage(fd, data)
		}
	}
}

//Close 关闭客户端
func (serv *Server) Close(fd int64) {
	serv.Clients[fd].Close()
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
