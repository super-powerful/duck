package client

import (
	"bytes"
	"errors"
	"github.com/super-powerful/duck/core"
	"log"
	"net"
	"sync"
	"time"
)

type _Message_ struct {
	Data     interface{}
	WaitChan chan struct{}
	IsOver   bool
	Err      error
	Mux      sync.Mutex
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
	m.Mux.Lock()
	defer m.Mux.Unlock()
	if m.IsOver {
		return
	}
	m.IsOver = true
	m.Err = err
	close(m.WaitChan)
}

func NewMessage(data interface{}) core.Message {
	return &_Message_{
		Data:     data,
		WaitChan: make(chan struct{}),
		Err:      nil,
	}
}

type EncodeEvent func(client core.UserClient, message core.Message) []byte
type DecodeEvent func(client core.UserClient, data []byte) (int, core.Message)
type MessageEvent func(message core.Message)
type ConnEvent func(client core.UserClient) bool
type DisConnEvent func(client core.UserClient)

type _Event_ struct {
	EncodeEvent  EncodeEvent
	DecodeEvent  DecodeEvent
	MessageEvent MessageEvent
	ConnEvent    ConnEvent
	DisConnEvent DisConnEvent
}

type _Client_ struct {
	Addr          string
	Conn          net.Conn
	Mux           sync.Mutex
	IsRun         bool
	WriteMessages chan *_Message_
	LastRead      time.Time
	LastWrite     time.Time
	_Event_
}

type Options func(c core.UserClient)

func New(options ...Options) core.UserClient {
	s := new(_Client_)
	for _, option := range options {
		option(s)
	}
	return s
}

func (c *_Client_) Close() error {
	c.Mux.Lock()
	defer c.Mux.Unlock()

	if c.IsRun {
		if err := c.close(); err != nil {
			return err
		}
	}
	return nil
}
func (c *_Client_) close() error {
	c.IsRun = false
	close(c.WriteMessages)
	if err := c.Conn.Close(); err != nil {
		return err
	}
	return nil
}
func (c *_Client_) SendMessage(data interface{}) core.Message {
	message := NewMessage(data)
	defer func() {
		if err := recover(); err != nil {
			message.(*_Message_).done(err.(error))
		}
	}()
	c.WriteMessages <- message.(*_Message_)
	return message
}

func (c *_Client_) Dial() error {
	c.Mux.Lock()
	defer c.Mux.Unlock()

	if c.IsRun {
		if err := c.close(); err != nil {
			return err
		}
	}
	if conn, err := net.Dial("tcp", c.Addr); err == nil {
		c.IsRun = true
		c.Conn = conn
		c.WriteMessages = make(chan *_Message_, 512)
		c.LastRead = time.Now()
		c.LastWrite = time.Now()
		log.Printf("建立%v成功!\n", c.Conn.RemoteAddr().String())
		if c.ConnEvent != nil {
			if c.ConnEvent(c) {
				go c.todo()
			} else {
				c.close()
			}
		} else {
			go c.todo()
		}
	} else {
		return err
	}
	return nil
}

func (c *_Client_) todo() {
	go c.reader()
	go c.writer()
}

func (c *_Client_) reader() {
	dataBuff := new(bytes.Buffer)
	data := make([]byte, 1024)
	for c.IsRun {
		if size, err := c.Conn.Read(data); err == nil {
			if c.DecodeEvent == nil {
				continue
			}
			dataBuff.Write(data[:size])
			useSize := 0
			for {
				if len, message := c.DecodeEvent(c, dataBuff.Bytes()); len != 0 {
					useSize += len
					dataBuff.Next(len)
					if c.MessageEvent != nil && message != nil {
						c.LastRead = time.Now()
						go c.MessageEvent(message)
					}
				} else {
					break
				}
			}
			if useSize != 0 {
				dataBuff = bytes.NewBuffer(dataBuff.Bytes())
			}
		} else {
			break
		}
	}
	c.Close()
}

func (c *_Client_) writer() {
	messages := c.WriteMessages
	for c.IsRun {
		if message, ok := <-messages; ok {
			if data := c.EncodeEvent(c, message); data != nil {
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
			message.done(errors.New("已断开远程连接"))
		} else {
			break
		}
	}
}
