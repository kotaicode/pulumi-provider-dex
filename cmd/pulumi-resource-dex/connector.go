package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	api "github.com/dexidp/dex/api/v2"
	"github.com/pulumi/pulumi-go-provider/infer"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ConnectorArgs defines the inputs for a dex.Connector resource.
type ConnectorArgs struct {
	ConnectorId string      `pulumi:"connectorId"`
	Type        string      `pulumi:"type"`
	Name        string      `pulumi:"name"`
	OIDCConfig  *OIDCConfig `pulumi:"oidcConfig,optional"`
	RawConfig   *string     `pulumi:"rawConfig,optional"`
}

// ConnectorState defines the outputs/state for a dex.Connector resource.
type ConnectorState struct {
	ConnectorArgs
}

// OIDCConfig mirrors Dex's OIDC connector JSON configuration.
// This is intentionally close to the config used in simple-client/types.go.
// Note: JSON tags match pulumi tags (camelCase) for proper decoding. We convert to Dex format in buildConnectorConfigBytes.
type OIDCConfig struct {
	Issuer                    string            `pulumi:"issuer" json:"issuer"`
	ClientId                  string            `pulumi:"clientId" json:"clientId"` // Match pulumi tag for decoder
	ClientSecret              string            `pulumi:"clientSecret" json:"clientSecret" provider:"secret"`
	RedirectUri               string            `pulumi:"redirectUri" json:"redirectUri"` // Match pulumi tag for decoder
	Scopes                    []string          `pulumi:"scopes,optional" json:"scopes,omitempty"`
	InsecureSkipEmailVerified *bool             `pulumi:"insecureSkipEmailVerified,optional" json:"insecureSkipEmailVerified,omitempty"`
	InsecureIssuer            *bool             `pulumi:"insecureIssuer,optional" json:"insecureIssuer,omitempty"`
	UserNameKey               *string            `pulumi:"userNameKey,optional" json:"userNameKey,omitempty"`
	ClaimMapping              *OIDCClaimMapping `pulumi:"claimMapping,optional" json:"claimMapping,omitempty"`
	Extra                     map[string]any    `pulumi:"extra,optional" json:"-"`
}

// OIDCClaimMapping represents claim mapping configuration.
type OIDCClaimMapping struct {
	EmailKey  *string `pulumi:"emailKey,optional" json:"email,omitempty"`
	GroupsKey *string `pulumi:"groupsKey,optional" json:"groups,omitempty"`
}

// Connector represents a Dex connector resource (generic).
type Connector struct{}

// Annotate provides schema metadata for the Connector resource.
func (c *Connector) Annotate(a infer.Annotator) {
	a.Describe(c, "Manages a connector (upstream identity provider) in Dex.")
}

