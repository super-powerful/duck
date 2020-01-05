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

type _UserMessage_ struct {
	Data     interface{}
	WaitChan chan struct{}
	Err      error
}

func (m *_UserMessage_) GetData() interface{} {
	return m.Data
}

func (m *_UserMessage_) Done() <-chan struct{} {
	return m.WaitChan
}

func (m *_UserMessage_) Error() error {
	return m.Err
}

func (m *_UserMessage_) done(err error) {
	m.Err = err
	close(m.WaitChan)
}

func NewMessage(data interface{}) core.Message {
	return &_UserMessage_{
		Data:     data,
		WaitChan: make(chan struct{}),
		Err:      nil,
	}
}

type EncodeEvent func(core.UserClient, core.Message) []byte
type DecodeEvent func(core.UserClient, []byte) (int, core.Message)
type MessageEvent func(core.Message)
type ConnEvent func(core.UserClient) bool
type DisConnEvent func(core.UserClient)

type _Event_ struct {
	EncodeEvent  EncodeEvent
	DecodeEvent  DecodeEvent
	MessageEvent MessageEvent
	ConnEvent    ConnEvent
	DisConnEvent DisConnEvent
}

type _UserClient_ struct {
	Addr          string
	Conn          net.Conn
	Mux           sync.Mutex
	IsRun         bool
	WriteMessages chan *_UserMessage_
	LastRead      time.Time
	LastWrite     time.Time
	_Event_
}

func (c *_UserClient_) Close() error {
	panic("implement me")
}
func (c *_UserClient_) close() error {
	panic("implement me")
}
func (c *_UserClient_) SendMessage(data interface{}) core.ServerMessage {
	panic("implement me")
}

func (c *_UserClient_) Dial() error {
	c.Mux.Lock()
	defer c.Mux.Unlock()

	if c.IsRun {
		if err := c.close(); err != nil {
			return err
		}
	}
	if conn, err := net.Dial("tcp", c.Addr); err == nil {
		log.Printf("建立%v成功!\n", c.Conn.RemoteAddr().String())
		c.IsRun = true
		c.Conn = conn
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

func (c *_UserClient_) todo() {
	go c.reader()
	go c.writer()
}

func (c *_UserClient_) reader() {
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
					if c.MessageEvent != nil && message != nil {
						c.LastRead = time.Now()
						go c.MessageEvent(message)
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

func (c *_UserClient_) writer() {
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
