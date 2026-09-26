package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const shutdownTimeout = 5 * time.Second

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, nil)).With("service", "events")
	slog.SetDefault(log)

	cfg, err := loadConfig()
	if err != nil {
		fatal(log, "failed to load config", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	prod, err := newProducer(cfg.KafkaBrokers, log)
	if err != nil {
		fatal(log, "failed to create producer", err)
	}
	defer closeQuietly(log, "producer", prod.close)

	cons, err := newConsumer(cfg.KafkaBrokers, cfg.ConsumerGroup, cfg.topics(), log)
	if err != nil {
		fatal(log, "failed to create consumer", err)
	}
	defer closeQuietly(log, "consumer", cons.close)

	consumerDone := make(chan struct{})
	go func() {
		defer close(consumerDone)
		log.Info("consumer started", "group", cfg.ConsumerGroup, "topics", cfg.topics())
		if err := cons.run(ctx); err != nil {
			log.Error("consumer stopped with error", "error", err)
		}
	}()

	h := handlers{cfg: cfg, producer: prod}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/events/health", h.health)
	mux.HandleFunc("POST /api/events/movie", h.createMovieEvent)
	mux.HandleFunc("POST /api/events/user", h.createUserEvent)
	mux.HandleFunc("POST /api/events/payment", h.createPaymentEvent)

	server := &http.Server{Addr: ":" + cfg.Port, Handler: mux}
	go func() {
		log.Info("events service listening", "address", server.Addr)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fatal(log, "server error", err)
		}
	}()

	<-ctx.Done()
	log.Info("shutting down gracefully")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer shutdownCancel()
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Error("server shutdown error", "error", err)
	}

	select {
	case <-consumerDone:
	case <-shutdownCtx.Done():
		log.Warn("consumer did not stop in time")
	}
	log.Info("events service stopped")
}

// fatal логирует ошибку и завершает процесс.
func fatal(log *slog.Logger, msg string, err error) {
	log.Error(msg, "error", err)
	os.Exit(1)
}

// closeQuietly закрывает ресурс и логирует ошибку, не прерывая завершение.
func closeQuietly(log *slog.Logger, name string, closeFn func() error) {
	if err := closeFn(); err != nil {
		log.Error("close error", "resource", name, "error", err)
	}
}
