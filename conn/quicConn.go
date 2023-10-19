package conn

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	quicbbr "github.com/DrakenLibra/gt-bbr"
	probing "github.com/prometheus-community/pro-bing"
	"github.com/quic-go/quic-go"
	"math/big"
	"net"
	"sync"
	"sync/atomic"
)

type QuicConnection struct {
	quic.Connection
	quic.Stream
}

type QuicBbrConnection struct {
	quicbbr.Session
	quicbbr.Stream
}

type QuicListener struct {
	quic.Listener
}

type QuicBbrListener struct {
	quicbbr.Listener
}

var _ net.Conn = &QuicConnection{}
var _ net.Listener = &QuicListener{}
var _ net.Conn = &QuicBbrConnection{}
var _ net.Listener = &QuicBbrListener{}

func (c *QuicBbrConnection) Close() error {
	err := c.Stream.Close()
	err = c.Session.Close()
	return err
}

func QuicDial(addr string, config *tls.Config) (net.Conn, error) {
	config.NextProtos = []string{"gt-quic"}
	conn, err := quic.DialAddr(context.Background(), addr, config, &quic.Config{})
	if err != nil {
		panic(err)
	}
	stream, err := conn.OpenStreamSync(context.Background())
	if err != nil {
		panic(err)
	}
	nc := &QuicConnection{
		Connection: conn,
		Stream:     stream,
	}
	return nc, err
}

func QuicBbrDial(addr string, config *tls.Config) (net.Conn, error) {
	config.NextProtos = []string{"gt-quic"}
	conn, err := quicbbr.DialAddr(addr, config, &quicbbr.Config{})
	if err != nil {
		panic(err)
	}
	stream, err := conn.OpenStreamSync()
	if err != nil {
		panic(err)
	}
	nc := &QuicBbrConnection{
		Session: conn,
		Stream:  stream,
	}
	return nc, err
}

func QuicListen(addr string, config *tls.Config) (net.Listener, error) {
	config.NextProtos = []string{"gt-quic"}
	listener, err := quic.ListenAddr(addr, config, &quic.Config{})
	if err != nil {
		panic(err)
	}
	ln := &QuicListener{
		Listener: *listener,
	}
	return ln, err
}

func QuicBbrListen(addr string, config *tls.Config) (net.Listener, error) {
	config.NextProtos = []string{"gt-quic"}
	listener, err := quicbbr.ListenAddr(addr, config, &quicbbr.Config{})
	if err != nil {
		panic(err)
	}
	ln := &QuicBbrListener{
		Listener: listener,
	}
	return ln, err
}

func (ln *QuicListener) Accept() (net.Conn, error) {
	conn, _ := ln.Listener.Accept(context.Background())
	stream, err := conn.AcceptStream(context.Background())
	nc := &QuicConnection{
		Connection: conn,
		Stream:     stream,
	}
	return nc, err
}

func (ln *QuicBbrListener) Accept() (net.Conn, error) {
	conn, _ := ln.Listener.Accept()
	stream, err := conn.AcceptStream()
	nc := &QuicBbrConnection{
		Session: conn,
		Stream:  stream,
	}
	return nc, err
}

func GenerateTLSConfig() *tls.Config {
	ecdsaKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}
	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &ecdsaKey.PublicKey, ecdsaKey)
	if err != nil {
		panic(err)
	}
	keyBytes, err := x509.MarshalECPrivateKey(ecdsaKey)
	if err != nil {
		panic(err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "ECDSA PRIVATE KEY", Bytes: keyBytes})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		panic(err)
	}
	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"gt-quic"},
	}
}

func GetAutoProbesResults(addr string) (avgRtt, pktLoss float64) {
	pureAddr, _, _ := net.SplitHostPort(addr)
	totalNum := 100

	var wg sync.WaitGroup
	wg.Add(totalNum)

	var totalLossRate int64 = 0
	var totalDelay int64 = 0
	for i := 0; i < totalNum; i++ {
		go func() {
			pinger, err := probing.NewPinger(pureAddr)
			if err != nil {
				panic(err)
			}
			pinger.Count = 3
			err = pinger.Run()
			if err != nil {
				panic(err)
			}
			stats := pinger.Statistics()
			avgRtt := stats.AvgRtt.Microseconds()
			pktLoss := int64(stats.PacketLoss * 100)
			atomic.AddInt64(&totalLossRate, pktLoss)
			atomic.AddInt64(&totalDelay, avgRtt)
			wg.Done()
		}()
	}
	wg.Wait()

	avgRtt = float64(atomic.LoadInt64(&totalDelay)) / (float64(1000 * totalNum))
	pktLoss = float64(atomic.LoadInt64(&totalLossRate)) / float64(totalNum*100)

	return
}

func GetQuicProbesResults(addr string) (avgRtt, pktLoss float64) {
	pureAddr, _, _ := net.SplitHostPort(addr)
	totalNum := 100

	var wg sync.WaitGroup
	wg.Add(totalNum)

	var totalLossRate int64 = 0
	var totalDelay int64 = 0
	for i := 0; i < totalNum; i++ {
		go func() {
			pinger, err := probing.NewPinger(pureAddr)
			if err != nil {
				panic(err)
			}
			pinger.Count = 3
			err = pinger.Run()
			if err != nil {
				panic(err)
			}
			stats := pinger.Statistics()
			avgRtt := stats.AvgRtt.Microseconds()
			pktLoss := int64(stats.PacketLoss * 100)
			atomic.AddInt64(&totalLossRate, pktLoss)
			atomic.AddInt64(&totalDelay, avgRtt)
			wg.Done()
		}()
	}
	wg.Wait()

	avgRtt = float64(atomic.LoadInt64(&totalDelay)) / (float64(1000 * totalNum))
	pktLoss = float64(atomic.LoadInt64(&totalLossRate)) / float64(totalNum*100)

	return
}
