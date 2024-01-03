package stream

import (
	"net/http"

	"github.com/quocbang/media-call-stream-hub/hub"
	"github.com/quocbang/media-call-stream-hub/hub/streams"
)

func start(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		w.Write([]byte(`failed to upgrade`))
	}

	// get stream id
	streamID := r.Form.Get("streamID")
	if streamID == "" {
		w.Write([]byte(`missing streamID`))
		return
	}

	// TODO: check is owner

	//
	hub.Mutex.Lock()
	if stream, ok := hub.Stream[streamID]; ok {
		hub.Mutex.Unlock()
		streams.StreamConn(conn, stream.Peers)
		return
	}

	hub.Mutex.Unlock()
}
