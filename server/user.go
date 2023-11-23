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
	Server *Server
}

// 创建用户 api
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		Server: server,
	}

	//启动当前 User channel 消息的 gorountine
	go user.ListenMessage()

	return user
}

func (this *User) Online() {
	// 用户上线，添加到onlineMap 中
	this.Server.mapLock.Lock()
	this.Server.OnlineMap[this.Name] = this
	this.Server.mapLock.Unlock()

	// 广播当前用户上线消息
	this.Server.BroadCast(this, "已上线")
}

func (this *User) Offline() {
	// 用户下线，将用户从onlineMap 中移除
	this.Server.mapLock.Lock()
	delete(this.Server.OnlineMap, this.Name)
	this.Server.mapLock.Unlock()

	// 广播当前用户下线消息
	this.Server.BroadCast(this, "已下线")
}

func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg))
}

func (this *User) doMessgae(msg string) {
	if msg == "who" {
		// 查询当前用户
		this.Server.mapLock.Lock()
		for _, user := range this.Server.OnlineMap {
			onlineMessage := "[" + user.Addr + "]" + user.Name + ":" + "Online" + "\n"
			//onlineMessage := fmt.Sprintf("[%v]%v:在线", user.Addr, user.Name)
			this.SendMsg(onlineMessage)
			//this.C <- onlineMessage
		}
		this.Server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "rename|" {
		//消息格式 rename|张三
		Newname := strings.Split(msg, "|")[1]
		// 判断该用户名是否被使用
		_, ok := this.Server.OnlineMap[Newname]
		if ok {
			this.SendMsg("当前用户名已被使用\n")
		} else {
			this.Server.mapLock.Lock()
			delete(this.Server.OnlineMap, this.Name)
			this.Server.OnlineMap[Newname] = this
			this.Server.mapLock.Unlock()

			//fmt.Println(this.Name)
			this.Name = Newname
			this.SendMsg("您已经更新用户名:" + this.Name + "\n")
		}

	} else if len(msg) > 3 && msg[:3] == "to|" {
		removeName := strings.Split(msg, "|")[1]
		removeUser, ok := this.Server.OnlineMap[removeName]
		if !ok {
			this.SendMsg("改用户姓名不存在！\n")
			return
		}
		content := strings.Split(msg, "|")[2]
		if content == "" {
			this.SendMsg("无效内容，请重新发送")
			return
		}
		removeUser.SendMsg(this.Name + "向您发送消息：" + content + "\n")
	} else {
		this.Server.BroadCast(this, msg)
	}
}

// 监听当前 User channel 的方法，一旦有消息，就直接发送给客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.conn.Write([]byte(msg + "\n"))
	}

}
