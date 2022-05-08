package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	cli "github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/metadata"
	"github.com/micro/go-micro/v2/util/log"

	client "github.com/lecex/user/core/client"
	"github.com/lecex/user/core/env"
)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second

	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second

	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10

	// Maximum message size allowed from peer.
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

var upgrader = websocket.Upgrader{
	CheckOrigin:     func(r *http.Request) bool { return true },
	ReadBufferSize:  1024 * 10,
	WriteBufferSize: 1024 * 10,
}

// Client is a middleman between the websocket connection and the hub.
type Client struct {
	hub *Hub
	// The websocket connection.
	conn *websocket.Conn

	// Buffered channel of outbound messages.
	send chan []byte
	// X-CSRF-Token
	token string
	// 用户ID
	UserId string
	// 设备信息
	DeviceInfo string
}

// readPump pumps messages from the websocket connection to the hub.
//
// The application runs readPump in a per-connection goroutine. The application
// ensures that there is at most one reader on a connection by executing all
// reads from this goroutine.
func (c *Client) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()
	c.conn.SetReadLimit(maxMessageSize)
	c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error { c.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Error("error: %v", err)
			}
			break
		}
		res, err := c.call(message)
		if err != nil {
			log.Error("error: %v", err)
		}
		c.send <- res
	}
}

func (c *Client) call(req []byte) (message []byte, err error) {
	r := make(map[string]interface{})
	err = json.Unmarshal(req, &r)
	if err != nil {
		return
	}
	// 获取设备信息
	if deviceInfo, ok := r["deviceInfo"]; ok {
		c.DeviceInfo = deviceInfo.(string)
	}
	var service, method string
	if m, ok := r["token"]; ok {
		c.token = m.(string)
	}
	if s, ok := r["service"]; ok {
		service = s.(string)
	}
	if m, ok := r["method"]; ok {
		method = m.(string)
	}
	if service == "" {
		return nil, fmt.Errorf("服务不允许为空")
	}
	if method == "" {
		return nil, fmt.Errorf("服务方法不允许为空")
	}
	// 构建上下文 context
	meta := map[string]string{}
	if c.token != "" {
		meta["X-Csrf-Token"] = c.token
	}
	ctx := metadata.NewContext(context.TODO(), meta)
	// 上下文构建完成
	request := map[string]interface{}{}
	if re, ok := r["request"]; ok {
		request = re.(map[string]interface{})
	}
	res := make(map[string]interface{})
	err = client.Call(ctx, env.Getenv("MICRO_API_NAMESPACE", "go.micro.api.")+service, method, &request, &res, cli.WithContentType("application/json"))
	if err != nil {
		return nil, err
	}
	if u, ok := res["user"]; ok {
		if i, ok := u.(map[string]interface{})["id"]; ok {
			c.UserId = i.(string)
		}
	}
	jsonRes, err := json.Marshal(res)
	if err != nil {
		return
	}
	jsonRes = bytes.TrimSpace(bytes.Replace([]byte(jsonRes), newline, space, -1))
	return jsonRes, err
}

// writePump pumps messages from the hub to the websocket connection.
//
// A goroutine running writePump is started for each connection. The
// application ensures that there is at most one writer to a connection by
// executing all writes from this goroutine.
func (c *Client) writePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// The hub closed the channel.
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message.
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(newline)
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// serveWs handles websocket requests from the peer.
func serveWs(hub *Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err)
		return
	}
	client := &Client{
		hub:  hub,
		conn: conn,
		send: make(chan []byte, 256),
	}
	client.hub.register <- client

	// Allow collection of memory referenced by the caller by doing all work in
	// new goroutines.
	go client.writePump()
	go client.readPump()
}
