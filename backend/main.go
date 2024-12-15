package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

func main() {
	// Setup WebSocket endpoint for signaling
	http.HandleFunc("/ws", HandleWebSocket)

	// Serve static files for the frontend
	fs := http.FileServer(http.Dir("./frontend"))
	http.Handle("/", fs)

	// Start the server
	port := ":8080"
	log.Println("Server running on http://localhost" + port)
	log.Fatal(http.ListenAndServe(port, nil))
}

// SIGNALING //
type Message struct {
	Type    string  `json:"type"`
	Payload Payload `json:"payload"`
}

type Payload struct {
	Sdp              string         `json:"sdp"`
	Type             webrtc.SDPType `json:"type"`
	Candidate        string         `json:"candidate"`
	SdpMLineIndex    *uint16        `json:"sdpMLineIndex"`
	SdpMid           string         `json:"sdpMid"`
	UsernameFragment string         `json:"usernameFragment"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var peerConnections = make(map[*websocket.Conn]*webrtc.PeerConnection)
var trackLocalMap = make(map[string]*webrtc.TrackLocalStaticRTP)

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	fmt.Println("HandleWebSocket")
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}
	defer conn.Close()

	pc, err := CreatePeerConnection()
	if err != nil {
		log.Println("Failed to create PeerConnection:", err)
		return
	}
	peerConnections[conn] = pc

	pc.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		log.Println("OnICECandidate")

		if candidate != nil {
			candidateJSON, _ := json.Marshal(candidate.ToJSON())
			conn.WriteMessage(websocket.TextMessage, candidateJSON)
		}
	})

	fmt.Println("HandleWebSocket1")

	pc.OnTrack(func(track *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		log.Println("Track received:", track.Kind(), track.ID())

		// Create a local track if not already present
		localTrack, ok := trackLocalMap[track.ID()]
		if !ok {
			localTrack, err = webrtc.NewTrackLocalStaticRTP(
				webrtc.RTPCodecCapability{
					MimeType:    track.Codec().MimeType,
					ClockRate:   track.Codec().ClockRate,
					Channels:    track.Codec().Channels,
					SDPFmtpLine: track.Codec().SDPFmtpLine,
				},
				track.ID(),
				track.StreamID(),
			)
			if err != nil {
				log.Println("Failed to create local track:", err)
				return
			}
			trackLocalMap[track.ID()] = localTrack
		}

		// Start relaying RTP packets from remote track
		go func() {
			buf := make([]byte, 1500) // Standard MTU size
			for {
				n, _, readErr := track.Read(buf)
				if readErr != nil {
					log.Println("Track read error:", readErr)
					return
				}

				// Broadcast to all PeerConnections
				for _, peer := range peerConnections {
					senders := peer.GetSenders()
					trackExists := false

					// Ensure the local track isn't added multiple times
					for _, sender := range senders {
						if sender.Track() != nil && sender.Track().ID() == localTrack.ID() {
							trackExists = true
							break
						}
					}

					if !trackExists {
						if _, err := peer.AddTrack(localTrack); err != nil {
							log.Println("Failed to add track to PeerConnection:", err)
							return
						}
					}

					if _, writeErr := localTrack.Write(buf[:n]); writeErr != nil {
						log.Println("Track write error:", writeErr)
					}
				}
			}
		}()
	})

	fmt.Println("HandleWebSocket3s")

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Read error:", err)
			delete(peerConnections, conn)
			break
		}

		fmt.Println("HandleWebSocket message: ", string(message))

		var msg Message
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Println("Failed to unmarshal message:", err)
			continue
		}
		fmt.Println("msg.Type", msg.Type)

		switch msg.Type {
		case "offer":
			var o = webrtc.SessionDescription{}
			o.SDP = msg.Payload.Sdp
			o.Type = msg.Payload.Type
			handleSDPOffer(conn, pc, o)
		case "answer":
			var a = webrtc.SessionDescription{}
			a.SDP = msg.Payload.Sdp
			a.Type = msg.Payload.Type
			handleSDPAnswer(conn, pc, a)
		case "ice-candidate":
			var i = webrtc.ICECandidateInit{}
			i.Candidate = msg.Payload.Candidate
			i.SDPMLineIndex = msg.Payload.SdpMLineIndex
			i.SDPMid = &msg.Payload.SdpMid
			i.UsernameFragment = &msg.Payload.UsernameFragment
			handleICECandidate(conn, pc, i)
		default:
			log.Println("Unknown message type:", msg.Type)
		}
	}
}

func handleSDPOffer(conn *websocket.Conn, pc *webrtc.PeerConnection, offer webrtc.SessionDescription) {
	// var offer webrtc.SessionDescription
	// if err := json.Unmarshal([]byte(payload), &offer); err != nil {
	// 	log.Println("Failed to unmarshal SDP offer:", err)
	// 	return
	// }

	if err := pc.SetRemoteDescription(offer); err != nil {
		log.Println("Failed to set remote description:", err)
		return
	}

	answer, err := pc.CreateAnswer(nil)
	if err != nil {
		log.Println("Failed to create SDP answer:", err)
		return
	}

	if err := pc.SetLocalDescription(answer); err != nil {
		log.Println("Failed to set local description:", err)
		return
	}

	answerJSON, _ := json.Marshal(answer)
	fmt.Println("answerJSON:: ", string(answerJSON))
	conn.WriteMessage(websocket.TextMessage, answerJSON)
}

func handleSDPAnswer(conn *websocket.Conn, pc *webrtc.PeerConnection, answer webrtc.SessionDescription) {
	// var answer webrtc.SessionDescription
	// if err := json.Unmarshal([]byte(payload), &answer); err != nil {
	// 	log.Println("Failed to unmarshal SDP answer:", err)
	// 	return
	// }

	if err := pc.SetRemoteDescription(answer); err != nil {
		log.Println("Failed to set remote description:", err)
	}
}

func handleICECandidate(conn *websocket.Conn, pc *webrtc.PeerConnection, candidate webrtc.ICECandidateInit) {
	// var candidate webrtc.ICECandidateInit
	// if err := json.Unmarshal([]byte(payload), &candidate); err != nil {
	// 	log.Println("Failed to unmarshal ICE candidate:", err)
	// 	return
	// }

	if err := pc.AddICECandidate(candidate); err != nil {
		log.Println("Failed to add ICE candidate:", err)
	}
}

// MEDIA //
var config = webrtc.Configuration{}

func CreatePeerConnection() (*webrtc.PeerConnection, error) {
	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return nil, err
	}

	// Add all existing tracks to the new PeerConnection
	for _, track := range trackLocalMap {
		if _, err := peerConnection.AddTrack(track); err != nil {
			log.Println("Failed to add track to PeerConnection:", err)
		}
	}

	return peerConnection, nil
}
