package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	addr := flag.String("addr", "localhost:8080", "bind address")
	flag.Parse()

	slog.Info("starting http server", slog.String("addr", *addr))
	if err := http.ListenAndServe(*addr, &Handler{}); err != nil {
		slog.Error("cannot serve http", slog.Any("err", err))
	}
}

type Handler struct{}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/":
		h.handleIndex(w, r)
	case "/connect":
		h.handleConnect(w, r)
	case "/metrics":
		promhttp.Handler().ServeHTTP(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *Handler) handleIndex(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(`Welcome to the fly-autoscaler-multiapp-example!`))
}

func (h *Handler) handleConnect(w http.ResponseWriter, r *http.Request) {
	connectionN.Inc()
	defer func() { connectionN.Dec() }()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	timer := time.NewTimer(10 * time.Second)
	defer timer.Stop()

	for {
		fmt.Fprintln(w, time.Now().UTC().Format(time.RFC3339))
		w.(http.Flusher).Flush()

		select {
		case <-r.Context().Done():
			return
		case <-timer.C:
			return
		case <-ticker.C:
		}
	}
}

var connectionN = promauto.NewGauge(prometheus.GaugeOpts{
	Name: "connection_count",
	Help: "The number of currently connected clients.",
})
