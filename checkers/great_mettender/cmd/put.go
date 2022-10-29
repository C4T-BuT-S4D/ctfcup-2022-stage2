package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"

	"gmtchecker/lib"
	pingerpb "gmtservice/pkg/proto/pinger"
	tenderspb "gmtservice/pkg/proto/tenders"

	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

func Put(ctx context.Context, host, _, flag, _ string) error {
	conn, err := lib.Connect(host)
	if err != nil {
		return fmt.Errorf("connecting to host: %w", err)
	}

	pingClient := pingerpb.NewPingerServiceClient(conn)
	bidsClient := tenderspb.NewBidServiceClient(conn)
	tendersClient := tenderspb.NewTendersServiceClient(conn)

	if _, err := pingClient.Ping(ctx, &pingerpb.PingRequest{}); err != nil {
		return lib.Down("ping error", "running ping: %v", err)
	}

	author := uuid.NewString()
	ctxAuthor := metadata.AppendToOutgoingContext(ctx, "user", author)

	var bidUsers []string
	var bidCtxs []context.Context
	for i := 0; i < lib.RandInt(5, 10); i++ {
		user := uuid.NewString()
		bidUsers = append(bidUsers, user)
		bidCtxs = append(bidCtxs, metadata.AppendToOutgoingContext(ctx, "user", user))
	}

	realInput := fmt.Sprintf("Here's your flag: %s", flag)
	tender := &tenderspb.Tender{
		Name:         lib.Lorem().Words(lib.RandInt(3, 5)),
		Description:  lib.Lorem().Paragraph(),
		Private:      true,
		ProgramInput: lib.EncodeFormat(realInput),
	}
	createTender1, err := tendersClient.Create(ctxAuthor, &tenderspb.Tender_CreateRequest{Tender: tender})
	if err != nil {
		return lib.Mumble("create error", "tender create: %v", err)
	}
	if createTender1.Tender == nil || createTender1.Tender.Id == "" {
		return lib.Mumble("invalid tender", "tender id empty")
	}

	tender.Id = createTender1.Tender.Id

	bidToUser := make(map[string]string)
	for i := 0; i < lib.RandInt(1, 3); i++ {
		bid := &tenderspb.Bid{
			TenderId:    tender.Id,
			Price:       rand.Float64() * 10000,
			Description: lib.Lorem().Words(lib.RandInt(3, 5)),
			Program:     lib.SampleCatProgram(),
		}
		userIndex := lib.RandInt(0, len(bidUsers)-1)
		createBid, err := bidsClient.Create(
			bidCtxs[userIndex],
			&tenderspb.Bid_CreateRequest{Bid: bid},
		)
		if err != nil {
			return lib.Mumble("create bid error", "create bid: %v", err)
		}
		if createBid.Bid == nil || createBid.Bid.Id == "" {
			return lib.Mumble("invalid bid", "bid empty")
		}
		bidToUser[createBid.Bid.Id] = bidUsers[userIndex]
	}

	data := lib.FlagData{
		Author:    author,
		TenderID:  tender.Id,
		BidToUser: bidToUser,
	}
	raw, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("marshalling flag data: %w", err)
	}
	// Public flag data is ID of the tender.
	return lib.OK(tender.Id, string(raw))
}
