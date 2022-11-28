package checker

import (
	"context"
	"math/rand"
	"sort"
	"time"

	"gmtchecker/client"
	pingerpb "gmtservice/pkg/proto/pinger"
	tenderspb "gmtservice/pkg/proto/tenders"

	"github.com/google/uuid"
	"github.com/pomo-mondreganto/go-checklib"
	"github.com/pomo-mondreganto/go-checklib/gen"
	"github.com/pomo-mondreganto/go-checklib/require"
	"google.golang.org/grpc/metadata"
)

func (ch *Checker) Check(c *checklib.C, host string) {
	conn := client.Connect(c, host)

	pingClient := pingerpb.NewPingerServiceClient(conn)
	bidsClient := tenderspb.NewBidServiceClient(conn)
	tendersClient := tenderspb.NewTendersServiceClient(conn)

	_, err := pingClient.Ping(c, &pingerpb.PingRequest{})
	require.NoError(c, err, "ping error")

	scenarios := []func(
		*checklib.C,
		tenderspb.TendersServiceClient,
		tenderspb.BidServiceClient,
	){
		checkScenario1,
		checkScenario2,
	}
	cnt := gen.RandInt(2, 6)
	for i := 0; i < cnt; i++ {
		scenario := gen.Sample(scenarios)
		scenario(c, tendersClient, bidsClient)
		time.Sleep(time.Millisecond * 100 * time.Duration(gen.RandInt(1, 5)))
	}
	checklib.OK(c, "OK", "")
}

func checkScenario1(
	c *checklib.C,
	tendersClient tenderspb.TendersServiceClient,
	bidsClient tenderspb.BidServiceClient,
) {
	user1 := uuid.NewString()
	ctx1 := metadata.AppendToOutgoingContext(c, "user", user1)

	var bidUsers []string
	var bidCtxs []context.Context
	for i := 0; i < gen.RandInt(5, 10); i++ {
		user := uuid.NewString()
		bidUsers = append(bidUsers, user)
		bidCtxs = append(bidCtxs, metadata.AppendToOutgoingContext(c, "user", user))
	}

	tender1 := &tenderspb.Tender{
		Name:         gen.String(33),
		Description:  gen.Paragraph(),
		Private:      true,
		ProgramInput: client.EncodeFormat(gen.Sentences(gen.RandInt(1, 3))),
	}
	createTender1, err := tendersClient.Create(ctx1, &tenderspb.Tender_CreateRequest{Tender: tender1})
	require.NoError(c, err, "create tender error")
	require.False(c, createTender1.Tender == nil || createTender1.Tender.Id == "", "invalid tender")

	timeDiff := time.Since(createTender1.Tender.CreatedAt.AsTime())
	require.False(c, timeDiff > time.Second*10 || timeDiff < -time.Second*10, "invalid tender")

	tender1.Id = createTender1.Tender.Id
	tender1.Author = user1
	createTender1.Tender.CreatedAt = nil
	require.EqualProto(c, tender1, createTender1.Tender, "invalid create tender")

	getTender1, err := tendersClient.Get(ctx1, &tenderspb.Tender_GetRequest{Id: tender1.Id})
	require.NoError(c, err, "get tender error")
	if getTender1.Tender != nil {
		getTender1.Tender.CreatedAt = nil
	}
	require.EqualProto(c, tender1, getTender1.Tender, "invalid get tender")

	var madeBids []*tenderspb.Bid
	bidToUser := make(map[string]int)
	for i := 0; i < gen.RandInt(5, 10); i++ {
		bid := &tenderspb.Bid{
			TenderId:    tender1.Id,
			Price:       rand.Float64() * 10000,
			Description: gen.Sentence(),
			Program:     client.SampleProgram(),
		}
		userIndex := gen.RandInt(0, len(bidUsers)-1)
		createBid, err := bidsClient.Create(
			bidCtxs[userIndex],
			&tenderspb.Bid_CreateRequest{Bid: bid},
		)
		require.NoError(c, err, "create bid error")
		require.False(c, createBid.Bid == nil || createBid.Bid.Id == "", "invalid bid")

		bid.Id = createBid.Bid.Id
		require.EqualProto(c, bid, createBid.Bid, "invalid create bid")

		madeBids = append(madeBids, bid)
		bidToUser[bid.Id] = userIndex
	}

	finishTender1, err := tendersClient.Close(ctx1, &tenderspb.Tender_CloseRequest{Id: tender1.Id})
	require.NoError(c, err, "close tender error")
	require.NotNil(c, finishTender1.WinningBid, "no winning bid")

	sort.Slice(madeBids, func(i, j int) bool {
		return madeBids[i].Price < madeBids[j].Price
	})
	require.EqualProto(c, madeBids[0], finishTender1.WinningBid, "invalid winning bid")

	execute1, err := bidsClient.Execute(ctx1, &tenderspb.Bid_ExecuteRequest{Id: finishTender1.WinningBid.Id})
	require.NoError(c, err, "execute error")
	require.NotEqual(c, "", execute1.Output, "invalid execute output")
	require.Equal(c, "", execute1.Error, "error in execute result")
}

