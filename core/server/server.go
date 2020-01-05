package server

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/super-powerful/duck/core"
	"log"
	"net"
	"sync"
	"time"
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

func NewMessage(client core.ServerClient, data interface{}) core.Message {
	return &_Message_{
		Client:   client,
		Data:     data,
		WaitChan: make(chan struct{}),
		Err:      nil,
	}
}

type EncodeEvent func(core.ServerClient, core.Message) []byte
type DecodeEvent func(core.ServerClient, []byte) (int, core.Message)
type MessageEvent func(core.Message)
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
	IsRun   bool
	Mux     sync.Mutex
	Clients sync.Map
	_Event_
}

type Options func(s core.Server)

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
	log.Printf("服务启动中...\n")

	if s.IsRun {
		if err := s.stop(); err != nil {
			return err
		}
	}

	if lis, err := net.Listen("tcp", s.Addr); err == nil {
		s.Lis = lis
		s.IsRun = true

		go s.run()
	} else {
		return err
	}

	return nil
}
func (s *_Server_) run() error {
	log.Printf("服务已启动,监听%s正在等待连接进入!\n", s.Lis.Addr().String())
	for s.IsRun {
		if conn, err := s.Lis.Accept(); err == nil {
			go func() {
				client := newServerClient(s, conn)
				log.Printf("客户端%v建立连接,ID为%v!\n", conn.RemoteAddr().String(), client.ID)
				if s.ClientConnFilterEvent != nil {
					if s.ClientConnFilterEvent(client) {
						s.Clients.Store(client.ID, client)
						client.todo()
					} else {
						if err := client.Close(); err != nil {
							fmt.Println(err)
						}
					}
				} else {
					s.Clients.Store(client.ID, client)
					client.todo()
				}
			}()
		}
	}

	return nil
}

func (s *_Server_) Stop() error {
	s.Mux.Lock()
	defer s.Mux.Unlock()

	if s.IsRun {
		s.stop()
		log.Printf("服务%v已经停止!\n", s.Lis.Addr().String())
	}

	return nil
}

func (s *_Server_) stop() error {
	s.IsRun = false
	if err := s.Lis.Close(); err != nil {
		fmt.Println(err)
	}
	s.Clients.Range(func(key, value interface{}) bool {
		if err := value.(*_ServerClient_).Close(); err != nil {
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

func (s *_Server_) GetClients(handle func(core.ServerClient)) {
	s.Clients.Range(func(key, value interface{}) bool {
		handle(value.(core.ServerClient))
		return true
	})
}
func (s *_Server_) SendMessage(client core.ServerClient, data interface{}) core.Message {
	message := NewMessage(client, data)
	client.(*_ServerClient_).WriterMessages <- message.(*_Message_)
	return message
}

type _ServerClient_ struct {
	ID             string
	Server         *_Server_
	Conn           net.Conn
	IsRun          bool
	Mux            sync.Mutex
	WriterMessages chan *_Message_
	LastRead       time.Time
	LastWrite      time.Time
}

func (c *_ServerClient_) GetID() string {
	return c.ID
}

func (c *_ServerClient_) Close() error {
	c.Mux.Lock()
	defer c.Mux.Unlock()

	if c.IsRun {
		log.Printf("客户端%v离开,ID为%v!\n", c.Conn.RemoteAddr().String(), c.ID)
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
					if c.Server.MessageEvent != nil && message != nil {
						c.LastRead = time.Now()
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
	c.Close()
}

func (c *_ServerClient_) writer() {
	messages := c.WriterMessages
	for c.IsRun {
		if message, ok := <-messages; ok {
			if data := c.Server.EncodeEvent(c, message); data != nil {
				if _, err := c.Conn.Write(data); err != nil {
					message.done(err)
					break
				} else {
					message.done(nil)
					c.LastWrite = time.Now()
				}
			}
		} else {
			break
		}
	}
	c.Close()
	for {
		if message, ok := <-messages; ok {
			message.done(errors.New("客户端断开连接"))
		} else {
			break
		}
	}
}

func (c *_ServerClient_) SendMessage(data interface{}) core.Message {
	message := NewMessage(c, data)
	c.WriterMessages <- message.(*_Message_)
	return message
}

func newServerClient(server *_Server_, conn net.Conn) *_ServerClient_ {
	return &_ServerClient_{
		ID:             uuid.New().String(),
		Server:         server,
		Conn:           conn,
		IsRun:          false,
		WriterMessages: make(chan *_Message_, 512),
		LastRead:       time.Now(),
		LastWrite:      time.Now(),
	}
}
