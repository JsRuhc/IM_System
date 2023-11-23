package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIP   string
	ServerPort int
	conn       net.Conn
	Name       string
	flag       int
}

func NewClient(serverIP string, serverPort int) *Client {
	//创建客户端对象
	clinet := &Client{
		ServerIP:   serverIP,
		ServerPort: serverPort,
		flag:       999,
	}

	// 连接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIP, serverPort))
	if err != nil {
		fmt.Println("net.Dial error", err)
		return nil
	}

	// 返回对象
	clinet.conn = conn

	return clinet
}

// 处理server回应的消息，直接显示到标准输出
func (clinet *Client) DealResponse() {

	//buf := make([]byte, 4096)
	//go func() {
	//	for {
	//		n, _ := clinet.conn.Read(buf)
	//		msg := string(buf[:n-1])
	//		fmt.Println(msg)
	//	}
	//}()

	// 一旦client.conn 有数据，就直接copy到stdout标准输出上，永久阻塞监听
	io.Copy(os.Stdout, clinet.conn)
}

func (client *Client) menu() bool {
	var flag int
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.更新用户名")
	fmt.Println("4.显示在线用户")
	fmt.Println("0.退出")

	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 4 {
		client.flag = flag
		return true
	} else {
		fmt.Println(">>>>> 请输出合法范围内的数字 <<<<<")
		return false
	}

}

func (clinet *Client) PublicChat() {
	var chatMsg string
	// 提示用户输入消息
	fmt.Println(">>>>> 请输入聊天内容，exit退出 <<<<<")
	fmt.Scanln(&chatMsg)

	for chatMsg != "exit" {
		// 发给服务器
		// 消息不为空
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := clinet.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn write err:", err)
				break
			}
		}
		chatMsg = ""
		fmt.Println(">>>>> 请输入聊天内容，exit退出 <<<<<")
		fmt.Scanln(&chatMsg)
	}
}

func (clinet *Client) UpdateName() bool {
	fmt.Println(">>>>> 请输入用户名: <<<<<")
	fmt.Scanln(&clinet.Name)

	sendMessage := "rename|" + clinet.Name + "\n"
	_, err := clinet.conn.Write([]byte(sendMessage))
	if err != nil {
		fmt.Println("conn write err:", err)
		return false
	}
	return true
}

func (clinet *Client) ShowUser() bool {
	fmt.Println(">>>>> 在线用户: <<<<<")
	sendMessage := "who\n"
	_, err := clinet.conn.Write([]byte(sendMessage))
	if err != nil {
		fmt.Println("conn write err:", err)
		return false
	}
	return true
}

func (client *Client) PrivateChat() {
	var remoteName string
	var chatMsg string

	client.ShowUser()
	fmt.Println(">>>>> 请输入聊天对象[用户名]，exit退出 <<<<<")
	fmt.Scanln(&remoteName)

	for remoteName != "exit" {
		fmt.Println(">>>>> 请输入消息内容，exit退出 <<<<<")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				sendMsg := "to|" + remoteName + "|" + chatMsg + "\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn Write err:", err)
					break
				}
			}
			chatMsg = ""
			fmt.Println(">>>>> 请输入消息内容，exit退出 <<<<<")
			fmt.Scanln(&chatMsg)
		}
		client.ShowUser()
		remoteName = ""
		fmt.Println(">>>>> 请输入聊天对象[用户名]，exit退出 <<<<<")
		fmt.Scanln(&remoteName)
	}

}

func (clinet *Client) Run() {
	for clinet.flag != 0 {
		for clinet.menu() != true {
		}

		// 根据不同模式处理不同业务
		switch clinet.flag {
		case 1:
			// 公聊模式
			clinet.PublicChat()
			break
		case 2:
			// 私聊模式
			clinet.PrivateChat()
			break
		case 3:
			// 更新用户名
			clinet.UpdateName()
			break
		case 4:
			//显示在线用户
			clinet.ShowUser()
			break
		}
	}
}

var serverIP string
var serverPort int

// client.exe -ip 127.0.0.1 -port 8888
func init() {
	flag.StringVar(&serverIP, "ip", "127.0.0.1", "设置服务器IP地址（默认为 127.0.0.1）")
	flag.IntVar(&serverPort, "port", 8888, "设置服务器端口（默认为 8888）")
}

func main() {
	// 命令解析
	flag.Parse()

	clinet := NewClient(serverIP, serverPort)
	if clinet == nil {
		fmt.Println(">>>>> 链接服务器失败 <<<<<")
		return
	}
	// 单独开启一个 goroutine 去处理server的回执消息
	go clinet.DealResponse()

	fmt.Println(">>>>> 链接服务器成功 <<<<<")

	// 启动客户端业务
	clinet.Run()
}
