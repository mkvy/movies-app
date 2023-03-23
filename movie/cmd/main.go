package main

import (
	"context"
	"fmt"
	"github.com/mkvy/movies-app/gen"
	"github.com/mkvy/movies-app/movie/internal/controller/movie"
	metadatagateway "github.com/mkvy/movies-app/movie/internal/gateway/metadata/http"
	ratinggateway "github.com/mkvy/movies-app/movie/internal/gateway/rating/http"
	grpchandler "github.com/mkvy/movies-app/movie/internal/handler/grpc"
	"github.com/mkvy/movies-app/pkg/discovery"
	"github.com/mkvy/movies-app/pkg/discovery/consul"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"gopkg.in/yaml.v3"
	"log"
	"net"
	"os"
	"time"
)

const serviceName = "movie"

func main() {
	// if not docker image:
	//f, err := os.Open("./metadata/configs/base.yaml")
	f, err := os.Open("base.yaml")
	if err != nil {
		panic(err)
	}
	var cfg config
	if err := yaml.NewDecoder(f).Decode(&cfg); err != nil {
		panic(err)
	}
	port := cfg.API.Port
	log.Printf("Starting the movie service on port %d", port)
	registry, err := consul.NewRegistry("host.docker.internal:8500")
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	instanceID := discovery.GenerateInstanceID(serviceName)
	if err := registry.Register(ctx, instanceID, serviceName, fmt.Sprintf("localhost:%d", port)); err != nil {
		panic(err)
	}
	go func() {
		for {
			if err := registry.ReportHealthyState(instanceID, serviceName); err != nil {
				log.Println("Failed to report healthy state: " + err.Error())
			}
			time.Sleep(1 * time.Second)
		}
	}()
	defer registry.Deregister(ctx, instanceID, serviceName)

	metadataGateway := metadatagateway.New(registry)
	ratingGateway := ratinggateway.New(registry)
	ctrl := movie.New(ratingGateway, metadataGateway)
	h := grpchandler.New(ctrl)
	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer()
	reflection.Register(srv)
	gen.RegisterMovieServiceServer(srv, h)
	if err := srv.Serve(lis); err != nil {
		panic(err)
	}
}
