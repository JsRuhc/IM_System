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

	//在线用户列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	//消息广播的channel
	Message chan string
}

func Newserver(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// 监听Mseeage广播消息 channel的 goroutine, 一旦有消息就会发送给全部在线的User
func (this *Server) ListenMessager() {
	for {

		msg := <-this.Message

		// 将msg发送给所有的在线User
		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}

}

// 广播的用户消息内容
func (this *Server) BroadCast(user *User, msg string) {
	//sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	sendMsg := fmt.Sprintf("[%v]%v:%v", user.Addr, user.Name, msg)

	this.Message <- sendMsg
}

func (this *Server) Hander(conn net.Conn) {
	//fmt.Println("链接成功！")

	// 用户上线
	user := NewUser(conn, this)
	fmt.Println(user.Addr, "上线")
	user.Online()

	// 监听用户是否活跃
	isLive := make(chan bool)

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				// 用户下线
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
			}

			// 提取用户消息(去掉 ‘\n’)
			msg := string(buf[:n-1])

			// 用户针对msg继续处理
			user.doMessgae(msg)

			// 用户发送信息后会
			isLive <- true
		}
	}()

	//当前hander 阻塞
	for {
		select {
		case <-isLive: // 将 isLive 管道中的数据取出，重置管道
			// 用户是活跃的，不做如何操作
		case <-time.After(time.Second * 40):
			// 已经超时
			// 向当前用户发送踢出信息
			user.SendMsg("你被踢了")
			// 销毁资源
			close(user.C)
			// 关闭链接
			conn.Close()
			// 退出hander
			return //runtine.Goexit()
		}
	}
}

func (this *Server) Start() {
	//socket Listen start
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
	}
	// close Listen socket
	defer listener.Close()

	//启动监听 Message 的 goroutine
	go this.ListenMessager()

	for {
		//accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Listener accpet err:", err)
			continue
		}

		// do Handler
		go this.Hander(conn)
	}

}
