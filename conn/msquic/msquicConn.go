package msquic

import (
	"crypto/tls"
	"github.com/isrc-cas/gt/quic"
	"net"
)

type QuicIscasConn struct {
	net.Conn // *quic.stream
	parent   *quic.Connection
}

func (q *QuicIscasConn) Close() (err error) {
	err1 := q.Conn.Close()
	err2 := q.parent.Close()
	if err1 != nil {
		return err1
	}
	return err2
}

func MsquicDial(addr string, config *tls.Config) (conn net.Conn, err error) {
	unsecure := config.InsecureSkipVerify
	parent, err := quic.NewConnection(addr, 10_000, "", unsecure)
	if err != nil {
		return
	}
	stream, err := parent.OpenStream()
	if err != nil {
		return
	}
	conn = &QuicIscasConn{
		Conn:   stream,
		parent: parent,
	}
	return
}
