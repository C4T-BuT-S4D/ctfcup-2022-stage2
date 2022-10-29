package bids

import (
	"context"
	"time"

	"great_mettender/internal/auth"
	"great_mettender/internal/controllers"
	"great_mettender/internal/executor"
	"great_mettender/internal/models"
	tenderspb "great_mettender/pkg/proto/tenders"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// maxProgramLength is 16kb.
	maxProgramLength = 16 * 1024
)

func NewService(
	tendersController *controllers.Tenders,
	bidsController *controllers.Bids,
	exe *executor.Executor,
) *Service {
	return &Service{
		tendersController: tendersController,
		bidsController:    bidsController,
		executor:          exe,
	}
}

type Service struct {
	tenderspb.UnimplementedBidServiceServer

	tendersController *controllers.Tenders
	bidsController    *controllers.Bids
	executor          *executor.Executor
}

func (s *Service) Create(ctx context.Context, request *tenderspb.Bid_CreateRequest) (*tenderspb.Bid_CreateResponse, error) {
	if request.Bid == nil {
		return nil, status.Error(codes.InvalidArgument, "bid required")
	}
	if len(request.Bid.Program) > maxProgramLength {
		return nil, status.Error(codes.InvalidArgument, "program too long")
	}

	user := auth.UserFromContext(ctx)

	bid := models.NewBidFromProto(request.Bid)
	bid.ID = uuid.NewString()
	bid.CreatedAt = time.Now()
	bid.Author = user

	tender, err := s.tendersController.Get(ctx, bid.TenderID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "fetching tender: %v", err)
	}
	if tender.Finished {
		return nil, status.Error(codes.InvalidArgument, "cannot bid on a finished tender")
	}

	reputation, err := s.bidsController.CalculateReputation(ctx, user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "calculating reputation: %v", err)
	}
	if reputation < tender.RequiredReputation {
		return nil, status.Errorf(codes.InvalidArgument, "bad reputation: %v < %v", reputation, tender.RequiredReputation)
	}

	if err := s.bidsController.Add(ctx, bid); err != nil {
		return nil, status.Errorf(codes.Internal, "adding bid: %v", err)
	}
	return &tenderspb.Bid_CreateResponse{Bid: bid.ToProto()}, nil
}

func (s *Service) Execute(ctx context.Context, request *tenderspb.Bid_ExecuteRequest) (*tenderspb.Bid_ExecuteResponse, error) {
	bid, err := s.bidsController.Get(ctx, request.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "fetching bid: %v", err)
	}

	tender, err := s.tendersController.Get(ctx, bid.TenderID)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "fetching tender: %v", err)
	}

	if !tender.Finished {
		return nil, status.Error(codes.PermissionDenied, "not finished")
	}

	user := auth.UserFromContext(ctx)
	if user != tender.Author && !bid.Won {
		return nil, status.Error(codes.PermissionDenied, "loser")
	}
	if user != bid.Author && user != tender.Author {
		return nil, status.Error(codes.PermissionDenied, "no access")
	}

	res, err := s.executor.Execute(ctx, bid.Program, tender.ProgramInput)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "execution: %v", err)
	}

	// NDA.
	if user != tender.Author {
		res.Output = ""
	}

	return res.ToProto(), nil
}
