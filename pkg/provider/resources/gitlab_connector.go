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
// GitLabConnector - GitLab OAuth2 connector (type: "gitlab")
// ============================================================================

// GitLabConnectorArgs defines inputs for GitLabConnector.
type GitLabConnectorArgs struct {
	ConnectorId         string   `pulumi:"connectorId"`
	Name                string   `pulumi:"name"`
	BaseURL             *string  `pulumi:"baseURL,optional"`
	ClientId            string   `pulumi:"clientId"`
	ClientSecret        string   `pulumi:"clientSecret" provider:"secret"`
	RedirectUri         string   `pulumi:"redirectUri"`
	Groups              []string `pulumi:"groups,optional"`
	UseLoginAsID        *bool    `pulumi:"useLoginAsID,optional"`
	GetGroupsPermission *bool    `pulumi:"getGroupsPermission,optional"`
}

// GitLabConnectorState defines outputs for GitLabConnector.
type GitLabConnectorState struct {
	GitLabConnectorArgs
}

// GitLabConnector manages a GitLab connector in Dex.
type GitLabConnector struct{}

// Annotate provides schema metadata.
func (c *GitLabConnector) Annotate(a infer.Annotator) {
	a.Describe(c, "Manages a GitLab connector in Dex. This connector allows users to authenticate using their GitLab accounts and supports group-based access control.")
}

// Annotate provides schema metadata for GitLabConnectorArgs.
func (c *GitLabConnectorArgs) Annotate(a infer.Annotator) {
	a.Describe(&c.ConnectorId, "Unique identifier for the GitLab connector.")
	a.Describe(&c.Name, "Human-readable name for the connector, displayed to users during login.")
	a.Describe(&c.BaseURL, "GitLab instance base URL. Defaults to 'https://gitlab.com' for GitLab.com.")
	a.Describe(&c.ClientId, "GitLab OAuth application client ID.")
	a.Describe(&c.ClientSecret, "GitLab OAuth application client secret.")
	a.Describe(&c.RedirectUri, "Redirect URI registered in GitLab OAuth app. Must match Dex's callback URL.")
	a.Describe(&c.Groups, "List of GitLab group names. Only users in these groups will be allowed to authenticate.")
	a.Describe(&c.UseLoginAsID, "If true, use GitLab username as the user ID. Defaults to false.")
	a.Describe(&c.GetGroupsPermission, "If true, request 'read_api' scope to fetch group memberships. Defaults to false.")
}

// Annotate provides schema metadata for GitLabConnectorState.
func (c *GitLabConnectorState) Annotate(a infer.Annotator) {
	// GitLabConnectorState embeds GitLabConnectorArgs, so field descriptions are inherited
}

// Check validates inputs.
func (c *GitLabConnector) Check(ctx context.Context, req infer.CheckRequest) (infer.CheckResponse[GitLabConnectorArgs], error) {
	args, failures, err := infer.DefaultCheck[GitLabConnectorArgs](ctx, req.NewInputs)
	if err != nil {
		return infer.CheckResponse[GitLabConnectorArgs]{Failures: failures}, err
	}

	// Apply defaults
	if args.BaseURL == nil || *args.BaseURL == "" {
		defaultURL := "https://gitlab.com"
		args.BaseURL = &defaultURL
	}
	if args.UseLoginAsID == nil {
		defaultUseLogin := false
		args.UseLoginAsID = &defaultUseLogin
	}
	if args.GetGroupsPermission == nil {
		defaultGetGroups := false
		args.GetGroupsPermission = &defaultGetGroups
	}

	return infer.CheckResponse[GitLabConnectorArgs]{
		Inputs:   args,
		Failures: failures,
	}, nil
}

