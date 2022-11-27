package tenders

import (
	"context"
	"time"

	"great_mettender/internal/auth"
	"great_mettender/internal/controllers"
	"great_mettender/internal/models"
	tenderspb "great_mettender/pkg/proto/tenders"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	// maxInputLength is 16kb.
	maxInputLength = 16 * 1024
)

func NewService(
	tendersController *controllers.Tenders,
	bidsController *controllers.Bids,
) *Service {
	return &Service{
		tendersController: tendersController,
		bidsController:    bidsController,
	}
}

type Service struct {
	tenderspb.UnimplementedTendersServiceServer

	tendersController *controllers.Tenders
	bidsController    *controllers.Bids
}

func (s *Service) Create(ctx context.Context, request *tenderspb.Tender_CreateRequest) (*tenderspb.Tender_CreateResponse, error) {
	logrus.Debug("Tenders/Create")

	if request.Tender == nil {
		return nil, status.Error(codes.InvalidArgument, "tender required")
	}

	if len(request.Tender.ProgramInput) > maxInputLength {
		return nil, status.Error(codes.InvalidArgument, "input too long")
	}

	tender := models.NewTenderFromProto(request.Tender)

	tender.ID = uuid.NewString()
	tender.Author = auth.UserFromContext(ctx)
	tender.CreatedAt = time.Now()

	if err := s.tendersController.Add(ctx, tender); err != nil {
		return nil, status.Errorf(codes.Internal, "adding tender: %v", err)
	}

	return &tenderspb.Tender_CreateResponse{Tender: tender.ToProto(true, true)}, nil
}

func (s *Service) Get(ctx context.Context, request *tenderspb.Tender_GetRequest) (*tenderspb.Tender_GetResponse, error) {
	logrus.Debugf("Tenders/Get %v", request)

	tender, err := s.tendersController.Get(ctx, request.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "getting tender: %v", err)
	}

	user := auth.UserFromContext(ctx)
	return &tenderspb.Tender_GetResponse{
		Tender: tender.ToProto(user == tender.Author || user == tender.Winner, user == tender.Author),
	}, nil
}

func (s *Service) Close(ctx context.Context, request *tenderspb.Tender_CloseRequest) (*tenderspb.Tender_CloseResponse, error) {
	logrus.Debugf("Tenders/Close %v", request)

	tender, err := s.tendersController.Get(ctx, request.Id)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "getting tender: %v", err)
	}

	user := auth.UserFromContext(ctx)
	if user != tender.Author {
		return nil, status.Error(codes.PermissionDenied, "only author")
	}

	bids, err := s.bidsController.ListByTender(ctx, tender.ID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "fetching bids: %v", err)
	}

	var bestBid *models.Bid
	for _, bid := range bids {
		if bestBid == nil || bestBid.Price > bid.Price {
			bestBid = bid
		}
	}

	var winningProto *tenderspb.Bid
	if bestBid != nil {
		if err := s.bidsController.SetWon(ctx, bestBid.ID); err != nil {
			return nil, status.Errorf(codes.Internal, "setting bid won: %v", err)
		}
		winningProto = bestBid.ToProto()
	}

	if err := s.tendersController.Finish(ctx, tender.ID); err != nil {
		return nil, status.Errorf(codes.Internal, "setting finished: %v", err)
	}

	return &tenderspb.Tender_CloseResponse{WinningBid: winningProto}, nil
}
