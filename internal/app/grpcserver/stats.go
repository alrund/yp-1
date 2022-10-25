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
		response.Code = http.StatusForbidden
		response.Message = http.StatusText(http.StatusForbidden)
		return &response, nil
	}

	if s.us.TrustedSubnet == nil {
		_, ipnet, err := net.ParseCIDR(cfg.TrustedSubnet)
		if err != nil {
			response.Code = http.StatusInternalServerError
			response.Message = http.StatusText(http.StatusInternalServerError)
			return &response, err
		}
		s.us.TrustedSubnet = ipnet
	}

	if in.XRealIP == "" {
		response.Code = http.StatusForbidden
		response.Message = http.StatusText(http.StatusForbidden)
		return &response, nil
	}

	realIP := net.ParseIP(in.XRealIP)

	if !s.us.TrustedSubnet.Contains(realIP) {
		response.Code = http.StatusForbidden
		response.Message = http.StatusText(http.StatusForbidden)
		return &response, nil
	}

	stat, err := s.us.GetStats()
	if err != nil {
		response.Code = http.StatusInternalServerError
		response.Message = http.StatusText(http.StatusInternalServerError)
		return &response, err
	}

	response.Code = http.StatusOK
	response.Message = http.StatusText(http.StatusOK)
	response.Users = int32(stat.Users)
	response.Urls = int32(stat.Urls)

	return &response, nil
}
