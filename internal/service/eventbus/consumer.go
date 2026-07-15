package eventbus

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/kkonst40/chat-service/internal/config"
	"github.com/twmb/franz-go/pkg/kgo"
)

type UserLoginCache interface {
	GetUserLogins(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]string, error)
	SetUserLogins(ctx context.Context, logins map[uuid.UUID]string) error
}

type Consumer struct {
	client     *kgo.Client
	loginCache UserLoginCache
	ctx        context.Context
}

const (
	topicConsumerGroup = "iapp-consumer-group"
	topicUserEvents    = "user-events"
)

func NewConsumer(cfg *config.Config, userLoginCache UserLoginCache) (*Consumer, error) {
	cl, err := kgo.NewClient(
		kgo.SeedBrokers(fmt.Sprintf("%s:%s", cfg.Kafka.Host, cfg.Kafka.Port)),
		kgo.ConsumerGroup(topicConsumerGroup),
		kgo.ConsumeTopics(topicUserEvents),
	)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		client:     cl,
		loginCache: userLoginCache,
	}, nil
}

func (c *Consumer) Run(ctx context.Context) {
	c.ctx = ctx
	defer c.client.Close()

	for {
		fetches := c.client.PollFetches(ctx)
		if errs := fetches.Errors(); len(errs) > 0 {
			if ctx.Err() != nil {
				break
			}
			continue
		}

		iter := fetches.RecordIter()
		for !iter.Done() {
			record := iter.Next()
			slog.Debug("Kafka message", "partition", record.Partition, "key", string(record.Key), "value", string(record.Value))

			err := c.handleMessage(record.Value)
			if err != nil {
				slog.Error("Handling message from Kafka", "key", string(record.Key), "error", err.Error())
				continue
			}

			commitCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			c.client.CommitRecords(commitCtx, record)
			cancel()
		}
	}
}

func (c *Consumer) handleMessage(data []byte) error {
	var event eventMessage
	if err := json.Unmarshal(data, &event); err != nil {
		slog.Error("event message unmarshaling", "error", err.Error())
		return err
	}

	switch event.Type {
	case eventTypeLoginUpdate:
		var p loginUpdatePayload
		json.Unmarshal(event.Payload, &p)

		if err := c.loginCache.SetUserLogins(c.ctx, map[uuid.UUID]string{p.UserID: p.Login}); err != nil {
			slog.Error("user login cache SetUserLogins error", "error", err)
			return err
		}

	default:
		slog.Error("unknown event", "type", event.Type)
		return fmt.Errorf("unknown event type")
	}

	return nil
}
