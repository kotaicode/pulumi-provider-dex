package commands

import (
	"context"
	"fmt"
	"time"

	api "github.com/dexidp/dex/api/v2"
	"google.golang.org/grpc/status"
)

// TestDeleteMyWebAppCmd tests DeleteClient API with "my-web-app" client.
type TestDeleteMyWebAppCmd struct {
	BaseCmd
}

// Run executes the test-delete-my-web-app command.
func (t *TestDeleteMyWebAppCmd) Run() error {
	host := t.GetHost()
	client, listCtx, cleanup := connectDex(host)
	defer cleanup()

	// Test with the exact client ID that Pulumi uses
	testID := "my-web-app"
	fmt.Printf("=== Testing DeleteClient with ID: %q ===\n", testID)

	// Step 1: Check if client exists
	fmt.Printf("\n=== Step 1: Checking if client exists ===\n")
	listResp, err := client.ListClients(listCtx, &api.ListClientReq{})
	if err != nil {
		return fmt.Errorf("failed to list clients: %w", err)
	}

	found := false
	for _, c := range listResp.Clients {
		if c.Id == testID {
			found = true
			fmt.Printf("  ✓ Found client: %q (Name: %q)\n", c.Id, c.Name)
			fmt.Printf("    RedirectURIs: %v\n", c.RedirectUris)
			fmt.Printf("    Public: %v\n", c.Public)
			break
		}
	}
	if !found {
		fmt.Printf("  ✗ Client %q not found - creating it first...\n", testID)
		createResp, err := client.CreateClient(listCtx, &api.CreateClientReq{
			Client: &api.Client{
				Id:           testID,
				Name:         "My Web App",
				RedirectUris: []string{"http://localhost:3000/callback"},
				Public:       false,
			},
		})
		if err != nil {
			return fmt.Errorf("failed to create client: %w", err)
		}
		fmt.Printf("  ✓ Client created (AlreadyExists: %v)\n", createResp.AlreadyExists)
	}

	// Step 2: Delete the client
	fmt.Printf("\n=== Step 2: Deleting client ===\n")
	deleteCtx, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()

	deleteResp, err := client.DeleteClient(deleteCtx, &api.DeleteClientReq{
		Id: testID,
	})
	if err != nil {
		st := status.Convert(err)
		fmt.Printf("  ✗ DeleteClient returned error:\n")
		fmt.Printf("    Code: %v\n", st.Code())
		fmt.Printf("    Message: %s\n", st.Message())
		return fmt.Errorf("delete failed: %w", err)
	}
	fmt.Printf("  ✓ DeleteClient returned success\n")
	if deleteResp != nil {
		fmt.Printf("    Response: %+v\n", deleteResp)
	}

	// Step 3: Wait a bit
	fmt.Printf("\n=== Step 3: Waiting 200ms ===\n")
	time.Sleep(200 * time.Millisecond)

	// Step 4: Verify deletion
	fmt.Printf("\n=== Step 4: Verifying deletion ===\n")
	verifyCtx, cancel3 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel3()

	listResp2, err := client.ListClients(verifyCtx, &api.ListClientReq{})
	if err != nil {
		return fmt.Errorf("failed to list clients for verification: %w", err)
	}

	fmt.Printf("  Total clients in Dex: %d\n", len(listResp2.Clients))
	fmt.Printf("  Client IDs: ")
	for i, c := range listResp2.Clients {
		if i > 0 {
			fmt.Printf(", ")
		}
		fmt.Printf("%q", c.Id)
	}
	fmt.Printf("\n")

	foundAfterDelete := false
	for _, c := range listResp2.Clients {
		if c.Id == testID {
			foundAfterDelete = true
			fmt.Printf("  ✗ VERIFICATION FAILED: Client %q still exists!\n", testID)
			break
		}
	}

	if !foundAfterDelete {
		fmt.Printf("  ✓ VERIFICATION SUCCESS: Client %q successfully deleted\n", testID)
		return nil
	}

	return fmt.Errorf("delete reported success but client still exists")
}

