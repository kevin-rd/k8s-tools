package main

import (
	"github.io/kevin-rd/k8s-tools/go-socks5/internal/socks5"

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
	log.Debug("welcome go socks demo")
	// config.LoadConfig("./config.json")
	socks5.Start(10086)
}
