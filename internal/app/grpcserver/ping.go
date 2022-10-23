package grpcserver

import (
	"context"
	"net/http"

	pb "github.com/alrund/yp-1/internal/proto"
)

// Ping checks the database connection.
func (s *Server) Ping(ctx context.Context, in *pb.PingRequest) (*pb.PingResponse, error) {
	var response pb.PingResponse

	if err := s.us.Ping(ctx); err != nil {
		response.ErrorCode = http.StatusInternalServerError
		response.Error = http.StatusText(http.StatusInternalServerError)
		return &response, err
	}

	response.ErrorCode = http.StatusOK
	response.Error = http.StatusText(http.StatusOK)
	return &response, nil
}
