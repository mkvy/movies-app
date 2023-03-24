package testutil

import (
	"github.com/mkvy/movies-app/gen"
	"github.com/mkvy/movies-app/movie/internal/controller/movie"
	metadatagateway "github.com/mkvy/movies-app/movie/internal/gateway/metadata/grpc"
	ratinggateway "github.com/mkvy/movies-app/movie/internal/gateway/rating/grpc"
	grpchandler "github.com/mkvy/movies-app/movie/internal/handler/grpc"
	"github.com/mkvy/movies-app/pkg/discovery"
)

// NewTestMovieGRPCServer creates a new movie gRPC server to be used in tests.
func NewTestMovieGRPCServer(registry discovery.Registry) gen.MovieServiceServer {
	metadataGateway := metadatagateway.New(registry)
	ratingGateway := ratinggateway.New(registry)
	ctrl := movie.New(ratingGateway, metadataGateway)
	return grpchandler.New(ctrl)
}
