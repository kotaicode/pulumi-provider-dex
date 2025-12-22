package commands

import (
	"context"
	"fmt"
	"time"

	api "github.com/dexidp/dex/api/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestDeleteDirectCmd tests DeleteClient API with a test client (creates, deletes, verifies).
type TestDeleteDirectCmd struct {
	BaseCmd
}

// Run executes the test-delete-direct command.
func (t *TestDeleteDirectCmd) Run() error {
	host := t.GetHost()
	client, createCtx, cleanup := connectDex(host)
	defer cleanup()

	// Step 1: Create a test client
	testID := "test-delete-direct"
	fmt.Printf("=== Step 1: Creating test client %q ===\n", testID)

	createResp, err := client.CreateClient(createCtx, &api.CreateClientReq{
		Client: &api.Client{
			Id:           testID,
			Name:         "Test Delete Direct",
			RedirectUris: []string{"http://localhost:3000/callback"},
			Public:       false,
		},
	})
	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			fmt.Printf("  Client already exists, continuing...\n")
		} else {
			return fmt.Errorf("failed to create client: %w", err)
		}
	} else {
		fmt.Printf("  ✓ Client created (AlreadyExists: %v)\n", createResp.AlreadyExists)
	}

	// Step 2: Verify client exists
	fmt.Printf("\n=== Step 2: Verifying client exists ===\n")
	listCtx1, cancel1 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel1()
	listResp, err := client.ListClients(listCtx1, &api.ListClientReq{})
	if err != nil {
		return fmt.Errorf("failed to list clients: %w", err)
	}

	found := false
	for _, c := range listResp.Clients {
		if c.Id == testID {
			found = true
			fmt.Printf("  ✓ Found client: %q (Name: %q)\n", c.Id, c.Name)
			break
		}
	}
	if !found {
		return fmt.Errorf("client %q not found after creation", testID)
	}

	// Step 3: Delete the client
	fmt.Printf("\n=== Step 3: Deleting client ===\n")
	deleteCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

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

	// Step 4: Wait a bit
	fmt.Printf("\n=== Step 4: Waiting 200ms ===\n")
	time.Sleep(200 * time.Millisecond)

	// Step 5: Verify deletion
	fmt.Printf("\n=== Step 5: Verifying deletion ===\n")
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

