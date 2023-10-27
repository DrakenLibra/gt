package msquic

/*
#include <stdlib.h>

#include "stream.h"
*/
import "C"

import (
	"errors"
	"io"
	"net"
	"sync"
	"time"
	"unsafe"

	"github.com/mattn/go-pointer"
)

type Stream struct {
	conn       *Connection
	cppStream  unsafe.Pointer
	pointerID  unsafe.Pointer
	onStarted  chan struct{}
	onClose    chan struct{}
	closeOnce  sync.Once
	onSend     chan struct{}
	onReceive  chan []byte
	receiveBuf []byte
}

func (s *Stream) Close() error {
	s.closeOnce.Do(func() {
		close(s.onClose)
	})
	pointer.Unref(s.pointerID)
	C.DeleteStream(s.cppStream)
	return nil
}

func (s *Stream) Read(b []byte) (n int, err error) {
	if len(s.receiveBuf) != 0 {
		n = copy(b, s.receiveBuf)
		s.receiveBuf = s.receiveBuf[n:]
		goto handleMsquic
	}

	select {
	case buf := <-s.onReceive:
		n = copy(b, buf)
		s.receiveBuf = buf[n:]
		goto handleMsquic
	case <-s.onClose:
		return 0, io.EOF
	}

handleMsquic:
	C.StreamReceiveComplete(s.cppStream, C.uint64_t(n))
	return
}

func (s *Stream) Write(b []byte) (n int, err error) {
	cBuf := C.CBytes(b)
	C.StreamSend(s.cppStream, cBuf, C.size_t(len(b)))
	select {
	case <-s.onSend:
		return len(b), nil
	case <-s.onClose:
		return 0, errors.New("msquic stream closed")
	}
}

func (s *Stream) LocalAddr() net.Addr {
	return s.conn.LocalAddr()
}

func (s *Stream) RemoteAddr() net.Addr {
	return s.conn.RemoteAddr()
}

func (s *Stream) SetDeadline(t time.Time) error {
	// stream 无法单独设置超时，只能设置 connection
	return s.conn.SetDeadline(t)
}

func (s *Stream) SetReadDeadline(t time.Time) error {
	// stream 无法单独设置超时，只能设置 connection
	return s.conn.SetReadDeadline(t)
}

func (s *Stream) SetWriteDeadline(t time.Time) error {
	// stream 无法单独设置超时，只能设置 connection
	return s.conn.SetWriteDeadline(t)
}

//export OnStreamShutdownComplete
func OnStreamShutdownComplete(cppConn, context unsafe.Pointer) {
	stream, ok := pointer.Restore(context).(*Stream)
	if !ok || stream == nil {
		return
	}
	stream.closeOnce.Do(func() {
		close(stream.onClose)
	})
}

//export OnStreamStartComplete
func OnStreamStartComplete(cppStream, context unsafe.Pointer) {
	stream, ok := pointer.Restore(context).(*Stream)
	if !ok || stream == nil {
		return
	}
	select {
	case stream.onStarted <- struct{}{}:
	case <-stream.onClose:
	}
}

//export OnStreamReceive
func OnStreamReceive(cppStream, context, data unsafe.Pointer, length C.size_t) {
	stream, ok := pointer.Restore(context).(*Stream)
	if !ok || stream == nil {
		return
	}
	buf := C.GoBytes(data, C.int(length))
	defer C.free(data)
	select {
	case stream.onReceive <- buf:
	case <-stream.onClose:
	}
}

//export OnStreamSendComplete
func OnStreamSendComplete(cppStream, context unsafe.Pointer) {
	stream, ok := pointer.Restore(context).(*Stream)
	if !ok || stream == nil {
		return
	}
	select {
	case stream.onSend <- struct{}{}:
	case <-stream.onClose:
	}
}
