// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package handler

import (
	"strings"

	pb "github.com/lecex/core/proto/event"
)

// Hub maintains the set of active clients and broadcasts messages to the
// clients.
type Hub struct {
	// Registered clients.
	clients map[*Client]bool

	// Inbound messages from the clients.
	broadcast chan *pb.Event

	// Register requests from the clients.
	register chan *Client

	// Unregister requests from clients.
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		// broadcast:  make(chan []byte),
		broadcast:  make(chan *pb.Event),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
			}
		case event := <-h.broadcast:
			for client := range h.clients {
				send := false
				if strings.Index(event.DeviceInfo, client.DeviceInfo) > -1 && event.DeviceInfo != "" {
					send = true
				}
				if strings.Index(event.UserId, client.UserId) > -1 && event.UserId != "" {
					send = true
				}
				if send {
					select {
					case client.send <- event.Data:
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
		}
	}
}
