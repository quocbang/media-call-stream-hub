package hub

import (
	"sync"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"
)

type Client struct {
	Mutex *sync.RWMutex
	Conn  *websocket.Conn
	Send  chan []byte
}

type ConnectionsState struct {
	PeerConnection *webrtc.PeerConnection
	WebSocketConn  *websocket.Conn
}

type Hub struct {
	clients    map[*Client]struct{}
	broadcast  chan []byte
	register   chan *Client
	unRegister chan *Client
}

type Room struct {
	Peers *Peers
	Hub   *Hub
}

var (
	Mutex  *sync.RWMutex
	Rooms  map[string]*Room
	Stream map[string]*Room
)

func init() {
	Rooms = make(map[string]*Room)
	Stream = make(map[string]*Room)
}

func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]struct{}),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unRegister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = struct{}{}
		case client := <-h.unRegister:
			close(client.Send)
			delete(h.clients, client)
		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
		}
	}
}
