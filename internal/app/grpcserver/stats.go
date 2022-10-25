package grpcserver

import (
	"context"
	"net"

	pb "github.com/alrund/yp-1/internal/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Stats get statistic information.
func (s *Server) Stats(ctx context.Context, in *pb.StatsRequest) (*pb.StatsResponse, error) {
	var response pb.StatsResponse

	cfg := s.us.GetConfig()
	if cfg.TrustedSubnet == "" {
		return &response, status.Error(codes.PermissionDenied, codes.PermissionDenied.String())
	}

	if s.us.TrustedSubnet == nil {
		_, ipnet, err := net.ParseCIDR(cfg.TrustedSubnet)
		if err != nil {
			return &response, status.Error(codes.Internal, codes.Internal.String())
		}
		s.us.TrustedSubnet = ipnet
	}

	if in.XRealIP == "" {
		return &response, status.Error(codes.PermissionDenied, codes.PermissionDenied.String())
	}

	realIP := net.ParseIP(in.XRealIP)

	if !s.us.TrustedSubnet.Contains(realIP) {
		return &response, status.Error(codes.PermissionDenied, codes.PermissionDenied.String())
	}

	stat, err := s.us.GetStats()
	if err != nil {
		return &response, status.Error(codes.PermissionDenied, err.Error())
	}

	response.Users = int32(stat.Users)
	response.Urls = int32(stat.Urls)

	return &response, nil
}
