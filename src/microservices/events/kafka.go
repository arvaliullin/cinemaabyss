package main

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/IBM/sarama"
)

const (
	kafkaConnectAttempts = 30
	kafkaConnectDelay    = 2 * time.Second
)

// kafkaVersion соответствует брокеру из docker-compose (wurstmeister/kafka 2.7.0).
var kafkaVersion = sarama.V2_7_0_0

// withRetry повторяет подключение к Kafka, пока брокер стартует.
func withRetry(log *slog.Logger, name string, connect func() error) error {
	var err error
	for attempt := 1; attempt <= kafkaConnectAttempts; attempt++ {
		if err = connect(); err == nil {
			return nil
		}
		log.Warn("kafka connect attempt failed",
			"client", name,
			"attempt", attempt,
			"max_attempts", kafkaConnectAttempts,
			"error", err,
		)
		time.Sleep(kafkaConnectDelay)
	}
	return fmt.Errorf("kafka %s: %w", name, err)
}
