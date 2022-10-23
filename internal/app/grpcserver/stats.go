package grpcserver

import (
	"context"
	"net"
	"net/http"

	pb "github.com/alrund/yp-1/internal/proto"
)

// Stats get statistic information.
func (s *Server) Stats(ctx context.Context, in *pb.StatsRequest) (*pb.StatsResponse, error) {
	var response pb.StatsResponse

	cfg := s.us.GetConfig()
	if cfg.TrustedSubnet == "" {
		response.ErrorCode = http.StatusForbidden
		response.Error = http.StatusText(http.StatusForbidden)
		return &response, nil
	}

	_, ipnet, err := net.ParseCIDR(cfg.TrustedSubnet)
	if err != nil {
		response.ErrorCode = http.StatusInternalServerError
		response.Error = http.StatusText(http.StatusInternalServerError)
		return &response, err
	}

	if in.XRealIP == "" {
		response.ErrorCode = http.StatusForbidden
		response.Error = http.StatusText(http.StatusForbidden)
		return &response, nil
	}

	realIP := net.ParseIP(in.XRealIP)

	if !ipnet.Contains(realIP) {
		response.ErrorCode = http.StatusForbidden
		response.Error = http.StatusText(http.StatusForbidden)
		return &response, nil
	}

	stat, err := s.us.GetStats()
	if err != nil {
		response.ErrorCode = http.StatusInternalServerError
		response.Error = http.StatusText(http.StatusInternalServerError)
		return &response, err
	}

	response.ErrorCode = http.StatusOK
	response.Error = http.StatusText(http.StatusOK)
	response.Users = int32(stat.Users)
	response.Urls = int32(stat.Urls)

	return &response, nil
}
