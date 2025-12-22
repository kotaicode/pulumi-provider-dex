1. High-level concept

Pulumi provider: dex
Plugin binary: pulumi-resource-dex

Pulumi engine ⇄ pulumi-resource-dex (gRPC) ⇄ Dex gRPC API (mTLS)

We’ll build a native Go provider using pulumi-go-provider and talk to Dex using its official api.proto definitions and gRPC stubs.

We’ll initially focus on:

Provider config (connection to Dex gRPC).

Core resources:

dex.Client

dex.Connector

Opinionated resources:

dex.AzureOidcConnector

dex.CognitoOidcConnector

2. Provider configuration design
2.1 Package & provider

Pulumi package name: dex

Provider resource type: dex:index:Provider

2.2 Provider inputs

These are applied when you instantiate the provider in your Pulumi program:

const dexProvider = new dex.Provider("dex", {
  host: "dex.internal.kotaicode:5557",
  tls: {
    caCert: pulumi.secret(fs.readFileSync("certs/ca.crt").toString()),
    clientCert: pulumi.secret(fs.readFileSync("certs/client.crt").toString()),
    clientKey: pulumi.secret(fs.readFileSync("certs/client.key").toString()),
  },
});


Schema (Provider inputs)

host: string – Dex gRPC host:port (required).
Equivalent to Terraform provider "dexidp" { host = "127.0.0.1:5557" }.

tls.caCert: string? – PEM CA cert

tls.clientCert: string? – PEM client cert

tls.clientKey: string? – PEM private key

tls.insecureSkipVerify: boolean? – allow self-signed / name mismatch in dev

timeoutSeconds: number? – optional per-RPC timeout

Provider behaviour

On Configure, establish a Dex gRPC client (with mTLS if TLS values provided).

Reuse a shared connection for all resource operations.

Validate connectivity on first use (e.g., lightweight API call) so misconfig fails early.

3. Core resources
3.1 dex.Client (OAuth2 client in Dex)

Type: dex:index:Client

Dex supports managing OAuth2 clients over gRPC (create/update/delete).

Inputs (proposed)

clientId: string (required, stable ID in Dex)

name: string (required)

secret: string (optional – generated if omitted; always a Pulumi secret)

redirectUris: string[] (required)

trustedPeers: string[]?

public: boolean? – public vs confidential

extensions: any? – escape hatch for less common Dex client fields

Outputs

id: string – Pulumi resource ID (mirror clientId)

clientId: string

secret: string (Pulumi secret)

createdAt: string? – if Dex exposes it (optional future enhancement)

CRUD mapping

Create: call Dex gRPC CreateClient with given fields; if secret absent, generate secure one in provider and send to Dex.

Read: GetClient or list/filter by clientId.

Update: UpdateClient for mutable fields; guard immutable ones (e.g. clientId).

Delete: DeleteClient(clientId).

Init example (TypeScript)

const webClient = new dex.Client("webClient", {
  clientId: "my-web-app",
  name: "My Web App",
  redirectUris: ["https://app.example.com/callback"],
}, { provider: dexProvider });

3.2 dex.Connector (generic)

Dex treats connectors as the abstraction for upstream IdPs.

Type: dex:index:Connector

We’ll keep this generic, but with a strongly typed oidcConfig for the common case.

Inputs (proposed)

connectorId: string – the Dex connector ID (required; stable)

type: "oidc" | "ldap" | "saml" | string – connector type (required)

name: string – label on login screen

Connector-specific configs (one-of style):

oidcConfig?: ConnectorOidcConfig

rawConfig?: string – JSON string for advanced/custom use

ConnectorOidcConfig shape (mirrors Dex OIDC connector):

issuer: string (required)

clientId: string (required)

clientSecret: string (required, secret)

redirectUri: string (required)

scopes?: string[] (default ["openid", "email", "profile"])

userNameKey?: string

insecureSkipEmailVerified?: boolean

insecureIssuer?: boolean

claimMapping?: { emailKey?: string; groupsKey?: string; }

extra?: Record<string, any> – escape hatch for future Dex fields

Outputs

id: string – Pulumi resource ID (mirror connectorId)

connectorId: string

type, name

oidcConfig?

rawConfig?

Rules

