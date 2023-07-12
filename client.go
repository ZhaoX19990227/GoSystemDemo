package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int // 当前client的模式
}

func NewClient(serverIp string, serverPort int) *Client {
	// 创建客户端对象
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}
	// 连接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("连接失败")
		return nil
	}
	client.conn = conn
	//返回对象
	return client
}

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器ip地址（默认是127.0.0.1）")
	flag.IntVar(&serverPort, "port", 8989, "设置服务器端口（默认是8989）")
}

func main() {
	flag.Parse()
	//client := NewClient("127.0.0.1", 1234)
	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println("连接失败")
		return
	}

	// 单独开启一个goroutine处理server响应的数据
	go client.DealResponse()

	fmt.Println("连接成功")

	client.Run()
}

func (client *Client) Run() {
	for client.flag != 0 {
		for client.menu() != true {
		}
		// 根据不同的模式 处理业务逻辑
		switch client.flag {
		case 1:
			fmt.Println("公聊模式")
			break
		case 2:
			fmt.Println("私聊模式")
			break
		case 3:
			//fmt.Println("更新用户名")
			client.UpdateName()
			break
		}
	}
}

func (client *Client) UpdateName() bool {
	fmt.Println("请输入用户名：")
	fmt.Scanln(&client.Name)

	s := "rename|" + client.Name + "\n"
	_, err := client.conn.Write([]byte(s))
	if err != nil {
		fmt.Println("client.conn.Write err:", err)
		return false
	}
	return true
}

func (client *Client) menu() bool {
	var flag int

	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("0.退出")

	// 等待用户输入一个整型
	fmt.Scanln(&flag)
	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println("输入非法")
		return false
	}

}

// 处理server响应的数据，直接输出
func (client *Client) DealResponse() {
	// 一旦有数据就直接copy到Stdout上标准输出上，永久阻塞监听
	io.Copy(os.Stdout, client.conn)
	//等价于
	//for  {
	//	buf := make([]byte,1024)
	//	client.conn.Read(buf)
	//	fmt.Println(buf)
	//}

}
