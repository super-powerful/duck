package core

type Message interface {
	GetData() interface{}
	Done() <-chan struct{}
	Error() error
}

type Client interface {
	Close() error
	SendMessage(data interface{}) Message
}

type Server interface {
	Run() error
	Stop() error
	GetClient(ID string) ServerClient
	GetClients(func(handle ServerClient))
	SendMessage(client ServerClient, data interface{}) Message
}

type ServerClient interface {
	Client
	GetID() string
	GetAddr() string
}

type ServerMessage interface {
	Message
	GetClient() ServerClient
}

type UserClient interface {
	Client
	Dial() error
}
