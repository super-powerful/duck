package core

type Server interface {
	Run() error
	Stop() error
	GetClient(ID string) ServerClient
}

type ServerClient interface {
	GetID() string
	Close() error
}

type ServerMessage interface {
	GetClient() ServerClient
	GetData() interface{}
	Done() <-chan struct{}
	Error() error
}
