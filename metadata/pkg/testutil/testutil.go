package testutil

import (
	"github.com/mkvy/movies-app/gen"
	"github.com/mkvy/movies-app/metadata/internal/controller/metadata"
	grpchandler "github.com/mkvy/movies-app/metadata/internal/handler/grpc"
	"github.com/mkvy/movies-app/metadata/internal/repository/memory"
)

// NewTestMetadataGRPCServer creates a new metadata gRPC server to be used in tests.
func NewTestMetadataGRPCServer() gen.MetadataServiceServer {
	r := memory.New()
	ctrl := metadata.New(r)
	return grpchandler.New(ctrl)
}
