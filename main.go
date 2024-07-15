package main

import (
	"log"
	"net/url"
	"strconv"
	"time"
	"math/rand"
	"encoding/json"

	"github.com/gorilla/websocket"
	"github.com/pion/webrtc/v3"
)

func main() {
	// Define the WebSocket server address
	serverAddr := "localhost:8090"
	u := url.URL{Scheme: "ws", Host: serverAddr, Path: "/"}

	// Create a new WebSocket connection
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	log.Printf("Connected to %s", serverAddr)

	Join(c, "tt", RandomNumber())

	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}

		// Parse the JSON message
		var msg map[string]interface{}
		if err := json.Unmarshal([]byte(message), &msg); err != nil {
			log.Println("unmarshal:", err)
			continue
		}

		// log.Printf("Received: %s", message)
		switch msg["type"] {
		case "joined":
			handleJoined(c, msg["body"])
		case "offer_sdp_received":
			handleOffer(c, msg["body"])
		default:
			log.Println("Unknown message type")
		}
	}
}

func Join(c *websocket.Conn, channelName string, userId int) {
	// {"type":"join","body":{"channelName":"tt","userId":267887}}
	c.WriteMessage(websocket.TextMessage, []byte(`{"type":"join","body":{"channelName":"`+channelName+`","userId":`+strconv.Itoa(userId)+`}}`))
}

// return 6-digit random number
func RandomNumber() int {	
	rand.Seed(time.Now().UnixNano())
	return rand.Intn(900000) + 100000
}

func sendOffer() {
	// Initialize the ICE configuration
	config := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"},
			},
		},
	}

	// Create a new RTCPeerConnection
	localPeerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		log.Fatalf("Failed to create the peer connection: %v", err)
	}

	// Set the OnICECandidate handler
	localPeerConnection.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate == nil {
			// Gathering of ICE candidates is complete
			log.Println("ICE gathering complete")
			return
		}

		// Convert the ICE candidate to JSON
		candidateJSON, err := json.Marshal(candidate.ToJSON())
		if err != nil {
			log.Fatalf("Failed to marshal the ICE candidate: %v", err)
			return
		}

		log.Printf("ICE candidate: %s", candidateJSON)
	})

	// Create an offer
	offer, err := localPeerConnection.CreateOffer(nil)
	if err != nil {
		log.Fatalf("Failed to create the offer: %v", err)
	}

	// Set the local description of the peer
	err = localPeerConnection.SetLocalDescription(offer)
	if err != nil {
		log.Fatalf("Failed to set the local description: %v", err)
	}

	// Now that gathering is complete, the local description is the final one
	updatedOffer := localPeerConnection.LocalDescription()
	if updatedOffer == nil {
		log.Fatalf("Failed to get updated offer: %v", err)
	}
	
	// Marshal the offer to JSON to the remote peer
	offerJSON, err := json.Marshal(updatedOffer)
	if err != nil {
		log.Fatalf("Failed to marshal the offer: %v", err)
	}

	log.Printf("Offer sent: %s", offerJSON)
	
}

func handleJoined(c *websocket.Conn, data interface{}) {
	body, ok := data.([]interface{})
	if !ok {
		log.Fatalf("Failed to parse the body: %v", data)
	}

	jsonData, err := json.Marshal(body)
	if err != nil {
		log.Fatalf("Error marshaling data: %v", err)
	}

	var uids []string
	err = json.Unmarshal([]byte(jsonData), &uids)
	if err != nil {
		log.Println("unmarshal:", err)	
	}

	log.Printf("uids: %v", uids)
	// sendOffer()
}

func handleOffer(c *websocket.Conn, data interface{}) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Fatalf("Error marshaling data: %v", err)
	}

	var sdp webrtc.SessionDescription
	if err := json.Unmarshal([]byte(jsonData), &sdp); err != nil {
		log.Fatalf("Failed to unmarshal the offer: %v", err)
	}
	// log.Printf("offer sdp: %v", sdp)

	// Create a new remote PeerConnection
	remotePeerConnection, err := webrtc.NewPeerConnection(webrtc.Configuration{})
	if err != nil {
		log.Fatalf("Failed to create the remote peer connection: %v", err)
	}

	// Set the remote SessionDescription
	err = remotePeerConnection.SetRemoteDescription(sdp)
	if err != nil {
		log.Fatalf("Failed to set the remote description: %v", err)
	}

	// Create an answer
	answer, err := remotePeerConnection.CreateAnswer(nil)
	if err != nil {
		log.Fatalf("Failed to create the answer: %v", err)
	}

	// Set the local description of the remote peer
	err = remotePeerConnection.SetLocalDescription(answer)
	if err != nil {
		log.Fatalf("Failed to set the local description: %v", err)
	}

	// Marshal the answer to JSON
	answerJSON, err := json.Marshal(remotePeerConnection.LocalDescription())
	if err != nil {
		log.Fatalf("Failed to marshal the answer: %v", err)
	}

	c.WriteMessage(websocket.TextMessage, []byte(`{"type":"send_answer","body":`+string(answerJSON)+`}`))
}