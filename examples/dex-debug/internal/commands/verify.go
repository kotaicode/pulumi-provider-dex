package commands

import (
	"encoding/json"
	"fmt"

	api "github.com/dexidp/dex/api/v2"
)

// VerifyCmd lists all clients and connectors in Dex.
type VerifyCmd struct {
	BaseCmd
}

// Run executes the verify command.
func (v *VerifyCmd) Run() error {
	host := v.GetHost()
	client, gctx, cleanup := connectDex(host)
	defer cleanup()

	// List clients
	fmt.Println("=== Dex Clients ===")
	clientsResp, err := client.ListClients(gctx, &api.ListClientReq{})
	if err != nil {
		return fmt.Errorf("failed to list clients: %w", err)
	}
	for _, c := range clientsResp.Clients {
		fmt.Printf("  ID: %s\n", c.Id)
		fmt.Printf("  Name: %s\n", c.Name)
		fmt.Printf("  Public: %v\n", c.Public)
		fmt.Printf("  RedirectURIs: %v\n", c.RedirectUris)
		fmt.Println()
	}

	// List connectors
	fmt.Println("=== Dex Connectors ===")
	connectorsResp, err := client.ListConnectors(gctx, &api.ListConnectorReq{})
	if err != nil {
		return fmt.Errorf("failed to list connectors: %w", err)
	}
	for _, con := range connectorsResp.Connectors {
		fmt.Printf("  ID: %s\n", con.Id)
		fmt.Printf("  Name: %s\n", con.Name)
		fmt.Printf("  Type: %s\n", con.Type)
		if len(con.Config) > 0 {
			var config map[string]interface{}
			if err := json.Unmarshal(con.Config, &config); err == nil {
				// Hide sensitive fields
				if _, ok := config["clientSecret"]; ok {
					config["clientSecret"] = "***REDACTED***"
				}
				configJSON, _ := json.MarshalIndent(config, "    ", "  ")
				fmt.Printf("  Config:\n%s\n", configJSON)
			}
		}
		fmt.Println()
	}

	return nil
}

