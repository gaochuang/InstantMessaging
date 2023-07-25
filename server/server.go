package server

import (
	"IM/user"
	"fmt"
	"net"
	"sync"
)

type Server struct {
	IP   string
	port int

	//online user
	OnlineMap map[string]*user.User
	MapLock   sync.RWMutex

	//广播channel
	Message chan string
}

func NewServer(ip string, port int) *Server {

	server := &Server{IP: ip, port: port,
		OnlineMap: make(map[string]*user.User),
		Message:   make(chan string),
	}

	return server
}

func (this *Server) ListenMessage() {
	for {
		msg := <-this.Message

		this.MapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.MapLock.Unlock()
	}
}

func (this *Server) BroadCast(user *user.User, msg string) {

	sendMsg := "[" + user.Name + "]" + ":" + msg

	this.Message <- sendMsg
}

func (this *Server) serverHandle(conn net.Conn) {
	user := user.NewUser(conn)
	this.MapLock.Lock()
	this.OnlineMap[user.Name] = user
	this.MapLock.Unlock()

	this.BroadCast(user, "online now")

	select {}
}

func (this *Server) StartServer() error {
	server, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.IP, this.port))
	if err != nil {
		fmt.Println("create server failed, error reason: ", err.Error())
		return err
	}

	defer server.Close()

	//listen message
	go this.ListenMessage()

	for {
		conn, err := server.Accept()
		if err != nil {
			fmt.Println("Accept failed, reason: ", err.Error())
			continue
		}

		go this.serverHandle(conn)
	}
}
