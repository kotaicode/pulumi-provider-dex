package commands

import (
	"context"

	api "github.com/dexidp/dex/api/v2"
	"github.com/kotaicode/pulumi-provider-dex/examples/dex-debug/internal/client"
)

// connectDex is a helper function that connects to Dex using the client package.
func connectDex(host string) (api.DexClient, context.Context, func()) {
	return client.Connect(host)
}

