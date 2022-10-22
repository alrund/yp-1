package grpcserver

import (
	"github.com/alrund/yp-1/internal/app"
	pb "github.com/alrund/yp-1/internal/proto"
)

type Server struct {
	pb.UnimplementedAppServer
	us *app.URLShortener
}

func New(us *app.URLShortener) *Server {
	return &Server{us: us}
}
