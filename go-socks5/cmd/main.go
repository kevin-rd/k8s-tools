package main

import (
	"context"
	"github.io/kevin-rd/k8s-tools/go-socks5/internal/metrics"
	"github.io/kevin-rd/k8s-tools/go-socks5/internal/socks5"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	nested "github.com/antonfisher/nested-logrus-formatter"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&nested.Formatter{
		NoColors: false,
	})
	log.SetReportCaller(true)
	log.SetLevel(log.DebugLevel)
}

func main() {
	log.Info("Welcome go socks5!")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	stopCh := make(chan os.Signal, 1)
	signal.Notify(stopCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		count := 0
		for sig := range stopCh {
			count++
			log.Debug("Receive signal: %v, count: %d", sig, count)

			if count == 1 {
				log.Info("First signal received, initiating graceful shutdown...")
				cancel()
			} else {
				log.Warn("Receive signal again, force exit.")
				os.Exit(1)
			}
		}
	}()

	// metrics server
	go func() {
		log.Info("Starting metrics server...")
		if err := metrics.StartServer(ctx, ":10081"); err != nil {
			if err == http.ErrServerClosed {
				log.Info("Metrics server has gracefully shutdown.")
			} else {
				log.Fatalf("metrics server error: %v", err)
			}
		}
	}()

	// socks5 server
	socks5.MustStart(ctx, 10080)
	log.Info("Shutdown done.")
}