// Create creates a new connector in Dex.
func (c *Connector) Create(ctx context.Context, req infer.CreateRequest[ConnectorArgs]) (infer.CreateResponse[ConnectorState], error) {
	args := req.Inputs

	// In preview/dry-run mode, skip actual Dex API calls and return expected state
	// This check MUST be first, before any other operations or config checks
	if req.DryRun {
		state := ConnectorState{
			ConnectorArgs: args,
		}
		return infer.CreateResponse[ConnectorState]{
			ID:     args.ConnectorId,
			Output: state,
		}, nil
	}

	cfg := infer.GetConfig[DexConfig](ctx)
	if cfg.client == nil {
		return infer.CreateResponse[ConnectorState]{}, fmt.Errorf("Dex client not configured")
	}

	if err := validateConnectorArgs(args); err != nil {
		return infer.CreateResponse[ConnectorState]{}, err
	}

	configBytes, err := buildConnectorConfigBytes(args)
	if err != nil {
		return infer.CreateResponse[ConnectorState]{}, err
	}

	conn := &api.Connector{
		Id:     args.ConnectorId,
		Type:   args.Type,
		Name:   args.Name,
		Config: configBytes,
	}

	callCtx, cancel := context.WithTimeout(ctx, time.Duration(ptrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	resp, err := cfg.client.CreateConnector(callCtx, &api.CreateConnectorReq{
		Connector: conn,
	})
	if err != nil {
		return infer.CreateResponse[ConnectorState]{}, wrapError("create", "connector", args.ConnectorId, err)
	}

	if resp.AlreadyExists {
		// Resource already exists - read it and return it so Pulumi can track it
		// This allows destroy to work properly even if the resource was created outside Pulumi
		readCtx, cancel := context.WithTimeout(ctx, time.Duration(ptrOr(cfg.TimeoutSeconds, 5))*time.Second)
		defer cancel()
		
		listResp, err := cfg.client.ListConnectors(readCtx, &api.ListConnectorReq{})
		if err != nil {
			return infer.CreateResponse[ConnectorState]{}, fmt.Errorf("connector already exists but failed to list connectors: %w", err)
		}
		
		var found *api.Connector
		for _, con := range listResp.Connectors {
			if con.Id == args.ConnectorId {
				found = con
				break
			}
		}
		
		if found == nil {
			return infer.CreateResponse[ConnectorState]{}, fmt.Errorf("connector already exists but not found in list")
		}
		
		// Decode the existing connector
		existingArgs, _, err := decodeConnector(found)
		if err != nil {
			return infer.CreateResponse[ConnectorState]{}, fmt.Errorf("failed to decode existing connector: %w", err)
		}
		
		state := ConnectorState{
			ConnectorArgs: existingArgs,
		}
		
		return infer.CreateResponse[ConnectorState]{
			ID:     args.ConnectorId,
			Output: state,
		}, nil
	}

	state := ConnectorState{
		ConnectorArgs: args,
	}

	return infer.CreateResponse[ConnectorState]{
		ID:     args.ConnectorId,
		Output: state,
	}, nil
}

// Read reads an existing connector from Dex.
func (c *Connector) Read(ctx context.Context, req infer.ReadRequest[ConnectorArgs, ConnectorState]) (infer.ReadResponse[ConnectorArgs, ConnectorState], error) {
	cfg := infer.GetConfig[DexConfig](ctx)
	if cfg.client == nil {
		return infer.ReadResponse[ConnectorArgs, ConnectorState]{}, fmt.Errorf("Dex client not configured")
	}

	// Dex API doesn't expose GetConnector; we list and filter by ID.
	callCtx, cancel := context.WithTimeout(ctx, time.Duration(ptrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	listResp, err := cfg.client.ListConnectors(callCtx, &api.ListConnectorReq{})
	if err != nil {
		return infer.ReadResponse[ConnectorArgs, ConnectorState]{}, fmt.Errorf("failed to list Dex connectors: %w", err)
	}

	var found *api.Connector
	for _, con := range listResp.Connectors {
		if con.Id == req.ID {
			found = con
			break
		}
	}

	if found == nil {
		// Connector not found => resource should be deleted.
		return infer.ReadResponse[ConnectorArgs, ConnectorState]{}, nil
	}

	args, state, err := decodeConnector(found)
	if err != nil {
		return infer.ReadResponse[ConnectorArgs, ConnectorState]{}, err
	}

	return infer.ReadResponse[ConnectorArgs, ConnectorState]{
		ID:     found.Id,
		Inputs: args,
		State:  state,
	}, nil
}

// Update updates an existing connector in Dex.
func (c *Connector) Update(ctx context.Context, req infer.UpdateRequest[ConnectorArgs, ConnectorState]) (infer.UpdateResponse[ConnectorState], error) {
	args := req.Inputs
	old := req.State

	// In preview/dry-run mode, skip actual Dex API calls and return expected state
	// This check MUST be first, before any other operations or config checks
	if req.DryRun {
		state := ConnectorState{
			ConnectorArgs: args,
		}
		return infer.UpdateResponse[ConnectorState]{Output: state}, nil
	}

	cfg := infer.GetConfig[DexConfig](ctx)
	if cfg.client == nil {
		return infer.UpdateResponse[ConnectorState]{}, fmt.Errorf("Dex client not configured")
	}

	if args.ConnectorId != old.ConnectorId {
		return infer.UpdateResponse[ConnectorState]{}, fmt.Errorf("connectorId cannot be changed (was %q, got %q)", old.ConnectorId, args.ConnectorId)
	}

	if err := validateConnectorArgs(args); err != nil {
		return infer.UpdateResponse[ConnectorState]{}, err
	}

	configBytes, err := buildConnectorConfigBytes(args)
	if err != nil {
		return infer.UpdateResponse[ConnectorState]{}, err
	}

	updateReq := &api.UpdateConnectorReq{
		Id:        args.ConnectorId,
		NewType:   args.Type,
		NewName:   args.Name,
		NewConfig: configBytes,
	}

	callCtx, cancel := context.WithTimeout(ctx, time.Duration(ptrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	_, err = cfg.client.UpdateConnector(callCtx, updateReq)
	if err != nil {
		return infer.UpdateResponse[ConnectorState]{}, wrapError("update", "connector", args.ConnectorId, err)
	}

	state := ConnectorState{
		ConnectorArgs: args,
	}

	return infer.UpdateResponse[ConnectorState]{Output: state}, nil
}

// Delete deletes a connector from Dex.
func (c *Connector) Delete(ctx context.Context, req infer.DeleteRequest[ConnectorState]) (infer.DeleteResponse, error) {
	cfg := infer.GetConfig[DexConfig](ctx)
	if cfg.client == nil {
		return infer.DeleteResponse{}, fmt.Errorf("Dex client not configured")
	}

	deleteID := req.ID
	if deleteID == "" && req.State.ConnectorId != "" {
		deleteID = req.State.ConnectorId
	}
	if deleteID == "" {
		return infer.DeleteResponse{}, fmt.Errorf("cannot delete connector: no ID provided in request or state")
	}

	// Note: Pulumi does not call Delete during preview, so no preview check needed

	callCtx, cancel := context.WithTimeout(ctx, time.Duration(ptrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	_, err := cfg.client.DeleteConnector(callCtx, &api.DeleteConnectorReq{
		Id: deleteID,
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			// Already deleted; treat as success.
			return infer.DeleteResponse{}, nil
		}
		return infer.DeleteResponse{}, wrapError("delete", "connector", deleteID, err)
	}

	return infer.DeleteResponse{}, nil
}

// validateConnectorArgs enforces high-level invariants for connectors.
func validateConnectorArgs(args ConnectorArgs) error {
	if args.ConnectorId == "" {
		return fmt.Errorf("connectorId is required")
	}
	if args.Type == "" {
		return fmt.Errorf("type is required")
	}
	if args.Name == "" {
		return fmt.Errorf("name is required")
	}

	oidcSet := args.OIDCConfig != nil
	rawSet := args.RawConfig != nil && *args.RawConfig != ""
	if oidcSet == rawSet {
		return fmt.Errorf("exactly one of oidcConfig or rawConfig must be set")
	}
	if args.Type != "oidc" && oidcSet {
		return fmt.Errorf("oidcConfig is only valid when type == \"oidc\"")
	}
	return nil
}

// buildConnectorConfigBytes produces the JSON config bytes to send to Dex.
func buildConnectorConfigBytes(args ConnectorArgs) ([]byte, error) {
	if args.OIDCConfig != nil {
		// Convert from Pulumi format (camelCase) to Dex format (PascalCase for clientID/redirectURI).
		base := map[string]any{}
		raw, err := json.Marshal(args.OIDCConfig)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal OIDC config: %w", err)
		}
		if err := json.Unmarshal(raw, &base); err != nil {
			return nil, fmt.Errorf("failed to rehydrate OIDC config: %w", err)
		}
		// Convert camelCase keys to Dex format
		if clientId, ok := base["clientId"]; ok {
			base["clientID"] = clientId
			delete(base, "clientId")
		}
		if redirectUri, ok := base["redirectUri"]; ok {
			base["redirectURI"] = redirectUri
			delete(base, "redirectUri")
		}
		// Merge Extra fields
		for k, v := range args.OIDCConfig.Extra {
			base[k] = v
		}
		out, err := json.Marshal(base)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal merged OIDC config: %w", err)
		}
		return out, nil
	}

	// Raw JSON path.
	var raw json.RawMessage
	if err := json.Unmarshal([]byte(ptrOr(args.RawConfig, "")), &raw); err != nil {
		return nil, fmt.Errorf("rawConfig must be valid JSON: %w", err)
	}
	return []byte(ptrOr(args.RawConfig, "")), nil
}

// decodeConnector converts a Dex Connector into ConnectorArgs/State.
func decodeConnector(con *api.Connector) (ConnectorArgs, ConnectorState, error) {
	args := ConnectorArgs{
		ConnectorId: con.Id,
		Type:        con.Type,
		Name:        con.Name,
	}

	// Try to parse as OIDC config when type == "oidc".
	if con.Type == "oidc" && len(con.Config) > 0 {
		var base map[string]any
		if err := json.Unmarshal(con.Config, &base); err == nil {
			// Attempt to map known fields into OIDCConfig.
			oidc := &OIDCConfig{}
			// Marshal back into JSON and unmarshal into typed struct for known fields.
			if data, err := json.Marshal(base); err == nil {
				_ = json.Unmarshal(data, oidc)
			}

			// Convert from Dex format (clientID, redirectURI) to Pulumi format (clientId, redirectUri)
			if clientID, ok := base["clientID"]; ok {
				base["clientId"] = clientID
				delete(base, "clientID")
			}
			if redirectURI, ok := base["redirectURI"]; ok {
				base["redirectUri"] = redirectURI
				delete(base, "redirectURI")
			}
			// Remove known fields from base, whatever remains goes into Extra.
			delete(base, "issuer")
			delete(base, "clientId")
			delete(base, "clientSecret")
			delete(base, "redirectUri")
			delete(base, "scopes")
			delete(base, "insecureSkipEmailVerified")
			delete(base, "insecureIssuer")
			delete(base, "userNameKey")
			delete(base, "claimMapping")

			if len(base) > 0 {
				oidc.Extra = base
			}

			args.OIDCConfig = oidc
		} else {
			// Fall back to rawConfig if JSON parsing fails.
			rc := string(con.Config)
			args.RawConfig = &rc
		}
	} else if len(con.Config) > 0 {
		rc := string(con.Config)
		args.RawConfig = &rc
	}

	state := ConnectorState{
		ConnectorArgs: args,
	}
	return args, state, nil
}
