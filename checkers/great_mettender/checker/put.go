package checker

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"

	"gmtchecker/client"
	pingerpb "gmtservice/pkg/proto/pinger"
	tenderspb "gmtservice/pkg/proto/tenders"

	"github.com/google/uuid"
	"github.com/pomo-mondreganto/go-checklib"
	"github.com/pomo-mondreganto/go-checklib/gen"
	"github.com/pomo-mondreganto/go-checklib/require"
	"google.golang.org/grpc/metadata"
)

func (ch *Checker) Put(c *checklib.C, host, _, flag string, _ int) {
	conn := client.Connect(c, host)

	pingClient := pingerpb.NewPingerServiceClient(conn)
	bidsClient := tenderspb.NewBidServiceClient(conn)
	tendersClient := tenderspb.NewTendersServiceClient(conn)

	_, err := pingClient.Ping(c, &pingerpb.PingRequest{})
	require.NoError(c, err, "ping error")

	author := uuid.NewString()
	ctxAuthor := metadata.AppendToOutgoingContext(c, "user", author)

	var bidUsers []string
	var bidCtxs []context.Context
	for i := 0; i < gen.RandInt(5, 10); i++ {
		user := uuid.NewString()
		bidUsers = append(bidUsers, user)
		bidCtxs = append(bidCtxs, metadata.AppendToOutgoingContext(c, "user", user))
	}

	realInput := fmt.Sprintf("Here's your flag: %s", flag)
	tender := &tenderspb.Tender{
		Name:         gen.String(33),
		Description:  gen.Paragraph(),
		Private:      true,
		ProgramInput: client.EncodeFormat(realInput),
	}
	createTender1, err := tendersClient.Create(ctxAuthor, &tenderspb.Tender_CreateRequest{Tender: tender})
	require.NoError(c, err, "create tender error")
	require.False(c, createTender1.Tender == nil || createTender1.Tender.Id == "", "invalid tender")

	tender.Id = createTender1.Tender.Id

	bidToUser := make(map[string]string)
	for i := 0; i < gen.RandInt(1, 3); i++ {
		bid := &tenderspb.Bid{
			TenderId:    tender.Id,
			Price:       rand.Float64() * 10000,
			Description: gen.Sentence(),
			Program:     client.SampleCatProgram(),
		}
		userIndex := gen.RandInt(0, len(bidUsers)-1)
		createBid, err := bidsClient.Create(
			bidCtxs[userIndex],
			&tenderspb.Bid_CreateRequest{Bid: bid},
		)
		require.NoError(c, err, "create bid error")
		require.False(c, createBid.Bid == nil || createBid.Bid.Id == "", "invalid bid")

		bidToUser[createBid.Bid.Id] = bidUsers[userIndex]
	}

	data := client.FlagData{
		Author:    author,
		TenderID:  tender.Id,
		BidToUser: bidToUser,
	}
	raw, _ := json.Marshal(data)
	// Public flag data is ID of the tender.
	checklib.OK(c, tender.Id, string(raw))
}
