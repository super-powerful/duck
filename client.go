package duck

import (
	"bytes"
	"fmt"
	"net"
	"sync"
)

// _Client_ 客户端
type _Client_ struct {
	ID           string        // 客户端ID
	Server       *_Server_     // 服务
	Conn         net.Conn      // 真实连接
	ConnWriteMux sync.Mutex    // 连接写锁
	Messages     chan *Message // 响应通道
}

func (c *_Client_) GetID() string {
	return c.ID
}

func (c *_Client_) listener() {
	go c.reader()
	go c.writer()

}

func (c *_Client_) reader() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	cacheData := new(bytes.Buffer)
	receiveData := make([]byte, 1024)
	for c.Server.IsRun {
		if size, err := c.Conn.Read(receiveData); err == nil {
			cacheData.Write(receiveData[:size])
			if messages := c.Server.DecodeMessageCallback(cacheData); messages != nil && len(messages) != 0 {
				for _, message := range messages {
					c.Server.Messages <- message
				}
			}
		} else {
			break
		}
	}
}

func (c *_Client_) writer() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Println(err)
		}
	}()
	cacheData := new(bytes.Buffer)
	for c.Server.IsRun {
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
