package stream

import (
	"net/http"

	"github.com/gorilla/mux"
)

func Init(router *mux.Router) {
	router.HandleFunc("/stream/:streamID", start)
	router.HandleFunc("/stream/viewer/:streamID", func(w http.ResponseWriter, r *http.Request) {})
}
