package commands

import (
	"context"
	"fmt"
	"time"

	api "github.com/dexidp/dex/api/v2"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// TestDeleteCmd tests deleting a specific client by ID.
type TestDeleteCmd struct {
	BaseCmd
	ClientID string `arg:"" help:"Client ID to delete"`
}

// Run executes the test-delete command.
func (t *TestDeleteCmd) Run() error {
	host := t.GetHost()
	client, _, cleanup := connectDex(host)
	defer cleanup()

	fmt.Printf("Attempting to delete client: %s\n", t.ClientID)

	deleteCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := client.DeleteClient(deleteCtx, &api.DeleteClientReq{Id: t.ClientID})
	if err != nil {
		st := status.Convert(err)
		fmt.Printf("Error: %v\n", err)
		fmt.Printf("Status code: %v\n", st.Code())
		if st.Code() == codes.NotFound {
			fmt.Println("Client not found (already deleted)")
			return nil
		}
		return fmt.Errorf("error details: %s", st.Message())
	}

	fmt.Println("✓ Successfully deleted")
	return nil
}

