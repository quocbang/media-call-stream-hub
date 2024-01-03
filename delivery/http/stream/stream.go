package stream

import (
	"github.com/labstack/echo/v4"
)

func Init(group *echo.Group) {
	s := &stream{}
	// create stream
	group.POST("", s.Create)
}
