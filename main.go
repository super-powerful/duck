package main

import (
	"github.com/super-powerful/duck/core/server"
)

func main() {
	server.NewServer(server.Addr(":8080")).Run()
	wait := make(chan struct{})
	<-wait
}
