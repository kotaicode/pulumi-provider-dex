package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	api "github.com/dexidp/dex/api/v2"
)

func ensureMsEntraConnector(
	ctx context.Context,
	dex api.DexClient,
	tenantID string,
	clientID string,
	clientSecret string,
) error {
	// 1. Build config from your YAML values
	cfg := OIDCConnectorConfig{
		Issuer:       fmt.Sprintf("https://login.microsoftonline.com/%s/v2.0", tenantID),
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURI:  "https://example.com/dex/callback", //TODO load from env var/config
		Scopes: []string{
			"openid",
			"email",
			"profile",
			"User.Read",
			"offline_access",
		},
		InsecureEnableGroups:      true,
		InsecureSkipEmailVerified: true,
		OverrideClaimMapping:      true,
		ClaimMapping: &OIDCClaimMapping{
			Groups: "roles",
		},
	}

	cfgBytes, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal connector config: %w", err)
	}

	// 2. Create the Connector message
	connector := &api.Connector{
		Id:     "microsoft-grpc", // must be unique; TODO: load from config
		Name:   "entraId-grpc",   //TODO: load from config
		Type:   "oidc",
		Config: cfgBytes,
	}

	// 3. Call CreateConnector
	cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := dex.CreateConnector(cctx, &api.CreateConnectorReq{
		Connector: connector,
	})
	if err != nil {
		return fmt.Errorf("CreateConnector: %w", err)
	}

	if resp.AlreadyExists {
		// Optional: you might want to UpdateConnector instead
		// to change clientSecret, scopes, etc.
		uctx, cancel2 := context.WithTimeout(ctx, 5*time.Second)
		defer cancel2()

		_, err := dex.UpdateConnector(uctx, &api.UpdateConnectorReq{
			Id:        connector.Id,
			NewConfig: connector.Config,
			NewType:   connector.Type,
			NewName:   connector.Name,
		})
		if err != nil {
			return fmt.Errorf("UpdateConnector: %w", err)
		}
	}

	return nil
}
