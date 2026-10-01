package main

import (
	"encoding/json"
	"log/slog"

	"github.com/IBM/sarama"
)

// logHandler обрабатывает сообщения consumer group записью в лог.
type logHandler struct {
	log *slog.Logger
}

func (logHandler) Setup(sarama.ConsumerGroupSession) error   { return nil }
func (logHandler) Cleanup(sarama.ConsumerGroupSession) error { return nil }

// ConsumeClaim логирует каждое прочитанное сообщение и подтверждает его.
func (h logHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		log := h.log.With(
			"topic", msg.Topic,
			"partition", msg.Partition,
			"offset", msg.Offset,
		)

		var e event
		if err := json.Unmarshal(msg.Value, &e); err != nil {
			log.Warn("consumed invalid message", "error", err, "raw", string(msg.Value))
		} else {
			log.Info("event consumed",
				"event_id", e.ID,
				"event_type", e.Type,
				"payload", e.Payload,
			)
		}
		session.MarkMessage(msg, "")
	}
	return nil
}
