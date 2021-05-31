package handler

import (
	"net/http"

	"github.com/micro/go-micro/v2/web"
)

// Register 注册
func Register(service web.Service) {
	service.Handle("/websocket/", http.StripPrefix("/websocket/", http.FileServer(http.Dir("html"))))
	hub := newHub()
	go hub.run()
	// websocket interface
	service.HandleFunc("/websocket/hi", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
}
