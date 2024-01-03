package stream

import "github.com/gorilla/websocket"

const (
	maxReadMessageSize  = 1024 * 1024
	maxWriteMessageSize = 1024 * 1024
)

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  maxReadMessageSize,
		WriteBufferSize: maxWriteMessageSize,
	}
)
