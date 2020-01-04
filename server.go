package duck

import (
	"bytes"
	"errors"
	"net"
	"sync"
)

// _Server_ ...
type _Server_ struct {
	Conf   _Config_     // 服务配置
	IsRun  bool         // 服务是否运行标识
	Lis    net.Listener // 服务监听
	LisMux sync.Mutex   // 服务监听锁

	Clients         sync.Map      // 客户端连接
	RequestMessages chan *Message // 消息

	EncodeMessageCallback func(buffer *bytes.Buffer, message *Message) // 消息编码
	DecodeMessageCallback func(buffer *bytes.Buffer) []*Message        // 消息解码

	ServerRunCallbacks    []ServerRunCallback      // 服务运行回调
	ServerStopCallbacks   []ServerStopCallback     // 服务停止回调
	ClientConnCallbacks   []ClientConnCallback     // 客户端建立连接回调
	ClientUnConnCallbacks []ClientUnConnCallback   // 客户端断开连接回调
	ClientReqCallbacks    []ClientRequestCallback  // 客户端请求回调
	ClientRpsCallbacks    []ClientResponseCallback // 客户端响应回调
}

// Run 启动服务
func (s *_Server_) Run() error {
	s.LisMux.Lock()
	defer s.LisMux.Unlock()

	if s.IsRun {
		return errors.New("the service is running")
	}

	if lis, err := net.Listen("tcp", s.Conf.Addr); err == nil {
		s.Lis = lis
		s.IsRun = true
		for _, callback := range s.ServerRunCallbacks {
			callback()
		}
		go s.listener()
	} else {
		return err
	}
	return nil
}

func (s *_Server_) listener() {
	s.RequestMessages = make(chan *Message, 65535)
	for s.IsRun {
		conn, _ := s.Lis.Accept()

		go func() {
			client := newClient(conn, s)
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
				client.Run()
			} else {
				client.Stop()
			}
		}()
	}
	close(s.RequestMessages)
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
		value.(*_Client_).Stop()
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
	server := new(_Server_)

	DefaultEncoding(server)
	DefaultDecoding(server)

	server.ServerRunCallbacks = make([]ServerRunCallback, 0)
	server.ServerStopCallbacks = make([]ServerStopCallback, 0)
	server.ClientConnCallbacks = make([]ClientConnCallback, 0)
	server.ClientUnConnCallbacks = make([]ClientUnConnCallback, 0)
	server.ClientReqCallbacks = make([]ClientRequestCallback, 0)
	server.ClientRpsCallbacks = make([]ClientResponseCallback, 0)

	for _, option := range options {
		option(server)
	}

	return server
}