// Create creates a new GitLab connector.
func (c *GitLabConnector) Create(ctx context.Context, req infer.CreateRequest[GitLabConnectorArgs]) (infer.CreateResponse[GitLabConnectorState], error) {
	args := req.Inputs

	// In preview/dry-run mode, skip actual Dex API calls and return expected state
	if req.DryRun {
		state := GitLabConnectorState{
			GitLabConnectorArgs: args,
		}
		return infer.CreateResponse[GitLabConnectorState]{
			ID:     args.ConnectorId,
			Output: state,
		}, nil
	}

	cfg := infer.GetConfig[provider.DexConfig](ctx)
	if cfg.Client == nil {
		return infer.CreateResponse[GitLabConnectorState]{}, fmt.Errorf("Dex client not configured")
	}

	// Build GitLab connector config
	gitlabConfig := map[string]any{
		"clientID":     args.ClientId,
		"clientSecret": args.ClientSecret,
		"redirectURI":  args.RedirectUri,
	}

	if args.BaseURL != nil && *args.BaseURL != "" {
		gitlabConfig["baseURL"] = *args.BaseURL
	}
	if len(args.Groups) > 0 {
		gitlabConfig["groups"] = args.Groups
	}
	if args.UseLoginAsID != nil {
		gitlabConfig["useLoginAsID"] = *args.UseLoginAsID
	}
	if args.GetGroupsPermission != nil {
		gitlabConfig["getGroupsPermission"] = *args.GetGroupsPermission
	}

	configBytes, err := json.Marshal(gitlabConfig)
	if err != nil {
		return infer.CreateResponse[GitLabConnectorState]{}, fmt.Errorf("failed to marshal GitLab config: %w", err)
	}

	connector := &api.Connector{
		Id:     args.ConnectorId,
		Type:   "gitlab",
		Name:   args.Name,
		Config: configBytes,
	}

	createCtx, cancel := context.WithTimeout(ctx, time.Duration(provider.PtrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	resp, err := cfg.Client.CreateConnector(createCtx, &api.CreateConnectorReq{
		Connector: connector,
	})
	if err != nil {
		return infer.CreateResponse[GitLabConnectorState]{}, provider.WrapError("create", "gitlab-connector", args.ConnectorId, err)
	}

	if resp.AlreadyExists {
		return infer.CreateResponse[GitLabConnectorState]{}, fmt.Errorf("connector with id %q already exists", args.ConnectorId)
	}

	state := GitLabConnectorState{
		GitLabConnectorArgs: args,
	}

	return infer.CreateResponse[GitLabConnectorState]{
		ID:     args.ConnectorId,
		Output: state,
	}, nil
}

// Read retrieves an existing GitLab connector.
func (c *GitLabConnector) Read(ctx context.Context, req infer.ReadRequest[GitLabConnectorArgs, GitLabConnectorState]) (infer.ReadResponse[GitLabConnectorArgs, GitLabConnectorState], error) {
	cfg := infer.GetConfig[provider.DexConfig](ctx)
	if cfg.Client == nil {
		return infer.ReadResponse[GitLabConnectorArgs, GitLabConnectorState]{}, fmt.Errorf("Dex client not configured")
	}

	listCtx, cancel := context.WithTimeout(ctx, time.Duration(provider.PtrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	listResp, err := cfg.Client.ListConnectors(listCtx, &api.ListConnectorReq{})
	if err != nil {
		return infer.ReadResponse[GitLabConnectorArgs, GitLabConnectorState]{}, fmt.Errorf("failed to list connectors: %w", err)
	}

	var found *api.Connector
	for _, conn := range listResp.Connectors {
		if conn.Id == req.ID {
			found = conn
			break
		}
	}

	if found == nil {
		return infer.ReadResponse[GitLabConnectorArgs, GitLabConnectorState]{}, nil
	}

	var configMap map[string]any
	if err := json.Unmarshal(found.Config, &configMap); err != nil {
		return infer.ReadResponse[GitLabConnectorArgs, GitLabConnectorState]{}, nil
	}

	// Parse groups array
	var groups []string
	if groupsVal, ok := configMap["groups"].([]any); ok {
		for _, g := range groupsVal {
			if str, ok := g.(string); ok {
				groups = append(groups, str)
			}
		}
	}

	baseURL := GetStringPtr(configMap, "baseURL")
	useLoginAsID := GetBoolPtr(configMap, "useLoginAsID")
	getGroupsPermission := GetBoolPtr(configMap, "getGroupsPermission")

	args := GitLabConnectorArgs{
		ConnectorId:         found.Id,
		Name:                found.Name,
		BaseURL:             baseURL,
		ClientId:            GetString(configMap, "clientID"),
		ClientSecret:        GetString(configMap, "clientSecret"),
		RedirectUri:         GetString(configMap, "redirectURI"),
		Groups:              groups,
		UseLoginAsID:        useLoginAsID,
		GetGroupsPermission: getGroupsPermission,
	}

	state := GitLabConnectorState{
		GitLabConnectorArgs: args,
	}

	return infer.ReadResponse[GitLabConnectorArgs, GitLabConnectorState]{
		ID:     found.Id,
		Inputs: args,
		State:  state,
	}, nil
}

// Update updates an existing GitLab connector.
func (c *GitLabConnector) Update(ctx context.Context, req infer.UpdateRequest[GitLabConnectorArgs, GitLabConnectorState]) (infer.UpdateResponse[GitLabConnectorState], error) {
	args := req.Inputs
	oldState := req.State

	// In preview/dry-run mode, skip actual Dex API calls and return expected state
	if req.DryRun {
		state := GitLabConnectorState{
			GitLabConnectorArgs: args,
		}
		return infer.UpdateResponse[GitLabConnectorState]{
			Output: state,
		}, nil
	}

	cfg := infer.GetConfig[provider.DexConfig](ctx)
	if cfg.Client == nil {
		return infer.UpdateResponse[GitLabConnectorState]{}, fmt.Errorf("Dex client not configured")
	}

	if args.ConnectorId != oldState.ConnectorId {
		return infer.UpdateResponse[GitLabConnectorState]{}, fmt.Errorf("connectorId cannot be changed")
	}
	if args.BaseURL != nil && oldState.BaseURL != nil && *args.BaseURL != *oldState.BaseURL {
		return infer.UpdateResponse[GitLabConnectorState]{}, fmt.Errorf("baseURL cannot be changed (would require replace)")
	}

	gitlabConfig := map[string]any{
		"clientID":     args.ClientId,
		"clientSecret": args.ClientSecret,
		"redirectURI":  args.RedirectUri,
	}

	if args.BaseURL != nil && *args.BaseURL != "" {
		gitlabConfig["baseURL"] = *args.BaseURL
	}
	if len(args.Groups) > 0 {
		gitlabConfig["groups"] = args.Groups
	}
	if args.UseLoginAsID != nil {
		gitlabConfig["useLoginAsID"] = *args.UseLoginAsID
	}
	if args.GetGroupsPermission != nil {
		gitlabConfig["getGroupsPermission"] = *args.GetGroupsPermission
	}

	configBytes, err := json.Marshal(gitlabConfig)
	if err != nil {
		return infer.UpdateResponse[GitLabConnectorState]{}, fmt.Errorf("failed to marshal GitLab config: %w", err)
	}

	updateCtx, cancel := context.WithTimeout(ctx, time.Duration(provider.PtrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	_, err = cfg.Client.UpdateConnector(updateCtx, &api.UpdateConnectorReq{
		Id:        args.ConnectorId,
		NewType:   "gitlab",
		NewName:   args.Name,
		NewConfig: configBytes,
	})
	if err != nil {
		return infer.UpdateResponse[GitLabConnectorState]{}, provider.WrapError("update", "gitlab-connector", args.ConnectorId, err)
	}

	state := GitLabConnectorState{
		GitLabConnectorArgs: args,
	}

	return infer.UpdateResponse[GitLabConnectorState]{
		Output: state,
	}, nil
}

// Delete deletes a GitLab connector.
func (c *GitLabConnector) Delete(ctx context.Context, req infer.DeleteRequest[GitLabConnectorState]) (infer.DeleteResponse, error) {
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
		return infer.DeleteResponse{}, provider.WrapError("delete", "gitlab-connector", deleteID, err)
	}

	return infer.DeleteResponse{}, nil
}
