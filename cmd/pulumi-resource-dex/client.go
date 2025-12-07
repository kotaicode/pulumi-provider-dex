package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	api "github.com/dexidp/dex/api/v2"
	"github.com/pulumi/pulumi-go-provider/infer"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ClientArgs defines the inputs for a dex.Client resource.
type ClientArgs struct {
	ClientId     string   `pulumi:"clientId"`
	Name         string   `pulumi:"name"`
	Secret       *string  `pulumi:"secret,optional" provider:"secret"`
	RedirectUris []string `pulumi:"redirectUris"`
	TrustedPeers []string `pulumi:"trustedPeers,optional"`
	Public       *bool    `pulumi:"public,optional"`
	LogoUrl      *string  `pulumi:"logoUrl,optional"`
}

// ClientState defines the outputs/state for a dex.Client resource.
type ClientState struct {
	ClientArgs
	CreatedAt *string `pulumi:"createdAt,optional"`
}

// Client represents a Dex OAuth2 client resource.
type Client struct{}

// Annotate provides schema metadata for the Client resource.
func (c *Client) Annotate(a infer.Annotator) {
	a.Describe(c, "Manages an OAuth2 client in Dex.")
}

// Create creates a new OAuth2 client in Dex.
func (c *Client) Create(ctx context.Context, req infer.CreateRequest[ClientArgs]) (infer.CreateResponse[ClientState], error) {
	args := req.Inputs

	// In preview/dry-run mode, skip actual Dex API calls and return expected state
	// This check MUST be first, before any other operations or config checks
	if req.DryRun {
		// For preview, we just mirror the inputs into state and do NOT call Dex.
		state := ClientState{
			ClientArgs: args,
		}
		return infer.CreateResponse[ClientState]{
			ID:     args.ClientId,
			Output: state,
		}, nil
	}

	cfg := infer.GetConfig[DexConfig](ctx)
	if cfg.client == nil {
		return infer.CreateResponse[ClientState]{}, fmt.Errorf("Dex client not configured")
	}

	// Generate secret if not provided
	secret := ""
	if args.Secret != nil && *args.Secret != "" {
		secret = *args.Secret
	} else {
		// Generate a secure random secret (32 bytes = 256 bits, base64 encoded)
		secretBytes := make([]byte, 32)
		if _, err := rand.Read(secretBytes); err != nil {
			return infer.CreateResponse[ClientState]{}, wrapError("create", "client", args.ClientId, fmt.Errorf("failed to generate secret: %w", err))
		}
		secret = base64.URLEncoding.EncodeToString(secretBytes)
	}

	// Build the Dex Client message
	client := &api.Client{
		Id:           args.ClientId,
		Secret:       secret,
		RedirectUris: args.RedirectUris,
		TrustedPeers: args.TrustedPeers,
		Name:         args.Name,
		LogoUrl:      ptrOr(args.LogoUrl, ""),
	}
	if args.Public != nil {
		client.Public = *args.Public
	}

	// Call Dex CreateClient
	createCtx, cancel := context.WithTimeout(ctx, time.Duration(ptrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	resp, err := cfg.client.CreateClient(createCtx, &api.CreateClientReq{
		Client: client,
	})
	if err != nil {
		return infer.CreateResponse[ClientState]{}, wrapError("create", "client", args.ClientId, err)
	}

	if resp.AlreadyExists {
		// Resource already exists - read it and return it so Pulumi can track it
		// This allows destroy to work properly even if the resource was created outside Pulumi
		readCtx, cancel := context.WithTimeout(ctx, time.Duration(ptrOr(cfg.TimeoutSeconds, 5))*time.Second)
		defer cancel()

		getResp, err := cfg.client.GetClient(readCtx, &api.GetClientReq{
			Id: args.ClientId,
		})
		if err != nil {
			return infer.CreateResponse[ClientState]{}, wrapError("read existing", "client", args.ClientId, err)
		}

		// Build state from existing client
		state := ClientState{
			ClientArgs: ClientArgs{
				ClientId:     getResp.Client.Id,
				Name:         getResp.Client.Name,
				Secret:       &getResp.Client.Secret,
				RedirectUris: getResp.Client.RedirectUris,
				TrustedPeers: getResp.Client.TrustedPeers,
				Public:       &getResp.Client.Public,
				LogoUrl:      &getResp.Client.LogoUrl,
			},
		}

		return infer.CreateResponse[ClientState]{
			ID:     args.ClientId,
			Output: state,
		}, nil
	}

	// Build the state
	now := time.Now().Format(time.RFC3339)
	state := ClientState{
		ClientArgs: ClientArgs{
			ClientId:     args.ClientId,
			Name:         args.Name,
			Secret:       &secret,
			RedirectUris: args.RedirectUris,
			TrustedPeers: args.TrustedPeers,
			Public:       args.Public,
			LogoUrl:      args.LogoUrl,
		},
		CreatedAt: &now,
	}

	return infer.CreateResponse[ClientState]{
		ID:     args.ClientId,
		Output: state,
	}, nil
}

// Read retrieves an existing OAuth2 client from Dex.
func (c *Client) Read(ctx context.Context, req infer.ReadRequest[ClientArgs, ClientState]) (infer.ReadResponse[ClientArgs, ClientState], error) {
	cfg := infer.GetConfig[DexConfig](ctx)
	if cfg.client == nil {
		return infer.ReadResponse[ClientArgs, ClientState]{}, fmt.Errorf("Dex client not configured")
	}

	// Call Dex GetClient
	getCtx, cancel := context.WithTimeout(ctx, time.Duration(ptrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	resp, err := cfg.client.GetClient(getCtx, &api.GetClientReq{
		Id: req.ID,
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			// Resource doesn't exist - return empty response to indicate deletion
			return infer.ReadResponse[ClientArgs, ClientState]{}, nil
		}
		return infer.ReadResponse[ClientArgs, ClientState]{}, fmt.Errorf("failed to get Dex client: %w", err)
	}

	if resp.Client == nil {
		return infer.ReadResponse[ClientArgs, ClientState]{}, nil
	}

	client := resp.Client

	// Build the state from Dex response
	state := ClientState{
		ClientArgs: ClientArgs{
			ClientId:     client.Id,
			Name:         client.Name,
			Secret:       &client.Secret,
			RedirectUris: client.RedirectUris,
			TrustedPeers: client.TrustedPeers,
			Public:       &client.Public,
			LogoUrl:      ptrOrString(client.LogoUrl),
		},
		// Note: Dex API doesn't expose createdAt, so we keep the existing value if present
		CreatedAt: req.State.CreatedAt,
	}

	// Build inputs from the state (for normalization)
	inputs := ClientArgs{
		ClientId:     state.ClientId,
		Name:         state.Name,
		Secret:       state.Secret,
		RedirectUris: state.RedirectUris,
		TrustedPeers: state.TrustedPeers,
		Public:       state.Public,
		LogoUrl:      state.LogoUrl,
	}

	return infer.ReadResponse[ClientArgs, ClientState]{
		ID:     client.Id,
		Inputs: inputs,
		State:  state,
	}, nil
}

// Update updates an existing OAuth2 client in Dex.
func (c *Client) Update(ctx context.Context, req infer.UpdateRequest[ClientArgs, ClientState]) (infer.UpdateResponse[ClientState], error) {
	args := req.Inputs
	oldState := req.State

	// In preview/dry-run mode, skip actual Dex API calls and return expected state
	// This check MUST be first, before any other operations or config checks
	if req.DryRun {
		state := ClientState{
			ClientArgs: args,
		}
		return infer.UpdateResponse[ClientState]{
			Output: state,
		}, nil
	}

	cfg := infer.GetConfig[DexConfig](ctx)
	if cfg.client == nil {
		return infer.UpdateResponse[ClientState]{}, fmt.Errorf("Dex client not configured")
	}

	// Validate that clientId hasn't changed (it's immutable)
	if args.ClientId != oldState.ClientId {
		return infer.UpdateResponse[ClientState]{}, fmt.Errorf("clientId cannot be changed (was %q, got %q)", oldState.ClientId, args.ClientId)
	}

	// Build the update request
	// Note: UpdateClientReq doesn't support Secret or Public changes - these are immutable
	updateReq := &api.UpdateClientReq{
		Id:           args.ClientId,
		Name:         args.Name,
		RedirectUris: args.RedirectUris,
		TrustedPeers: args.TrustedPeers,
		LogoUrl:      ptrOr(args.LogoUrl, ""),
	}

	// Call Dex UpdateClient
	updateCtx, cancel := context.WithTimeout(ctx, time.Duration(ptrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	_, err := cfg.client.UpdateClient(updateCtx, updateReq)
	if err != nil {
		return infer.UpdateResponse[ClientState]{}, fmt.Errorf("failed to update Dex client: %w", err)
	}

	// Build the updated state
	// Keep the existing secret since it can't be updated via UpdateClient
	state := ClientState{
		ClientArgs: ClientArgs{
			ClientId:     args.ClientId,
			Name:         args.Name,
			Secret:       oldState.Secret, // Keep existing secret
			RedirectUris: args.RedirectUris,
			TrustedPeers: args.TrustedPeers,
			Public:       args.Public,
			LogoUrl:      args.LogoUrl,
		},
		CreatedAt: oldState.CreatedAt, // Preserve createdAt
	}

	return infer.UpdateResponse[ClientState]{
		Output: state,
	}, nil
}

// Delete deletes an OAuth2 client from Dex.
func (c *Client) Delete(ctx context.Context, req infer.DeleteRequest[ClientState]) (infer.DeleteResponse, error) {
	cfg := infer.GetConfig[DexConfig](ctx)
	if cfg.client == nil {
		return infer.DeleteResponse{}, fmt.Errorf("Dex client not configured")
	}

	// Use the ID from the request, or fall back to the state if available
	deleteID := req.ID
	if deleteID == "" && req.State.ClientId != "" {
		deleteID = req.State.ClientId
	}
	if deleteID == "" {
		return infer.DeleteResponse{}, fmt.Errorf("cannot delete client: no ID provided in request or state (req.ID=%q, req.State.ClientId=%q)", req.ID, req.State.ClientId)
	}

	// Note: Pulumi does not call Delete during preview, so no preview check needed

	// Call Dex DeleteClient
	deleteCtx, cancel := context.WithTimeout(ctx, time.Duration(ptrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	_, err := cfg.client.DeleteClient(deleteCtx, &api.DeleteClientReq{
		Id: deleteID,
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			// Already deleted, treat as success
			return infer.DeleteResponse{}, nil
		}
		return infer.DeleteResponse{}, fmt.Errorf("failed to delete Dex client %q: %w", deleteID, err)
	}

	// Verify the delete actually happened by checking if the client still exists
	// This helps catch cases where DeleteClient returns success but doesn't actually delete
	// Use ListClients instead of GetClient for more reliable verification
	// Add a small delay to allow Dex to process the delete
	time.Sleep(200 * time.Millisecond)

	listCtx, listCancel := context.WithTimeout(ctx, time.Duration(ptrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer listCancel()

	listResp, listErr := cfg.client.ListClients(listCtx, &api.ListClientReq{})
	if listErr != nil {
		// Can't verify, but delete call succeeded - return error to be safe
		// This ensures we don't silently ignore verification failures
		return infer.DeleteResponse{}, fmt.Errorf("delete reported success but verification failed (ListClients error): %w", listErr)
	}

	// Check if the client still exists in the list
	for _, client := range listResp.Clients {
		if client.Id == deleteID {
			// Client still exists - the delete didn't work
			// This is a critical error - the delete call succeeded but the resource still exists
			errMsg := fmt.Sprintf("delete reported success but client %q still exists in Dex (found in ListClients response with %d total clients)", deleteID, len(listResp.Clients))
			return infer.DeleteResponse{}, fmt.Errorf(errMsg)
		}
	}

	// Client not found in list - delete was successful
	return infer.DeleteResponse{}, nil
}

// ptrOrString returns the value pointed to by p, or nil if p is empty or nil.
func ptrOrString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
