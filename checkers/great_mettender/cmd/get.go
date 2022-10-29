package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"gmtchecker/lib"
	pingerpb "gmtservice/pkg/proto/pinger"
	tenderspb "gmtservice/pkg/proto/tenders"

	"github.com/google/go-cmp/cmp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func Get(ctx context.Context, host, flagID, flag, _ string) error {
	var data lib.FlagData
	if err := json.Unmarshal([]byte(flagID), &data); err != nil {
		return fmt.Errorf("unmarshalling flag data: %w", err)
	}

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

	ctxAuthor := metadata.AppendToOutgoingContext(ctx, "user", data.Author)

	tender, err := tendersClient.Get(ctxAuthor, &tenderspb.Tender_GetRequest{Id: data.TenderID})
	if err != nil {
		return lib.Corrupt("missing tender", "tender get error: %v", err)
	}
	if tender.Tender == nil {
		return lib.Corrupt("missing tender", "tender get nil")
	}

	canCheckExecute := false

	// Finish old tender.
	if time.Since(tender.Tender.CreatedAt.AsTime()) > time.Minute {
		canCheckExecute = true

		if !tender.Tender.Finished {
			if _, err := tendersClient.Close(ctxAuthor, &tenderspb.Tender_CloseRequest{Id: data.TenderID}); err != nil {
				return lib.Mumble("close tender error", "tender close error: %v", err)
			}
		}
	}

	input, err := lib.DecodeFormat(tender.Tender.ProgramInput)
	if err != nil {
		return lib.Corrupt("invalid tender input", "tender input decode error: %v", err)
	}
	if !strings.Contains(input, flag) {
		return lib.Corrupt(
			"invalid tender input",
			"tender input is missing flag: %v",
			input,
		)
	}

	// Not yet.
	if !canCheckExecute {
		return lib.OK("OK", "")
	}

	for bidID, user := range data.BidToUser {
		execute1, err := bidsClient.Execute(ctxAuthor, &tenderspb.Bid_ExecuteRequest{Id: bidID})
		if err != nil {
			return lib.Corrupt("exec bid error", "exec1 error: %v", err)
		}
		if diff := cmp.Diff(input, execute1.Output); diff != "" {
			return lib.Corrupt("exec bid invalid", "exec1 res: %v; diff %v", execute1, diff)
		}
		if execute1.Error != "" {
			return lib.Corrupt("exec bid invalid", "exec1 res: %v", execute1)
		}

		// Bid can be lost, handle PermissionDenied error in such case.
		ctxBid := metadata.AppendToOutgoingContext(ctx, "user", user)
		execute2, err := bidsClient.Execute(ctxBid, &tenderspb.Bid_ExecuteRequest{Id: bidID})
		if err != nil {
			if st, ok := status.FromError(err); ok && st.Code() == codes.PermissionDenied {
				continue
			}
			return lib.Corrupt("exec bid error", "exec2 error: %v", err)
		}
		if execute2.Error != "" {
			return lib.Corrupt("exec bid invalid", "exec2 res: %v", execute1)
		}
	}

	return lib.OK("OK", "")
}
