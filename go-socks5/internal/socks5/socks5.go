package socks5

import (
	"bufio"
	"context"
	"errors"
	"net"
	"strconv"
	"sync"

	log "github.com/sirupsen/logrus"
)

const SOCKS5VERSION uint8 = 5

const (
	MethodNoAuth uint8 = iota
	MethodGSSAPI
	MethodUserPass
	MethodNoAcceptable uint8 = 0xFF
)

const (
	RequestConnect uint8 = iota + 1
	RequestBind
	RequestUDP
)

const (
	RequestAtypIPV4       uint8 = iota
	RequestAtypDomainname uint8 = 3
	RequestAtypIPV6       uint8 = 4
)

const (
	Succeeded uint8 = iota
	Failure
	Allowed
	NetUnreachable
	HostUnreachable
	ConnRefused
	TTLExpired
	CmdUnsupported
	AddrUnsupported
)

type Proxy struct {
	Inbound struct {
		reader *bufio.Reader
		writer net.Conn
	}
	Request struct {
		atyp uint8
		addr string
	}
	OutBound struct {
		reader *bufio.Reader
		writer net.Conn
	}
}

func MustStart(ctx context.Context, port int) {
	log.Debug("Socks5 server start.")
	listenIp := "0.0.0.0"
	listenPort := port

	// 创建监听
	addr, err := net.ResolveTCPAddr("tcp", listenIp+":"+strconv.Itoa(listenPort))
	if err != nil {
		log.Fatalf("fail in resolve tcp addr: %v", err)
	}
	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		log.Fatalf("fail in listen port: %v", err)
	}
	defer listener.Close()

	go func() {
		<-ctx.Done()
		log.Debug("Close socks5 listener...")
		_ = listener.Close()
	}()

	var wg sync.WaitGroup
	sem := make(chan struct{}, 1000)

	for {
		select {
		case <-ctx.Done():
			log.Warn("Shutting down server...")
			close(sem)
			wg.Wait()
			log.Info("Server gracefully shut down.")
			return
		default:
			conn, err := listener.Accept()
			if err != nil {
				if errors.Is(err, net.ErrClosed) {
					return
				}
				log.Warn("fail in accept", err)
				continue
			}

			// limit goroutine pool
			sem <- struct{}{}
			wg.Add(1)

			go func(conn net.Conn) {
				defer func() {
					_ = conn.Close()
					<-sem
					wg.Done()
				}()

				// todo: handle error and context
				handle(ctx, conn)
			}(conn)
		}
	}
}
