package http

import (
	"github.com/labstack/echo/v4"
	"github.com/quocbang/media-call-stream-hub/delivery/http/stream"
)

func NewHTTPHandlers(echo *echo.Echo) {
	// api
	api := echo.Group("/api")

	// stream
	gStream := api.Group("/stream")
	stream.Init(gStream)
}
