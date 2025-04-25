package socks5

import (
	"bufio"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
	"github.io/kevin-rd/k8s-tools/go-socks5/internal/metrics"
	"io"
	"net"
)

func handle(ctx context.Context, conn net.Conn) {
	labels := prometheus.Labels{"host": conn.RemoteAddr().(*net.TCPAddr).IP.String()}
	metrics.ConnectGauge.With(labels).Inc()
	metrics.ConnectCounter.With(labels).Inc()
	defer metrics.ConnectGauge.With(labels).Dec()

	proxy := &Proxy{}
	proxy.Inbound.reader = bufio.NewReader(conn)
	proxy.Inbound.writer = conn

	err := handshake(proxy)
	if err != nil {
		log.Warn("fail in handshake: ", err)
		return
	}
	transport(ctx, proxy)
}

func handshake(proxy *Proxy) error {
	err := auth(proxy)
	if err != nil {
		log.Warn(err)
		return err
	}

	err = readRequest(proxy)
	if err != nil {
		log.Warn(err)
		return err
	}

	err = replay(proxy)
	if err != nil {
		log.Warn(err)
		return err
	}
	return err
}

func auth(proxy *Proxy) error {
	/*
		Read
		   +-----+----------+-----------+
		   | VER | NMETHODS |  METHODS  |
		   +-----+----------+-----------+
		   |  1  |    1     |  1 to 255 |
		   +-----+----------+-----------+
	*/
	buf := bufPool512.Get().([]byte)
	defer bufPool512.Put(buf)

	n, err := io.ReadFull(proxy.Inbound.reader, buf[:2])
	if err != nil || n != 2 {
		return fmt.Errorf("failed to read 2 bytes from client: read %d bytes, err: %v", n, err)
	}

	ver, nmethods := buf[0], int(buf[1])
	if ver != SOCKS5VERSION {
		return errors.New("only support socks5 version")
	}
	_, err = io.ReadFull(proxy.Inbound.reader, buf[:nmethods])
	if err != nil {
		return errors.New("fail to read methods" + err.Error())
	}
	var supportNoAuth bool
	for _, m := range buf[:nmethods] {
		switch m {
		case MethodNoAuth:
			supportNoAuth = true
		default:
			supportNoAuth = false
		}
	}
	if !supportNoAuth {
		return errors.New("no only support no auth")
	}

	/*
		replay
		+-----+--------+
		| VER | METHOD |
		+-----+--------+
		|  1  |   1    |
		+-----+--------+
	*/
	// 无需认证
	n, err = proxy.Inbound.writer.Write([]byte{0x05, 0x00})
	if n != 2 {
		return fmt.Errorf("fail to write socks method: %v", err)
	}

	return nil
}

func readRequest(proxy *Proxy) error {
	/*
		Read
		   +----+-----+-------+------+----------+----------+
		   |VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
		   +----+-----+-------+------+----------+----------+
		   | 1  |  1  | X'00' |  1   | Variable |    2     |
		   +----+-----+-------+------+----------+----------+
	*/
	buf := bufPool512.Get().([]byte)
	defer bufPool512.Put(buf)

	n, err := io.ReadFull(proxy.Inbound.reader, buf[:4])
	if n != 4 {
		return errors.New("fail to read request " + err.Error())
	}
	ver, cmd, _, atyp := uint8(buf[0]), uint8(buf[1]), uint8(buf[2]), uint8(buf[3])
	if ver != SOCKS5VERSION {
		return errors.New("only support socks5 version")
	}
	if cmd != RequestConnect {
		return errors.New("only support connect requests")
	}
	var addr string
	switch atyp {
	case RequestAtypIPV4:
		_, err = io.ReadFull(proxy.Inbound.reader, buf[:4])
		if err != nil {
			return errors.New("fail in read requests ipv4 " + err.Error())
		}
		addr = string(buf[:4])
	case RequestAtypDomainname:
		_, err = io.ReadFull(proxy.Inbound.reader, buf[:1])
		if err != nil {
			return errors.New("fail in read requests domain len" + err.Error())
		}
		domainLen := int(buf[0])
		_, err = io.ReadFull(proxy.Inbound.reader, buf[:domainLen])
		if err != nil {
			return errors.New("fail in read requests domain " + err.Error())
		}
		addr = string(buf[:domainLen])
	case RequestAtypIPV6:
		_, err = io.ReadFull(proxy.Inbound.reader, buf[:16])
		if err != nil {
			return errors.New("fail in read requests ipv4 " + err.Error())
		}
		addr = string(buf[:16])
	}
	_, err = io.ReadFull(proxy.Inbound.reader, buf[:2])
	if err != nil {
		return errors.New("fail in read requests port " + err.Error())
	}
	port := binary.BigEndian.Uint16(buf[:2])
	proxy.Request.atyp = atyp
	proxy.Request.addr = fmt.Sprintf("%s:%d", addr, port)
	log.Debug("request is", proxy.Request)
	return nil
}

func replay(proxy *Proxy) error {
	/*
		write
		   +----+-----+-------+------+----------+----------+
		   |VER | REP |  RSV  | ATYP | BND.ADDR | BND.PORT |
		   +----+-----+-------+------+----------+----------+
		   | 1  |  1  | X'00' |  1   | Variable |    2     |
		   +----+-----+-------+------+----------+----------+
	*/
	conn, err := net.Dial("tcp", proxy.Request.addr)
	if err != nil {
		log.Warn("fail to connect ", proxy.Request.addr)
		_, rerr := proxy.Inbound.writer.Write([]byte{SOCKS5VERSION, HostUnreachable, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
		if rerr != nil {
			return errors.New("fail in replay " + err.Error())
		}
		return errors.New("fail in connect addr " + proxy.Request.addr + err.Error())
	}
	_, err = proxy.Inbound.writer.Write([]byte{SOCKS5VERSION, Succeeded, 0x00, 0x01, 0, 0, 0, 0, 0, 0})
	if err != nil {
		return errors.New("fail in replay " + err.Error())
	}
	proxy.OutBound.reader = bufio.NewReader(conn)
	proxy.OutBound.writer = conn
	return nil
}

func transport(ctx context.Context, proxy *Proxy) {
	done := make(chan struct{}, 2)

	go func() {
		defer func() { done <- struct{}{} }()
		_, err := copyWithCtx(ctx, proxy.OutBound.writer, proxy.Inbound.reader)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				log.Debugf("copy inbound -> outbound failed: %v", err)
			} else {
				log.Warnf("copy inbound -> outbound failed: %v", err)
			}
		}
	}()

	go func() {
		defer func() { done <- struct{}{} }()
		_, err := copyWithCtx(ctx, proxy.Inbound.writer, proxy.OutBound.reader)
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				log.Debugf("copy outbound -> inbound failed: %v", err)
			} else {
				log.Warnf("copy outbound -> inbound failed: %v", err)
			}
		}
	}()

	go func() {
		<-ctx.Done()
		_ = proxy.Inbound.writer.Close()
		_ = proxy.OutBound.writer.Close()
	}()

	<-done
	_ = proxy.Inbound.writer.Close()
	_ = proxy.OutBound.writer.Close()

	// finally wait for both copy routines to complete
	<-done
	log.Debug("transport has completed: ", proxy.Inbound.writer.RemoteAddr().String())
}
