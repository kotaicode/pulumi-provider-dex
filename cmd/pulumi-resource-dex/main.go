package main

import (
	"context"
	"log"

	"github.com/kotaicode/pulumi-provider-dex/pkg/provider"
	"github.com/kotaicode/pulumi-provider-dex/pkg/provider/resources"
	"github.com/pulumi/pulumi-go-provider/infer"
)

// providerName is the logical name Pulumi will use for this provider.
const providerName = "dex"

func main() {
	prov, err := infer.NewProviderBuilder().
		WithNamespace("dex").
		WithDisplayName("Dex Provider").
		WithPublisher("Kotaicode GmbH").
		WithKeywords("category/cloud").
		WithDescription("A Pulumi provider for managing Dex resources via the Dex gRPC Admin API").
		WithPluginDownloadURL("github://api.github.com/kotaicode/pulumi-provider-dex").
		WithLanguageMap(map[string]any{
			"go": map[string]any{
				"importBasePath":                 "github.com/kotaicode/pulumi-provider-dex/sdk/go/dex",
				"respectSchemaVersion":           true,
				"generateResourceContainerTypes": true,
			},
			"nodejs": map[string]any{"packageName": "@kotaicode/pulumi-dex", "respectSchemaVersion": true},
			"python": map[string]any{"packageName": "pulumi_kotaicode_dex", "respectSchemaVersion": true},
		}).
		WithRepository("github.com/kotaicode/pulumi-provider-dex").
		WithResources(
			infer.Resource(&resources.Client{}),
			infer.Resource(&resources.Connector{}),
			infer.Resource(&resources.AzureOidcConnector{}),
			infer.Resource(&resources.AzureMicrosoftConnector{}),
			infer.Resource(&resources.CognitoOidcConnector{}),
			infer.Resource(&resources.GitLabConnector{}),
			infer.Resource(&resources.GitHubConnector{}),
			infer.Resource(&resources.GoogleConnector{}),
			infer.Resource(&resources.LocalConnector{}),
		).
		WithConfig(infer.Config(&provider.DexConfig{})).
		Build()
	if err != nil {
		log.Fatalf("failed to build dex provider: %v", err)
	}

	prov.Run(context.Background(), providerName, provider.Version)
}
