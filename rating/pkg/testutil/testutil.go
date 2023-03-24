package testutil

import (
	"github.com/mkvy/movies-app/gen"
	"github.com/mkvy/movies-app/rating/internal/controller/rating"
	grpchandler "github.com/mkvy/movies-app/rating/internal/handler/grpc"
	"github.com/mkvy/movies-app/rating/internal/repository/memory"
)

// NewTestRatingGRPCServer creates a new rating gRPC server to be used in tests.
func NewTestRatingGRPCServer() gen.RatingServiceServer {
	r := memory.New()
	ctrl := rating.New(r, nil)
	return grpchandler.New(ctrl)
}
