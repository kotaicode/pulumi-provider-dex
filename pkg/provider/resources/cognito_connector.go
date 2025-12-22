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
	a.Describe(c, "Manages an AWS Cognito user pool connector in Dex using the generic OIDC connector (type: oidc). This connector allows users to authenticate using their AWS Cognito credentials.")
}

// Annotate provides schema metadata for CognitoOidcConnectorArgs.
func (c *CognitoOidcConnectorArgs) Annotate(a infer.Annotator) {
	a.Describe(&c.ConnectorId, "Unique identifier for the Cognito connector.")
	a.Describe(&c.Name, "Human-readable name for the connector, displayed to users during login.")
	a.Describe(&c.Region, "AWS region where the Cognito user pool is located (e.g., 'us-east-1', 'eu-west-1').")
	a.Describe(&c.UserPoolId, "AWS Cognito user pool ID.")
	a.Describe(&c.ClientId, "Cognito app client ID.")
	a.Describe(&c.ClientSecret, "Cognito app client secret.")
	a.Describe(&c.RedirectUri, "Redirect URI registered in Cognito. Must match Dex's callback URL.")
	a.Describe(&c.Scopes, "OIDC scopes to request from Cognito. Defaults to ['openid', 'email', 'profile'] if not specified.")
	a.Describe(&c.UserNameSource, "Source for the username claim. Valid values: 'email' or 'sub' (subject).")
	a.Describe(&c.ExtraOidc, "Additional OIDC configuration fields as key-value pairs for advanced scenarios.")
}

// Annotate provides schema metadata for CognitoOidcConnectorState.
func (c *CognitoOidcConnectorState) Annotate(a infer.Annotator) {
	// CognitoOidcConnectorState embeds CognitoOidcConnectorArgs, so field descriptions are inherited
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

	cfg := infer.GetConfig[provider.DexConfig](ctx)
	if cfg.Client == nil {
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

	createCtx, cancel := context.WithTimeout(ctx, time.Duration(provider.PtrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	resp, err := cfg.Client.CreateConnector(createCtx, &api.CreateConnectorReq{
		Connector: connector,
	})
	if err != nil {
		return infer.CreateResponse[CognitoOidcConnectorState]{}, provider.WrapError("create", "cognito-oidc-connector", args.ConnectorId, err)
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
	cfg := infer.GetConfig[provider.DexConfig](ctx)
	if cfg.Client == nil {
		return infer.ReadResponse[CognitoOidcConnectorArgs, CognitoOidcConnectorState]{}, fmt.Errorf("Dex client not configured")
	}

	listCtx, cancel := context.WithTimeout(ctx, time.Duration(provider.PtrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	listResp, err := cfg.Client.ListConnectors(listCtx, &api.ListConnectorReq{})
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
		ClientId:       GetString(configMap, "clientID"),
		ClientSecret:   GetString(configMap, "clientSecret"),
		RedirectUri:    GetString(configMap, "redirectURI"),
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

	cfg := infer.GetConfig[provider.DexConfig](ctx)
	if cfg.Client == nil {
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

	updateCtx, cancel := context.WithTimeout(ctx, time.Duration(provider.PtrOr(cfg.TimeoutSeconds, 5))*time.Second)
	defer cancel()

	_, err = cfg.Client.UpdateConnector(updateCtx, &api.UpdateConnectorReq{
		Id:        args.ConnectorId,
		NewType:   "oidc",
		NewName:   args.Name,
		NewConfig: configBytes,
	})
	if err != nil {
		return infer.UpdateResponse[CognitoOidcConnectorState]{}, provider.WrapError("update", "cognito-oidc-connector", args.ConnectorId, err)
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
		return infer.DeleteResponse{}, provider.WrapError("delete", "cognito-oidc-connector", deleteID, err)
	}

	return infer.DeleteResponse{}, nil
}
