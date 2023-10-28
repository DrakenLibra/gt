package test

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"github.com/isrc-cas/gt/conn/msquic"
	"io"
	"math/big"
	"net"
	"os"
	"strings"
	"testing"
	"time"
)

const (
	keyFile  = "tls.key"
	certFile = "tls.crt"
)

func TestQUIC(t *testing.T) {
	// 生成 TLS 证书
	err := GenerateTLSKeyAndCert("*.example.com,localhost", keyFile, certFile)
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
	listener, err := msquic.NewListenr("127.0.0.1:0", 10_000, keyFile, certFile, "")
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

func quicServerStart(listener *msquic.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			panic(err)
		}

		go func() {
			conn := conn.(*msquic.Connection)
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
	conn, err := msquic.NewConnection("localhost:"+port, 10_000, certFile, true)
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

func GenerateTLSKeyAndCert(hosts, keyPath, certPath string) (err error) {
	notBefore := time.Now()
	notAfter := notBefore.Add(validityPeriod)
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		return
	}
	x590Cert := &x509.Certificate{
		SerialNumber:          serialNumber,
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}
	hostSlice := strings.Split(hosts, ",")
	for _, host := range hostSlice {
		ip := net.ParseIP(host)
		if ip != nil {
			x590Cert.IPAddresses = append(x590Cert.IPAddresses, ip)
		} else {
			x590Cert.DNSNames = append(x590Cert.DNSNames, host)
		}
	}

	// key
	ecdsaKey, err := ecdsa.GenerateKey(ecdsaCurve, rand.Reader)
	if err != nil {
		return
	}
	keyFile, err := os.Create(keyPath)
	if err != nil {
		return
	}
	keyBytes, err := x509.MarshalECPrivateKey(ecdsaKey)
	if err != nil {
		return
	}
	err = pem.Encode(keyFile, &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})
	if err != nil {
		return
	}
	err = keyFile.Close()
	if err != nil {
		return
	}

	// crt
	certBytes, err := x509.CreateCertificate(
		rand.Reader,
		x590Cert,
		x590Cert,
		&ecdsaKey.PublicKey,
		ecdsaKey,
	)
	if err != nil {
		return
	}
	certFile, err := os.Create(certPath)
	if err != nil {
		return
	}
	err = pem.Encode(certFile, &pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	if err != nil {
		return
	}
	err = certFile.Close()
	return
}
