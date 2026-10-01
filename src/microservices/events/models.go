package main

import (
	"errors"
	"time"
)

// movieEvent описывает событие фильма из запроса.
type movieEvent struct {
	MovieID     *int     `json:"movie_id"`
	Title       string   `json:"title"`
	Action      string   `json:"action"`
	UserID      *int     `json:"user_id,omitempty"`
	Rating      *float64 `json:"rating,omitempty"`
	Genres      []string `json:"genres,omitempty"`
	Description string   `json:"description,omitempty"`
}

// validate проверяет обязательные поля события фильма.
func (e *movieEvent) validate() error {
	if e.MovieID == nil || e.Title == "" || e.Action == "" {
		return errors.New("movie_id, title and action are required")
	}
	return nil
}

// userEvent описывает событие пользователя из запроса.
type userEvent struct {
	UserID    *int       `json:"user_id"`
	Username  string     `json:"username,omitempty"`
	Email     string     `json:"email,omitempty"`
	Action    string     `json:"action"`
	Timestamp *time.Time `json:"timestamp"`
}

// validate проверяет обязательные поля события пользователя.
func (e *userEvent) validate() error {
	if e.UserID == nil || e.Action == "" || e.Timestamp == nil {
		return errors.New("user_id, action and timestamp are required")
	}
	return nil
}

// paymentEvent описывает событие платежа из запроса.
type paymentEvent struct {
	PaymentID  *int       `json:"payment_id"`
	UserID     *int       `json:"user_id"`
	Amount     *float64   `json:"amount"`
	Status     string     `json:"status"`
	Timestamp  *time.Time `json:"timestamp"`
	MethodType string     `json:"method_type,omitempty"`
}

// validate проверяет обязательные поля события платежа.
func (e *paymentEvent) validate() error {
	if e.PaymentID == nil || e.UserID == nil || e.Amount == nil || e.Status == "" || e.Timestamp == nil {
		return errors.New("payment_id, user_id, amount, status and timestamp are required")
	}
	return nil
}

// event описывает сообщение, публикуемое в Kafka и возвращаемое в ответе API.
type event struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Payload   any       `json:"payload"`
}

// eventResponse описывает успешный ответ API на создание события.
type eventResponse struct {
	Status    string `json:"status"`
	Partition int32  `json:"partition"`
	Offset    int64  `json:"offset"`
	Event     event  `json:"event"`
}

// errorResponse описывает ответ API с ошибкой.
type errorResponse struct {
	Error string `json:"error"`
}
