package swissknife

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	pb "github.com/siriusfreak/swiss-knife/internal/pkg/generated/api/swiss-knife"
	"google.golang.org/grpc"
)

func StartGateway(ctx context.Context, grpcAddress, gatewayAddress string) error {
	mux := runtime.NewServeMux()
	opts := []grpc.DialOption{grpc.WithInsecure()}

	withCors := cors.New(cors.Options{
		AllowOriginFunc:  func(origin string) bool { return true },
		AllowedMethods:   []string{"GET", "POST", "PATCH", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"ACCEPT", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}).Handler(mux)

	if err := pb.RegisterSwissKnifeHandlerFromEndpoint(ctx, mux, grpcAddress, opts); err != nil {
		return err
	}

	return http.ListenAndServe(gatewayAddress, withCors)
}
