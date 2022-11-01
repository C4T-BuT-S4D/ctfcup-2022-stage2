package client

import (
	"crypto/tls"
	"fmt"

	qnet "gmtservice/pkg/quicrpc"

	"github.com/pomo-mondreganto/go-checklib"
	"github.com/pomo-mondreganto/go-checklib/require"
	"google.golang.org/grpc"
)

const servicePort = 9090

func Connect(c *checklib.C, host string) *grpc.ClientConn {
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

	conn, err := grpc.Dial(fmt.Sprintf("%s:%d", host, servicePort), grpcOpts...)
	require.NoError(c, err, "connection error")
	return conn
}
