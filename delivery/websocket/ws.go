package websocket

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/quocbang/media-call-stream-hub/delivery/websocket/stream"
)

func NewWebsocketHandlers() http.Handler {
	router := mux.NewRouter()

	// stream
	stream.Init(router)

	return router
}
