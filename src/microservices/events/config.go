package main

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
)

// config содержит настройки events-сервиса.
type config struct {
	Port          string   `envconfig:"PORT" default:"8082"`
	KafkaBrokers  []string `envconfig:"KAFKA_BROKERS" required:"true"`
	ConsumerGroup string   `envconfig:"KAFKA_CONSUMER_GROUP" default:"events-service"`
	MovieTopic    string   `envconfig:"MOVIE_EVENTS_TOPIC" default:"movie-events"`
	UserTopic     string   `envconfig:"USER_EVENTS_TOPIC" default:"user-events"`
	PaymentTopic  string   `envconfig:"PAYMENT_EVENTS_TOPIC" default:"payment-events"`
}

// loadConfig загружает конфигурацию из переменных окружения.
func loadConfig() (config, error) {
	var cfg config
	if err := envconfig.Process("", &cfg); err != nil {
		return config{}, fmt.Errorf("not able to read environment variables: %w", err)
	}
	return cfg, nil
}

// topics возвращает список топиков, которые читает consumer.
func (c config) topics() []string {
	return []string{c.MovieTopic, c.UserTopic, c.PaymentTopic}
}
