package quic

/*
#include "quic.h"
*/
import "C"

// TODO ALPN
// TODO resumption ticket，使得后续的连接可以复用之前的 TLS 握手结果，减少握手时间
// TODO DatagramSend

func init() {
	ok := C.Init()
	if !ok {
		panic("msquic init failed")
	}
}
