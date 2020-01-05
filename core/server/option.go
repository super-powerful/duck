package server

func Addr(addr string) Options {
	return func(s *_Server_) {
		s.Addr = addr
	}
}

func Encode(encodeEvent EncodeEvent) Options {
	return func(s *_Server_) {
		s.EncodeEvent = encodeEvent
	}
}

func Decode(decodeEvent DecodeEvent) Options {
	return func(s *_Server_) {
		s.DecodeEvent = decodeEvent
	}
}

func Message(messageEvent MessageEvent) Options {
	return func(s *_Server_) {
		s.MessageEvent = messageEvent
	}
}

func ClientConnFilter(clientConnFilterEvent ClientConnFilterEvent) Options {
	return func(s *_Server_) {
		s.ClientConnFilterEvent = clientConnFilterEvent
	}
}

func ClientDisConn(clientDisConnEvent ClientDisConnEvent) Options {
	return func(s *_Server_) {
		s.ClientDisConnEvent = clientDisConnEvent
	}
}

