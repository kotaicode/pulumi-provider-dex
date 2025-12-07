package main

import (
	"context"
	"fmt"
	"log"
	"time"

	api "github.com/dexidp/dex/api/v2"
	"github.com/pulumi/pulumi-go-provider/infer"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// providerName is the logical name Pulumi will use for this provider.
const providerName = "dex"

func main() {
	prov, err := infer.NewProviderBuilder().
		WithResources(
			infer.Resource(&Client{}),
			infer.Resource(&Connector{}),
			infer.Resource(&AzureOidcConnector{}),
			infer.Resource(&AzureMicrosoftConnector{}),
			infer.Resource(&CognitoOidcConnector{}),
		).
		WithConfig(infer.Config(&DexConfig{})).
		Build()
	if err != nil {
		log.Fatalf("failed to build dex provider: %v", err)
	}

	prov.Run(context.Background(), providerName, "0.1.0")
}

// DexConfig describes provider-level configuration (connection to Dex gRPC).
// This struct doubles as the configured client object passed to resources.
// Note: Environment variables can be set but are not automatically read by the provider.
// Users should set them in their Pulumi program or use Pulumi config.
type DexConfig struct {
	Host            string  `pulumi:"host"`
	CACertPEM       *string `pulumi:"caCert,optional" provider:"secret"`
	ClientCertPEM   *string `pulumi:"clientCert,optional" provider:"secret"`
	ClientKeyPEM    *string `pulumi:"clientKey,optional" provider:"secret"`
	InsecureSkipTLS *bool   `pulumi:"insecureSkipVerify,optional"`
	TimeoutSeconds  *int    `pulumi:"timeoutSeconds,optional"`

	// internal fields are not exposed in schema and are used at runtime only.
	client api.DexClient
}

// Annotate config fields with descriptions & defaults for the schema.
func (c *DexConfig) Annotate(a infer.Annotator) {
	a.Describe(&c.Host, "Dex gRPC host:port, e.g. dex.internal.kotaicode:5557.")
	a.Describe(&c.CACertPEM, "PEM-encoded CA certificate for validating Dex's TLS certificate.")
	a.Describe(&c.ClientCertPEM, "PEM-encoded client certificate for mTLS to Dex.")
	a.Describe(&c.ClientKeyPEM, "PEM-encoded private key for the client certificate.")
	a.Describe(&c.InsecureSkipTLS, "If true, disables TLS verification (development only).")
	a.Describe(&c.TimeoutSeconds, "Per-RPC timeout in seconds when talking to Dex.")
}

// Configure is called once per provider instance to establish a Dex gRPC client.
// It satisfies infer.CustomConfigure via pointer receiver.
func (c *DexConfig) Configure(ctx context.Context) error {
	if c.Host == "" {
		return fmt.Errorf("host is required")
	}

	// TODO: Optionally make Configure preview-safe by checking runInfo.Preview
	// For now, we'll let Configure connect to Dex even in preview mode.
	// The Create/Update methods will short-circuit based on req.DryRun before making API calls.
	// TODO: implement proper TLS (mTLS) based on CACertPEM / ClientCertPEM / ClientKeyPEM.
	// For now, use insecure transport so we can quickly iterate against a dev Dex.
	dialCtx, cancel := context.WithTimeout(ctx, time.Duration(ptrOr(c.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		dialCtx,
		c.Host,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return fmt.Errorf("failed to connect to Dex at %s: %w", c.Host, err)
	}

	c.client = api.NewDexClient(conn)

	// TODO: optionally perform a lightweight health check call here to fail fast.
	return nil
}

// ptrOr returns the value pointed to by p, or def if p is nil.
func ptrOr[T any](p *T, def T) T {
	if p == nil {
		return def
	}
	return *p
}
