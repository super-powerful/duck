package server

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestNewServer(t *testing.T) {
	beginTime := time.Now()
	for i := 0; i < 1000; i++ {
		if _, err := net.Dial("tcp", "127.0.0.1:8080"); err != nil {
			t.Log(err)
		}
	}
	overTime := time.Now()
	fmt.Printf("一万连接建立完成使用了%v纳秒", overTime.Sub(beginTime).Nanoseconds())
}