func checkScenario2(
	c *checklib.C,
	tendersClient tenderspb.TendersServiceClient,
	bidsClient tenderspb.BidServiceClient,
) {
	user1 := uuid.NewString()
	ctx1 := metadata.AppendToOutgoingContext(c, "user", user1)

	var bidUsers []string
	var bidCtxs []context.Context
	for i := 0; i < gen.RandInt(5, 10); i++ {
		user := uuid.NewString()
		bidUsers = append(bidUsers, user)
		bidCtxs = append(bidCtxs, metadata.AppendToOutgoingContext(c, "user", user))
	}

	realInput := gen.Sentences(gen.RandInt(1, 3))
	tender1 := &tenderspb.Tender{
		Name:         gen.String(33),
		Description:  gen.Paragraph(),
		Private:      true,
		ProgramInput: client.EncodeFormat(realInput),
	}
	createTender1, err := tendersClient.Create(ctx1, &tenderspb.Tender_CreateRequest{Tender: tender1})
	require.NoError(c, err, "create tender error")
	require.False(c, createTender1.Tender == nil || createTender1.Tender.Id == "", "invalid tender")

	tender1.Id = createTender1.Tender.Id
	tender1.Author = user1

	var madeBids []*tenderspb.Bid
	bidToCtx := make(map[string]context.Context)
	bidToOutput := make(map[string]string)
	for i := 0; i < gen.RandInt(5, 10); i++ {
		program, output := client.SampleProgramWithOutput(realInput)
		bid := &tenderspb.Bid{
			TenderId:    tender1.Id,
			Price:       rand.Float64() * 10000,
			Description: gen.Sentence(),
			Program:     program,
		}
		userIndex := gen.RandInt(0, len(bidUsers)-1)
		createBid, err := bidsClient.Create(
			bidCtxs[userIndex],
			&tenderspb.Bid_CreateRequest{Bid: bid},
		)
		require.NoError(c, err, "create bid error")
		require.False(c, createBid.Bid == nil || createBid.Bid.Id == "", "invalid bid")

		bid.Id = createBid.Bid.Id
		require.EqualProto(c, bid, createBid.Bid, "invalid create bid")

		madeBids = append(madeBids, bid)
		bidToCtx[bid.Id] = bidCtxs[userIndex]
		bidToOutput[bid.Id] = output
	}

	finishTender1, err := tendersClient.Close(ctx1, &tenderspb.Tender_CloseRequest{Id: tender1.Id})
	require.NoError(c, err, "close tender error")
	require.NotNil(c, finishTender1.WinningBid, "no winning bid")

	sort.Slice(madeBids, func(i, j int) bool {
		return madeBids[i].Price < madeBids[j].Price
	})
	require.EqualProto(c, madeBids[0], finishTender1.WinningBid, "invalid winning bid")

	winningID := finishTender1.WinningBid.Id
	expectedOutput := bidToOutput[winningID]

	execute1, err := bidsClient.Execute(ctx1, &tenderspb.Bid_ExecuteRequest{Id: winningID})
	require.NoError(c, err, "execute error")
	require.Equal(c, expectedOutput, execute1.Output, "invalid execute output")
	require.Equal(c, "", execute1.Error, "execute error not empty")

	winningCtx := bidToCtx[winningID]
	execute2, err := bidsClient.Execute(winningCtx, &tenderspb.Bid_ExecuteRequest{Id: winningID})
	require.NoError(c, err, "execute error")
	require.Equal(c, "", execute2.Error, "execute error not empty")
	require.Equal(c, "", execute2.Output, "invalid execute output")

	for bidID, output := range bidToOutput {
		if bidID == winningID {
			continue
		}

		execute3, err := bidsClient.Execute(ctx1, &tenderspb.Bid_ExecuteRequest{Id: bidID})
		require.NoError(c, err, "execute error")
		require.Equal(c, output, execute3.Output, "invalid execute output")
		require.Equal(c, "", execute3.Error, "invalid execute error")
	}
}
