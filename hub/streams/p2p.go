package streams

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v4"

	"github.com/quocbang/media-call-stream-hub/hub"
)

func StreamConn(conn *websocket.Conn, p *hub.Peers) {
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.1.google.com:19302"},
			},
		},
		ICETransportPolicy: webrtc.ICETransportPolicyRelay,
	}

	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		log.Println(err) // TODO: should resolve error
		return
	}
	defer peerConnection.Close()

	codecTypes := []webrtc.RTPCodecType{webrtc.RTPCodecTypeVideo, webrtc.RTPCodecTypeAudio}
	for _, ct := range codecTypes {
		if _, err := peerConnection.AddTransceiverFromKind(ct, webrtc.RTPTransceiverInit{
			Direction: webrtc.RTPTransceiverDirectionRecvonly,
		}); err != nil {
			log.Println(err) // TODO: should resolve error
			return
		}
	}

	newPeer := hub.ConnectionsState{
		PeerConnection: peerConnection,
		WebSocketConn:  conn,
	}

	// add new p2p
	p.Mutex.Lock()
	p.Connections = append(p.Connections, newPeer)
	p.Mutex.Unlock()

	log.Printf("conn %v \n", p.Connections)

	peerConnection.OnICECandidate(func(i *webrtc.ICECandidate) {
		if i == nil {
			return
		}

		candidateString, err := json.Marshal(i.ToJSON())
		if err != nil {
			log.Println(err) // TODO: should resolve error
			return
		}

		if err := newPeer.WebSocketConn.WriteJSON(&hub.WebsocketMessage{
			Event: hub.Candidate,
			Data:  string(candidateString),
		}); err != nil {
			log.Println(err) // TODO: should resolve error
		}
	})

	peerConnection.OnConnectionStateChange(func(pcs webrtc.PeerConnectionState) {
		switch pcs {
		case webrtc.PeerConnectionStateFailed:
			if err := peerConnection.Close(); err != nil {
				log.Println(err)
			}
		case webrtc.PeerConnectionStateClosed:
			p.SignalPeerConnections()
		}
	})

	p.SignalPeerConnections()
	message := &hub.WebsocketMessage{}
	for {
		_, raw, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		} else if err := json.Unmarshal(raw, &message); err != nil {
			log.Println(err)
			return
		}

		switch message.Event {
		case hub.Candidate:
			candidate := webrtc.ICECandidateInit{}
			if err := json.Unmarshal([]byte(message.Data), &candidate); err != nil {
				log.Println(err)
				return
			}

			if err := peerConnection.AddICECandidate(candidate); err != nil {
				log.Println(err)
				return
			}
		case hub.Answer:
			answer := webrtc.SessionDescription{}
			if err := json.Unmarshal([]byte(message.Data), &answer); err != nil {
				log.Println(err)
				return
			}

			if err := peerConnection.SetRemoteDescription(answer); err != nil {
				log.Println(err)
				return
			}
		}
	}
}
