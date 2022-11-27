package models

import (
	"time"

	tenderspb "great_mettender/pkg/proto/tenders"

	"google.golang.org/protobuf/types/known/durationpb"
)

type ExecutionResult struct {
	Output  string        `json:"output,omitempty"`
	Ops     uint32        `json:"ops,omitempty"`
	Elapsed time.Duration `json:"elapsed,omitempty"`
	Error   string        `json:"error,omitempty"`
}

func (r *ExecutionResult) ToProto() *tenderspb.Bid_ExecuteResponse {
	return &tenderspb.Bid_ExecuteResponse{
		Output:  r.Output,
		Ops:     r.Ops,
		Elapsed: durationpb.New(r.Elapsed),
		Error:   r.Error,
	}
}
