package main

type OIDCClaimMapping struct {
	Groups string `json:"groups,omitempty"`
}

type OIDCConnectorConfig struct {
	Issuer                    string            `json:"issuer"`
	ClientID                  string            `json:"clientID"`
	ClientSecret              string            `json:"clientSecret"`
	RedirectURI               string            `json:"redirectURI"`
	Scopes                    []string          `json:"scopes,omitempty"`
	InsecureEnableGroups      bool              `json:"insecureEnableGroups,omitempty"`
	InsecureSkipEmailVerified bool              `json:"insecureSkipEmailVerified,omitempty"`
	OverrideClaimMapping      bool              `json:"overrideClaimMapping,omitempty"`
	ClaimMapping              *OIDCClaimMapping `json:"claimMapping,omitempty"`
}
