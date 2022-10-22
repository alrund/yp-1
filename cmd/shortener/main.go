package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/alrund/yp-1/internal/app"
	"github.com/alrund/yp-1/internal/app/config"
	"github.com/alrund/yp-1/internal/app/encryption"
	"github.com/alrund/yp-1/internal/app/grpcserver"
	"github.com/alrund/yp-1/internal/app/handler"
	"github.com/alrund/yp-1/internal/app/middleware"
	"github.com/alrund/yp-1/internal/app/storage"
	"github.com/alrund/yp-1/internal/app/token/generator"
	pb "github.com/alrund/yp-1/internal/proto"
	"github.com/gorilla/mux"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const defaultBuildValue string = "N/A"

var (
	buildVersion = defaultBuildValue
	buildDate    = defaultBuildValue
	buildCommit  = defaultBuildValue
)

func main() {
	printBuildInfo()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	cfg := config.GetConfig()

	us := &app.URLShortener{
		Config:         cfg,
		Storage:        getStorage(cfg),
		TokenGenerator: generator.NewSimple(),
	}

	closeBothCh := make(chan struct{})
	go func() {
		<-sigCh
		close(closeBothCh)
	}()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		if err := run(us, closeBothCh); err != nil {
			log.Fatalf("HTTP server ListenAndServe: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		if err := runGRPC(us, closeBothCh); err != nil {
			log.Fatalf("GRPC server ListenAndServe: %v", err)
		}
	}()

	wg.Wait()
}

func printBuildInfo() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}

func run(us *app.URLShortener, closeBothCh chan struct{}) error {
	httpShutdownCh := make(chan struct{})

	cfg := us.Config

	server := &http.Server{
		Addr:              cfg.ServerAddress,
		Handler:           getRouter(us, cfg),
		ReadHeaderTimeout: 1 * time.Second,
	}

	go func() {
		<-closeBothCh

		if err := server.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP server Shutdown: %v", err)
		}

		close(httpShutdownCh)
	}()

	log.Println("Starting HTTP server", cfg.ServerAddress)

	var err error
	if cfg.EnableHTTPS && cfg.CertFile != "" && cfg.KeyFile != "" {
		err = server.ListenAndServeTLS(cfg.CertFile, cfg.KeyFile)
	} else {
		err = server.ListenAndServe()
	}

	if err != http.ErrServerClosed {
		return err
	}

	<-httpShutdownCh
	fmt.Println("HTTP Server Shutdown gracefully")

	return nil
}

func runGRPC(us *app.URLShortener, closeBothCh chan struct{}) error {
	var (
		err        error
		serverGRPC *grpc.Server
		cfg        = us.Config
	)

	if cfg.EnableHTTPS && cfg.CertFile != "" && cfg.KeyFile != "" {
		creds, err := credentials.NewServerTLSFromFile(cfg.CertFile, cfg.KeyFile)
		if err != nil {
			return err
		}
		serverGRPC = grpc.NewServer(grpc.Creds(creds))
	} else {
		serverGRPC = grpc.NewServer()
	}

	pb.RegisterAppServer(serverGRPC, grpcserver.New(us))

	grpcShutdownCh := make(chan struct{})

	go func() {
		<-closeBothCh

		serverGRPC.GracefulStop()

		close(grpcShutdownCh)
	}()

	log.Println("Starting GRPC server", cfg.GrpcServerAddress)

	listen, err := net.Listen("tcp", cfg.GrpcServerAddress)
	if err != nil {
		return err
	}

	err = serverGRPC.Serve(listen)
	if err != nil {
		return err
	}

	<-grpcShutdownCh
	fmt.Println("GRPC server Shutdown gracefully")

	return nil
}

func getStorage(cfg *config.Config) app.Storage {
	var (
		err  error
		strg app.Storage = storage.NewMap()
	)

	if cfg.FileStoragePath != "" {
		strg, err = storage.NewFile(cfg.FileStoragePath)
		if err != nil {
			log.Fatal(err)
		}
	}
	if cfg.DatabaseDsn != "" {
		strg, err = storage.NewDB(cfg.DatabaseDsn)
		if err != nil {
			log.Fatal(err)
		}
	}

	return strg
}

func getRouter(us *app.URLShortener, cfg *config.Config) *mux.Router {
	r := mux.NewRouter()

	r.HandleFunc("/", handler.Add(us)).Methods(http.MethodPost)
	r.HandleFunc("/api/shorten", handler.AddJSON(us)).Methods(http.MethodPost)
	r.HandleFunc("/api/shorten/batch", handler.AddBatchJSON(us)).Methods(http.MethodPost)
	r.HandleFunc("/ping", handler.Ping(us)).Methods(http.MethodGet)
	r.HandleFunc("/{id}", handler.Get(us)).Methods(http.MethodGet)
	r.HandleFunc("/api/user/urls", handler.GetUserURLs(us)).Methods(http.MethodGet)
	r.HandleFunc("/api/user/urls", handler.DeleteURLs(us)).Methods(http.MethodDelete)
	r.HandleFunc("/api/internal/stats", handler.Stats(us)).Methods(http.MethodGet)

	subRouter := r.PathPrefix("/debug/pprof").Subrouter()
	subRouter.HandleFunc("/", pprof.Index)
	subRouter.HandleFunc("/cmdline", pprof.Cmdline)
	subRouter.HandleFunc("/profile", pprof.Profile)
	subRouter.HandleFunc("/symbol", pprof.Symbol)
	subRouter.HandleFunc("/trace", pprof.Trace)
	subRouter.HandleFunc("/{name}", pprof.Index)

	r.Use(middleware.Compress)
	r.Use(middleware.Decompress)
	r.Use(middleware.Auth(encryption.NewEncryption(cfg.CipherPass)))

	return r
}
