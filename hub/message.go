package hub

type WebsocketMessage struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}
