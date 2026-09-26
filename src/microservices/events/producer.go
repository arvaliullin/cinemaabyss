package main

import (
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/IBM/sarama"
)

// producer публикует события в Kafka.
type producer struct {
	sync sarama.SyncProducer
	log  *slog.Logger
}

// newProducer создаёт синхронный producer, ожидая готовности брокера.
func newProducer(brokers []string, log *slog.Logger) (*producer, error) {
	cfg := sarama.NewConfig()
	cfg.Version = kafkaVersion
	cfg.Producer.Return.Successes = true
	cfg.Producer.RequiredAcks = sarama.WaitForLocal

	var sp sarama.SyncProducer
	err := withRetry(log, "producer", func() error {
		var err error
		sp, err = sarama.NewSyncProducer(brokers, cfg)
		return err
	})
	if err != nil {
		return nil, err
	}
	return &producer{sync: sp, log: log}, nil
}

// publish сериализует событие и отправляет его в топик.
func (p *producer) publish(topic string, e event) (partition int32, offset int64, err error) {
	value, err := json.Marshal(e)
	if err != nil {
		return 0, 0, fmt.Errorf("marshal event: %w", err)
	}

	partition, offset, err = p.sync.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(e.ID),
		Value: sarama.ByteEncoder(value),
	})
	if err != nil {
		return 0, 0, fmt.Errorf("send message to %s: %w", topic, err)
	}

	p.log.Info("event published",
		"event_id", e.ID,
		"event_type", e.Type,
		"topic", topic,
		"partition", partition,
		"offset", offset,
	)
	return partition, offset, nil
}

// close освобождает ресурсы producer.
func (p *producer) close() error {
	return p.sync.Close()
}
