package transport

import (
	"github.com/hashicorp/raft"
	"io"
	"net"
	"time"
)

// streamLayer implements StreamLayer interface for plain TCP.
type streamLayer struct {
	advertise net.Addr
	listener  net.Listener
}

// Dial implements the StreamLayer interface.
func (t *streamLayer) Dial(address raft.ServerAddress, timeout time.Duration) (net.Conn, error) {
	return net.DialTimeout("tcp", string(address), timeout)
}

// Accept implements the net.Listener interface.
func (t *streamLayer) Accept() (c net.Conn, err error) {
	return t.listener.Accept()
}

// Close implements the net.Listener interface.
func (t *streamLayer) Close() (err error) {
	return t.listener.Close()
}

// Addr implements the net.Listener interface.
func (t *streamLayer) Addr() net.Addr {
	if t.advertise != nil {
		return t.advertise
	}
	return t.listener.Addr()
}

func NewNetworkTransport(lis net.Listener, advertise net.Addr, maxPool int, timeout time.Duration, logOutput io.Writer) *raft.NetworkTransport {
	stream := &streamLayer{
		advertise: advertise,
		listener:  lis,
	}
	return raft.NewNetworkTransport(stream, maxPool, timeout, logOutput)
}
