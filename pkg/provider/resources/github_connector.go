package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	api "github.com/dexidp/dex/api/v2"
	"github.com/kotaicode/pulumi-provider-dex/pkg/provider"
	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ============================================================================
// GitHubConnector - GitHub OAuth2 connector (type: "github")
// ============================================================================

// GitHubOrg represents a GitHub organization with optional teams.
type GitHubOrg struct {
	Name  string   `pulumi:"name"`
	Teams []string `pulumi:"teams,optional"`
}

// GitHubConnectorArgs defines inputs for GitHubConnector.
type GitHubConnectorArgs struct {
	ConnectorId          string      `pulumi:"connectorId"`
	Name                 string      `pulumi:"name"`
	ClientId             string      `pulumi:"clientId"`
	ClientSecret         string      `pulumi:"clientSecret" provider:"secret"`
	RedirectUri          string      `pulumi:"redirectUri"`
	Orgs                 []GitHubOrg `pulumi:"orgs,optional"`
	LoadAllGroups        *bool       `pulumi:"loadAllGroups,optional"`
	TeamNameField        *string     `pulumi:"teamNameField,optional"`
	UseLoginAsID         *bool       `pulumi:"useLoginAsID,optional"`
	PreferredEmailDomain *string     `pulumi:"preferredEmailDomain,optional"`
	HostName             *string     `pulumi:"hostName,optional"` // For GitHub Enterprise
	RootCA               *string     `pulumi:"rootCA,optional"`   // For GitHub Enterprise
}

// GitHubConnectorState defines outputs for GitHubConnector.
type GitHubConnectorState struct {
	GitHubConnectorArgs
}

// GitHubConnector manages a GitHub connector in Dex.
type GitHubConnector struct{}

// Annotate provides schema metadata.
func (c *GitHubConnector) Annotate(a infer.Annotator) {
	a.Describe(c, "Manages a GitHub connector in Dex. This connector allows users to authenticate using their GitHub accounts and supports organization and team-based access control.")
}

// Annotate provides schema metadata for GitHubConnectorArgs.
func (c *GitHubConnectorArgs) Annotate(a infer.Annotator) {
	a.Describe(&c.ConnectorId, "Unique identifier for the GitHub connector.")
	a.Describe(&c.Name, "Human-readable name for the connector, displayed to users during login.")
	a.Describe(&c.ClientId, "GitHub OAuth app client ID.")
	a.Describe(&c.ClientSecret, "GitHub OAuth app client secret.")
	a.Describe(&c.RedirectUri, "Redirect URI registered in GitHub OAuth app. Must match Dex's callback URL.")
	a.Describe(&c.Orgs, "List of GitHub organizations with optional team restrictions. Only users in these orgs/teams will be allowed to authenticate.")
	a.Describe(&c.LoadAllGroups, "If true, load all groups (teams) the user is a member of. Defaults to false.")
	a.Describe(&c.TeamNameField, "Field to use for team names in group claims. Valid values: 'name', 'slug', or 'both'. Defaults to 'slug'.")
	a.Describe(&c.UseLoginAsID, "If true, use GitHub login username as the user ID. Defaults to false.")
	a.Describe(&c.PreferredEmailDomain, "Preferred email domain. If set, users with emails in this domain will be preferred.")
	a.Describe(&c.HostName, "GitHub Enterprise hostname (e.g., 'github.example.com'). Leave empty for github.com.")
	a.Describe(&c.RootCA, "Root CA certificate for GitHub Enterprise (PEM format). Required if using self-signed certificates.")
}

// Annotate provides schema metadata for GitHubOrg.
func (c *GitHubOrg) Annotate(a infer.Annotator) {
	a.Describe(&c.Name, "GitHub organization name.")
	a.Describe(&c.Teams, "List of team names within the organization. If empty, all members of the organization can authenticate.")
}

// Annotate provides schema metadata for GitHubConnectorState.
func (c *GitHubConnectorState) Annotate(a infer.Annotator) {
	// GitHubConnectorState embeds GitHubConnectorArgs, so field descriptions are inherited
}

// Check validates inputs.
func (c *GitHubConnector) Check(ctx context.Context, req infer.CheckRequest) (infer.CheckResponse[GitHubConnectorArgs], error) {
	args, failures, err := infer.DefaultCheck[GitHubConnectorArgs](ctx, req.NewInputs)
	if err != nil {
		return infer.CheckResponse[GitHubConnectorArgs]{Failures: failures}, err
	}

	// Validate teamNameField
	if args.TeamNameField != nil {
		valid := map[string]bool{"name": true, "slug": true, "both": true}
		if !valid[*args.TeamNameField] {
			failures = append(failures, p.CheckFailure{
				Property: "teamNameField",
				Reason:   "must be one of: name, slug, both",
			})
		}
	}

	// Apply defaults
	if args.LoadAllGroups == nil {
		defaultLoadAll := false
		args.LoadAllGroups = &defaultLoadAll
	}
	if args.TeamNameField == nil {
		defaultTeamNameField := "slug"
		args.TeamNameField = &defaultTeamNameField
	}
	if args.UseLoginAsID == nil {
		defaultUseLogin := false
		args.UseLoginAsID = &defaultUseLogin
	}

	return infer.CheckResponse[GitHubConnectorArgs]{
		Inputs:   args,
		Failures: failures,
	}, nil
}

