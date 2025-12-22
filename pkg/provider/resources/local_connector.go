package resources

import (
	"context"
	"fmt"
	"time"

	api "github.com/dexidp/dex/api/v2"
	"github.com/kotaicode/pulumi-provider-dex/pkg/provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ============================================================================
// LocalConnector - Builtin/local connector (type: "local")
// ============================================================================

// LocalConnectorArgs defines inputs for LocalConnector.
type LocalConnectorArgs struct {
	ConnectorId string `pulumi:"connectorId"`
	Name        string `pulumi:"name"`
	Enabled     *bool  `pulumi:"enabled,optional"`
}

// LocalConnectorState defines outputs for LocalConnector.
type LocalConnectorState struct {
	LocalConnectorArgs
}

// LocalConnector manages a local/builtin connector in Dex.
type LocalConnector struct{}

// Annotate provides schema metadata.
func (c *LocalConnector) Annotate(a infer.Annotator) {
	a.Describe(c, "Manages a local/builtin connector in Dex. The local connector provides username/password authentication stored in Dex's database. This is useful for testing or when you don't have an external identity provider.")
}

// Annotate provides schema metadata for LocalConnectorArgs.
func (c *LocalConnectorArgs) Annotate(a infer.Annotator) {
	a.Describe(&c.ConnectorId, "Unique identifier for the local connector.")
	a.Describe(&c.Name, "Human-readable name for the connector, displayed to users during login.")
	a.Describe(&c.Enabled, "Whether the local connector is enabled. Defaults to true.")
}

// Annotate provides schema metadata for LocalConnectorState.
func (c *LocalConnectorState) Annotate(a infer.Annotator) {
	// LocalConnectorState embeds LocalConnectorArgs, so field descriptions are inherited
}

// Check validates inputs.
func (c *LocalConnector) Check(ctx context.Context, req infer.CheckRequest) (infer.CheckResponse[LocalConnectorArgs], error) {
	args, failures, err := infer.DefaultCheck[LocalConnectorArgs](ctx, req.NewInputs)
	if err != nil {
		return infer.CheckResponse[LocalConnectorArgs]{Failures: failures}, err
	}

	// Apply defaults
	if args.Enabled == nil {
		defaultEnabled := true
		args.Enabled = &defaultEnabled
	}

	return infer.CheckResponse[LocalConnectorArgs]{
		Inputs:   args,
		Failures: failures,
	}, nil
}

// Create creates a new local connector.
func (c *LocalConnector) Create(ctx context.Context, req infer.CreateRequest[LocalConnectorArgs]) (infer.CreateResponse[LocalConnectorState], error) {
	args := req.Inputs

	// In preview/dry-run mode, skip actual Dex API calls and return expected state
	if req.DryRun {
		state := LocalConnectorState{
			LocalConnectorArgs: args,
		}
		return infer.CreateResponse[LocalConnectorState]{
			ID:     args.ConnectorId,
			Output: state,
		}, nil
	}

	cfg := infer.GetConfig[provider.DexConfig](ctx)
	if cfg.Client == nil {
		return infer.CreateResponse[LocalConnectorState]{}, fmt.Errorf("Dex client not configured")
	}

	// Local connector has minimal config - just an empty JSON object
	configBytes := []byte("{}")

	connector := &api.Connector{
		Id:     args.ConnectorId,
		Type:   "local",
		Name:   args.Name,
		Config: configBytes,
	}

	createCtx, cancel := context.WithTimeout(ctx, time.Duration(provider.PtrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	resp, err := cfg.Client.CreateConnector(createCtx, &api.CreateConnectorReq{
		Connector: connector,
	})
	if err != nil {
		return infer.CreateResponse[LocalConnectorState]{}, provider.WrapError("create", "local-connector", args.ConnectorId, err)
	}

	if resp.AlreadyExists {
		return infer.CreateResponse[LocalConnectorState]{}, fmt.Errorf("connector with id %q already exists", args.ConnectorId)
	}

	state := LocalConnectorState{
		LocalConnectorArgs: args,
	}

	return infer.CreateResponse[LocalConnectorState]{
		ID:     args.ConnectorId,
		Output: state,
	}, nil
}

// Read retrieves an existing local connector.
func (c *LocalConnector) Read(ctx context.Context, req infer.ReadRequest[LocalConnectorArgs, LocalConnectorState]) (infer.ReadResponse[LocalConnectorArgs, LocalConnectorState], error) {
	cfg := infer.GetConfig[provider.DexConfig](ctx)
	if cfg.Client == nil {
		return infer.ReadResponse[LocalConnectorArgs, LocalConnectorState]{}, fmt.Errorf("Dex client not configured")
	}

	listCtx, cancel := context.WithTimeout(ctx, time.Duration(provider.PtrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	listResp, err := cfg.Client.ListConnectors(listCtx, &api.ListConnectorReq{})
	if err != nil {
		return infer.ReadResponse[LocalConnectorArgs, LocalConnectorState]{}, fmt.Errorf("failed to list connectors: %w", err)
	}

	var found *api.Connector
	for _, conn := range listResp.Connectors {
		if conn.Id == req.ID {
			found = conn
			break
		}
	}

	if found == nil {
		return infer.ReadResponse[LocalConnectorArgs, LocalConnectorState]{}, nil
	}

	// Local connector has minimal config, so we just use defaults
	enabled := true
	args := LocalConnectorArgs{
		ConnectorId: found.Id,
		Name:        found.Name,
		Enabled:     &enabled,
	}

	state := LocalConnectorState{
		LocalConnectorArgs: args,
	}

	return infer.ReadResponse[LocalConnectorArgs, LocalConnectorState]{
		ID:     found.Id,
		Inputs: args,
		State:  state,
	}, nil
}

// Update updates an existing local connector.
func (c *LocalConnector) Update(ctx context.Context, req infer.UpdateRequest[LocalConnectorArgs, LocalConnectorState]) (infer.UpdateResponse[LocalConnectorState], error) {
	args := req.Inputs
	oldState := req.State

	// In preview/dry-run mode, skip actual Dex API calls and return expected state
	if req.DryRun {
		state := LocalConnectorState{
			LocalConnectorArgs: args,
		}
		return infer.UpdateResponse[LocalConnectorState]{
			Output: state,
		}, nil
	}

	cfg := infer.GetConfig[provider.DexConfig](ctx)
	if cfg.Client == nil {
		return infer.UpdateResponse[LocalConnectorState]{}, fmt.Errorf("Dex client not configured")
	}

	if args.ConnectorId != oldState.ConnectorId {
		return infer.UpdateResponse[LocalConnectorState]{}, fmt.Errorf("connectorId cannot be changed")
	}

	configBytes := []byte("{}")

	updateCtx, cancel := context.WithTimeout(ctx, time.Duration(provider.PtrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	_, err := cfg.Client.UpdateConnector(updateCtx, &api.UpdateConnectorReq{
		Id:        args.ConnectorId,
		NewType:   "local",
		NewName:   args.Name,
		NewConfig: configBytes,
	})
	if err != nil {
		return infer.UpdateResponse[LocalConnectorState]{}, provider.WrapError("update", "local-connector", args.ConnectorId, err)
	}

	state := LocalConnectorState{
		LocalConnectorArgs: args,
	}

	return infer.UpdateResponse[LocalConnectorState]{
		Output: state,
	}, nil
}

// Delete deletes a local connector.
func (c *LocalConnector) Delete(ctx context.Context, req infer.DeleteRequest[LocalConnectorState]) (infer.DeleteResponse, error) {
	cfg := infer.GetConfig[provider.DexConfig](ctx)
	if cfg.Client == nil {
		return infer.DeleteResponse{}, fmt.Errorf("Dex client not configured")
	}

	deleteID := req.ID
	if deleteID == "" && req.State.ConnectorId != "" {
		deleteID = req.State.ConnectorId
	}

	deleteCtx, cancel := context.WithTimeout(ctx, time.Duration(provider.PtrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	_, err := cfg.Client.DeleteConnector(deleteCtx, &api.DeleteConnectorReq{
		Id: deleteID,
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return infer.DeleteResponse{}, nil
		}
		return infer.DeleteResponse{}, provider.WrapError("delete", "local-connector", deleteID, err)
	}

	return infer.DeleteResponse{}, nil
}
