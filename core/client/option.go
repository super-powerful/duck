package client

import "github.com/super-powerful/duck/core"

func Addr(addr string) Options {
	return func(c core.UserClient) {
		c.(*_Client_).Addr = addr
	}
}

func Encode(encodeEvent EncodeEvent) Options {
	return func(c core.UserClient) {
		c.(*_Client_).EncodeEvent = encodeEvent
	}
}

func Decode(decodeEvent DecodeEvent) Options {
	return func(c core.UserClient) {
		c.(*_Client_).DecodeEvent = decodeEvent
	}
}

func Message(messageEvent MessageEvent) Options {
	return func(c core.UserClient) {
		c.(*_Client_).MessageEvent = messageEvent
	}
}

func Conn(connEvent ConnEvent) Options {
	return func(c core.UserClient) {
		c.(*_Client_).ConnEvent = connEvent
	}
}

func DisConn(disConnEvent DisConnEvent) Options {
	return func(c core.UserClient) {
		c.(*_Client_).DisConnEvent = disConnEvent
	}
}
