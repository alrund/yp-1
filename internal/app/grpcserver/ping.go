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
		response.Code = http.StatusInternalServerError
		response.Message = http.StatusText(http.StatusInternalServerError)
		return &response, err
	}

	response.Code = http.StatusOK
	response.Message = http.StatusText(http.StatusOK)
	return &response, nil
}
