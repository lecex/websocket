package handler

import (
	"context"
	"log"
	"net/http"

	"github.com/lecex/core/env"
	pb "github.com/lecex/core/proto/event"

	"github.com/micro/go-micro/v2"
	"github.com/micro/go-micro/v2/web"
)

const topic = "event"

type Subscriber struct {
	hub *Hub
}

// Register 注册
func Register(service web.Service) {
	service.Handle("/", http.StripPrefix("/", http.FileServer(http.Dir("html"))))
	hub := newHub()
	go hub.run()
	// websocket interface
	service.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	go func() {
		srv := micro.NewService(
			micro.Name(env.Getenv("MICRO_SRV_NAMESPACE", "go.micro.srv.")+"websocket"),
			micro.Version("latest"),
			micro.Address(env.Getenv("SERVER_SRV_ADDRESS", ":8989")),
		)
		srv.Init()

		sub := &Subscriber{
			hub: hub,
		}
		micro.RegisterSubscriber(topic, srv.Server(), sub)
		if err := srv.Run(); err != nil {
			log.Fatalf("websocketSubscriberSrv run error: %v\n", err)
		}
	}()
}

func (sub *Subscriber) Process(ctx context.Context, event *pb.Event) error {
	sub.hub.broadcast <- event
	return nil
}
