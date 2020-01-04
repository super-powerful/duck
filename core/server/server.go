package server

import (
	"bytes"
	"fmt"
	"github.com/google/uuid"
	"github.com/super-powerful/duck/core"
	"net"
	"sync"
)

type _Message_ struct {
	Client   core.ServerClient
	Data     interface{}
	WaitChan chan struct{}
	Err      error
}

func (m *_Message_) GetClient() core.ServerClient {
	return m.Client
}

func (m *_Message_) GetData() interface{} {
	return m.Data
}

func (m *_Message_) Done() <-chan struct{} {
	return m.WaitChan
}

func (m *_Message_) Error() error {
	return m.Err
}

func (m *_Message_) done(err error) {
	m.Err = err
	close(m.WaitChan)
}

func NewMessage(client core.ServerClient, data interface{}) core.ServerMessage {
	return &_Message_{
		Client:   client,
		Data:     data,
		WaitChan: make(chan struct{}),
		Err:      nil,
	}
}

type EncodeEvent func(core.ServerClient, *_Message_) []byte
type DecodeEvent func(core.ServerClient, []byte) (int, *_Message_)
type MessageEvent func(*_Message_)
type ClientConnFilterEvent func(core.ServerClient) bool
type ClientDisConnEvent func(core.ServerClient)

type _Event_ struct {
	EncodeEvent           EncodeEvent
	DecodeEvent           DecodeEvent
	MessageEvent          MessageEvent
	ClientConnFilterEvent ClientConnFilterEvent
	ClientDisConnEvent    ClientDisConnEvent
}

type _Server_ struct {
	Addr    string
	Lis     net.Listener
	RunFlag bool
	Mux     sync.Mutex
	Clients sync.Map
	_Event_
}

type Options func(s *_Server_)

func NewServer(options ...Options) core.Server {
	s := new(_Server_)
	for _, option := range options {
		option(s)
	}
	return s
}

func (s *_Server_) Run() error {
	s.Mux.Lock()
	defer s.Mux.Unlock()

	if s.RunFlag {
		if err := s.stop(); err != nil {
			return err
		}
	}

	if lis, err := net.Listen("tcp", s.Addr); err == nil {
		s.Lis = lis
		s.RunFlag = true

		go s.run()
	} else {
		return err
	}

	return nil
}
func (s *_Server_) run() error {

	for s.RunFlag {
		if conn, err := s.Lis.Accept(); err == nil {
			client := newServerClient(s, conn)
			if s.ClientConnFilterEvent != nil {
				if s.ClientConnFilterEvent(client) {
					s.Clients.Store(client.ID, client)
					client.todo()
				} else {
					if err := client.Close(); err != nil {
						fmt.Println(err)
					}
				}
			}
		}
	}

	return nil
}

func (s *_Server_) Stop(signal int) error {
	s.Mux.Lock()
	defer s.Mux.Unlock()

	if s.RunFlag {
		s.stop()
	}

	return nil
}

func (s *_Server_) stop() error {
	s.RunFlag = false
	if err := s.Lis.Close(); err != nil {
		fmt.Println(err)
	}
	s.Clients.Range(func(key, value interface{}) bool {
		if err := value.(_ServerClient_).Close(); err != nil {
			fmt.Println(err)
		}
		s.Clients.Delete(key)
		return true
	})
	return nil
}

func (s *_Server_) GetClient(ID string) core.ServerClient {
	if client, ok := s.Clients.Load(ID); ok {
		return client.(core.ServerClient)
	} else {
		return nil
	}
}

type _ServerClient_ struct {
	ID          string
	Server      *_Server_
	Conn        net.Conn
	IsRun       bool
	Mux         sync.Mutex
	WriterBytes chan *_Message_
}

func (c *_ServerClient_) GetID() string {
	return c.ID
}

func (c *_ServerClient_) Close() error {
	c.Mux.Lock()
	defer c.Mux.Unlock()

	if c.IsRun {
		c.IsRun = false
		if err := c.Conn.Close(); err != nil {
			fmt.Println(err)
		}
		c.Server.Clients.Delete(c.GetID())
		if c.Server.ClientDisConnEvent != nil {
			c.Server.ClientDisConnEvent(c)
		}
	}

	return nil
}

func (c *_ServerClient_) todo() {
	c.Mux.Lock()
	defer c.Mux.Unlock()
	c.IsRun = true
	go c.reader()
	go c.writer()
}

func (c *_ServerClient_) reader() {
	defer func() {
		c.Close()
	}()
	dataBuff := new(bytes.Buffer)
	data := make([]byte, 1024)
	for c.IsRun {
		if size, err := c.Conn.Read(data); err == nil {
			if c.Server.DecodeEvent == nil {
				continue
			}
			dataBuff.Write(data[:size])
			useSize := 0
			for {
				if len, message := c.Server.DecodeEvent(c, dataBuff.Bytes()); len != 0 {
					useSize += len
					if c.Server.MessageEvent != nil {
						go c.Server.MessageEvent(message)
					}
				} else {
					break
				}
			}
			if useSize != 0 {
				dataBuff.Next(useSize)
				dataBuff = bytes.NewBuffer(dataBuff.Bytes())
			}
		} else {
			break
		}
	}
}

func (c *_ServerClient_) writer() {
	defer func() {
		c.Close()
	}()
	for c.IsRun {
		if message, ok := <-c.WriterBytes; ok {
			if data := c.Server.EncodeEvent(c, message); data != nil {
				if _, err := c.Conn.Write(data); err != nil {
					message.done(err)
					break
				} else {
					message.done(nil)
				}
			}
		} else {
			break
		}
	}
}

func newServerClient(server *_Server_, conn net.Conn) *_ServerClient_ {
	return &_ServerClient_{
		ID:          uuid.New().String(),
		Server:      server,
		Conn:        conn,
		IsRun:       false,
		WriterBytes: make(chan *_Message_, 512),
	}
}
