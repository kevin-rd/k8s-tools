package socks5

import (
	"bufio"
	"net"
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
