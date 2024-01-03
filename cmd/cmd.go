package cmd

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	repoHTTP "github.com/quocbang/media-call-stream-hub/delivery/http"
	"github.com/quocbang/media-call-stream-hub/delivery/websocket"
)

const (
	HTTP_PORT      int = 9090
	WEBSOCKET_PORT int = 9091
)

func Run() {
	echo := echo.New()

	// logger

	// middleware
	echo.Use(middleware.CORS())
	echo.Use(middleware.Recover())
	echo.Use(middleware.Logger())

	// http handler
	repoHTTP.NewHTTPHandlers(echo)

	// websocket handler
	handlers := websocket.NewWebsocketHandlers()

	// start notification
	log.Printf("listen http server on port: [%d] \n", HTTP_PORT)
	log.Printf("listen websocket server on port: [%d] \n", WEBSOCKET_PORT)

	if err := echo.Start(fmt.Sprintf(":%d", HTTP_PORT)); err != nil {
		log.Fatalf("failed to start echo, error: %v", err)
	}

	wg := sync.WaitGroup{}
	wg.Add(2)
	// start http
	go func() {
		defer wg.Done()
		echo.Logger.Fatal(echo.Start(fmt.Sprintf(":%d", HTTP_PORT)))
	}()
	// start websocket
	go func() {
		defer wg.Done()
		if err := http.ListenAndServe(fmt.Sprintf(":%d", WEBSOCKET_PORT), handlers); err != nil {
			log.Fatalf("failed to listen and serve websocket, error: %v", err)
		}
	}()
	wg.Wait()
}
