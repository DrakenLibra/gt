package quic_test

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"testing"
	"time"

	"github.com/isrc-cas/gt/quic"
	"github.com/isrc-cas/gt/util"
)

const (
	keyFile  = "tls.key"
	certFile = "tls.crt"
)

func TestQUIC(t *testing.T) {
	// 生成 TLS 证书
	err := util.GenerateTLSKeyAndCert("*.example.com,localhost", keyFile, certFile)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err = os.Remove(keyFile)
		if err != nil {
			t.Fatal(err)
		}
		err = os.Remove(certFile)
		if err != nil {
			t.Fatal(err)
		}
	}()

	// 启动服务端、客户端
	listener, err := quic.NewListenr("127.0.0.1:0", 10_000, keyFile, certFile, "")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		err := listener.Close()
		if err != nil {
			t.Fatal(err)
		}
	}()
	listenerAddr := listener.Addr().String()
	t.Logf("server listen on %s", listenerAddr)
	go quicServerStart(listener)
	done := make(chan struct{})
	go quicClientStart(listenerAddr, done)

	<-done
}

func quicServerStart(listener *quic.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}

		go func() {
			conn := conn.(*quic.Connection)
			fmt.Printf(
				"new conn localAddr: %s, remoteAddr: %s\n",
				conn.LocalAddr().String(),
				conn.RemoteAddr().String(),
			)
			defer func() {
				err = conn.Close()
				if err != nil {
					panic(err)
				}
			}()
			stream, err := conn.PeerStreamStarted()
			if err != nil {
				panic(err)
			}
			fmt.Printf(
				"new stream localAddr: %s, remoteAddr: %s\n",
				stream.LocalAddr().String(),
				stream.RemoteAddr().String(),
			)

			// echo 服务器
			buf := make([]byte, 1024)
			for {
				nread, err := stream.Read(buf)
				if nread > 0 {
					_, err = stream.Write(buf[:nread])
					if err != nil {
						panic(err)
					}
				}
				if err != nil {
					if errors.Is(err, io.EOF) {
						break
					}
					panic(err)
				}
			}
		}()
	}
}

func quicClientStart(listenerAddr string, done chan struct{}) {
	_, port, err := net.SplitHostPort(listenerAddr) // 证书使用了 localhost，而没有使用 127.0.0.1
	if err != nil {
		panic(err)
	}
	conn, err := quic.NewConnection("localhost:"+port, 10_000, certFile, true)
	if err != nil {
		panic(err)
	}
	defer func() {
		err = conn.Close()
		if err != nil {
			panic(err)
		}
	}()
	stream, err := conn.OpenStream()
	if err != nil {
		panic(err)
	}

	randomBuf := make([]byte, 1024)
	_, err = rand.Read(randomBuf)
	if err != nil {
		panic(err)
	}
	buf := make([]byte, len(randomBuf))
	const N = 50 * 1024
	startTime := time.Now()
	for i := 0; i < N; i++ {
		_, err := stream.Write(randomBuf)
		if err != nil {
			panic(err)
		}

		_, err = io.ReadFull(stream, buf)
		if err != nil {
			panic(err)
		}

		if !bytes.Equal(randomBuf, buf) {
			panic("randomBuf != buf")
		}
	}

	// FIXME 用 sample.c 里面的代码或者其他的方式测试传输速度，我测试只有 18MB/s
	fmt.Printf("client loop %d times done, per time write and read %d bytes, cost %s\n",
		N, len(randomBuf), time.Since(startTime))
	close(done)
}
