package cmd

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"gmtchecker/lib"
	pingerpb "gmtservice/pkg/proto/pinger"
	tenderspb "gmtservice/pkg/proto/tenders"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

func Check(ctx context.Context, host string) error {
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

	scenarios := []func(
		context.Context,
		tenderspb.TendersServiceClient,
		tenderspb.BidServiceClient,
	) error{
		checkScenario1,
		checkScenario2,
	}
	cnt := lib.RandInt(2, 6)
	for i := 0; i < cnt; i++ {
		scenario := lib.Sample(scenarios)
		if err := scenario(ctx, tendersClient, bidsClient); err != nil {
			return err
		}
		time.Sleep(time.Millisecond * 100 * time.Duration(lib.RandInt(1, 5)))
	}
	return lib.OK("OK", "")
}

func checkScenario1(
	ctx context.Context,
	tendersClient tenderspb.TendersServiceClient,
	bidsClient tenderspb.BidServiceClient,
) error {
	user1 := uuid.NewString()
	ctx1 := metadata.AppendToOutgoingContext(ctx, "user", user1)

	var bidUsers []string
	var bidCtxs []context.Context
	for i := 0; i < lib.RandInt(5, 10); i++ {
		user := uuid.NewString()
		bidUsers = append(bidUsers, user)
		bidCtxs = append(bidCtxs, metadata.AppendToOutgoingContext(ctx, "user", user))
	}

	tender1 := &tenderspb.Tender{
		Name:         lib.Lorem().Words(lib.RandInt(3, 5)),
		Description:  lib.Lorem().Paragraph(),
		Private:      true,
		ProgramInput: lib.EncodeFormat(lib.Lorem().Sentences(lib.RandInt(1, 3))),
	}
	createTender1, err := tendersClient.Create(ctx1, &tenderspb.Tender_CreateRequest{Tender: tender1})
	if err != nil {
		return lib.Mumble("create error", "tender1 create: %v", err)
	}
	if createTender1.Tender == nil {
		return lib.Mumble("invalid tender", "tender1 empty")
	}
	if createTender1.Tender.Id == "" {
		return lib.Mumble("invalid tender", "tender1 empty id")
	}

	timeDiff := time.Since(createTender1.Tender.CreatedAt.AsTime())
	if timeDiff > time.Second*10 || timeDiff < -time.Second*10 {
		return lib.Mumble("invalid tender", "tender1 created_at diff: %v", timeDiff)
	}

	tender1.Id = createTender1.Tender.Id
	tender1.Author = user1
	createTender1.Tender.CreatedAt = nil
	if bad, diff := lib.DiffProto(tender1, createTender1.Tender); bad {
		return lib.Mumble("invalid create tender", "tender1 create diff: %v", diff)
	}

	getTender1, err := tendersClient.Get(ctx1, &tenderspb.Tender_GetRequest{Id: tender1.Id})
	if err != nil {
		return lib.Mumble("get error", "tender1 get: %v", err)
	}
	if getTender1.Tender != nil {
		getTender1.Tender.CreatedAt = nil
	}
	if bad, diff := lib.DiffProto(tender1, getTender1.Tender); bad {
		return lib.Mumble("invalid get tender", "tender1 get diff: %v", diff)
	}

	var madeBids []*tenderspb.Bid
	bidToUser := make(map[string]int)
	for i := 0; i < lib.RandInt(5, 10); i++ {
		bid := &tenderspb.Bid{
			TenderId:    tender1.Id,
			Price:       rand.Float64() * 10000,
			Description: lib.Lorem().Words(lib.RandInt(3, 5)),
			Program:     lib.SampleProgram(),
		}
		userIndex := lib.RandInt(0, len(bidUsers)-1)
		createBid, err := bidsClient.Create(
			bidCtxs[userIndex],
			&tenderspb.Bid_CreateRequest{Bid: bid},
		)
		if err != nil {
			return lib.Mumble("create bid error", "create bid: %v", err)
		}
		if createBid.Bid == nil {
			return lib.Mumble("invalid bid", "bid empty")
		}
		if createBid.Bid.Id == "" {
			return lib.Mumble("invalid bid", "bid empty id")
		}
		bid.Id = createBid.Bid.Id
		if bad, diff := lib.DiffProto(bid, createBid.Bid); bad {
			return lib.Mumble("invalid create bid", "bid create diff: %v", diff)
		}
		madeBids = append(madeBids, bid)
		bidToUser[bid.Id] = userIndex
	}

	finishTender1, err := tendersClient.Close(ctx1, &tenderspb.Tender_CloseRequest{Id: tender1.Id})
	if err != nil {
		return lib.Mumble("close error", "tender1 close: %v", err)
	}
	if finishTender1.WinningBid == nil {
		return lib.Mumble("winning bid", "tender1 winning bid empty")
	}

	sort.Slice(madeBids, func(i, j int) bool {
		return madeBids[i].Price < madeBids[j].Price
	})
	if bad, diff := lib.DiffProto(madeBids[0], finishTender1.WinningBid); bad {
		return lib.Mumble("invalid winning bid", "bid winning diff: %v", diff)
	}

	execute1, err := bidsClient.Execute(ctx1, &tenderspb.Bid_ExecuteRequest{Id: finishTender1.WinningBid.Id})
	if err != nil {
		return lib.Mumble("execute error", "execute1: %v", err)
	}

	if execute1.Output == "" {
		return lib.Mumble("invalid execution", "execute1 output empty: %+v", execute1)
	}
	if execute1.Error != "" {
		return lib.Mumble("invalid execution", "execute1 error: %+v", execute1)
	}

	return nil
}

