package grpcserver

import (
	"context"

	"github.com/alrund/yp-1/internal/app/encryption"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// AuthInterceptor authenticates the user.
func AuthInterceptor(enc *encryption.Encryption) func(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		var (
			err    error
			userID string
		)

		md, ok := metadata.FromIncomingContext(ctx)
		if ok {
			values := md.Get(string(UserIDContextKey))
			if len(values) > 0 {
				userID, err = enc.Decrypt(values[0])
				if err != nil {
					return nil, err
				}
			}
		}

		if userID == "" {
			userID = uuid.New().String()
		}

		ctx = context.WithValue(ctx, UserIDContextKey, userID)

		return handler(ctx, req)
	}
}
