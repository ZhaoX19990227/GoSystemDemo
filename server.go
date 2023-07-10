package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	// 在线用户列表 key:用户名 value:user对象
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	// 消息广播的channel
	Message chan string
}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// 广播消息
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg

	this.Message <- sendMsg
}

func (this *Server) Handler(conn net.Conn) {
	// 当前链接的业务
	// fmt.Println("链接建立成功")

	user := NewUser(conn, this)

	// 用户上线，将用户加入到onlineMap中
	//this.mapLock.Lock()
	//this.OnlineMap[user.Name] = user
	//this.mapLock.Unlock()
	user.OnLine()

	// 广播当前用户上线消息
	this.BroadCast(user, user.Name+"已上线")

	isLive := make(chan bool)

	go func() {
		buf := make([]byte, 4096)
		n, err := conn.Read(buf)
		if n == 0 {
			//this.BroadCast(user, user.Name+"下线")
			user.OffLine()
			return
		}
		if err != nil && err != io.EOF {
			fmt.Println("Conn Read Error...", err)
			return
		}
		//提取用户的输入(去除\n)
		msg := string(buf[:n-1])
		//消息广播
		//this.BroadCast(user, msg)
		user.DoMessage(msg)

		// 用户任意消息代表处于活跃状态
		isLive <- true
	}()

	// 当前handler阻塞
	for {
		//select - case穿透
		select {
		// 当前用户活跃，重置计时器，下面的也会执行
		case <-isLive:
		case <-time.After(time.Minute * 30):
			//超时 将user关闭
			user.SendMsgToCurClient("你被踢了")

			//销毁
			close(user.C)
			conn.Close()
			//退出当前handler
			return //或者runtime.Goexit()
		}
	}
}

// 启动服务器的接口
func (this *Server) Start() {
	//socket listen
	listen, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.listen.err:", err)
		return
	}
	//close listen socket
	defer listen.Close()

	// 启动监听Message的goroutine
	go this.ListenMessager()

	for {
		//accept
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}
		//do handler
		go this.Handler(conn)
	}
}

// 监听Message广播消息channel的goroutine，一旦有消息就发给所有的user
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message
		// 将msg发送给在线的user
		this.mapLock.Lock()
		for _, client := range this.OnlineMap {
			client.C <- msg
		}
		this.mapLock.Unlock()
	}
}
