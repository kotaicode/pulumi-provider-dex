package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	api "github.com/dexidp/dex/api/v2"
	"github.com/kotaicode/pulumi-provider-dex/pkg/provider"
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
	a.Describe(c, "Manages an Azure AD/Entra ID connector in Dex using the generic OIDC connector (type: oidc). This connector allows users to authenticate using their Azure AD/Entra ID credentials.")
}

// Annotate provides schema metadata for AzureOidcConnectorArgs.
func (c *AzureOidcConnectorArgs) Annotate(a infer.Annotator) {
	a.Describe(&c.ConnectorId, "Unique identifier for the Azure connector.")
	a.Describe(&c.Name, "Human-readable name for the connector, displayed to users during login.")
	a.Describe(&c.TenantId, "Azure AD tenant ID (UUID format). This identifies your Azure AD organization.")
	a.Describe(&c.ClientId, "Azure AD application (client) ID.")
	a.Describe(&c.ClientSecret, "Azure AD application client secret.")
	a.Describe(&c.RedirectUri, "Redirect URI registered in Azure AD. Must match Dex's callback URL (typically 'https://dex.example.com/callback').")
	a.Describe(&c.Scopes, "OIDC scopes to request from Azure AD. Defaults to ['openid', 'profile', 'email', 'offline_access'] if not specified.")
	a.Describe(&c.UserNameSource, "Source for the username claim. Valid values: 'preferred_username' (default), 'upn' (User Principal Name), or 'email'.")
	a.Describe(&c.ExtraOidc, "Additional OIDC configuration fields as key-value pairs for advanced scenarios.")
}

