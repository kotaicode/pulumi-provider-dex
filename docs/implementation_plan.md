5.1 Tech choices

Language: Go

Provider framework: pulumi-go-provider (Layer 3)

Dex client: use the official gRPC stubs generated from Dex api.proto.

Optionally study / reuse patterns from terraform-provider-dexidp for how they handle TLS and CRUD.

5.2 Phases & milestones

Phase 0 – PoC / Spike (time-boxed)

Stand up a local Dex with gRPC API & DEX_API_CONNECTORS_CRUD=true.

Write a tiny Go program that:

Connects via mTLS.

Calls CreateClient, ListClients, CreateConnector, ListConnectors.

Confirm API semantics and get comfortable with Dex behaviour.

Phase 1 – Provider bootstrap

Use pulumi-go-provider “Hello, Pulumi” example as template.

Create cmd/pulumi-resource-dex/main.go.

Implement provider configuration (host, tls, etc.).

Wire gRPC client creation in provider.

Deliverable: provider binary that can be pulumi plugin install’d and does basic Check/Configure.

Phase 2 – dex.Client resource

Define Client schema in Go (inputs/outputs).

Implement:

Create: Dex CreateClient

Read: Dex ListClients/GetClient by clientId

Update: Dex UpdateClient

Delete: Dex DeleteClient

Handle secrets properly (Pulumi secret type).

Add unit tests + integration tests against a test Dex instance.

Deliverable: Pulumi program can create/update/delete Dex clients.

Phase 3 – dex.Connector resource (generic)

Add Connector schema with type, connectorId, name, oidcConfig, rawConfig.

Implement mapping from oidcConfig → Dex JSON; Dex JSON → oidcConfig (best-effort).

Implement full CRUD.

Validate invariants (one of oidcConfig/rawConfig, type === "oidc" when using oidcConfig, etc).

Deliverable: Pulumi program can define arbitrary connectors including OIDC connectors.

Phase 4 – Opinionated resources

Implement AzureOidcConnector and CognitoOidcConnector:

Derive issuer, userNameKey, defaults.

Under the hood, either:

Call Dex gRPC directly, or

Reuse internal helper that builds an OIDC connector JSON and calls the same logic as Connector.

Add nice validation & error messages (e.g., invalid region / tenantId pattern).

Deliverable: sample Pulumi stack that creates Azure + Cognito connectors in Dex.

Phase 5 – Packaging & DX

Generate language SDKs (TS, Go; optionally Python/C#).

Publish to private or public registry:

NPM: @kotaicode/pulumi-dex or @pulumi/dex

Go module: github.com/kotaicode/pulumi-dex/sdk/go/dex

Documentation:

README with install instructions.

Examples: multi-tenant Azure + multi-pool Cognito.

CI:

Build & test provider.

Run integration tests with Dex in Docker.