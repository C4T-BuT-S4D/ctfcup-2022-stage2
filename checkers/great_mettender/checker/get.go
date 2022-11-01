package checker

import (
	"encoding/json"
	"strings"
	"time"

	"gmtchecker/client"
	pingerpb "gmtservice/pkg/proto/pinger"
	tenderspb "gmtservice/pkg/proto/tenders"

	"github.com/pomo-mondreganto/go-checklib"
	"github.com/pomo-mondreganto/go-checklib/require"
	o "github.com/pomo-mondreganto/go-checklib/require/options"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func (ch *Checker) Get(c *checklib.C, host, flagID, flag string, _ int) {
	var data client.FlagData

	require.NoError(
		c,
		json.Unmarshal([]byte(flagID), &data),
		"bad flag data",
		o.CheckFailed(),
	)

	conn := client.Connect(c, host)

	pingClient := pingerpb.NewPingerServiceClient(conn)
	bidsClient := tenderspb.NewBidServiceClient(conn)
	tendersClient := tenderspb.NewTendersServiceClient(conn)

	_, err := pingClient.Ping(c, &pingerpb.PingRequest{})
	require.NoError(c, err, "ping error")

	ctxAuthor := metadata.AppendToOutgoingContext(c, "user", data.Author)

	tender, err := tendersClient.Get(ctxAuthor, &tenderspb.Tender_GetRequest{Id: data.TenderID})
	require.NoError(c, err, "missing tender", o.Corrupt())
	require.NotNil(c, tender.Tender, "missing tender", o.Corrupt())

	canCheckExecute := time.Since(tender.Tender.CreatedAt.AsTime()) > time.Minute

	// Finish old tender.
	if canCheckExecute && !tender.Tender.Finished {
		_, err := tendersClient.Close(ctxAuthor, &tenderspb.Tender_CloseRequest{Id: data.TenderID})
		require.NoError(c, err, "close tender error")
	}

	input, err := client.DecodeFormat(tender.Tender.ProgramInput)
	require.NoError(c, err, "invalid tender input", o.Corrupt())

	require.True(c, strings.Contains(input, flag), "missing flag", o.Corrupt())

	// Not yet.
	if !canCheckExecute {
		checklib.OK(c, "OK", "")
	}

	for bidID, user := range data.BidToUser {
		execute1, err := bidsClient.Execute(ctxAuthor, &tenderspb.Bid_ExecuteRequest{Id: bidID})
		require.NoError(c, err, "exec bid error", o.Corrupt())
		require.Equal(c, input, execute1.Output, "exec bid bad output", o.Corrupt())

		require.Equal(c, "", execute1.Error, "exec bid error", o.Corrupt())

		// Bid can be lost, handle PermissionDenied error in such case.
		ctxBid := metadata.AppendToOutgoingContext(c, "user", user)
		execute2, err := bidsClient.Execute(ctxBid, &tenderspb.Bid_ExecuteRequest{Id: bidID})
		if err != nil {
			if st, ok := status.FromError(err); ok && st.Code() == codes.PermissionDenied {
				continue
			}
			require.NoError(c, err, "exec bid error", o.Corrupt())
		}
		require.Equal(c, "", execute2.Error, "exec bid error", o.Corrupt())
	}

	checklib.OK(c, "OK", "")
}
