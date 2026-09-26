package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// handlers содержит зависимости HTTP-обработчиков.
type handlers struct {
	cfg      config
	producer *producer
}

// health сообщает о доступности сервиса.
func (handlers) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]bool{"status": true})
}

// createMovieEvent публикует событие фильма в Kafka.
func (h handlers) createMovieEvent(w http.ResponseWriter, r *http.Request) {
	var in movieEvent
	if !decode(w, r, &in, in.validate) {
		return
	}
	h.publish(w, h.cfg.MovieTopic, event{
		ID:        fmt.Sprintf("movie-%d-%s", *in.MovieID, in.Action),
		Type:      "movie",
		Timestamp: time.Now().UTC(),
		Payload:   in,
	})
}

// createUserEvent публикует событие пользователя в Kafka.
func (h handlers) createUserEvent(w http.ResponseWriter, r *http.Request) {
	var in userEvent
	if !decode(w, r, &in, in.validate) {
		return
	}
	h.publish(w, h.cfg.UserTopic, event{
		ID:        fmt.Sprintf("user-%d-%s", *in.UserID, in.Action),
		Type:      "user",
		Timestamp: *in.Timestamp,
		Payload:   in,
	})
}

// createPaymentEvent публикует событие платежа в Kafka.
func (h handlers) createPaymentEvent(w http.ResponseWriter, r *http.Request) {
	var in paymentEvent
	if !decode(w, r, &in, in.validate) {
		return
	}
	h.publish(w, h.cfg.PaymentTopic, event{
		ID:        fmt.Sprintf("payment-%d-%s", *in.PaymentID, in.Status),
		Type:      "payment",
		Timestamp: *in.Timestamp,
		Payload:   in,
	})
}

// publish отправляет событие и формирует ответ API.
func (h handlers) publish(w http.ResponseWriter, topic string, e event) {
	partition, offset, err := h.producer.publish(topic, e)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, errorResponse{Error: err.Error()})
		return
	}
	writeJSON(w, http.StatusCreated, eventResponse{
		Status:    "success",
		Partition: partition,
		Offset:    offset,
		Event:     e,
	})
}

// decode читает JSON-тело и проверяет его; при ошибке отвечает 400.
func decode(w http.ResponseWriter, r *http.Request, dst any, validate func() error) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: "invalid JSON body: " + err.Error()})
		return false
	}
	if err := validate(); err != nil {
		writeJSON(w, http.StatusBadRequest, errorResponse{Error: err.Error()})
		return false
	}
	return true
}

// writeJSON сериализует ответ с указанным статусом.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
