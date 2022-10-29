package e2e

import (
	"context"
	"testing"

	pingerpb "great_mettender/pkg/proto/pinger"

	"github.com/stretchr/testify/require"
)

func TestPinger(t *testing.T) {
	addr := startServer(t)
	conn := createChannel(t, addr)
	client := pingerpb.NewPingerServiceClient(conn)

	_, err := client.Ping(context.TODO(), &pingerpb.PingRequest{})
	require.NoError(t, err)
}
