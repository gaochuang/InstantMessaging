package user

import "net"

type User struct {
	Name string
	Addr string
	Conn net.Conn
	C    chan string
}

func NewUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String()
	user := User{
		Name: userAddr,
		Addr: userAddr,
		Conn: conn,
		C:    make(chan string),
	}

	//使用一个go routine监听channel
	go user.ListerMessage()

	return &user
}

//监听user channel，一旦channel有消息发生给client
func (this *User) ListerMessage() {
	for {
		select {
		case msg := <-this.C:
			this.Conn.Write([]byte(msg + "\n"))
		default:
		}
	}
}