Exactly one of oidcConfig or rawConfig must be set.

If type !== "oidc", oidcConfig is invalid.

CRUD mapping

Create:

If oidcConfig is set: build the JSON matching Dex OIDC connector config and send as config bytes in Dex CreateConnector.

If rawConfig is set: validate it is valid JSON; send as bytes unchanged.

Read:

Fetch connector by connectorId, decode config bytes:

If JSON matches OIDC schema, populate oidcConfig.

Otherwise, store in rawConfig.

Update: same mapping, use UpdateConnector.

Delete: DeleteConnector(connectorId).

Example (TS)

const genericConnector = new dex.Connector("github-connector", {
  connectorId: "github",
  type: "github",
  name: "GitHub",
  rawConfig: JSON.stringify({
    clientID: "...",
    clientSecret: "...",
    redirectURI: "https://dex.example.com/callback",
    orgs: ["kotaicode"],
  }),
}, { provider: dexProvider });

4. Opinionated high-level resources

These are convenience wrappers that:

Take “business language” fields (tenant IDs, regions, pool IDs).

Derive the Dex OIDC config for you.

Under the hood, they still call Dex via CreateConnector like the generic connector.

4.1 dex.AzureOidcConnector

Type: dex:azure:OidcConnector

Inputs

name: string – display name on Dex login screen (default auto-generated from tenant?)

connectorId: string – stable ID in Dex

tenantId: string

clientId: string

clientSecret: string (secret)

redirectUri: string (default from provider or explicit)

scopes?: string[] (default ["openid","profile","email","offline_access"])

userNameSource?: "preferred_username" | "upn" | "email" (default preferred_username)

extraOidc?: Partial<ConnectorOidcConfig> for advanced override

Derived fields

issuer = "https://login.microsoftonline.com/<tenantId>/v2.0"

The provider then internally creates a Dex OIDC connector with:

{
  "issuer": "https://login.microsoftonline.com/<tenantId>/v2.0",
  "clientID": "<clientId>",
  "clientSecret": "<clientSecret>",
  "redirectURI": "<redirectUri>",
  "scopes": [...],
  "userNameKey": "<derived from userNameSource>",
  ...extra
}


Implementation detail

Option A (simplest): implement AzureOidcConnector as its own Pulumi resource that directly calls Dex gRPC and stores its own state.

Option B: implement AzureOidcConnector as a macro that internally instantiates a dex.Connector inside the SDK. That’s nicer logically, but Pulumi’s model for “composed resources” is easier from user code than inside the provider. For a first iteration, I’d make it a real provider resource that just wraps the same Dex calls as Connector but with derived config.

Example

const azureTenantA = new dex.AzureOidcConnector("azureTenantA", {
  connectorId: "azure-tenant-a",
  name: "Azure AD (Tenant A)",
  tenantId: "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
  clientId: "...",
  clientSecret: "...",
  redirectUri: "https://dex.example.com/callback",
}, { provider: dexProvider });

4.2 dex.CognitoOidcConnector

Type: dex:cognito:OidcConnector

Inputs

name: string

connectorId: string

region: string – e.g. eu-central-1

userPoolId: string – full pool ID

clientId: string

clientSecret: string

redirectUri: string

scopes?: string[] (default ["openid","email","profile"])

userNameSource?: "email" | "sub" (default email)

extraOidc?: Partial<ConnectorOidcConfig>

Derived fields

issuer = "https://cognito-idp.<region>.amazonaws.com/<userPoolId>"

Then the Dex OIDC config looks like:

{
  "issuer": "https://cognito-idp.eu-central-1.amazonaws.com/<pool>",
  "clientID": "<clientId>",
  "clientSecret": "<clientSecret>",
  "redirectURI": "<redirectUri>",
  "scopes": ["openid","email","profile"],
  "userNameKey": "email"
}


Example

const cognitoEu = new dex.CognitoOidcConnector("cognitoEu", {
  connectorId: "cognito-eu",
  name: "Cognito (EU)",
  region: "eu-central-1",
  userPoolId: "eu-central-1_XXXXXXX",
  clientId: "...",
  clientSecret: "...",
  redirectUri: "https://dex.example.com/callback",
}, { provider: dexProvider });