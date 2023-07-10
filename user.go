package main

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

// 创建一个用户
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()
	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	// 启动监听当前user channel消息的goroutine
	go user.ListenMessage()

	return user
}

// 用户上线
func (this *User) OnLine() {
	// 用户上线，将用户加入到onlineMap中
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	// 广播当前用户上线消息
	this.server.BroadCast(this, this.Name+"已上线")
}

// 用户下线
func (this *User) OffLine() {
	// 用户下线，将用户从onlineMap移除
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	// 广播当前用户上线消息
	this.server.BroadCast(this, this.Name+"已下线")
}

// 给当前user对应的客户端发消息
func (this *User) SendMsgToCurClient(msg string) {
	this.conn.Write([]byte(msg))
}

// 用户处理消息
func (this *User) DoMessage(msg string) {
	if "who" == msg {
		//查询当前有哪些用户在线
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onLineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线\n"
			this.SendMsgToCurClient(onLineMsg)
		}
		this.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		//rename|张三
		name := strings.Split(msg, "|")[1]
		user := this.server.OnlineMap[name]
		// 说明已经存在
		if user != nil {
			this.SendMsgToCurClient("当前用户已存在")
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.server.OnlineMap[name] = this
			this.server.mapLock.Unlock()

			this.Name = name
			this.SendMsgToCurClient("更新成功：" + this.Name + "\n")
		}
	} else if len(msg) > 4 && msg[:3] == "to|" {
		// to｜张三｜消息内容
		// 获取对方的用户名
		remoteName := strings.Split(msg, "|")[1]
		if remoteName == "" {
			this.SendMsgToCurClient("消息格式不正确,请使用 \"to｜张三｜消息内容\"格式.\n")
			return
		}
		// 根据用户名获取user对象
		remoteUser, ok := this.server.OnlineMap[remoteName]
		if !ok {
			this.SendMsgToCurClient("该用户不存在\n")
			return
		}
		// 获取消息内容，通过user对象发送消息
		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.SendMsgToCurClient("请重发")
			return
		}
		remoteUser.SendMsgToCurClient(this.Name + "对您说：" + content)
	} else {
		this.server.BroadCast(this, msg)
	}
}

// 监听当前user channel，一旦有消息，发送给对应的客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}
}
