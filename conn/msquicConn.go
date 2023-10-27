package conn

import (
	"crypto/tls"
	"github.com/isrc-cas/gt/conn/msquic"
	"net"
	"time"
)

type msquicStruct struct {
	msquicConn   *msquic.Connection
	msquicStream *msquic.Stream
}

var _ net.Conn = &msquicStruct{}
var _ net.Listener = &msquic.Listener{}

func MsquicDial(addr string, config *tls.Config) (net.Conn, error) {
	unsecure := config.InsecureSkipVerify
	parent, err := msquic.NewConnection(addr, 10_000, "", unsecure)
	if err != nil {
		return nil, err
	}
	stream, err := parent.OpenStream()
	if err != nil {
		return nil, err
	}
	conn := &msquicStruct{
		msquicStream: stream,
		msquicConn:   parent,
	}
	return conn, err
}

func (q *msquicStruct) Read(b []byte) (n int, err error) {
	return q.msquicStream.Read(b)
}

func (q *msquicStruct) Write(b []byte) (n int, err error) {
	return q.msquicStream.Write(b)
}

func (q *msquicStruct) LocalAddr() net.Addr {
	return q.msquicStream.LocalAddr()
}

func (q *msquicStruct) RemoteAddr() net.Addr {
	return q.msquicStream.RemoteAddr()
}

func (q *msquicStruct) SetDeadline(t time.Time) error {
	return q.msquicStream.SetDeadline(t)
}

func (q *msquicStruct) SetReadDeadline(t time.Time) error {
	return q.msquicStream.SetReadDeadline(t)
}

func (q *msquicStruct) SetWriteDeadline(t time.Time) error {
	return q.msquicStream.SetWriteDeadline(t)
}

func (q *msquicStruct) Close() (err error) {
	err1 := q.msquicStream.Close()
	err2 := q.msquicConn.Close()
	if err1 != nil {
		return err1
	}
	return err2
}
