package main

import (
	"encoding/json"
	"fmt"
	"time"
)

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type EventHandler func(event Event, c *Client) error

const (
	EventSendMessage = "send_message"
	EventChangeRoom  = "change_room"
	EventNewMessage  = "new_message"
)

type SendMessageEvent struct {
	Message string `json:"message"`
	From    string `json:"from"`
}

type ChangeRoomEvent struct {
	Name string `json:"name"`
}

type NewMessageEvent struct {
	SendMessageEvent
	Sent time.Time `json:"sent"`
}

func SendMessageHandler(event Event, c *Client) error {
	var chatEvent SendMessageEvent
	if err := json.Unmarshal(event.Payload, &chatEvent); err != nil {
		return fmt.Errorf("bad payload in request: %v", err)
	}

	var broadcastMessage NewMessageEvent

	broadcastMessage.Sent = time.Now()
	broadcastMessage.Message = chatEvent.Message
	broadcastMessage.From = chatEvent.From

	data, err := json.Marshal(broadcastMessage)

	if err != nil {
		return fmt.Errorf("failed to marshal broadcast message: %v", err)
	}
	// place payload into an event
	var outgoingEvent Event
	outgoingEvent.Payload = data
	outgoingEvent.Type = EventNewMessage

	// broadcast to all other clients
	for client := range c.manager.clients {
		if client.chatroom == c.chatroom {
			client.egress <- outgoingEvent
		}
	}

	return nil
}

// ChatRoomHandler will handle switching of chatrooms between clients
func ChatRoomHandler(event Event, c *Client) error {
	// Marshal payload into ChangeRoomEvent format
	var changeRoomEvent ChangeRoomEvent
	if err := json.Unmarshal([]byte(event.Payload), &changeRoomEvent); err != nil {
		return fmt.Errorf("bad payload in the request: %v", err)
	}
	c.chatroom = changeRoomEvent.Name
	return nil
}
