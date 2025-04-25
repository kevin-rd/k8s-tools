package metrics

import (
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"time"
)

var (
	// ConnectGauge is the current number of active SOCKS5 connections.
	ConnectGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "connect_gauge",
		Help: "Current number of active SOCKS5 connections",
	}, []string{"host"})

	// ConnectCounter is the total number of SOCKS5 connections.
	ConnectCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "connect_counter",
		Help: "Total number of SOCKS5 connections",
	}, []string{"host"})
)

func init() {
	// Register the metrics.
	prometheus.MustRegister(ConnectGauge, ConnectCounter)
}

func StartServer(ctx context.Context, addr string) error {
	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// 优雅关闭
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		_ = server.Shutdown(shutdownCtx)
	}()

	return server.ListenAndServe()
}
