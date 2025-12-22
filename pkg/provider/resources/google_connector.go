package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	api "github.com/dexidp/dex/api/v2"
	"github.com/kotaicode/pulumi-provider-dex/pkg/provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ============================================================================
// GoogleConnector - Google OpenID Connect connector (type: "google")
// ============================================================================

// GoogleConnectorArgs defines inputs for GoogleConnector.
type GoogleConnectorArgs struct {
	ConnectorId            string            `pulumi:"connectorId"`
	Name                   string            `pulumi:"name"`
	ClientId               string            `pulumi:"clientId"`
	ClientSecret           string            `pulumi:"clientSecret" provider:"secret"`
	RedirectUri            string            `pulumi:"redirectUri"`
	PromptType             *string           `pulumi:"promptType,optional"`
	HostedDomains          []string          `pulumi:"hostedDomains,optional"`
	Groups                 []string          `pulumi:"groups,optional"`
	ServiceAccountFilePath *string           `pulumi:"serviceAccountFilePath,optional"`
	DomainToAdminEmail     map[string]string `pulumi:"domainToAdminEmail,optional"`
}

// GoogleConnectorState defines outputs for GoogleConnector.
type GoogleConnectorState struct {
	GoogleConnectorArgs
}

// GoogleConnector manages a Google connector in Dex.
type GoogleConnector struct{}

// Annotate provides schema metadata.
func (c *GoogleConnector) Annotate(a infer.Annotator) {
	a.Describe(c, "Manages a Google connector in Dex. This connector allows users to authenticate using their Google accounts and supports domain and group-based access control.")
}

// Annotate provides schema metadata for GoogleConnectorArgs.
func (c *GoogleConnectorArgs) Annotate(a infer.Annotator) {
	a.Describe(&c.ConnectorId, "Unique identifier for the Google connector.")
	a.Describe(&c.Name, "Human-readable name for the connector, displayed to users during login.")
	a.Describe(&c.ClientId, "Google OAuth client ID.")
	a.Describe(&c.ClientSecret, "Google OAuth client secret.")
	a.Describe(&c.RedirectUri, "Redirect URI registered in Google OAuth app. Must match Dex's callback URL.")
	a.Describe(&c.PromptType, "OAuth prompt type. Valid values: 'consent' (default) or 'select_account'.")
	a.Describe(&c.HostedDomains, "List of Google Workspace domains. Only users with email addresses in these domains will be allowed to authenticate.")
	a.Describe(&c.Groups, "List of Google Groups. Only users in these groups will be allowed to authenticate.")
	a.Describe(&c.ServiceAccountFilePath, "Path to Google service account JSON file. Required for group-based access control.")
	a.Describe(&c.DomainToAdminEmail, "Map of domain names to admin email addresses. Used for group lookups in Google Workspace.")
}

// Annotate provides schema metadata for GoogleConnectorState.
func (c *GoogleConnectorState) Annotate(a infer.Annotator) {
	// GoogleConnectorState embeds GoogleConnectorArgs, so field descriptions are inherited
}

// Check validates inputs.
func (c *GoogleConnector) Check(ctx context.Context, req infer.CheckRequest) (infer.CheckResponse[GoogleConnectorArgs], error) {
	args, failures, err := infer.DefaultCheck[GoogleConnectorArgs](ctx, req.NewInputs)
	if err != nil {
		return infer.CheckResponse[GoogleConnectorArgs]{Failures: failures}, err
	}

	// Apply defaults
	if args.PromptType == nil || *args.PromptType == "" {
		defaultPrompt := "consent"
		args.PromptType = &defaultPrompt
	}

	return infer.CheckResponse[GoogleConnectorArgs]{
		Inputs:   args,
		Failures: failures,
	}, nil
}

