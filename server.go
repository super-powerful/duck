package duck

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"net"
	"sync"
)

// _Server_ ...
type _Server_ struct {
	Config                  *_Config_                                    // 服务配置
	IsRun                   bool                                         // 服务是否运行标识
	Lis                     net.Listener                                 // 服务监听
	LisMux                  sync.Mutex                                   // 服务监听锁
	Clients                 *sync.Map                                    // 客户端连接
	Messages                chan *Message                                // 消息
	EncodeMessageCallback   func(buffer *bytes.Buffer, message *Message) // 消息编码
	DecodeMessageCallback   func(buffer *bytes.Buffer) []*Message        // 消息解码
	ServerRunCallbacks      []ServerRunCallback                          // 服务运行回调
	ServerStopCallbacks     []ServerStopCallback                         // 服务停止回调
	ClientConnCallbacks     []ClientConnCallback                         // 客户端建立连接回调
	ClientUnConnCallbacks   []ClientUnConnCallback                       // 客户端断开连接回调
	ClientRequestCallbacks  []ClientRequestCallback                      // 客户端请求回调
	ClientResponseCallbacks []ClientResponseCallback                     // 客户端响应回调
}

var _ErrServerIsRun_ = errors.New("the service is running") // 服务已经运行

// Run 启动服务
func (s *_Server_) Run() error {
	s.LisMux.Lock()
	defer s.LisMux.Unlock()

	if s.IsRun {
		return _ErrServerIsRun_
	}

	if lis, err := net.Listen("tcp", s.Config.Addr); err == nil {
		s.Lis = lis
		s.IsRun = true
		for _, callback := range s.ServerRunCallbacks {
			callback()
		}
		go s.run()
	} else {
		return err
	}
	return nil
}

func (s *_Server_) run() {
	for s.IsRun {
		conn, _ := s.Lis.Accept()

		go func() {
			client := new(_Client_)
			client.Conn = conn
			client.Server = s
			client.ID = uuid.New().String()
			allowClient := true

			for _, callback := range s.ClientConnCallbacks {
				next, allow := callback(client)
				if !allow {
					allowClient = false
				}
				if !next {
					break
				}
			}

			if allowClient {
				s.Clients.Store(client.ID, client)
				go func() {

				}()
			} else {
				if err := conn.Close(); err != nil {
					fmt.Println(err)
				}
			}
		}()
	}
}

// Stop 停止服务
func (s *_Server_) Stop(signal int) error {
	s.LisMux.Lock()
	defer s.LisMux.Unlock()

	if s.IsRun {
		if err := s.Lis.Close(); err != nil {
			return err
		}
		s.Lis = nil
		s.IsRun = false
	}

	s.Clients.Range(func(key, value interface{}) bool {
		s.Clients.Delete(key)
		value.(*_Client_).Conn.Close()
		return true
	})

	for _, callback := range s.ServerStopCallbacks {
		callback(signal)
	}
	return nil
}

// _Config_ 服务配置
type _Config_ struct {
	Addr string // 监听地址
}

// _Option_ 服务配置选项
type _Option_ func(s *_Server_)

// New 创建服务
func New(options ..._Option_) Server {

	return nil
}
