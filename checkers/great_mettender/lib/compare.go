package lib

import (
	"github.com/golang/protobuf/proto"
	"github.com/google/go-cmp/cmp"
	"google.golang.org/protobuf/testing/protocmp"
)

func DiffProto(p1, p2 proto.Message) (bool, string) {
	if diff := cmp.Diff(p1, p2, protocmp.Transform()); diff != "" {
		return true, diff
	}
	return false, ""
}
