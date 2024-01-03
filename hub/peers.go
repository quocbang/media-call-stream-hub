package hub

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/pion/rtcp"
	"github.com/pion/webrtc/v4"
)

type Peers struct {
	Mutex       *sync.RWMutex
	Connections []ConnectionsState
	Track       map[string]*webrtc.TrackLocalStaticRTP
}

func (p *Peers) AddTrack(t *webrtc.TrackRemote) *webrtc.TrackLocalStaticRTP {
	p.Mutex.Lock()
	defer func() {
		p.Mutex.Unlock()
		p.SignalPeerConnections()
	}()

	trackLocal, err := webrtc.NewTrackLocalStaticRTP(t.Codec().RTPCodecCapability, t.ID(), t.StreamID())
	if err != nil {
		log.Println(err.Error())
		return nil
	}

	p.Track[t.ID()] = trackLocal
	return trackLocal
}

func (p *Peers) RemoveTrack(t *webrtc.TrackLocalStaticRTP) {
	p.Mutex.Lock()
	defer func() {
		p.Mutex.Unlock()
		p.SignalPeerConnections()
	}()

	delete(p.Track, t.ID())
}

func (p *Peers) SignalPeerConnections() {
	p.Mutex.Lock()
	defer func() {
		p.Mutex.Unlock()
		p.DispatchKeyFrame()
	}()

	attemptSync := func() (tryAgain bool) {
		for i := range p.Connections {
			if p.Connections[i].PeerConnection.ConnectionState() == webrtc.PeerConnectionStateClosed {
				p.Connections = append(p.Connections[:i], p.Connections[i+1:]...)
				log.Println("a", p.Connections)
				return true
			}

			existingSenders := map[string]bool{}
			for _, sender := range p.Connections[i].PeerConnection.GetSenders() {
				if sender.Track() == nil {
					continue
				}

				existingSenders[sender.Track().ID()] = true

				if _, ok := p.Track[sender.Track().ID()]; !ok {
					if err := p.Connections[i].PeerConnection.RemoveTrack(sender); err != nil {
						return true
					}
				}
			}

			for _, receiver := range p.Connections[i].PeerConnection.GetReceivers() {
				if receiver.Track() == nil {
					continue
				}

				existingSenders[receiver.Track().ID()] = true
			}

			for trackID := range p.Track {
				if _, ok := existingSenders[trackID]; !ok {
					if _, err := p.Connections[i].PeerConnection.AddTrack(p.Track[trackID]); err != nil {
						return true
					}
				}
			}

			offer, err := p.Connections[i].PeerConnection.CreateOffer(nil)
			if err != nil {
				return true
			}

			if err = p.Connections[i].PeerConnection.SetLocalDescription(offer); err != nil {
				return true
			}

			offerString, err := json.Marshal(offer)
			if err != nil {
				return true
			}

			if err = p.Connections[i].WebSocketConn.WriteJSON(&WebsocketMessage{
				Event: "offer",
				Data:  string(offerString),
			}); err != nil {
				return true
			}
		}

		return
	}

	for syncAttempt := 0; ; syncAttempt++ {
		if syncAttempt == 25 {
			go func() {
				time.Sleep(time.Second * 3)
				p.SignalPeerConnections()
			}()
			return
		}

		if !attemptSync() {
			break
		}
	}
}

func (p *Peers) DispatchKeyFrame() {
	p.Mutex.Lock()
	defer p.Mutex.Unlock()

	for i := range p.Connections {
		for _, receiver := range p.Connections[i].PeerConnection.GetReceivers() {
			if receiver.Track() == nil {
				continue
			}

			_ = p.Connections[i].PeerConnection.WriteRTCP([]rtcp.Packet{
				&rtcp.PictureLossIndication{
					MediaSSRC: uint32(receiver.Track().SSRC()),
				},
			})
		}
	}
}
