package main

import (
	"log"
	"net/url"

	"github.com/gorilla/websocket"
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

	// Connection is established, you can now read and write messages
	// Example: c.WriteMessage(websocket.TextMessage, []byte("your message here"))
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			return
		}
		log.Printf("Received: %s", message)
	}
}
