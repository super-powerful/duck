package core

type Server interface {
	Run() error
	Stop() error
	GetClient(ID string) ServerClient
	GetClients(func(handle ServerClient))
	SendMessage(client ServerClient, data interface{}) ServerMessage
}

type ServerClient interface {
	GetID() string
	Close() error
	SendMessage(data interface{}) ServerMessage
}

type ServerMessage interface {
	GetClient() ServerClient
	GetData() interface{}
	Done() <-chan struct{}
	Error() error
}

type Client interface {
	Dial() error
	ReDial() error
	Close() error
}
