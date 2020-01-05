package duck

import (
	"bytes"
	"encoding/binary"
	"github.com/super-powerful/duck/core"
	"github.com/super-powerful/duck/core/client"
	"github.com/super-powerful/duck/core/server"
	"log"
	"testing"
	"time"
)

func TestTcp(t *testing.T) {
	initServer()
	client := initClient()
	go sendMsg(client)
	<-make(chan struct{})
}

func sendMsg(client core.UserClient) {
	client.SendMessage(time.Now().Format("2006-01-02 15:04:05"))
	//// 空字符
	//emptyDataTime := time.NewTicker(time.Second * 3)
	//// 日期字符
	//strDataTime := time.NewTicker(time.Second * 1)
	//// 结束发送
	//endSend := time.After(time.Second * 2)
	//for {
	//	select {
	//	case <-emptyDataTime.C:
	//		client.SendMessage(nil)
	//	case currTime := <-strDataTime.C:
	//		client.SendMessage(currTime.Format("2006-01-02 15:04:05"))
	//	case <-endSend:
	//		client.Close()
	//		break
	//	}
	//}
}

func initClient() core.UserClient {
	client := client.New(
		// 服务连接地址
		client.Addr("127.0.0.1:8888"),
		// 简单的消息编码
		client.Encode(func(client core.UserClient, message core.Message) []byte {
			return MyEncode(message)
		}),
		// 简单的消息解码
		client.Decode(func(client core.UserClient, data []byte) (i int, message core.Message) {
			return MyDecode(client, data)
		}),
		client.Conn(func(client core.UserClient) bool {
			// true:允许连接 false:为拒绝连接
			return true
		}),
		client.DisConn(func(client core.UserClient) {
			// 客户端断开连接
		}),
	)
	client.Dial()
	return client
}

func initServer() {
	server.New(
		// 服务监听地址
		server.Addr(":8888"),
		// 简单的消息编码
		server.Encode(func(client core.ServerClient, message core.Message) []byte {
			return MyEncode(message)
		}),
		// 简单的消息解码
		server.Decode(func(client core.ServerClient, data []byte) (i int, message core.Message) {
			return MyDecode(client, data)
		}),
		server.Message(MyServerMessage),
		server.ClientConnFilter(func(client core.ServerClient) bool {
			// true:允许连接 false:为拒绝连接
			return true
		}),
		server.ClientDisConn(func(client core.ServerClient) {
			// 客户端断开连接
		}),
	).Run()
}

// 消息处理
// 此处为简单实现,可用作消息分发
func MyServerMessage(message core.Message) {
	serverMsg := message.(core.ServerMessage)
	if serverMsg.GetData() == nil {
		log.Printf("收到客户端[%v]一条空消息\n", serverMsg.GetClient().GetAddr())
		return
	}
	switch message.GetData().(type) {
	case string:
		log.Printf("收到客户端[%v]一条字符消息:%s\n", serverMsg.GetClient().GetAddr(), message.GetData().(string))
		return
	default:
		log.Printf("收到客户端[%v]一条未知协议消息\n", serverMsg.GetClient().GetAddr())
		return
	}
}

func MyEncode(message core.Message) []byte {
	buff := new(bytes.Buffer)
	var data []byte
	if message.GetData() == nil {
		data = make([]byte, 0)
	} else {
		data = []byte( message.GetData().(string))
	}
	allLen := make([]byte, 4)
	binary.BigEndian.PutUint32(allLen, uint32(len(data)))
	buff.Write(allLen)
	buff.Write(data)
	return buff.Bytes()
}

func MyDecode(who interface{}, data []byte) (int, core.Message) {
	if len(data) >= 4 {
		allLen := binary.BigEndian.Uint32(data[:4])
		if allLen == 0 {
			switch who.(type) {
			case core.ServerClient:
				return 4, server.NewMessage(who.(core.ServerClient), nil)
			case core.UserClient:
				return 4, client.NewMessage(nil)
			}
		} else if len(data) < int(allLen)+4 {
			return 0, nil
		}
		// 默认字符串请求,可执行扩展协议
		msgData := data[4 : allLen+4]
		msg := string(msgData)
		switch who.(type) {
		case core.ServerClient:
			return int(allLen) + 4, server.NewMessage(who.(core.ServerClient), msg)
		case core.UserClient:
			return int(allLen) + 4, client.NewMessage(msg)
		}
	}
	return 0, nil
}
