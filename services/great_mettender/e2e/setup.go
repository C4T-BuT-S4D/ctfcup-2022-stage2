package e2e

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"testing"
	"time"

	"great_mettender/internal/pinger"
	pingerpb "great_mettender/pkg/proto/pinger"
	qnet "great_mettender/pkg/quicrpc"

	"github.com/lucas-clemente/quic-go"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

func startServer(t *testing.T) net.Addr {
	t.Helper()

	tlsConfig := generateTLSConfig(t)

	ql, err := quic.ListenAddr(":0", tlsConfig, nil)
	require.NoError(t, err)

	t.Logf("server listening on %s", ql.Addr())

	lis := qnet.Listen(ql)

	s := grpc.NewServer()
	pingerpb.RegisterPingerServiceServer(s, &pinger.Service{})
	go func() {
		require.NoError(t, s.Serve(lis))
	}()
	return lis.Addr()
}

func createChannel(t *testing.T, addr net.Addr) (conn *grpc.ClientConn) {
	t.Helper()

	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		NextProtos:         []string{"great_mettender"},
	}

	creds := qnet.NewCredentials(tlsConf)

	dialer := qnet.NewQuicDialer(tlsConf)
	grpcOpts := []grpc.DialOption{
		grpc.WithContextDialer(dialer),
		grpc.WithTransportCredentials(creds),
	}

	start := time.Now()

	var err error
	for {
		conn, err = grpc.Dial(
			fmt.Sprintf("127.0.0.1:%d", addr.(*net.UDPAddr).AddrPort().Port()),
			grpcOpts...,
		)
		if err != nil && time.Since(start) < time.Second*5 {
			continue
		}
		require.NoError(t, err)
		break
	}
	return conn
}

func generateTLSConfig(t *testing.T) *tls.Config {
	t.Helper()

	key, err := rsa.GenerateKey(rand.Reader, 1024)
	require.NoError(t, err)

	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	require.NoError(t, err)

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})

	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	require.NoError(t, err)

	return &tls.Config{
		Certificates: []tls.Certificate{tlsCert},
		NextProtos:   []string{"great_mettender"},
	}
}
