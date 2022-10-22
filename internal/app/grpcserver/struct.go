package grpcserver

import (
	"context"
	"github.com/alrund/yp-1/internal/app"
	pb "github.com/alrund/yp-1/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"log"
	"net"
)

const bufSize = 1024 * 1024

type TestGenerator struct{}

func (st *TestGenerator) Generate() (string, error) {
	return "qwerty", nil
}

func dialer(us *app.URLShortener) func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(bufSize)

	server := grpc.NewServer()
	pb.RegisterAppServer(server, New(us))

	go func() {
		if err := server.Serve(listener); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()

	return func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}
}
