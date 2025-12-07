package main

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	api "github.com/dexidp/dex/api/v2"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ============================================================================
// AzureOidcConnector - Uses generic OIDC connector (type: "oidc")
// ============================================================================

// AzureOidcConnectorArgs defines inputs for AzureOidcConnector using generic OIDC.
type AzureOidcConnectorArgs struct {
	ConnectorId    string         `pulumi:"connectorId"`
	Name           string         `pulumi:"name"`
	TenantId       string         `pulumi:"tenantId"`
	ClientId       string         `pulumi:"clientId"`
	ClientSecret   string         `pulumi:"clientSecret" provider:"secret"`
	RedirectUri    string         `pulumi:"redirectUri"`
	Scopes         []string       `pulumi:"scopes,optional"`
	UserNameSource *string        `pulumi:"userNameSource,optional"` // "preferred_username" | "upn" | "email"
	ExtraOidc      map[string]any `pulumi:"extraOidc,optional"`      // Additional OIDC config fields
}

// AzureOidcConnectorState defines outputs for AzureOidcConnector.
type AzureOidcConnectorState struct {
	AzureOidcConnectorArgs
}

// AzureOidcConnector manages an Azure/Entra ID connector using Dex's generic OIDC connector.
type AzureOidcConnector struct{}

// Annotate provides schema metadata.
func (c *AzureOidcConnector) Annotate(a infer.Annotator) {
	a.Describe(c, "Manages an Azure AD/Entra ID connector in Dex using the generic OIDC connector (type: oidc).")
}

// Check validates inputs before creation/update.
func (c *AzureOidcConnector) Check(ctx context.Context, req infer.CheckRequest) (infer.CheckResponse[AzureOidcConnectorArgs], error) {
	args, failures, err := infer.DefaultCheck[AzureOidcConnectorArgs](ctx, req.NewInputs)
	if err != nil {
		return infer.CheckResponse[AzureOidcConnectorArgs]{Failures: failures}, err
	}

	// Validate tenantId format (UUID)
	if args.TenantId != "" {
		uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
		if !uuidRegex.MatchString(strings.ToLower(args.TenantId)) {
			failures = append(failures, p.CheckFailure{
				Property: "tenantId",
				Reason:   "must be a valid UUID",
			})
		}
	}

	// Validate userNameSource
	if args.UserNameSource != nil {
		valid := map[string]bool{"preferred_username": true, "upn": true, "email": true}
		if !valid[*args.UserNameSource] {
			failures = append(failures, p.CheckFailure{
				Property: "userNameSource",
				Reason:   "must be one of: preferred_username, upn, email",
			})
		}
	}

	// Apply defaults
	if len(args.Scopes) == 0 {
		args.Scopes = []string{"openid", "profile", "email", "offline_access"}
	}

	return infer.CheckResponse[AzureOidcConnectorArgs]{
		Inputs:   args,
		Failures: failures,
	}, nil
}