func checkScenario2(
	ctx context.Context,
	tendersClient tenderspb.TendersServiceClient,
	bidsClient tenderspb.BidServiceClient,
) error {
	user1 := uuid.NewString()
	ctx1 := metadata.AppendToOutgoingContext(ctx, "user", user1)

	var bidUsers []string
	var bidCtxs []context.Context
	for i := 0; i < lib.RandInt(5, 10); i++ {
		user := uuid.NewString()
		bidUsers = append(bidUsers, user)
		bidCtxs = append(bidCtxs, metadata.AppendToOutgoingContext(ctx, "user", user))
	}

	realInput := lib.Lorem().Sentences(lib.RandInt(1, 3))
	tender1 := &tenderspb.Tender{
		Name:         lib.Lorem().Words(lib.RandInt(3, 5)),
		Description:  lib.Lorem().Paragraph(),
		Private:      true,
		ProgramInput: lib.EncodeFormat(realInput),
	}
	createTender1, err := tendersClient.Create(ctx1, &tenderspb.Tender_CreateRequest{Tender: tender1})
	if err != nil {
		return lib.Mumble("create error", "tender1 create: %v", err)
	}
	if createTender1.Tender == nil || createTender1.Tender.Id == "" {
		return lib.Mumble("invalid tender", "tender1 invalid")
	}

	tender1.Id = createTender1.Tender.Id
	tender1.Author = user1

	var madeBids []*tenderspb.Bid
	bidToCtx := make(map[string]context.Context)
	bidToOutput := make(map[string]string)
	for i := 0; i < lib.RandInt(5, 10); i++ {
		program, output := lib.SampleProgramWithOutput(realInput)
		bid := &tenderspb.Bid{
			TenderId:    tender1.Id,
			Price:       rand.Float64() * 10000,
			Description: lib.Lorem().Words(lib.RandInt(3, 5)),
			Program:     program,
		}
		userIndex := lib.RandInt(0, len(bidUsers)-1)
		createBid, err := bidsClient.Create(
			bidCtxs[userIndex],
			&tenderspb.Bid_CreateRequest{Bid: bid},
		)
		if err != nil {
			return lib.Mumble("create bid error", "create bid: %v", err)
		}
		if createBid.Bid == nil {
			return lib.Mumble("invalid bid", "bid empty")
		}
		if createBid.Bid.Id == "" {
			return lib.Mumble("invalid bid", "bid empty id")
		}
		bid.Id = createBid.Bid.Id
		if bad, diff := lib.DiffProto(bid, createBid.Bid); bad {
			return lib.Mumble("invalid create bid", "bid create diff: %v", diff)
		}
		madeBids = append(madeBids, bid)
		bidToCtx[bid.Id] = bidCtxs[userIndex]
		bidToOutput[bid.Id] = output
	}

	finishTender1, err := tendersClient.Close(ctx1, &tenderspb.Tender_CloseRequest{Id: tender1.Id})
	if err != nil {
		return lib.Mumble("close error", "tender1 close: %v", err)
	}
	if finishTender1.WinningBid == nil {
		return lib.Mumble("winning bid", "tender1 winning bid empty")
	}

	sort.Slice(madeBids, func(i, j int) bool {
		return madeBids[i].Price < madeBids[j].Price
	})
	if bad, diff := lib.DiffProto(madeBids[0], finishTender1.WinningBid); bad {
		return lib.Mumble("invalid winning bid", "bid winning diff: %v", diff)
	}

	winningID := finishTender1.WinningBid.Id
	execute1, err := bidsClient.Execute(ctx1, &tenderspb.Bid_ExecuteRequest{Id: winningID})
	if err != nil {
		return lib.Mumble("execute error", "execute1: %v", err)
	}

	expectedOutput := bidToOutput[winningID]
	if diff := cmp.Diff(expectedOutput, execute1.Output); diff != "" {
		return lib.Mumble("invalid output", "execute1 result: %+v; diff: %v", execute1, diff)
	}

	if execute1.Error != "" {
		return lib.Mumble("invalid execute error", "execute1 result: %+v", execute1)
	}

	winningCtx := bidToCtx[winningID]
	execute2, err := bidsClient.Execute(winningCtx, &tenderspb.Bid_ExecuteRequest{Id: winningID})
	if err != nil {
		return lib.Mumble("execute error", "execute1: %v", err)
	}
	if execute2.Output != "" {
		return lib.Mumble("invalid output", "leaking flag in output: %+v", execute2)
	}

	if execute2.Error != "" {
		return lib.Mumble("invalid execute error", "execute2 result: %+v", execute1)
	}

	for bidID, output := range bidToOutput {
		if bidID == winningID {
			continue
		}

		execute3, err := bidsClient.Execute(ctx1, &tenderspb.Bid_ExecuteRequest{Id: bidID})
		if err != nil {
			return lib.Mumble("execute error", "execute1: %v", err)
		}

		if diff := cmp.Diff(output, execute3.Output); diff != "" {
			return lib.Mumble("invalid output", "execute3 result: %+v; diff: %v", execute1, diff)
		}

		if execute3.Error != "" {
			return lib.Mumble("invalid execute error", "execute3 result: %+v", execute1)
		}

		break
	}

	return nil
}
