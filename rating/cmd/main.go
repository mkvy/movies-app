package main

import (
	"context"
	"fmt"
	"github.com/mkvy/movies-app/gen"
	"github.com/mkvy/movies-app/pkg/discovery"
	"github.com/mkvy/movies-app/pkg/discovery/consul"
	"github.com/mkvy/movies-app/pkg/tracing"
	"github.com/mkvy/movies-app/rating/internal/controller/rating"
	grpchandler "github.com/mkvy/movies-app/rating/internal/handler/grpc"
	"github.com/mkvy/movies-app/rating/internal/repository/mysql"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v3"
	"net"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const serviceName = "rating"
const registryConsulAddr = "localhost:8500"

func main() {
	logger, _ := zap.NewProduction()
	// if not docker image:
	f, err := os.Open("./rating/configs/base.yaml")
	//if docker:
	//f, err := os.Open("base.yaml")
	if err != nil {
		logger.Fatal("Failed to open configuration", zap.Error(err))
	}
	var cfg config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		logger.Fatal("Failed to parse configuration", zap.Error(err))
	}
	port := cfg.API.Port
	logger.Info("Starting the rating service", zap.Int("port", port))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	tp, err := tracing.NewJaegerProvider(cfg.Jaeger.URL, serviceName)
	if err != nil {
		logger.Fatal("Failed to initialize Jaeger provider", zap.Error(err))
	}
	defer func() {
		if err := tp.Shutdown(ctx); err != nil {
			logger.Fatal("Failed to shutdown Jaeger provider", zap.Error(err))
		}
	}()
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	registry, err := consul.NewRegistry(registryConsulAddr)
	if err != nil {
		panic(err)
	}

	instanceID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("localhost:%d", port)); err != nil {
		panic(err)
	}
	go func() {
		for {
			if err := registry.ReportHealthyState(instanceID, serviceName); err != nil {
				logger.Error("Failed to report healthy state", zap.Error(err))
			}
			time.Sleep(1 * time.Second)
		}
	}()
	defer registry.Deregister(ctx, instanceID, serviceName)
	repo, err := mysql.New()
	if err != nil {
		logger.Fatal("Error while initializing repository", zap.Error(err))
	}
	ctrl := rating.New(repo, nil)
	h := grpchandler.New(ctrl)
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%v", port))
	if err != nil {
		logger.Fatal("Failed to listen", zap.Error(err))
	}
	srv := grpc.NewServer(grpc.UnaryInterceptor(otelgrpc.UnaryServerInterceptor()))
	reflection.Register(srv)
	gen.RegisterRatingServiceServer(srv, h)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		_ = <-sigChan
		cancel()
		logger.Info("Recieved signal, attempting graceful shutdown")
		srv.GracefulStop()
		logger.Info("Gracefully stopped the gRPC server")
	}()
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
	wg.Wait()
}
