package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"github.com/kkonst40/ichat/internal/config"
	"github.com/segmentio/kafka-go"
)

type UserLoginCache interface {
	GetUserLogins(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]string, error)
	SetUserLogins(ctx context.Context, logins map[uuid.UUID]string) error
}

type Consumer struct {
	reader     kafka.Reader
	loginCache UserLoginCache
	ctx        context.Context
}

func NewConsumer(cfg *config.Config, userLoginCache UserLoginCache) *Consumer {
	return &Consumer{
		reader: *kafka.NewReader(kafka.ReaderConfig{
			Brokers:  []string{fmt.Sprintf("%s:%s", cfg.Kafka.Host, cfg.Kafka.Port)},
			GroupID:  "iapp-consumer-group",
			Topic:    topicUserEvents,
			MinBytes: 10e3,
			MaxBytes: 10e6,
		}),
		loginCache: userLoginCache,
	}
}

func (c *Consumer) Start(ctx context.Context) error {
	c.ctx = ctx
	for {
		msg, err := c.reader.ReadMessage(c.ctx)
		if err != nil {
			if c.ctx.Err() != nil {
				return nil
			}
			return err
		}

		c.handleMessage(msg.Value)
	}
}

func (c *Consumer) Close() error {
	return c.reader.Close()
}

func (c *Consumer) handleMessage(data []byte) {
	var event eventMessage
	if err := json.Unmarshal(data, &event); err != nil {
		slog.Error("event message unmarshaling", "error", err.Error())
	}

	switch event.Type {
	case eventTypeLoginUpdate:
		var p loginUpdatePayload
		json.Unmarshal(event.Payload, &p)

		if err := c.loginCache.SetUserLogins(c.ctx, map[uuid.UUID]string{p.UserID: p.Login}); err != nil {
			slog.Error("user login cache SetUserLogins error", "error", err)
		}

	default:
		slog.Error("unknown event", "type", event.Type)
	}
}
