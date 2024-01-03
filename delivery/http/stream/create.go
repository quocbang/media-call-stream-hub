package stream

import (
	"crypto/sha256"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/pion/webrtc/v4"

	"github.com/quocbang/media-call-stream-hub/hub"
)

type stream struct {
}

type CreateStreamRoomResponse struct {
	RoomID   uuid.UUID `json:"room_id"`
	StreamID string    `json:"stream_id"`
}

func (s *stream) Create(ctx echo.Context) error {
	// create room id
	id := uuid.New()

	// start stream room
	streamID, err := startStreamRoom(id)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, err.Error()) // TODO: resolve error
	}

	return ctx.JSON(http.StatusOK, CreateStreamRoomResponse{
		RoomID:   uuid.New(),
		StreamID: streamID,
	})
}

func startStreamRoom(roomID uuid.UUID) (string, error) {
	hub.Mutex.Lock()
	defer hub.Mutex.Unlock()

	// create hashed stream id
	hash := sha256.New()
	hash.Write([]byte(roomID.String()))
	streamID := fmt.Sprintf("%x", hash.Sum(nil))

	// if room id was existed return error
	if _, ok := hub.Rooms[roomID.String()]; ok {
		return "", fmt.Errorf("unauthorized for this room")
	}

	h := hub.NewHub()
	p := &hub.Peers{}
	p.Track = make(map[string]*webrtc.TrackLocalStaticRTP)
	room := &hub.Room{
		Peers: p,
		Hub:   h,
	}

	hub.Rooms[roomID.String()] = room
	hub.Stream[streamID] = room

	go h.Run()
	return streamID, nil
}
