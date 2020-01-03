package duck

import (
	"bytes"
	"fmt"
	"github.com/google/uuid"
	"net"
	"sync"
)

// _Client_ 客户端
type _Client_ struct {
	ID           string        // 客户端ID
	Enable       bool          // 是否可用
	EnableMux    sync.RWMutex  // 是否可用锁
	Conn         net.Conn      // 真实连接
	ConnWriteMux sync.Mutex    // 连接写锁
	Messages     chan *Message // 响应通道

	Server *_Server_ // 服务
}

func newClient(conn net.Conn, s *_Server_) *_Client_ {
	return &_Client_{
		ID:       uuid.New().String(),
		Enable:   true,
		Server:   s,
		Conn:     conn,
		Messages: make(chan *Message, 65535),
	}
}

func (c *_Client_) GetID() string {
	return c.ID
}

func (c *_Client_) Run() {
	go c.reader()
	go c.writer()
}

func (c *_Client_) Stop() {
	c.EnableMux.Lock()
	defer c.EnableMux.Unlock()

	if c.Enable {
		c.Enable = false
		close(c.Messages)
		if err := c.Conn.Close(); err != nil && c.Server.IsRun {
			fmt.Println(err)
		}
	}
}

func (c *_Client_) reader() {
	c.EnableMux.RLock()
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
		c.EnableMux.RUnlock()
		c.Stop()
	}()

	cacheData := new(bytes.Buffer)
	receiveData := make([]byte, 1024)
	for c.Server.IsRun && c.Enable {
		if size, err := c.Conn.Read(receiveData); err == nil {
			cacheData.Write(receiveData[:size])
			if messages := c.Server.DecodeMessageCallback(cacheData); messages != nil && len(messages) != 0 {
				for _, message := range messages {
					c.Server.RequestMessages <- message
				}
			}
		} else {
			break
		}
	}
}

func (c *_Client_) writer() {
	c.EnableMux.RLock()
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
		c.EnableMux.RUnlock()
		c.Stop()
	}()

	cacheData := new(bytes.Buffer)
	for c.Server.IsRun && c.Enable {
		if message, err := <-c.Messages; err {
			c.Server.EncodeMessageCallback(cacheData, message)
			if _, err := c.Conn.Write(cacheData.Bytes()); err == nil {
				cacheData.Reset()
			} else {
				break
			}
		} else {
			break
		}
	}
}
