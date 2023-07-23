package server

import (
	"fmt"
	"net"
)

type Server struct {
	IP   string
	port int
}

func NewServer(ip string, port int) *Server {
	return &Server{IP: ip, port: port}
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
