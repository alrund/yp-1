package grpcserver

import (
	"context"

	pb "github.com/alrund/yp-1/internal/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Ping checks the database connection.
func (s *Server) Ping(ctx context.Context, in *pb.PingRequest) (*pb.PingResponse, error) {
	var response pb.PingResponse

	if err := s.us.Ping(ctx); err != nil {
		return &response, status.Error(codes.Internal, codes.Internal.String())
	}

	return &response, nil
}
