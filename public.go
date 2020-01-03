package duck

import (
	"context"
)

// Server 服务
type Server interface {
	Run() error
	Stop(signal int) error
	SendMessage(client Client, message *Message) error
}

// ServerRunCallback 服务启动回调
type ServerRunCallback func()

// ServerStopCallback 服务停止回调
type ServerStopCallback func(signal int)

// ClientConnCallback 客户端建立连接回调
type ClientConnCallback func(client Client) (next, allow bool)

// ClientUnConnCallback 客户端断开连接回调
type ClientUnConnCallback func(client Client)

// ClientRequestCallback 客户端请求回调
type ClientRequestCallback func(client Client, message *Message)

// ClientResponseCallback 客户端响应回调
type ClientResponseCallback func(client Client, message *Message)

type Client interface {
	GetID() string
}

type Message struct {
	Data    []byte
	Context context.Context
	Err     error
}
