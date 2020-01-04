package server

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func init() {
	server := NewServer(Addr(":8080"))
	if err := server.Run(); err != nil {
		fmt.Println(err)
	}
	time.Sleep(time.Second)
}

func TestNewServer(t *testing.T) {
	conn, _ := net.Dial("tcp", "127.0.0.1:8080")
	conn.Close()
	wait := make(chan struct{})
	<-wait
}
