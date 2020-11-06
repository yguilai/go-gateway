package tcp_server

import (
	"context"
	"log"
	"net"
	"runtime"
)

type tcpKeepAliveListener struct {
	*net.TCPListener
}

func (l tcpKeepAliveListener) Accept() (net.Conn, error) {
	tc, err := l.AcceptTCP()
	if err != nil {
		return nil, err
	}
	return tc, nil
}

type contextKey struct {
	name string
}

func (k *contextKey) String() string {
	return "tcp proxy context value" + k.name
}

type conn struct {
	server     *TcpServer
	cancelCtx  context.CancelFunc
	rwc        net.Conn
	remoteAddr string
}

func (c *conn) close() {
	c.rwc.Close()
}

func (c *conn) serve(ctx context.Context)  {
	defer func() {
		if err := recover(); err != nil && err != ErrAbortHandler {
			const size = 64 << 10
			buf :=make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Printf("tcp: panic serving %v: %v\n%s", c.remoteAddr, err, buf)
		}
		c.close()
	}()

	c.remoteAddr = c.rwc.RemoteAddr().String()
	ctx = context.WithValue(ctx, LocalAddrContextKey, c.rwc.LocalAddr())
	if c.server.Handler == nil {
		panic("handler is empty")
	}

	c.server.Handler.ServeTCP(ctx, c.rwc)
}