// Create creates a new Google connector.
func (c *GoogleConnector) Create(ctx context.Context, req infer.CreateRequest[GoogleConnectorArgs]) (infer.CreateResponse[GoogleConnectorState], error) {
	args := req.Inputs

	// In preview/dry-run mode, skip actual Dex API calls and return expected state
	if req.DryRun {
		state := GoogleConnectorState{
			GoogleConnectorArgs: args,
		}
		return infer.CreateResponse[GoogleConnectorState]{
			ID:     args.ConnectorId,
			Output: state,
		}, nil
	}

	cfg := infer.GetConfig[provider.DexConfig](ctx)
	if cfg.Client == nil {
		return infer.CreateResponse[GoogleConnectorState]{}, fmt.Errorf("Dex client not configured")
	}

	// Build Google connector config
	googleConfig := map[string]any{
		"clientID":     args.ClientId,
		"clientSecret": args.ClientSecret,
		"redirectURI":  args.RedirectUri,
	}

	if args.PromptType != nil {
		googleConfig["promptType"] = *args.PromptType
	}
	if len(args.HostedDomains) > 0 {
		googleConfig["hostedDomains"] = args.HostedDomains
	}
	if len(args.Groups) > 0 {
		googleConfig["groups"] = args.Groups
	}
	if args.ServiceAccountFilePath != nil {
		googleConfig["serviceAccountFilePath"] = *args.ServiceAccountFilePath
	}
	if len(args.DomainToAdminEmail) > 0 {
		googleConfig["domainToAdminEmail"] = args.DomainToAdminEmail
	}

	configBytes, err := json.Marshal(googleConfig)
	if err != nil {
		return infer.CreateResponse[GoogleConnectorState]{}, fmt.Errorf("failed to marshal Google config: %w", err)
	}

	connector := &api.Connector{
		Id:     args.ConnectorId,
		Type:   "google",
		Name:   args.Name,
		Config: configBytes,
	}

	createCtx, cancel := context.WithTimeout(ctx, time.Duration(provider.PtrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	resp, err := cfg.Client.CreateConnector(createCtx, &api.CreateConnectorReq{
		Connector: connector,
	})
	if err != nil {
		return infer.CreateResponse[GoogleConnectorState]{}, provider.WrapError("create", "google-connector", args.ConnectorId, err)
	}

	if resp.AlreadyExists {
		return infer.CreateResponse[GoogleConnectorState]{}, fmt.Errorf("connector with id %q already exists", args.ConnectorId)
	}

	state := GoogleConnectorState{
		GoogleConnectorArgs: args,
	}

	return infer.CreateResponse[GoogleConnectorState]{
		ID:     args.ConnectorId,
		Output: state,
	}, nil
}

// Read retrieves an existing Google connector.
func (c *GoogleConnector) Read(ctx context.Context, req infer.ReadRequest[GoogleConnectorArgs, GoogleConnectorState]) (infer.ReadResponse[GoogleConnectorArgs, GoogleConnectorState], error) {
	cfg := infer.GetConfig[provider.DexConfig](ctx)
	if cfg.Client == nil {
		return infer.ReadResponse[GoogleConnectorArgs, GoogleConnectorState]{}, fmt.Errorf("Dex client not configured")
	}

	listCtx, cancel := context.WithTimeout(ctx, time.Duration(provider.PtrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	listResp, err := cfg.Client.ListConnectors(listCtx, &api.ListConnectorReq{})
	if err != nil {
		return infer.ReadResponse[GoogleConnectorArgs, GoogleConnectorState]{}, fmt.Errorf("failed to list connectors: %w", err)
	}

	var found *api.Connector
	for _, conn := range listResp.Connectors {
		if conn.Id == req.ID {
			found = conn
			break
		}
	}

	if found == nil {
		return infer.ReadResponse[GoogleConnectorArgs, GoogleConnectorState]{}, nil
	}

	var configMap map[string]any
	if err := json.Unmarshal(found.Config, &configMap); err != nil {
		return infer.ReadResponse[GoogleConnectorArgs, GoogleConnectorState]{}, nil
	}

	// Parse arrays
	var hostedDomains []string
	if domainsVal, ok := configMap["hostedDomains"].([]any); ok {
		for _, d := range domainsVal {
			if str, ok := d.(string); ok {
				hostedDomains = append(hostedDomains, str)
			}
		}
	}

	var groups []string
	if groupsVal, ok := configMap["groups"].([]any); ok {
		for _, g := range groupsVal {
			if str, ok := g.(string); ok {
				groups = append(groups, str)
			}
		}
	}

	// Parse domainToAdminEmail map
	domainToAdminEmail := make(map[string]string)
	if domainMap, ok := configMap["domainToAdminEmail"].(map[string]any); ok {
		for k, v := range domainMap {
			if str, ok := v.(string); ok {
				domainToAdminEmail[k] = str
			}
		}
	}

	args := GoogleConnectorArgs{
		ConnectorId:            found.Id,
		Name:                   found.Name,
		ClientId:               GetString(configMap, "clientID"),
		ClientSecret:           GetString(configMap, "clientSecret"),
		RedirectUri:            GetString(configMap, "redirectURI"),
		PromptType:             GetStringPtr(configMap, "promptType"),
		HostedDomains:          hostedDomains,
		Groups:                 groups,
		ServiceAccountFilePath: GetStringPtr(configMap, "serviceAccountFilePath"),
		DomainToAdminEmail:     domainToAdminEmail,
	}

	state := GoogleConnectorState{
		GoogleConnectorArgs: args,
	}

	return infer.ReadResponse[GoogleConnectorArgs, GoogleConnectorState]{
		ID:     found.Id,
		Inputs: args,
		State:  state,
	}, nil
}

// Update updates an existing Google connector.
func (c *GoogleConnector) Update(ctx context.Context, req infer.UpdateRequest[GoogleConnectorArgs, GoogleConnectorState]) (infer.UpdateResponse[GoogleConnectorState], error) {
	args := req.Inputs
	oldState := req.State

	// In preview/dry-run mode, skip actual Dex API calls and return expected state
	if req.DryRun {
		state := GoogleConnectorState{
			GoogleConnectorArgs: args,
		}
		return infer.UpdateResponse[GoogleConnectorState]{
			Output: state,
		}, nil
	}

	cfg := infer.GetConfig[provider.DexConfig](ctx)
	if cfg.Client == nil {
		return infer.UpdateResponse[GoogleConnectorState]{}, fmt.Errorf("Dex client not configured")
	}

	if args.ConnectorId != oldState.ConnectorId {
		return infer.UpdateResponse[GoogleConnectorState]{}, fmt.Errorf("connectorId cannot be changed")
	}

	googleConfig := map[string]any{
		"clientID":     args.ClientId,
		"clientSecret": args.ClientSecret,
		"redirectURI":  args.RedirectUri,
	}

	if args.PromptType != nil {
		googleConfig["promptType"] = *args.PromptType
	}
	if len(args.HostedDomains) > 0 {
		googleConfig["hostedDomains"] = args.HostedDomains
	}
	if len(args.Groups) > 0 {
		googleConfig["groups"] = args.Groups
	}
	if args.ServiceAccountFilePath != nil {
		googleConfig["serviceAccountFilePath"] = *args.ServiceAccountFilePath
	}
	if len(args.DomainToAdminEmail) > 0 {
		googleConfig["domainToAdminEmail"] = args.DomainToAdminEmail
	}

	configBytes, err := json.Marshal(googleConfig)
	if err != nil {
		return infer.UpdateResponse[GoogleConnectorState]{}, fmt.Errorf("failed to marshal Google config: %w", err)
	}

	updateCtx, cancel := context.WithTimeout(ctx, time.Duration(provider.PtrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	_, err = cfg.Client.UpdateConnector(updateCtx, &api.UpdateConnectorReq{
		Id:        args.ConnectorId,
		NewType:   "google",
		NewName:   args.Name,
		NewConfig: configBytes,
	})
	if err != nil {
		return infer.UpdateResponse[GoogleConnectorState]{}, provider.WrapError("update", "google-connector", args.ConnectorId, err)
	}

	state := GoogleConnectorState{
		GoogleConnectorArgs: args,
	}

	return infer.UpdateResponse[GoogleConnectorState]{
		Output: state,
	}, nil
}

// Delete deletes a Google connector.
func (c *GoogleConnector) Delete(ctx context.Context, req infer.DeleteRequest[GoogleConnectorState]) (infer.DeleteResponse, error) {
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
		return infer.DeleteResponse{}, provider.WrapError("delete", "google-connector", deleteID, err)
	}

	return infer.DeleteResponse{}, nil
}
