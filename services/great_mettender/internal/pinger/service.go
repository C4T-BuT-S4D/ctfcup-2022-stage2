package pinger

import (
	"context"

	pingerpb "great_mettender/pkg/proto/pinger"

	"github.com/sirupsen/logrus"
)

type Service struct {
	pingerpb.UnimplementedPingerServiceServer
}

func (s *Service) Ping(context.Context, *pingerpb.PingRequest) (*pingerpb.PingResponse, error) {
	logrus.Debug("Pinger/Ping")
	return &pingerpb.PingResponse{}, nil
}