// Annotate provides schema metadata for AzureOidcConnectorState.
func (c *AzureOidcConnectorState) Annotate(a infer.Annotator) {
	// AzureOidcConnectorState embeds AzureOidcConnectorArgs, so field descriptions are inherited
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

	cfg := infer.GetConfig[provider.DexConfig](ctx)
	if cfg.Client == nil {
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

	createCtx, cancel := context.WithTimeout(ctx, time.Duration(provider.PtrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	resp, err := cfg.Client.CreateConnector(createCtx, &api.CreateConnectorReq{
		Connector: connector,
	})
	if err != nil {
		return infer.CreateResponse[AzureOidcConnectorState]{}, provider.WrapError("create", "azure-oidc-connector", args.ConnectorId, err)
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
	cfg := infer.GetConfig[provider.DexConfig](ctx)
	if cfg.Client == nil {
		return infer.ReadResponse[AzureOidcConnectorArgs, AzureOidcConnectorState]{}, fmt.Errorf("Dex client not configured")
	}

	// List connectors and find by ID
	listCtx, cancel := context.WithTimeout(ctx, time.Duration(provider.PtrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	listResp, err := cfg.Client.ListConnectors(listCtx, &api.ListConnectorReq{})
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
		ClientId:       GetString(configMap, "clientID"),
		ClientSecret:   GetString(configMap, "clientSecret"),
		RedirectUri:    GetString(configMap, "redirectURI"),
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

	cfg := infer.GetConfig[provider.DexConfig](ctx)
	if cfg.Client == nil {
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

	updateCtx, cancel := context.WithTimeout(ctx, time.Duration(provider.PtrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	_, err = cfg.Client.UpdateConnector(updateCtx, &api.UpdateConnectorReq{
		Id:        args.ConnectorId,
		NewType:   "oidc",
		NewName:   args.Name,
		NewConfig: configBytes,
	})
	if err != nil {
		return infer.UpdateResponse[AzureOidcConnectorState]{}, provider.WrapError("update", "azure-oidc-connector", args.ConnectorId, err)
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
	cfg := infer.GetConfig[provider.DexConfig](ctx)
	if cfg.Client == nil {
		return infer.DeleteResponse{}, fmt.Errorf("Dex client not configured")
	}

	deleteID := req.ID
	if deleteID == "" && req.State.ConnectorId != "" {
		deleteID = req.State.ConnectorId
	}

	// Note: Pulumi does not call Delete during preview, so no preview check needed

	deleteCtx, cancel := context.WithTimeout(ctx, time.Duration(provider.PtrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	_, err := cfg.Client.DeleteConnector(deleteCtx, &api.DeleteConnectorReq{
		Id: deleteID,
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return infer.DeleteResponse{}, nil // Already deleted
		}
		return infer.DeleteResponse{}, provider.WrapError("delete", "azure-oidc-connector", deleteID, err)
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
	a.Describe(c, "Manages an Azure AD/Entra ID connector in Dex using the Microsoft-specific connector (type: microsoft). This connector provides Microsoft-specific features like group filtering and domain restrictions.")
}

// Annotate provides schema metadata for AzureMicrosoftConnectorArgs.
func (c *AzureMicrosoftConnectorArgs) Annotate(a infer.Annotator) {
	a.Describe(&c.ConnectorId, "Unique identifier for the Azure Microsoft connector.")
	a.Describe(&c.Name, "Human-readable name for the connector, displayed to users during login.")
	a.Describe(&c.Tenant, "Azure AD tenant identifier. Can be 'common' (any Azure AD account), 'organizations' (any organizational account), or a specific tenant ID (UUID format).")
	a.Describe(&c.ClientId, "Azure AD application (client) ID.")
	a.Describe(&c.ClientSecret, "Azure AD application client secret.")
	a.Describe(&c.RedirectUri, "Redirect URI registered in Azure AD. Must match Dex's callback URL.")
	a.Describe(&c.Groups, "Name of the claim that contains group memberships (e.g., 'groups'). Used for group-based access control.")
}

// Annotate provides schema metadata for AzureMicrosoftConnectorState.
func (c *AzureMicrosoftConnectorState) Annotate(a infer.Annotator) {
	// AzureMicrosoftConnectorState embeds AzureMicrosoftConnectorArgs, so field descriptions are inherited
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

	cfg := infer.GetConfig[provider.DexConfig](ctx)
	if cfg.Client == nil {
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

	createCtx, cancel := context.WithTimeout(ctx, time.Duration(provider.PtrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	resp, err := cfg.Client.CreateConnector(createCtx, &api.CreateConnectorReq{
		Connector: connector,
	})
	if err != nil {
		return infer.CreateResponse[AzureMicrosoftConnectorState]{}, provider.WrapError("create", "azure-microsoft-connector", args.ConnectorId, err)
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
	cfg := infer.GetConfig[provider.DexConfig](ctx)
	if cfg.Client == nil {
		return infer.ReadResponse[AzureMicrosoftConnectorArgs, AzureMicrosoftConnectorState]{}, fmt.Errorf("Dex client not configured")
	}

	listCtx, cancel := context.WithTimeout(ctx, time.Duration(provider.PtrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	listResp, err := cfg.Client.ListConnectors(listCtx, &api.ListConnectorReq{})
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

	groups := GetStringPtr(configMap, "groups")

	args := AzureMicrosoftConnectorArgs{
		ConnectorId:  found.Id,
		Name:         found.Name,
		Tenant:       GetString(configMap, "tenant"),
		ClientId:     GetString(configMap, "clientID"),
		ClientSecret: GetString(configMap, "clientSecret"),
		RedirectUri:  GetString(configMap, "redirectURI"),
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

	cfg := infer.GetConfig[provider.DexConfig](ctx)
	if cfg.Client == nil {
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

	updateCtx, cancel := context.WithTimeout(ctx, time.Duration(provider.PtrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	_, err = cfg.Client.UpdateConnector(updateCtx, &api.UpdateConnectorReq{
		Id:        args.ConnectorId,
		NewType:   "microsoft",
		NewName:   args.Name,
		NewConfig: configBytes,
	})
	if err != nil {
		return infer.UpdateResponse[AzureMicrosoftConnectorState]{}, provider.WrapError("update", "azure-microsoft-connector", args.ConnectorId, err)
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
	cfg := infer.GetConfig[provider.DexConfig](ctx)
	if cfg.Client == nil {
		return infer.DeleteResponse{}, fmt.Errorf("Dex client not configured")
	}

	deleteID := req.ID
	if deleteID == "" && req.State.ConnectorId != "" {
		deleteID = req.State.ConnectorId
	}

	// Note: Pulumi does not call Delete during preview, so no preview check needed

	deleteCtx, cancel := context.WithTimeout(ctx, time.Duration(provider.PtrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	_, err := cfg.Client.DeleteConnector(deleteCtx, &api.DeleteConnectorReq{
		Id: deleteID,
	})
	if err != nil {
		if status.Code(err) == codes.NotFound {
			return infer.DeleteResponse{}, nil
		}
		return infer.DeleteResponse{}, provider.WrapError("delete", "azure-microsoft-connector", deleteID, err)
	}

	return infer.DeleteResponse{}, nil
}
