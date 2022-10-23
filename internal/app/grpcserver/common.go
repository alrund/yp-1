package grpcserver

import (
	"context"
	"log"
	"net"

	"github.com/alrund/yp-1/internal/app"
	"github.com/alrund/yp-1/internal/app/encryption"
	pb "github.com/alrund/yp-1/internal/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/test/bufconn"
)

type ContextKey string

const (
	UserIDContextKey ContextKey = "userID"
	bufSize                     = 1024 * 1024
)

type TestGenerator struct{}

func (st *TestGenerator) Generate() (string, error) {
	return "qwerty", nil
}

func dialer(us *app.URLShortener) func(context.Context, string) (net.Conn, error) {
	listener := bufconn.Listen(bufSize)

	enc := encryption.NewEncryption(us.Config.CipherPass)
	server := grpc.NewServer(grpc.UnaryInterceptor(AuthInterceptor(enc)))
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

func getContextWithUserID(userID string, enc *encryption.Encryption) context.Context {
	encrypted, err := enc.Encrypt(userID)
	if err != nil {
		log.Fatal(err)
	}

	md := metadata.Pairs(string(UserIDContextKey), encrypted)
	return metadata.NewOutgoingContext(context.Background(), md)
}
