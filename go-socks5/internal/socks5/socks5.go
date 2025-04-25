package socks5

import (
	"context"
	"errors"
	"net"
	"strconv"
	"sync"

	log "github.com/sirupsen/logrus"
)

func MustStart(ctx context.Context, port int) {
	address := "0.0.0.0:" + strconv.Itoa(port)
	log.Debug("Socks5 server start at: ", address)

	// 创建监听
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		log.Fatalf("fail in resolve tcp addr: %v", err)
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatalf("fail in listen port: %v", err)
	}
	defer func() { _ = listener.Close() }()

	go func() {
		<-ctx.Done()
		log.Info("Close socks5 listener...")
		_ = listener.Close()
	}()

	var wg sync.WaitGroup
	sem := make(chan struct{}, 1000)

	defer func() {
		wg.Wait()
		close(sem)
		log.Info("Server has gracefully shutdown.")
	}()

	for {
		select {
		case <-ctx.Done():
			log.Warn("Shutting down server...")
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					log.Debug("Server has gracefully shutdown from listener status")
					return
				}
				log.Warn("fail in accept", err)
				continue
			}

			// limit goroutine pool and wait for goroutine to finish
			sem <- struct{}{}
			wg.Add(1)

			go func(conn net.Conn) {
				defer func() {
					_ = conn.Close()
					wg.Done()
					<-sem
					log.Infof("Connection closed: %v", conn.RemoteAddr())
				}()

				log.Infof("New connection: %v", conn.RemoteAddr())
				handle(ctx, conn)
			}(conn)
		}
	}
}
