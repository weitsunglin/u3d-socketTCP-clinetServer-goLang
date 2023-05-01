package main

import (
	"fmt"
	"io"
	"net"
	"sync"
)

type Server struct {
	Ip   string
	Port int

	// 在线用户容器
	OnlineMap map[string]*User
	// 用户列容器锁，对容器进行操作时进行加锁
	mapLock sync.RWMutex

	// 消息广播的管道
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

// 启动服务器的接口
func (this *Server) Start() {
	// socket监听
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}

	// 程序退出时，关闭监听，注意defer关键字的用途
	defer listener.Close()

	// 启动一个协程来执行ListenMessager
	go this.ListenMessager()

	// 注意for循环不加条件，相当于while循环
	for {
		// Accept，此处会阻塞，当有客户端连接时才会往后执行
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}

		// 启动一个协程去处理
		go this.Handler(conn)

	}
}

func (this *Server) Handler(conn net.Conn) {

	// 构造User对象，NewUser全局方法在user.go脚本中
	user := NewUser(conn, this)

	// 用户上线
	user.Online()

	// 启动一个协程
	go func() {
		buf := make([]byte, 4096)
		for {
			// 从Conn中读取消息
			len, err := conn.Read(buf)
			if 0 == len {
				// 用户下线
				user.Offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}

			// 用户针对msg进行消息处理
			user.DoMessage(buf, len)
		}
	}()
}

func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]: " + msg

	this.Message <- sendMsg
}

func (this *Server) ListenMessager() {
	for {
		// 从Message管道中读取消息
		msg := <-this.Message

		// 加锁
		this.mapLock.Lock()
		// 遍历在线用户，把广播消息同步给在线用户
		for _, user := range this.OnlineMap {
			// 把要广播的消息写到用户管道中
			user.Channel <- msg
		}
		// 解锁
		this.mapLock.Unlock()
	}
}
