package main

import (
	"log"
	"math/rand/v2"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// newProxyHandler создаёт обработчик маршрутизации между upstream-сервисами.
func newProxyHandler(cfg config) http.Handler {
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

		log.Printf("proxying %s %s to %s", r.Method, r.URL.RequestURI(), target)
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
