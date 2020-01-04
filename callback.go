package duck

import (
	"bytes"
	"encoding/binary"
)

// DefaultEncoding 默认编码
func DefaultEncoding(s *_Server_) {
	s.EncodeMessageCallback = func(buffer *bytes.Buffer, message *Message) {
		size := len(message.Data)
		allLenBytes := make([]byte, 4)
		binary.BigEndian.PutUint32(allLenBytes, uint32(size))

		buffer.Write(allLenBytes)

		buffer.Write(message.Data)
	}
}

// DefaultDecoding 默认解码
func DefaultDecoding(s *_Server_) {
	s.DecodeMessageCallback = func(buffer *bytes.Buffer) []*Message {
		messages := make([]*Message, 0)
		for {
			if buffer.Len() < 4 {
				break
			}
			allLenBytes := buffer.Bytes()[:4]
			allLen := binary.BigEndian.Uint32(allLenBytes)
			if uint32(buffer.Len()-4) < allLen {
				break
			}
			buffer.Next(4)
			dataBytes := make([]byte, allLen)
			buffer.Read(dataBytes)
			message := &Message{
				Data:    dataBytes,
				Context: nil,
				Err:     nil,
			}
			messages = append(messages, message)
		}
		return messages
	}
}
