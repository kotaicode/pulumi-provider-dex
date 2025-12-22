package client

import (
	"context"
	"log"
	"time"

	api "github.com/dexidp/dex/api/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Connect establishes a connection to Dex gRPC server and returns a client, context, and cleanup function.
func Connect(host string) (api.DexClient, context.Context, func()) {
	conn, err := grpc.Dial(host, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("Failed to connect to Dex: %v", err)
	}

	client := api.NewDexClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	cleanup := func() {
		cancel()
		conn.Close()
	}

	return client, ctx, cleanup
}
