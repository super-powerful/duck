package main

import (
	"github.com/super-powerful/duck/core/server"
)

func main() {
	server.New(server.Addr(":8080")).Run()
	wait := make(chan struct{})
	<-wait
}
