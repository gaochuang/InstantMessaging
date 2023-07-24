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

func (this *Server) serverHandle(conn net.Conn) {

}
func (this *Server) StartServer() error {
	server, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.IP, this.port))
	if err != nil {
		fmt.Println("create server failed, error reason: ", err.Error())
		return err
	}

	defer server.Close()

	for {
		conn, err := server.Accept()
		if err != nil {
			fmt.Println("Accept failed, reason: ", err.Error())
			continue
		}

		go this.serverHandle(conn)
	}
}
