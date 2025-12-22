package commands

import (
	"context"
	"fmt"
	"time"

	api "github.com/dexidp/dex/api/v2"
)

// TestVerificationCmd tests delete verification logic with "my-web-app" client.
type TestVerificationCmd struct {
	BaseCmd
}

// Run executes the test-verification command.
func (t *TestVerificationCmd) Run() error {
	host := t.GetHost()
	client, _, cleanup := connectDex(host)
	defer cleanup()

	// Test: Delete a client, then immediately check if it still exists
	testID := "my-web-app"
	fmt.Printf("Step 1: Deleting client: %s\n", testID)

	deleteCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.DeleteClient(deleteCtx, &api.DeleteClientReq{Id: testID})
	if err != nil {
		return fmt.Errorf("delete failed: %w", err)
	}
	fmt.Println("  ✓ DeleteClient returned success")

	// Wait a bit
	time.Sleep(100 * time.Millisecond)

	// Step 2: List clients to see if it still exists
	fmt.Println("\nStep 2: Listing clients to verify deletion...")
	listCtx, listCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer listCancel()
	listResp, err := client.ListClients(listCtx, &api.ListClientReq{})
	if err != nil {
		return fmt.Errorf("failed to list clients: %w", err)
	}

	found := false
	for _, c := range listResp.Clients {
		if c.Id == testID {
			found = true
			fmt.Printf("  ✗ Client %s still exists!\n", testID)
			break
		}
	}

	if !found {
		fmt.Printf("  ✓ Client %s successfully deleted\n", testID)
		return nil
	}

	return fmt.Errorf("delete reported success but client still exists")
}

