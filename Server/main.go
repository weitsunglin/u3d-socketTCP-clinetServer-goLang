package main

import (
	"fmt"
	"time"
)

func main() {
	// 启动Server
	go StartServer()

	// TODO 你可以写其他逻辑\
	fmt.Println("Server已啟動")

	// 防止主线程退出
	for {
		time.Sleep(1 * time.Second)
	}
}

func StartServer() {
	server := NewServer("127.0.0.1", 8888)
	server.Start()
}