// Create creates a new Azure OIDC connector.
func (c *AzureOidcConnector) Create(ctx context.Context, req infer.CreateRequest[AzureOidcConnectorArgs]) (infer.CreateResponse[AzureOidcConnectorState], error) {
	args := req.Inputs

	// In preview/dry-run mode, skip actual Dex API calls and return expected state
	// This check MUST be first, before any other operations or config checks
	if req.DryRun {
		state := AzureOidcConnectorState{
			AzureOidcConnectorArgs: args,
		}
		return infer.CreateResponse[AzureOidcConnectorState]{
			ID:     args.ConnectorId,
			Output: state,
		}, nil
	}

	cfg := infer.GetConfig[DexConfig](ctx)
	if cfg.client == nil {
		return infer.CreateResponse[AzureOidcConnectorState]{}, fmt.Errorf("Dex client not configured")
	}

	// Derive issuer from tenantId
	issuer := fmt.Sprintf("https://login.microsoftonline.com/%s/v2.0", args.TenantId)

	// Derive userNameKey from userNameSource
	userNameKey := "preferred_username" // default
	if args.UserNameSource != nil {
		userNameKey = *args.UserNameSource
	}

	// Build OIDC config
	oidcConfig := map[string]any{
		"issuer":       issuer,
		"clientID":     args.ClientId,
		"clientSecret": args.ClientSecret,
		"redirectURI":  args.RedirectUri,
		"scopes":       args.Scopes,
		"userNameKey":  userNameKey,
	}

	// Merge extraOidc fields
	for k, v := range args.ExtraOidc {
		oidcConfig[k] = v
	}

	configBytes, err := json.Marshal(oidcConfig)
	if err != nil {
		return infer.CreateResponse[AzureOidcConnectorState]{}, fmt.Errorf("failed to marshal OIDC config: %w", err)
	}

	// Create connector
	connector := &api.Connector{
		Id:     args.ConnectorId,
		Type:   "oidc",
		Name:   args.Name,
		Config: configBytes,
	}

	createCtx, cancel := context.WithTimeout(ctx, time.Duration(ptrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	resp, err := cfg.client.CreateConnector(createCtx, &api.CreateConnectorReq{
		Connector: connector,
	})
	if err != nil {
		return infer.CreateResponse[AzureOidcConnectorState]{}, wrapError("create", "azure-oidc-connector", args.ConnectorId, err)
	}

	if resp.AlreadyExists {
		return infer.CreateResponse[AzureOidcConnectorState]{}, fmt.Errorf("connector with id %q already exists", args.ConnectorId)
	}

	state := AzureOidcConnectorState{
		AzureOidcConnectorArgs: args,
	}

	return infer.CreateResponse[AzureOidcConnectorState]{
		ID:     args.ConnectorId,
		Output: state,
	}, nil
}

// Read retrieves an existing Azure OIDC connector.
func (c *AzureOidcConnector) Read(ctx context.Context, req infer.ReadRequest[AzureOidcConnectorArgs, AzureOidcConnectorState]) (infer.ReadResponse[AzureOidcConnectorArgs, AzureOidcConnectorState], error) {
	cfg := infer.GetConfig[DexConfig](ctx)
	if cfg.client == nil {
		return infer.ReadResponse[AzureOidcConnectorArgs, AzureOidcConnectorState]{}, fmt.Errorf("Dex client not configured")
	}

	// List connectors and find by ID
	listCtx, cancel := context.WithTimeout(ctx, time.Duration(ptrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	listResp, err := cfg.client.ListConnectors(listCtx, &api.ListConnectorReq{})
	if err != nil {
		return infer.ReadResponse[AzureOidcConnectorArgs, AzureOidcConnectorState]{}, fmt.Errorf("failed to list connectors: %w", err)
	}

	var found *api.Connector
	for _, conn := range listResp.Connectors {
		if conn.Id == req.ID {
			found = conn
			break
		}
	}

	if found == nil {
		// Not found - return empty to indicate deletion
		return infer.ReadResponse[AzureOidcConnectorArgs, AzureOidcConnectorState]{}, nil
	}

	// Parse config back to args
	var configMap map[string]any
	if err := json.Unmarshal(found.Config, &configMap); err != nil {
		// If we can't parse, return empty (connector exists but config is invalid)
		return infer.ReadResponse[AzureOidcConnectorArgs, AzureOidcConnectorState]{}, nil
	}

	// Extract tenantId from issuer
	issuer, _ := configMap["issuer"].(string)
	tenantId := ""
	if strings.HasPrefix(issuer, "https://login.microsoftonline.com/") {
		parts := strings.Split(issuer, "/")
		if len(parts) >= 4 {
			tenantId = parts[3]
		}
	}

	// Extract userNameKey and map to userNameSource
	userNameKey, _ := configMap["userNameKey"].(string)
	userNameSource := &userNameKey
	if userNameKey == "preferred_username" {
		// This is the default, so we can leave it as is
	}

	scopes, _ := configMap["scopes"].([]any)
	scopesStr := make([]string, 0, len(scopes))
	for _, s := range scopes {
		if str, ok := s.(string); ok {
			scopesStr = append(scopesStr, str)
		}
	}

	// Build args from config
	args := AzureOidcConnectorArgs{
		ConnectorId:    found.Id,
		Name:           found.Name,
		TenantId:       tenantId,
		ClientId:       getString(configMap, "clientID"),
		ClientSecret:   getString(configMap, "clientSecret"),
		RedirectUri:    getString(configMap, "redirectURI"),
		Scopes:         scopesStr,
		UserNameSource: userNameSource,
	}

	state := AzureOidcConnectorState{
		AzureOidcConnectorArgs: args,
	}

	return infer.ReadResponse[AzureOidcConnectorArgs, AzureOidcConnectorState]{
		ID:     found.Id,
		Inputs: args,
		State:  state,
	}, nil
}

// Update updates an existing Azure OIDC connector.
func (c *AzureOidcConnector) Update(ctx context.Context, req infer.UpdateRequest[AzureOidcConnectorArgs, AzureOidcConnectorState]) (infer.UpdateResponse[AzureOidcConnectorState], error) {
	args := req.Inputs
	oldState := req.State

	// In preview/dry-run mode, skip actual Dex API calls and return expected state
	// This check MUST be first, before any other operations or config checks
	if req.DryRun {
		state := AzureOidcConnectorState{
			AzureOidcConnectorArgs: args,
		}
		return infer.UpdateResponse[AzureOidcConnectorState]{
			Output: state,
		}, nil
	}

	cfg := infer.GetConfig[DexConfig](ctx)
	if cfg.client == nil {
		return infer.UpdateResponse[AzureOidcConnectorState]{}, fmt.Errorf("Dex client not configured")
	}

	// Validate immutable fields
	if args.ConnectorId != oldState.ConnectorId {
		return infer.UpdateResponse[AzureOidcConnectorState]{}, fmt.Errorf("connectorId cannot be changed")
	}
	if args.TenantId != oldState.TenantId {
		return infer.UpdateResponse[AzureOidcConnectorState]{}, fmt.Errorf("tenantId cannot be changed (would require replace)")
	}

	// Rebuild config (same as Create)
	issuer := fmt.Sprintf("https://login.microsoftonline.com/%s/v2.0", args.TenantId)
	userNameKey := "preferred_username"
	if args.UserNameSource != nil {
		userNameKey = *args.UserNameSource
	}

	oidcConfig := map[string]any{
		"issuer":       issuer,
		"clientID":     args.ClientId,
		"clientSecret": args.ClientSecret,
		"redirectURI":  args.RedirectUri,
		"scopes":       args.Scopes,
		"userNameKey":  userNameKey,
	}

	for k, v := range args.ExtraOidc {
		oidcConfig[k] = v
	}

	configBytes, err := json.Marshal(oidcConfig)
	if err != nil {
		return infer.UpdateResponse[AzureOidcConnectorState]{}, fmt.Errorf("failed to marshal OIDC config: %w", err)
	}

	updateCtx, cancel := context.WithTimeout(ctx, time.Duration(ptrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	_, err = cfg.client.UpdateConnector(updateCtx, &api.UpdateConnectorReq{
		Id:        args.ConnectorId,
		NewType:   "oidc",
		NewName:   args.Name,
		NewConfig: configBytes,
	})
	if err != nil {
		return infer.UpdateResponse[AzureOidcConnectorState]{}, wrapError("update", "azure-oidc-connector", args.ConnectorId, err)
	}

	state := AzureOidcConnectorState{
		AzureOidcConnectorArgs: args,
	}

	return infer.UpdateResponse[AzureOidcConnectorState]{
		Output: state,
	}, nil
}

// Delete deletes an Azure OIDC connector.
func (c *AzureOidcConnector) Delete(ctx context.Context, req infer.DeleteRequest[AzureOidcConnectorState]) (infer.DeleteResponse, error) {
	cfg := infer.GetConfig[DexConfig](ctx)
	if cfg.client == nil {
		return infer.DeleteResponse{}, fmt.Errorf("Dex client not configured")
	}

	deleteID := req.ID
	if deleteID == "" && req.State.ConnectorId != "" {
		deleteID = req.State.ConnectorId
	}

	// Note: Pulumi does not call Delete during preview, so no preview check needed

	deleteCtx, cancel := context.WithTimeout(ctx, time.Duration(ptrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	_, err := cfg.client.DeleteConnector(deleteCtx, &api.DeleteConnectorReq{
		Id: deleteID,
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return infer.DeleteResponse{}, nil // Already deleted
		}
		return infer.DeleteResponse{}, wrapError("delete", "azure-oidc-connector", deleteID, err)
	}

	return infer.DeleteResponse{}, nil
}

// ============================================================================
// AzureMicrosoftConnector - Uses Dex's Microsoft-specific connector (type: "microsoft")
// ============================================================================

// AzureMicrosoftConnectorArgs defines inputs for AzureMicrosoftConnector using Microsoft connector.
type AzureMicrosoftConnectorArgs struct {
	ConnectorId  string  `pulumi:"connectorId"`
	Name         string  `pulumi:"name"`
	Tenant       string  `pulumi:"tenant"` // "common", "organizations", or tenant ID
	ClientId     string  `pulumi:"clientId"`
	ClientSecret string  `pulumi:"clientSecret" provider:"secret"`
	RedirectUri  string  `pulumi:"redirectUri"`
	Groups       *string `pulumi:"groups,optional"` // Group claim name, e.g., "groups"
}

// AzureMicrosoftConnectorState defines outputs for AzureMicrosoftConnector.
type AzureMicrosoftConnectorState struct {
	AzureMicrosoftConnectorArgs
}

// AzureMicrosoftConnector manages an Azure/Entra ID connector using Dex's Microsoft-specific connector.
type AzureMicrosoftConnector struct{}

// Annotate provides schema metadata.
func (c *AzureMicrosoftConnector) Annotate(a infer.Annotator) {
	a.Describe(c, "Manages an Azure AD/Entra ID connector in Dex using the Microsoft-specific connector (type: microsoft).")
}

// Check validates inputs.
func (c *AzureMicrosoftConnector) Check(ctx context.Context, req infer.CheckRequest) (infer.CheckResponse[AzureMicrosoftConnectorArgs], error) {
	args, failures, err := infer.DefaultCheck[AzureMicrosoftConnectorArgs](ctx, req.NewInputs)
	if err != nil {
		return infer.CheckResponse[AzureMicrosoftConnectorArgs]{Failures: failures}, err
	}

	// Validate tenant format
	if args.Tenant != "" && args.Tenant != "common" && args.Tenant != "organizations" {
		// Check if it's a UUID
		uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
		if !uuidRegex.MatchString(strings.ToLower(args.Tenant)) {
			failures = append(failures, p.CheckFailure{
				Property: "tenant",
				Reason:   "must be \"common\", \"organizations\", or a valid UUID",
			})
		}
	}

	return infer.CheckResponse[AzureMicrosoftConnectorArgs]{
		Inputs:   args,
		Failures: failures,
	}, nil
}

// Create creates a new Azure Microsoft connector.
func (c *AzureMicrosoftConnector) Create(ctx context.Context, req infer.CreateRequest[AzureMicrosoftConnectorArgs]) (infer.CreateResponse[AzureMicrosoftConnectorState], error) {
	args := req.Inputs

	// In preview/dry-run mode, skip actual Dex API calls and return expected state
	// This check MUST be first, before any other operations or config checks
	if req.DryRun {
		state := AzureMicrosoftConnectorState{
			AzureMicrosoftConnectorArgs: args,
		}
		return infer.CreateResponse[AzureMicrosoftConnectorState]{
			ID:     args.ConnectorId,
			Output: state,
		}, nil
	}

	cfg := infer.GetConfig[DexConfig](ctx)
	if cfg.client == nil {
		return infer.CreateResponse[AzureMicrosoftConnectorState]{}, fmt.Errorf("Dex client not configured")
	}

	// Build Microsoft connector config
	microsoftConfig := map[string]any{
		"clientID":     args.ClientId,
		"clientSecret": args.ClientSecret,
		"redirectURI":  args.RedirectUri,
		"tenant":       args.Tenant,
	}

	if args.Groups != nil {
		microsoftConfig["groups"] = *args.Groups
	}

	configBytes, err := json.Marshal(microsoftConfig)
	if err != nil {
		return infer.CreateResponse[AzureMicrosoftConnectorState]{}, fmt.Errorf("failed to marshal Microsoft config: %w", err)
	}

	connector := &api.Connector{
		Id:     args.ConnectorId,
		Type:   "microsoft",
		Name:   args.Name,
		Config: configBytes,
	}

	createCtx, cancel := context.WithTimeout(ctx, time.Duration(ptrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	resp, err := cfg.client.CreateConnector(createCtx, &api.CreateConnectorReq{
		Connector: connector,
	})
	if err != nil {
		return infer.CreateResponse[AzureMicrosoftConnectorState]{}, wrapError("create", "azure-microsoft-connector", args.ConnectorId, err)
	}

	if resp.AlreadyExists {
		return infer.CreateResponse[AzureMicrosoftConnectorState]{}, fmt.Errorf("connector with id %q already exists", args.ConnectorId)
	}

	state := AzureMicrosoftConnectorState{
		AzureMicrosoftConnectorArgs: args,
	}

	return infer.CreateResponse[AzureMicrosoftConnectorState]{
		ID:     args.ConnectorId,
		Output: state,
	}, nil
}

// Read retrieves an existing Azure Microsoft connector.
func (c *AzureMicrosoftConnector) Read(ctx context.Context, req infer.ReadRequest[AzureMicrosoftConnectorArgs, AzureMicrosoftConnectorState]) (infer.ReadResponse[AzureMicrosoftConnectorArgs, AzureMicrosoftConnectorState], error) {
	cfg := infer.GetConfig[DexConfig](ctx)
	if cfg.client == nil {
		return infer.ReadResponse[AzureMicrosoftConnectorArgs, AzureMicrosoftConnectorState]{}, fmt.Errorf("Dex client not configured")
	}

	listCtx, cancel := context.WithTimeout(ctx, time.Duration(ptrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	listResp, err := cfg.client.ListConnectors(listCtx, &api.ListConnectorReq{})
	if err != nil {
		return infer.ReadResponse[AzureMicrosoftConnectorArgs, AzureMicrosoftConnectorState]{}, fmt.Errorf("failed to list connectors: %w", err)
	}

	var found *api.Connector
	for _, conn := range listResp.Connectors {
		if conn.Id == req.ID {
			found = conn
			break
		}
	}

	if found == nil {
		return infer.ReadResponse[AzureMicrosoftConnectorArgs, AzureMicrosoftConnectorState]{}, nil
	}

	var configMap map[string]any
	if err := json.Unmarshal(found.Config, &configMap); err != nil {
		return infer.ReadResponse[AzureMicrosoftConnectorArgs, AzureMicrosoftConnectorState]{}, nil
	}

	groups := getStringPtr(configMap, "groups")

	args := AzureMicrosoftConnectorArgs{
		ConnectorId:  found.Id,
		Name:         found.Name,
		Tenant:       getString(configMap, "tenant"),
		ClientId:     getString(configMap, "clientID"),
		ClientSecret: getString(configMap, "clientSecret"),
		RedirectUri:  getString(configMap, "redirectURI"),
		Groups:       groups,
	}

	state := AzureMicrosoftConnectorState{
		AzureMicrosoftConnectorArgs: args,
	}

	return infer.ReadResponse[AzureMicrosoftConnectorArgs, AzureMicrosoftConnectorState]{
		ID:     found.Id,
		Inputs: args,
		State:  state,
	}, nil
}

// Update updates an existing Azure Microsoft connector.
func (c *AzureMicrosoftConnector) Update(ctx context.Context, req infer.UpdateRequest[AzureMicrosoftConnectorArgs, AzureMicrosoftConnectorState]) (infer.UpdateResponse[AzureMicrosoftConnectorState], error) {
	args := req.Inputs
	oldState := req.State

	// In preview/dry-run mode, skip actual Dex API calls and return expected state
	// This check MUST be first, before any other operations or config checks
	if req.DryRun {
		state := AzureMicrosoftConnectorState{
			AzureMicrosoftConnectorArgs: args,
		}
		return infer.UpdateResponse[AzureMicrosoftConnectorState]{
			Output: state,
		}, nil
	}

	cfg := infer.GetConfig[DexConfig](ctx)
	if cfg.client == nil {
		return infer.UpdateResponse[AzureMicrosoftConnectorState]{}, fmt.Errorf("Dex client not configured")
	}

	if args.ConnectorId != oldState.ConnectorId {
		return infer.UpdateResponse[AzureMicrosoftConnectorState]{}, fmt.Errorf("connectorId cannot be changed")
	}
	if args.Tenant != oldState.Tenant {
		return infer.UpdateResponse[AzureMicrosoftConnectorState]{}, fmt.Errorf("tenant cannot be changed (would require replace)")
	}

	microsoftConfig := map[string]any{
		"clientID":     args.ClientId,
		"clientSecret": args.ClientSecret,
		"redirectURI":  args.RedirectUri,
		"tenant":       args.Tenant,
	}

	if args.Groups != nil {
		microsoftConfig["groups"] = *args.Groups
	}

	configBytes, err := json.Marshal(microsoftConfig)
	if err != nil {
		return infer.UpdateResponse[AzureMicrosoftConnectorState]{}, fmt.Errorf("failed to marshal Microsoft config: %w", err)
	}

	updateCtx, cancel := context.WithTimeout(ctx, time.Duration(ptrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	_, err = cfg.client.UpdateConnector(updateCtx, &api.UpdateConnectorReq{
		Id:        args.ConnectorId,
		NewType:   "microsoft",
		NewName:   args.Name,
		NewConfig: configBytes,
	})
	if err != nil {
		return infer.UpdateResponse[AzureMicrosoftConnectorState]{}, wrapError("update", "azure-microsoft-connector", args.ConnectorId, err)
	}

	state := AzureMicrosoftConnectorState{
		AzureMicrosoftConnectorArgs: args,
	}

	return infer.UpdateResponse[AzureMicrosoftConnectorState]{
		Output: state,
	}, nil
}

// Delete deletes an Azure Microsoft connector.
func (c *AzureMicrosoftConnector) Delete(ctx context.Context, req infer.DeleteRequest[AzureMicrosoftConnectorState]) (infer.DeleteResponse, error) {
	cfg := infer.GetConfig[DexConfig](ctx)
	if cfg.client == nil {
		return infer.DeleteResponse{}, fmt.Errorf("Dex client not configured")
	}

	deleteID := req.ID
	if deleteID == "" && req.State.ConnectorId != "" {
		deleteID = req.State.ConnectorId
	}

	// Note: Pulumi does not call Delete during preview, so no preview check needed

	deleteCtx, cancel := context.WithTimeout(ctx, time.Duration(ptrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	_, err := cfg.client.DeleteConnector(deleteCtx, &api.DeleteConnectorReq{
		Id: deleteID,
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return infer.DeleteResponse{}, nil
		}
		return infer.DeleteResponse{}, wrapError("delete", "azure-microsoft-connector", deleteID, err)
	}

	return infer.DeleteResponse{}, nil
}

// ============================================================================
// CognitoOidcConnector - Uses generic OIDC connector (type: "oidc")
// ============================================================================

// CognitoOidcConnectorArgs defines inputs for CognitoOidcConnector.
type CognitoOidcConnectorArgs struct {
	ConnectorId    string         `pulumi:"connectorId"`
	Name           string         `pulumi:"name"`
	Region         string         `pulumi:"region"`
	UserPoolId     string         `pulumi:"userPoolId"`
	ClientId       string         `pulumi:"clientId"`
	ClientSecret   string         `pulumi:"clientSecret" provider:"secret"`
	RedirectUri    string         `pulumi:"redirectUri"`
	Scopes         []string       `pulumi:"scopes,optional"`
	UserNameSource *string        `pulumi:"userNameSource,optional"` // "email" | "sub"
	ExtraOidc      map[string]any `pulumi:"extraOidc,optional"`
}

// CognitoOidcConnectorState defines outputs for CognitoOidcConnector.
type CognitoOidcConnectorState struct {
	CognitoOidcConnectorArgs
}

// CognitoOidcConnector manages an AWS Cognito connector using Dex's generic OIDC connector.
type CognitoOidcConnector struct{}

// Annotate provides schema metadata.
func (c *CognitoOidcConnector) Annotate(a infer.Annotator) {
	a.Describe(c, "Manages an AWS Cognito user pool connector in Dex using the generic OIDC connector (type: oidc).")
}

// Check validates inputs.
func (c *CognitoOidcConnector) Check(ctx context.Context, req infer.CheckRequest) (infer.CheckResponse[CognitoOidcConnectorArgs], error) {
	args, failures, err := infer.DefaultCheck[CognitoOidcConnectorArgs](ctx, req.NewInputs)
	if err != nil {
		return infer.CheckResponse[CognitoOidcConnectorArgs]{Failures: failures}, err
	}

	// Validate region format (basic check)
	if args.Region != "" {
		regionRegex := regexp.MustCompile(`^[a-z0-9-]+$`)
		if !regionRegex.MatchString(args.Region) {
			failures = append(failures, p.CheckFailure{
				Property: "region",
				Reason:   "must be a valid AWS region identifier",
			})
		}
	}

	// Validate userNameSource
	if args.UserNameSource != nil {
		valid := map[string]bool{"email": true, "sub": true}
		if !valid[*args.UserNameSource] {
			failures = append(failures, p.CheckFailure{
				Property: "userNameSource",
				Reason:   "must be one of: email, sub",
			})
		}
	}

	// Apply defaults
	if len(args.Scopes) == 0 {
		args.Scopes = []string{"openid", "email", "profile"}
	}

	return infer.CheckResponse[CognitoOidcConnectorArgs]{
		Inputs:   args,
		Failures: failures,
	}, nil
}

// Create creates a new Cognito OIDC connector.
func (c *CognitoOidcConnector) Create(ctx context.Context, req infer.CreateRequest[CognitoOidcConnectorArgs]) (infer.CreateResponse[CognitoOidcConnectorState], error) {
	args := req.Inputs

	// In preview/dry-run mode, skip actual Dex API calls and return expected state
	// This check MUST be first, before any other operations or config checks
	if req.DryRun {
		state := CognitoOidcConnectorState{
			CognitoOidcConnectorArgs: args,
		}
		return infer.CreateResponse[CognitoOidcConnectorState]{
			ID:     args.ConnectorId,
			Output: state,
		}, nil
	}

	cfg := infer.GetConfig[DexConfig](ctx)
	if cfg.client == nil {
		return infer.CreateResponse[CognitoOidcConnectorState]{}, fmt.Errorf("Dex client not configured")
	}

	// Derive issuer from region and userPoolId
	issuer := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s", args.Region, args.UserPoolId)

	// Derive userNameKey from userNameSource
	userNameKey := "email" // default
	if args.UserNameSource != nil {
		userNameKey = *args.UserNameSource
	}

	// Build OIDC config
	oidcConfig := map[string]any{
		"issuer":       issuer,
		"clientID":     args.ClientId,
		"clientSecret": args.ClientSecret,
		"redirectURI":  args.RedirectUri,
		"scopes":       args.Scopes,
		"userNameKey":  userNameKey,
	}

	for k, v := range args.ExtraOidc {
		oidcConfig[k] = v
	}

	configBytes, err := json.Marshal(oidcConfig)
	if err != nil {
		return infer.CreateResponse[CognitoOidcConnectorState]{}, fmt.Errorf("failed to marshal OIDC config: %w", err)
	}

	connector := &api.Connector{
		Id:     args.ConnectorId,
		Type:   "oidc",
		Name:   args.Name,
		Config: configBytes,
	}

	createCtx, cancel := context.WithTimeout(ctx, time.Duration(ptrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	resp, err := cfg.client.CreateConnector(createCtx, &api.CreateConnectorReq{
		Connector: connector,
	})
	if err != nil {
		return infer.CreateResponse[CognitoOidcConnectorState]{}, wrapError("create", "cognito-oidc-connector", args.ConnectorId, err)
	}

	if resp.AlreadyExists {
		return infer.CreateResponse[CognitoOidcConnectorState]{}, fmt.Errorf("connector with id %q already exists", args.ConnectorId)
	}

	state := CognitoOidcConnectorState{
		CognitoOidcConnectorArgs: args,
	}

	return infer.CreateResponse[CognitoOidcConnectorState]{
		ID:     args.ConnectorId,
		Output: state,
	}, nil
}

// Read retrieves an existing Cognito OIDC connector.
func (c *CognitoOidcConnector) Read(ctx context.Context, req infer.ReadRequest[CognitoOidcConnectorArgs, CognitoOidcConnectorState]) (infer.ReadResponse[CognitoOidcConnectorArgs, CognitoOidcConnectorState], error) {
	cfg := infer.GetConfig[DexConfig](ctx)
	if cfg.client == nil {
		return infer.ReadResponse[CognitoOidcConnectorArgs, CognitoOidcConnectorState]{}, fmt.Errorf("Dex client not configured")
	}

	listCtx, cancel := context.WithTimeout(ctx, time.Duration(ptrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	listResp, err := cfg.client.ListConnectors(listCtx, &api.ListConnectorReq{})
	if err != nil {
		return infer.ReadResponse[CognitoOidcConnectorArgs, CognitoOidcConnectorState]{}, fmt.Errorf("failed to list connectors: %w", err)
	}

	var found *api.Connector
	for _, conn := range listResp.Connectors {
		if conn.Id == req.ID {
			found = conn
			break
		}
	}

	if found == nil {
		return infer.ReadResponse[CognitoOidcConnectorArgs, CognitoOidcConnectorState]{}, nil
	}

	var configMap map[string]any
	if err := json.Unmarshal(found.Config, &configMap); err != nil {
		return infer.ReadResponse[CognitoOidcConnectorArgs, CognitoOidcConnectorState]{}, nil
	}

	// Extract region and userPoolId from issuer
	issuer, _ := configMap["issuer"].(string)
	region := ""
	userPoolId := ""
	if strings.HasPrefix(issuer, "https://cognito-idp.") && strings.Contains(issuer, ".amazonaws.com/") {
		parts := strings.Split(issuer, "/")
		if len(parts) >= 4 {
			domainParts := strings.Split(parts[2], ".")
			if len(domainParts) >= 2 {
				region = domainParts[1]
			}
			userPoolId = parts[3]
		}
	}

	userNameKey, _ := configMap["userNameKey"].(string)
	userNameSource := &userNameKey
	if userNameKey == "email" {
		// This is the default
	}

	scopes, _ := configMap["scopes"].([]any)
	scopesStr := make([]string, 0, len(scopes))
	for _, s := range scopes {
		if str, ok := s.(string); ok {
			scopesStr = append(scopesStr, str)
		}
	}

	args := CognitoOidcConnectorArgs{
		ConnectorId:    found.Id,
		Name:           found.Name,
		Region:         region,
		UserPoolId:     userPoolId,
		ClientId:       getString(configMap, "clientID"),
		ClientSecret:   getString(configMap, "clientSecret"),
		RedirectUri:    getString(configMap, "redirectURI"),
		Scopes:         scopesStr,
		UserNameSource: userNameSource,
	}

	state := CognitoOidcConnectorState{
		CognitoOidcConnectorArgs: args,
	}

	return infer.ReadResponse[CognitoOidcConnectorArgs, CognitoOidcConnectorState]{
		ID:     found.Id,
		Inputs: args,
		State:  state,
	}, nil
}

// Update updates an existing Cognito OIDC connector.
func (c *CognitoOidcConnector) Update(ctx context.Context, req infer.UpdateRequest[CognitoOidcConnectorArgs, CognitoOidcConnectorState]) (infer.UpdateResponse[CognitoOidcConnectorState], error) {
	args := req.Inputs
	oldState := req.State

	// In preview/dry-run mode, skip actual Dex API calls and return expected state
	// This check MUST be first, before any other operations or config checks
	if req.DryRun {
		state := CognitoOidcConnectorState{
			CognitoOidcConnectorArgs: args,
		}
		return infer.UpdateResponse[CognitoOidcConnectorState]{
			Output: state,
		}, nil
	}

	cfg := infer.GetConfig[DexConfig](ctx)
	if cfg.client == nil {
		return infer.UpdateResponse[CognitoOidcConnectorState]{}, fmt.Errorf("Dex client not configured")
	}

	if args.ConnectorId != oldState.ConnectorId {
		return infer.UpdateResponse[CognitoOidcConnectorState]{}, fmt.Errorf("connectorId cannot be changed")
	}
	if args.Region != oldState.Region || args.UserPoolId != oldState.UserPoolId {
		return infer.UpdateResponse[CognitoOidcConnectorState]{}, fmt.Errorf("region and userPoolId cannot be changed (would require replace)")
	}

	issuer := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s", args.Region, args.UserPoolId)
	userNameKey := "email"
	if args.UserNameSource != nil {
		userNameKey = *args.UserNameSource
	}

	oidcConfig := map[string]any{
		"issuer":       issuer,
		"clientID":     args.ClientId,
		"clientSecret": args.ClientSecret,
		"redirectURI":  args.RedirectUri,
		"scopes":       args.Scopes,
		"userNameKey":  userNameKey,
	}

	for k, v := range args.ExtraOidc {
		oidcConfig[k] = v
	}

	configBytes, err := json.Marshal(oidcConfig)
	if err != nil {
		return infer.UpdateResponse[CognitoOidcConnectorState]{}, fmt.Errorf("failed to marshal OIDC config: %w", err)
	}

	updateCtx, cancel := context.WithTimeout(ctx, time.Duration(ptrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	_, err = cfg.client.UpdateConnector(updateCtx, &api.UpdateConnectorReq{
		Id:        args.ConnectorId,
		NewType:   "oidc",
		NewName:   args.Name,
		NewConfig: configBytes,
	})
	if err != nil {
		return infer.UpdateResponse[CognitoOidcConnectorState]{}, wrapError("update", "cognito-oidc-connector", args.ConnectorId, err)
	}

	state := CognitoOidcConnectorState{
		CognitoOidcConnectorArgs: args,
	}

	return infer.UpdateResponse[CognitoOidcConnectorState]{
		Output: state,
	}, nil
}

// Delete deletes a Cognito OIDC connector.
func (c *CognitoOidcConnector) Delete(ctx context.Context, req infer.DeleteRequest[CognitoOidcConnectorState]) (infer.DeleteResponse, error) {
	cfg := infer.GetConfig[DexConfig](ctx)
	if cfg.client == nil {
		return infer.DeleteResponse{}, fmt.Errorf("Dex client not configured")
	}

	deleteID := req.ID
	if deleteID == "" && req.State.ConnectorId != "" {
		deleteID = req.State.ConnectorId
	}

	// Note: Pulumi does not call Delete during preview, so no preview check needed

	deleteCtx, cancel := context.WithTimeout(ctx, time.Duration(ptrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	_, err := cfg.client.DeleteConnector(deleteCtx, &api.DeleteConnectorReq{
		Id: deleteID,
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return infer.DeleteResponse{}, nil
		}
		return infer.DeleteResponse{}, wrapError("delete", "cognito-oidc-connector", deleteID, err)
	}

	return infer.DeleteResponse{}, nil
}

// ============================================================================
// Helper functions
// ============================================================================

// getString extracts a string value from a map, returning empty string if not found.
func getString(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if str, ok := v.(string); ok {
			return str
		}
	}
	return ""
}

// getStringPtr extracts a string value from a map, returning nil if not found or empty.
func getStringPtr(m map[string]any, key string) *string {
	if v, ok := m[key]; ok {
		if str, ok := v.(string); ok && str != "" {
			return &str
		}
	}
	return nil
}
