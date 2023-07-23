package main

import (
	"IM/server"
	"flag"
	"fmt"
)

func main() {

	var ip string
	var port int

	flag.IntVar(&port, "port", 8888, "server port")
	flag.StringVar(&ip, "ip", "127.0.0.1", "server ip address")
	flag.Parse()

	fmt.Println("server ip: ", ip, " port: ", port)
	newServer := server.NewServer(ip, port)

	newServer.StartServer()
}
