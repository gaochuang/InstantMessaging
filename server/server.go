package server

import (
	"IM/user"
	"fmt"
	"io"
	"net"
	"strings"
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

func (this *Server) SendMsg(user *user.User, msg string) {
	user.Conn.Write([]byte(msg))
}

func (this *Server) HandlerMessage(msg string, user *user.User) {
	if msg == "who" {
		this.MapLock.Lock()
		for _, user := range this.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "online"
			this.SendMsg(user, onlineMsg)
		}
		this.MapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {

		newName := strings.Split(msg, "|")[1]

		_, ok := this.OnlineMap[newName]
		if ok {
			this.SendMsg(user, "user already exist")
		} else {
			this.MapLock.Lock()
			delete(this.OnlineMap, user.Name)
			user.Name = newName
			this.OnlineMap[newName] = user
			this.MapLock.Unlock()

			this.SendMsg(user, "update user name successful "+newName+"\n")

		}
	} else {

		//广播
		this.BroadCast(user, msg)
	}
}
func (this *Server) serverHandle(conn net.Conn) {
	user := user.NewUser(conn)
	this.MapLock.Lock()
	this.OnlineMap[user.Name] = user
	this.MapLock.Unlock()

	this.BroadCast(user, "online now")

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				this.BroadCast(user, "offline now")
				return
			}
			//EOF文件末尾
			if err != nil && err != io.EOF {
				fmt.Println("read failed, err: ", err)
				return
			}
			//去掉\n
			msg := string(buf[:n-1])
			this.HandlerMessage(msg, user)
		}

	}()

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
