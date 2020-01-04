package server

import (
	"fmt"
	"testing"
)

func TestNewServer(t *testing.T) {
	server := NewServer(Addr(":8080"))
	if err := server.Run(); err != nil {
		fmt.Println(err)
	}

	wait := make(chan struct{})
	<-wait
}
