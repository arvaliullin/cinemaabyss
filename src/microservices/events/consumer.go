package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/IBM/sarama"
)

// consumer читает события из Kafka и записывает их в лог.
type consumer struct {
	group  sarama.ConsumerGroup
	topics []string
	log    *slog.Logger
}

// newConsumer создаёт consumer group, ожидая готовности брокера.
func newConsumer(brokers []string, groupID string, topics []string, log *slog.Logger) (*consumer, error) {
	cfg := sarama.NewConfig()
	cfg.Version = kafkaVersion
	cfg.Consumer.Offsets.Initial = sarama.OffsetOldest
	cfg.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRoundRobin()}

	var group sarama.ConsumerGroup
	err := withRetry(log, "consumer", func() error {
		var err error
		group, err = sarama.NewConsumerGroup(brokers, groupID, cfg)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &consumer{group: group, topics: topics, log: log}, nil
}

// run блокирующе читает топики до отмены контекста.
func (c *consumer) run(ctx context.Context) error {
	handler := &logHandler{log: c.log}
	for {
		if err := c.group.Consume(ctx, c.topics, handler); err != nil {
			if errors.Is(err, sarama.ErrClosedConsumerGroup) {
				return nil
			}
			return fmt.Errorf("consume: %w", err)
		}
		if ctx.Err() != nil {
			return nil
		}
	}
}

// close освобождает ресурсы consumer.
func (c *consumer) close() error {
	return c.group.Close()
}
