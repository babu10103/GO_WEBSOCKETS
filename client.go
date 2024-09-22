package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

/*
	websocket connection does not support actually allows only one concurrent
	concurrent writer at a time.
	we can use unbuffered channel to prevent the connection from getting
	too many writes at the same time.
*/

var (
	pongWait     = 10 * time.Second
	pingInterval = (pongWait * 9) / 10
)

type ClientList map[*Client]bool

type Client struct {
	connection *websocket.Conn
	manager    *Manager
	chatroom   string
	// egress is used to avoid concurrent writes to the websocket connection
	egress chan Event
}

func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		connection: conn,
		manager:    manager,
		egress:     make(chan Event),
	}
}

func (c *Client) readMessages() {
	defer func() {
		c.manager.removeClient(c)
		log.Printf("Client %p disconnected", c)
	}()

	if err := c.connection.SetReadDeadline(time.Now().Add(pongWait)); err != nil {
		log.Printf("Error setting read deadline for client %p: %v", c, err)
		return
	}

	c.connection.SetReadLimit(512)

	c.connection.SetPongHandler(c.pongHandler)

	log.Printf("Start reading messages for client %p", c)

	for {
		_, payload, err := c.connection.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error reading message: %v", err)
			} else {
				log.Printf("[readMessages]: Error: %+v", err)
			}
			break
		}

		var request Event

		if err := json.Unmarshal(payload, &request); err != nil {
			log.Printf("Error unmarshalling message to event: %v", err)
			break
		}
		// Route the event
		if err := c.manager.routeEvent(request, c); err != nil {
			log.Println("Error routing event to handler:", err)
		}

	}
}

func (c *Client) writeMessages() {
	ticker := time.NewTicker(pingInterval)
	defer func() {
		ticker.Stop()
		c.manager.removeClient(c)
	}()

	log.Printf("Start writing messages for client %p", c)

	for {
		select {
		case message, ok := <-c.egress:
			if !ok {
				if err := c.connection.WriteMessage(websocket.CloseMessage, nil); err != nil {
					log.Printf("connection closed: %v", err)
				}
				return
			}
			// // var decodedPayload map[string]interface{}
			// var newMsg NewMessageEvent
			// if err := json.Unmarshal(message.Payload, &newMsg); err != nil {
			// 	log.Printf("Error decoding message payload: %v", err)
			// } else {
			// 	log.Printf("Decoded message payload: %+v", newMsg)
			// }

			// temp, err := json.Marshal(newMsg)
			// log.Printf("Temp: %+v", string(temp))

			data, err := json.Marshal(message)
			if err != nil {
				log.Printf("Error marshalling message to json: %v", err)
				break
			}
			if err := c.connection.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("Error writing message to client %p: %v", c, err)
			}
			log.Println("message sent")
		case <-ticker.C:
			log.Println("ping")
			if err := c.connection.WriteMessage(websocket.PingMessage, []byte(``)); err != nil {
				log.Printf("Error sending ping message: %v", err)
				return
			}
		}

	}

}

func (c *Client) pongHandler(pongMessage string) error {
	log.Printf("pong: %v", pongMessage)
	return c.connection.SetReadDeadline(time.Now().Add(pongWait))
}
