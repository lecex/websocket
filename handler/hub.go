// Copyright 2013 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package handler

import (
	"encoding/json"
	"strings"
	"time"

	pb "github.com/lecex/user/core/proto/event"

	"github.com/go-redis/redis"
	newRedis "github.com/lecex/websocket/providers/redis"
	"github.com/lecex/websocket/service/util"
	"github.com/micro/go-micro/v2/util/log"
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
	Redis      *redis.Client
}

func newHub() *Hub {
	return &Hub{
		// broadcast:  make(chan []byte),
		broadcast:  make(chan *pb.Event),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		clients:    make(map[*Client]bool),
		Redis:      newRedis.NewClient(),
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
				clientKey := ""
				if strings.Index(event.DeviceInfo, client.DeviceInfo) > -1 && event.DeviceInfo != "" && client.DeviceInfo != "" {
					send = true
					// clientKey = client.DeviceInfo
				}
				if event.UserId != "" && client.UserId != "" {
					if strings.Index(event.UserId, client.UserId) > -1 {
						send = true
						// clientKey = client.UserId
					}
				}
				// 执行锁
				if event.Lock != "" && send {
					lock := &util.Lock{
						Redis: h.Redis,
					}
					if !lock.Set("Websocket:"+clientKey+":"+event.Lock, 24*time.Hour) {
						log.Error("Websocket:" + clientKey + ":" + event.Lock + ":被锁定1天")
						send = false
					}
				}

				if send {
					log.Info("Websocket:Info:", clientKey, event)
					b, err := json.Marshal(event)
					if err != nil {
						log.Error("Hub.run.json.Marshal", err)
					}
					select {
					case client.send <- b:
					default:
						close(client.send)
						delete(h.clients, client)
					}
				}
			}
		}
	}
}