// Create creates a new GitHub connector.
func (c *GitHubConnector) Create(ctx context.Context, req infer.CreateRequest[GitHubConnectorArgs]) (infer.CreateResponse[GitHubConnectorState], error) {
	args := req.Inputs

	// In preview/dry-run mode, skip actual Dex API calls and return expected state
	if req.DryRun {
		state := GitHubConnectorState{
			GitHubConnectorArgs: args,
		}
		return infer.CreateResponse[GitHubConnectorState]{
			ID:     args.ConnectorId,
			Output: state,
		}, nil
	}

	cfg := infer.GetConfig[provider.DexConfig](ctx)
	if cfg.Client == nil {
		return infer.CreateResponse[GitHubConnectorState]{}, fmt.Errorf("Dex client not configured")
	}

	// Build GitHub connector config
	githubConfig := map[string]any{
		"clientID":     args.ClientId,
		"clientSecret": args.ClientSecret,
		"redirectURI":  args.RedirectUri,
	}

	if len(args.Orgs) > 0 {
		orgsConfig := make([]map[string]any, 0, len(args.Orgs))
		for _, org := range args.Orgs {
			orgConfig := map[string]any{"name": org.Name}
			if len(org.Teams) > 0 {
				orgConfig["teams"] = org.Teams
			}
			orgsConfig = append(orgsConfig, orgConfig)
		}
		githubConfig["orgs"] = orgsConfig
	}
	if args.LoadAllGroups != nil {
		githubConfig["loadAllGroups"] = *args.LoadAllGroups
	}
	if args.TeamNameField != nil {
		githubConfig["teamNameField"] = *args.TeamNameField
	}
	if args.UseLoginAsID != nil {
		githubConfig["useLoginAsID"] = *args.UseLoginAsID
	}
	if args.PreferredEmailDomain != nil {
		githubConfig["preferredEmailDomain"] = *args.PreferredEmailDomain
	}
	if args.HostName != nil {
		githubConfig["hostName"] = *args.HostName
	}
	if args.RootCA != nil {
		githubConfig["rootCA"] = *args.RootCA
	}

	configBytes, err := json.Marshal(githubConfig)
	if err != nil {
		return infer.CreateResponse[GitHubConnectorState]{}, fmt.Errorf("failed to marshal GitHub config: %w", err)
	}

	connector := &api.Connector{
		Id:     args.ConnectorId,
		Type:   "github",
		Name:   args.Name,
		Config: configBytes,
	}

	createCtx, cancel := context.WithTimeout(ctx, time.Duration(provider.PtrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	resp, err := cfg.Client.CreateConnector(createCtx, &api.CreateConnectorReq{
		Connector: connector,
	})
	if err != nil {
		return infer.CreateResponse[GitHubConnectorState]{}, provider.WrapError("create", "github-connector", args.ConnectorId, err)
	}

	if resp.AlreadyExists {
		return infer.CreateResponse[GitHubConnectorState]{}, fmt.Errorf("connector with id %q already exists", args.ConnectorId)
	}

	state := GitHubConnectorState{
		GitHubConnectorArgs: args,
	}

	return infer.CreateResponse[GitHubConnectorState]{
		ID:     args.ConnectorId,
		Output: state,
	}, nil
}

// Read retrieves an existing GitHub connector.
func (c *GitHubConnector) Read(ctx context.Context, req infer.ReadRequest[GitHubConnectorArgs, GitHubConnectorState]) (infer.ReadResponse[GitHubConnectorArgs, GitHubConnectorState], error) {
	cfg := infer.GetConfig[provider.DexConfig](ctx)
	if cfg.Client == nil {
		return infer.ReadResponse[GitHubConnectorArgs, GitHubConnectorState]{}, fmt.Errorf("Dex client not configured")
	}

	listCtx, cancel := context.WithTimeout(ctx, time.Duration(provider.PtrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	listResp, err := cfg.Client.ListConnectors(listCtx, &api.ListConnectorReq{})
	if err != nil {
		return infer.ReadResponse[GitHubConnectorArgs, GitHubConnectorState]{}, fmt.Errorf("failed to list connectors: %w", err)
	}

	var found *api.Connector
	for _, conn := range listResp.Connectors {
		if conn.Id == req.ID {
			found = conn
			break
		}
	}

	if found == nil {
		return infer.ReadResponse[GitHubConnectorArgs, GitHubConnectorState]{}, nil
	}

	var configMap map[string]any
	if err := json.Unmarshal(found.Config, &configMap); err != nil {
		return infer.ReadResponse[GitHubConnectorArgs, GitHubConnectorState]{}, nil
	}

	// Parse orgs array
	var orgs []GitHubOrg
	if orgsVal, ok := configMap["orgs"].([]any); ok {
		for _, o := range orgsVal {
			if orgMap, ok := o.(map[string]any); ok {
				org := GitHubOrg{
					Name: GetString(orgMap, "name"),
				}
				if teamsVal, ok := orgMap["teams"].([]any); ok {
					for _, t := range teamsVal {
						if teamStr, ok := t.(string); ok {
							org.Teams = append(org.Teams, teamStr)
						}
					}
				}
				orgs = append(orgs, org)
			}
		}
	}

	args := GitHubConnectorArgs{
		ConnectorId:          found.Id,
		Name:                 found.Name,
		ClientId:             GetString(configMap, "clientID"),
		ClientSecret:         GetString(configMap, "clientSecret"),
		RedirectUri:          GetString(configMap, "redirectURI"),
		Orgs:                 orgs,
		LoadAllGroups:        GetBoolPtr(configMap, "loadAllGroups"),
		TeamNameField:        GetStringPtr(configMap, "teamNameField"),
		UseLoginAsID:         GetBoolPtr(configMap, "useLoginAsID"),
		PreferredEmailDomain: GetStringPtr(configMap, "preferredEmailDomain"),
		HostName:             GetStringPtr(configMap, "hostName"),
		RootCA:               GetStringPtr(configMap, "rootCA"),
	}

	state := GitHubConnectorState{
		GitHubConnectorArgs: args,
	}

	return infer.ReadResponse[GitHubConnectorArgs, GitHubConnectorState]{
		ID:     found.Id,
		Inputs: args,
		State:  state,
	}, nil
}

// Update updates an existing GitHub connector.
func (c *GitHubConnector) Update(ctx context.Context, req infer.UpdateRequest[GitHubConnectorArgs, GitHubConnectorState]) (infer.UpdateResponse[GitHubConnectorState], error) {
	args := req.Inputs
	oldState := req.State

	// In preview/dry-run mode, skip actual Dex API calls and return expected state
	if req.DryRun {
		state := GitHubConnectorState{
			GitHubConnectorArgs: args,
		}
		return infer.UpdateResponse[GitHubConnectorState]{
			Output: state,
		}, nil
	}

	cfg := infer.GetConfig[provider.DexConfig](ctx)
	if cfg.Client == nil {
		return infer.UpdateResponse[GitHubConnectorState]{}, fmt.Errorf("Dex client not configured")
	}

	if args.ConnectorId != oldState.ConnectorId {
		return infer.UpdateResponse[GitHubConnectorState]{}, fmt.Errorf("connectorId cannot be changed")
	}
	if args.HostName != nil && oldState.HostName != nil && *args.HostName != *oldState.HostName {
		return infer.UpdateResponse[GitHubConnectorState]{}, fmt.Errorf("hostName cannot be changed (would require replace)")
	}

	githubConfig := map[string]any{
		"clientID":     args.ClientId,
		"clientSecret": args.ClientSecret,
		"redirectURI":  args.RedirectUri,
	}

	if len(args.Orgs) > 0 {
		orgsConfig := make([]map[string]any, 0, len(args.Orgs))
		for _, org := range args.Orgs {
			orgConfig := map[string]any{"name": org.Name}
			if len(org.Teams) > 0 {
				orgConfig["teams"] = org.Teams
			}
			orgsConfig = append(orgsConfig, orgConfig)
		}
		githubConfig["orgs"] = orgsConfig
	}
	if args.LoadAllGroups != nil {
		githubConfig["loadAllGroups"] = *args.LoadAllGroups
	}
	if args.TeamNameField != nil {
		githubConfig["teamNameField"] = *args.TeamNameField
	}
	if args.UseLoginAsID != nil {
		githubConfig["useLoginAsID"] = *args.UseLoginAsID
	}
	if args.PreferredEmailDomain != nil {
		githubConfig["preferredEmailDomain"] = *args.PreferredEmailDomain
	}
	if args.HostName != nil {
		githubConfig["hostName"] = *args.HostName
	}
	if args.RootCA != nil {
		githubConfig["rootCA"] = *args.RootCA
	}

	configBytes, err := json.Marshal(githubConfig)
	if err != nil {
		return infer.UpdateResponse[GitHubConnectorState]{}, fmt.Errorf("failed to marshal GitHub config: %w", err)
	}

	updateCtx, cancel := context.WithTimeout(ctx, time.Duration(provider.PtrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	_, err = cfg.Client.UpdateConnector(updateCtx, &api.UpdateConnectorReq{
		Id:        args.ConnectorId,
		NewType:   "github",
		NewName:   args.Name,
		NewConfig: configBytes,
	})
	if err != nil {
		return infer.UpdateResponse[GitHubConnectorState]{}, provider.WrapError("update", "github-connector", args.ConnectorId, err)
	}

	state := GitHubConnectorState{
		GitHubConnectorArgs: args,
	}

	return infer.UpdateResponse[GitHubConnectorState]{
		Output: state,
	}, nil
}

// Delete deletes a GitHub connector.
func (c *GitHubConnector) Delete(ctx context.Context, req infer.DeleteRequest[GitHubConnectorState]) (infer.DeleteResponse, error) {
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
		return infer.DeleteResponse{}, provider.WrapError("delete", "github-connector", deleteID, err)
	}

	return infer.DeleteResponse{}, nil
}
