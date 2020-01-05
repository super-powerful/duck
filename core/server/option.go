package server

import "github.com/super-powerful/duck/core"

func Addr(addr string) Options {
	return func(s core.Server) {
		s.(*_Server_).Addr = addr
	}
}

func Encode(encodeEvent EncodeEvent) Options {
	return func(s core.Server) {
		s.(*_Server_).EncodeEvent = encodeEvent
	}
}

func Decode(decodeEvent DecodeEvent) Options {
	return func(s core.Server) {
		s.(*_Server_).DecodeEvent = decodeEvent
	}
}

func Message(messageEvent MessageEvent) Options {
	return func(s core.Server) {
		s.(*_Server_).MessageEvent = messageEvent
	}
}

func ClientConnFilter(clientConnFilterEvent ClientConnFilterEvent) Options {
	return func(s core.Server) {
		s.(*_Server_).ClientConnFilterEvent = clientConnFilterEvent
	}
}

func ClientDisConn(clientDisConnEvent ClientDisConnEvent) Options {
	return func(s core.Server) {
		s.(*_Server_).ClientDisConnEvent = clientDisConnEvent
	}
}
