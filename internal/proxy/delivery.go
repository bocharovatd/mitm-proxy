package proxy

import (
	"net"
)

type Handlers interface {
	HandleConnection(conn net.Conn)
}