package auth

import (
	"context"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

var whitelistedMethods = []string{
	"/pinger.PingerService/Ping",
	"/tenders.TendersService/List",
}

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		for _, wlm := range whitelistedMethods {
			if wlm == info.FullMethod {
				return handler(ctx, req)
			}
		}
		userIDMeta := metadata.ValueFromIncomingContext(ctx, "user")
		if len(userIDMeta) == 0 {
			return nil, status.Error(codes.Unauthenticated, "user is required in metadata")
		}
		userID := userIDMeta[0]
		if _, err := uuid.Parse(userID); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid user format: %v", err)
		}
		return handler(context.WithValue(ctx, "user", userID), req)
	}
}

func UserFromContext(ctx context.Context) string {
	if val, ok := ctx.Value("user").(string); ok {
		return val
	}
	return ""
}
