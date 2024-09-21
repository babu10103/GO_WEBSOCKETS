package main

import (
	"log"

	"github.com/gorilla/websocket"
)

/*
	websocket connection does not support actually allows only one concurrent
	concurrent writer at a time.
	we can use unbuffered channel to prevent the connection from getting
	too many writes at the same time.
*/

type ClientList map[*Client]bool

type Client struct {
	connection *websocket.Conn
	manager    *Manager
}

func NewClient(conn *websocket.Conn, manager *Manager) *Client {
	return &Client{
		connection: conn,
		manager:    manager,
	}
}

func (c *Client) readMessages() {
	defer func() {
		c.manager.removeClient(c)
	}()
	for {
		// messageTypes : ping, ping, data, binary
		messageType, paylaod, err := c.connection.ReadMessage()

		if err != nil {
			// we are trying to log the error message only when there is something wrong happened to connection
			// if connection was closed for normal reasons only like client only closed the connection we won't log the error
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error reading message: %v", err)
			}
			break
		}
		log.Println(messageType)
		log.Println(string(paylaod))
	}
}
