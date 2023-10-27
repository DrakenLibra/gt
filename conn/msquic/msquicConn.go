package msquic

import (
	"crypto/tls"
	"net"
)

type MsquicConn struct {
	net.Conn // *quic.stream
	Parent   *Connection
}

func (q *MsquicConn) Close() (err error) {
	//err1 := q.Conn.Close()
	//err2 := q.Parent.Close()
	//if err1 != nil {
	//	return err1
	//}
	//return err2
	err = q.Conn.Close()
	return
}

func MsquicDial(addr string, config *tls.Config) (conn net.Conn, err error) {
	unsecure := config.InsecureSkipVerify
	parent, err := NewConnection(addr, 10_000, "", unsecure)
	if err != nil {
		return
	}
	stream, err := parent.OpenStream()
	if err != nil {
		return
	}
	conn = &MsquicConn{
		Conn:   stream,
		Parent: parent,
	}
	return
}

func MsquicListen(addr string, idleTimeoutMs uint64, keyFile string, certFile string, password string) (net.Listener, error) {
	return NewListenr(addr, idleTimeoutMs, keyFile, certFile, password)
}
