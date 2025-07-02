package handler

import (
	"context"
	"log"
	"time"
	"websocket/socket"
	"websocket/worker"
)

type ChatHandler struct {
	consumer *worker.StreamConsumer
}

func NewNotiHandler(c *worker.StreamConsumer) *ChatHandler {
	return &ChatHandler{
		consumer: c,
	}
}

func (w *ChatHandler) Run(ctx context.Context) error {
	// Chạy goroutine scan pending định kỳ
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()

		// Chạy lần đầu luôn (không cần đợi 30 phút)
		log.Println("Checking pending noti...")
		w.consumer.RecoverPending(func(ctx context.Context, data []map[string]string) error {
			return socket.SendNoti(ctx, data)
		})

		for {
			select {
			case <-ctx.Done():
				log.Println("Stop recovering pending noti.")
				return
			case <-ticker.C:
				log.Println("Checking pending noti...")
				err := w.consumer.RecoverPending(func(ctx context.Context, data []map[string]string) error {
					return socket.SendNoti(ctx, data)
				})
				if err != nil {
					log.Printf("RecoverPending error: %v\n", err)
				}
			}
		}
	}()

	// Chạy consumer chính
	return w.consumer.Consume(func(ctx context.Context, data []map[string]string) error {
		return socket.SendNoti(ctx, data)
	})
	// return nil
}
