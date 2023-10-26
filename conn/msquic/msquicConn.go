package msquic

import (
	"crypto/tls"
	"net"
	"time"
)

type MsquicConn struct {
}

type MsquicListener struct {
}

var _ net.Conn = &MsquicConn{}
var _ net.Listener = &MsquicListener{}

func MsquicDial(addr string, config *tls.Config) (net.Conn, error) {
	//TODO implement me
	panic("implement me")
}

func MsquicListen(addr string, config *tls.Config) (net.Listener, error) {
	//TODO implement me
	panic("implement me")
}

func (mc *MsquicConn) Read(b []byte) (n int, err error) {
	//TODO implement me
	panic("implement me")
}

func (mc *MsquicConn) Write(b []byte) (n int, err error) {
	//TODO implement me
	panic("implement me")
}

func (mc *MsquicConn) Close() error {
	//TODO implement me
	panic("implement me")
}

func (mc *MsquicConn) LocalAddr() net.Addr {
	//TODO implement me
	//Because this function has not used in gt yet, we have not implemented it.
	panic("implement me")
}

func (mc *MsquicConn) RemoteAddr() net.Addr {
	//TODO implement me
	panic("implement me")
}

func (mc *MsquicConn) SetDeadline(t time.Time) error {
	//TODO implement me
	//Because this function has not used in gt yet, we have not implemented it.
	panic("implement me")
}

func (mc *MsquicConn) SetReadDeadline(t time.Time) error {
	//TODO implement me
	panic("implement me")
}

func (mc *MsquicConn) SetWriteDeadline(t time.Time) error {
	//TODO implement me
	panic("implement me")
}

func (m *MsquicListener) Accept() (net.Conn, error) {
	//TODO implement me
	panic("implement me")
}

func (m *MsquicListener) Close() error {
	//TODO implement me
	panic("implement me")
}

func (m *MsquicListener) Addr() net.Addr {
	//TODO implement me
	panic("implement me")
}
