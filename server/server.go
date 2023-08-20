package server

import (
	"IM/user"
	"fmt"
	"io"
	"net"
	"runtime"
	"strings"
	"sync"
	"time"
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
	} else if len(msg) > 4 && msg[:3] == "to|" {
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			this.SendMsg(user, "message format error, please use \"to|Bob|hello\"format\n")
			return
		}
		remoteUser, ok := this.OnlineMap[remoteName]
		if !ok {
			this.SendMsg(user, "user doesn't exit")
			return
		}

		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.SendMsg(user, "the message is empty, please send again")
			return
		}

		msg := user.Name + "said to you" + content
		remoteUser.Conn.Write([]byte(msg))
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

	isLive := make(chan bool)

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
			isLive <- true
		}

	}()

	for {
		select {

		//isLive必须在上面，因为isLive不触发，设置定时器也会执行
		case <-isLive:
			//当前用户活跃应该重置定时器，为了激活select更新定时器

		case <-time.After(10 * time.Second):
			//定时器超时，将用户提出
			this.SendMsg(user, "you were forced offline")
			close(user.C)
			conn.Close()
			runtime.Goexit()
		}
	}

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
