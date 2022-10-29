package lib

import (
	"crypto/tls"
	"fmt"
	"time"

	qnet "gmtservice/pkg/quicrpc"

	"google.golang.org/grpc"
)

const servicePort = 9090

const ActionTimeout = time.Second * 10

func Connect(host string) (*grpc.ClientConn, error) {
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
	if err != nil {
		return nil, Down("connect error", "connecting to host %s: %v", host, err)
	}
	return conn, nil
}
