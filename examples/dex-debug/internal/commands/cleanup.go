package commands

import (
	"fmt"

	api "github.com/dexidp/dex/api/v2"
)

// CleanupCmd cleans up test clients and connectors (excluding static ones).
type CleanupCmd struct {
	BaseCmd
}

// Run executes the cleanup command.
func (c *CleanupCmd) Run() error {
	host := c.GetHost()
	client, gctx, cleanup := connectDex(host)
	defer cleanup()

	// List and delete leftover test clients (excluding static ones)
	fmt.Println("=== Cleaning up leftover test clients ===")
	clientsResp, err := client.ListClients(gctx, &api.ListClientReq{})
	if err != nil {
		return fmt.Errorf("failed to list clients: %w", err)
	}

	for _, cl := range clientsResp.Clients {
		// Skip static clients from config
		if cl.Id == "test-client" {
			continue
		}
		// Delete test clients
		if cl.Id == "my-web-app" || len(cl.Id) > 20 { // Timestamp-based IDs are longer
			fmt.Printf("Deleting client: %s\n", cl.Id)
			_, err := client.DeleteClient(gctx, &api.DeleteClientReq{Id: cl.Id})
			if err != nil {
				fmt.Printf("  Error deleting %s: %v\n", cl.Id, err)
			} else {
				fmt.Printf("  ✓ Deleted %s\n", cl.Id)
			}
		}
	}

	// List and delete leftover test connectors (excluding static ones)
	fmt.Println("\n=== Cleaning up leftover test connectors ===")
	connectorsResp, err := client.ListConnectors(gctx, &api.ListConnectorReq{})
	if err != nil {
		return fmt.Errorf("failed to list connectors: %w", err)
	}

	for _, con := range connectorsResp.Connectors {
		// Skip static connectors from config
		if con.Id == "local" {
			continue
		}
		// Delete test connectors
		if con.Id == "generic-oidc" || len(con.Id) > 20 { // Timestamp-based IDs are longer
			fmt.Printf("Deleting connector: %s\n", con.Id)
			_, err := client.DeleteConnector(gctx, &api.DeleteConnectorReq{Id: con.Id})
			if err != nil {
				fmt.Printf("  Error deleting %s: %v\n", con.Id, err)
			} else {
				fmt.Printf("  ✓ Deleted %s\n", con.Id)
			}
		}
	}

	fmt.Println("\n=== Cleanup complete ===")
	return nil
}

