package main

import (
	"log/slog"
	"math/rand/v2"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// newProxyHandler создаёт обработчик маршрутизации между upstream-сервисами.
func newProxyHandler(cfg config, log *slog.Logger) http.Handler {
	monolithProxy := newReverseProxy(cfg.monolithTarget)
	moviesProxy := newReverseProxy(cfg.moviesServiceTarget)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxy := monolithProxy
		target := cfg.monolithTarget

		if isMoviesPath(r.URL.Path) &&
			(!cfg.GradualMigration || rand.IntN(100) < cfg.MoviesMigrationPercent) {
			proxy = moviesProxy
			target = cfg.moviesServiceTarget
		}

		log.Info("proxying request",
			"method", r.Method,
			"uri", r.URL.RequestURI(),
			"target", target.String(),
		)
		proxy.ServeHTTP(w, r)
	})
}

// newReverseProxy создаёт прокси для указанного upstream-сервиса.
func newReverseProxy(target *url.URL) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Rewrite: func(request *httputil.ProxyRequest) {
			request.SetURL(target)
			request.SetXForwarded()
			request.Out.Host = target.Host
		},
	}
}
