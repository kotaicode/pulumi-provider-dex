# Pulumi Provider for Dex

A Pulumi provider for managing [Dex](https://dexidp.io/) resources via the Dex gRPC Admin API. This provider allows you to manage Dex OAuth2 clients and connectors (Identity Providers) as infrastructure-as-code.

## Installation

```bash
npm install @kotaicode/pulumi-dex
```

## Prerequisites

- [Pulumi CLI](https://www.pulumi.com/docs/get-started/install/) installed
- Access to a Dex instance with gRPC API enabled
- Node.js 14+ and npm

## Quick Start

### 1. Configure the Provider

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as dex from "@kotaicode/pulumi-dex";
import * as fs from "fs";

// Configure the Dex provider
const provider = new dex.Provider("dex", {
    host: "dex.internal:5557", // Dex gRPC host:port
    // For development with insecure Dex:
    insecureSkipVerify: true,
    // Or for production with mTLS:
    // caCert: fs.readFileSync("certs/ca.crt", "utf-8"),
    // clientCert: fs.readFileSync("certs/client.crt", "utf-8"),
    // clientKey: fs.readFileSync("certs/client.key", "utf-8"),
});
```

### 2. Create an OAuth2 Client

```typescript
const webClient = new dex.Client("webClient", {
    clientId: "my-web-app",
    name: "My Web App",
    redirectUris: ["https://app.example.com/callback"],
    // secret is optional - will be auto-generated if omitted
}, { provider });

export const clientId = webClient.clientId;
export const clientSecret = webClient.secret; // Pulumi secret
```

### 3. Create a Connector

```typescript
// Example: Azure AD connector
const azureConnector = new dex.AzureOidcConnector("azure-tenant", {
    connectorId: "azure-tenant",
    name: "Azure AD",
    tenantId: "your-tenant-id",
    clientId: "your-azure-app-client-id",
    clientSecret: "your-azure-app-client-secret",
    redirectUri: "https://dex.example.com/callback",
}, { provider });
```

## Features

- **OAuth2 Client Management**: Create, update, and delete Dex OAuth2 clients
- **Generic Connector Support**: Manage any Dex connector type (OIDC, LDAP, SAML, etc.)
- **OIDC Connector Support**: First-class support for OIDC connectors with typed configuration
- **Azure/Entra ID Integration**: 
  - `AzureOidcConnector` - Uses generic OIDC connector (type: `oidc`)
  - `AzureMicrosoftConnector` - Uses Dex's Microsoft-specific connector (type: `microsoft`)
- **AWS Cognito Integration**: `CognitoOidcConnector` for managing Cognito user pools as IdPs
- **GitLab Integration**: `GitLabConnector` for GitLab.com and self-hosted GitLab instances
- **GitHub Integration**: `GitHubConnector` for GitHub.com and GitHub Enterprise
- **Google Integration**: `GoogleConnector` for Google Workspace and Google accounts
- **Local/Builtin Connector**: `LocalConnector` for local user authentication

## Resources

### `dex.Client`

Manages an OAuth2 client in Dex. OAuth2 clients are applications that can authenticate users through Dex.

**Example:**
```typescript
const client = new dex.Client("myClient", {
    clientId: "my-app",
    name: "My Application",
    redirectUris: ["https://app.example.com/callback"],
    public: false,
    logoUrl: "https://app.example.com/logo.png",
}, { provider });
```

**Inputs:**
- `clientId` (string, required) - Unique identifier for the OAuth2 client. Used as the client_id in OAuth2 flows.
- `name` (string, required) - Human-readable name for the OAuth2 client.
- `secret` (string, optional, secret) - Client secret for the OAuth2 client. If not provided, a secure random secret will be generated automatically.
- `redirectUris` (string[], required) - List of allowed redirect URIs for OAuth2 authorization flows. Must be valid HTTP/HTTPS URLs.
- `trustedPeers` (string[], optional) - List of trusted peer client IDs that can exchange tokens with this client.
- `public` (boolean, optional) - If true, this client is a public client (e.g., mobile app) and does not require a client secret.
- `logoUrl` (string, optional) - URL to a logo image for the OAuth2 client. Used in consent screens.

**Outputs:**
- `id` - Resource ID (same as clientId)
- `clientId` - The client ID
- `secret` - The client secret (Pulumi secret)
- `createdAt` - Creation timestamp (RFC3339 format)

### `dex.Connector`

Manages a generic connector (upstream identity provider) in Dex. Use this resource for connectors not covered by specific connector types, or when you need full control over the connector configuration.

**Example (OIDC):**
```typescript
const connector = new dex.Connector("github-oidc", {
    connectorId: "github-oidc",
    type: "oidc",
    name: "GitHub OIDC",
    oidcConfig: {
        issuer: "https://token.actions.githubusercontent.com",
        clientId: "your-client-id",
        clientSecret: "your-secret",
        redirectUri: "https://dex.example.com/callback",
        scopes: ["openid", "email", "profile"],
    },
}, { provider });
```

**Inputs:**
- `connectorId` (string, required) - Unique identifier for the connector.
- `type` (string, required) - Type of connector (e.g., 'oidc', 'saml', 'ldap'). Must match a connector type supported by Dex.
- `name` (string, required) - Human-readable name for the connector, displayed to users during login.
- `oidcConfig` (OIDCConfig, optional) - OIDC-specific configuration. Use this for OIDC-based connectors.
- `rawConfig` (string, optional) - Raw JSON configuration for the connector. Use this for advanced configurations or connector types not directly supported. If provided, this takes precedence over OIDCConfig.

**Note:** Exactly one of `oidcConfig` or `rawConfig` must be provided.

### `dex.AzureOidcConnector`

Manages an Azure AD/Entra ID connector in Dex using the generic OIDC connector (type: oidc). This connector allows users to authenticate using their Azure AD/Entra ID credentials.

**Example:**
```typescript
const azure = new dex.AzureOidcConnector("azure", {
    connectorId: "azure",
    name: "Azure AD",
    tenantId: "your-tenant-id", // UUID format
    clientId: "your-app-id",
    clientSecret: "your-secret",
    redirectUri: "https://dex.example.com/callback",
    scopes: ["openid", "profile", "email", "offline_access"], // Optional, defaults to these
    userNameSource: "preferred_username", // Optional: "preferred_username" (default), "upn", or "email"
}, { provider });
```

**Inputs:**
- `connectorId` (string, required) - Unique identifier for the Azure connector.
- `name` (string, required) - Human-readable name for the connector.
- `tenantId` (string, required) - Azure AD tenant ID (UUID format). This identifies your Azure AD organization.
- `clientId` (string, required) - Azure AD application (client) ID.
- `clientSecret` (string, required, secret) - Azure AD application client secret.
- `redirectUri` (string, required) - Redirect URI registered in Azure AD. Must match Dex's callback URL.
- `scopes` (string[], optional) - OIDC scopes to request from Azure AD. Defaults to `["openid", "profile", "email", "offline_access"]` if not specified.
- `userNameSource` (string, optional) - Source for the username claim. Valid values: 'preferred_username' (default), 'upn' (User Principal Name), or 'email'.
- `extraOidc` (map, optional) - Additional OIDC configuration fields as key-value pairs for advanced scenarios.

### `dex.AzureMicrosoftConnector`

Manages an Azure AD/Entra ID connector in Dex using the Microsoft-specific connector (type: microsoft). This connector provides Microsoft-specific features like group filtering and domain restrictions.

**Example:**
```typescript
const azureMs = new dex.AzureMicrosoftConnector("azure-ms", {
    connectorId: "azure-ms",
    name: "Azure AD (Microsoft)",
    tenant: "common", // or "organizations" or specific tenant ID (UUID)
    clientId: "your-app-id",
    clientSecret: "your-secret",
    redirectUri: "https://dex.example.com/callback",
    groups: "groups", // Optional: group claim name, e.g., "groups"
}, { provider });
```

**Inputs:**
- `connectorId` (string, required) - Unique identifier for the Azure Microsoft connector.
- `name` (string, required) - Human-readable name for the connector.
- `tenant` (string, required) - Azure AD tenant identifier. Can be 'common' (any Azure AD account), 'organizations' (any organizational account), or a specific tenant ID (UUID format).
- `clientId` (string, required) - Azure AD application (client) ID.
- `clientSecret` (string, required, secret) - Azure AD application client secret.
- `redirectUri` (string, required) - Redirect URI registered in Azure AD. Must match Dex's callback URL.
- `groups` (string, optional) - Name of the claim that contains group memberships (e.g., 'groups'). Used for group-based access control.

### `dex.CognitoOidcConnector`

Manages an AWS Cognito user pool connector in Dex using the generic OIDC connector (type: oidc). This connector allows users to authenticate using their AWS Cognito credentials.

**Example:**
```typescript
const cognito = new dex.CognitoOidcConnector("cognito", {
    connectorId: "cognito",
    name: "AWS Cognito",
    region: "us-east-1",
    userPoolId: "us-east-1_XXXXXXXXX",
    clientId: "your-cognito-client-id",
    clientSecret: "your-secret",
    redirectUri: "https://dex.example.com/callback",
    userNameSource: "email", // Optional: "email" (default) or "sub"
}, { provider });
```

**Inputs:**
- `connectorId` (string, required) - Unique identifier for the Cognito connector.
- `name` (string, required) - Human-readable name for the connector.
- `region` (string, required) - AWS region where the Cognito user pool is located (e.g., 'us-east-1', 'eu-west-1').
- `userPoolId` (string, required) - AWS Cognito user pool ID.
- `clientId` (string, required) - Cognito app client ID.
- `clientSecret` (string, required, secret) - Cognito app client secret.
- `redirectUri` (string, required) - Redirect URI registered in Cognito. Must match Dex's callback URL.
- `scopes` (string[], optional) - OIDC scopes to request from Cognito. Defaults to `["openid", "email", "profile"]` if not specified.
- `userNameSource` (string, optional) - Source for the username claim. Valid values: 'email' or 'sub' (subject).
- `extraOidc` (map, optional) - Additional OIDC configuration fields as key-value pairs for advanced scenarios.

### `dex.GitLabConnector`

Manages a GitLab connector in Dex. This connector allows users to authenticate using their GitLab accounts and supports group-based access control.

**Example:**
```typescript
const gitlab = new dex.GitLabConnector("gitlab", {
    connectorId: "gitlab",
    name: "GitLab",
    clientId: "your-gitlab-client-id",
    clientSecret: "your-gitlab-client-secret",
    redirectUri: "https://dex.example.com/callback",
    baseURL: "https://gitlab.com", // Optional, defaults to https://gitlab.com
    groups: ["my-group"], // Optional: groups whitelist
    useLoginAsID: false, // Optional: use username as ID instead of internal ID, default: false
    getGroupsPermission: false, // Optional: include group permissions in groups claim, default: false
}, { provider });
```

**Inputs:**
- `connectorId` (string, required) - Unique identifier for the GitLab connector.
- `name` (string, required) - Human-readable name for the connector.
- `clientId` (string, required) - GitLab OAuth application client ID.
- `clientSecret` (string, required, secret) - GitLab OAuth application client secret.
- `redirectUri` (string, required) - Redirect URI registered in GitLab OAuth app. Must match Dex's callback URL.
- `baseURL` (string, optional) - GitLab instance base URL. Defaults to 'https://gitlab.com' for GitLab.com.
- `groups` (string[], optional) - List of GitLab group names. Only users in these groups will be allowed to authenticate.
- `useLoginAsID` (boolean, optional) - If true, use GitLab username as the user ID. Defaults to false.
- `getGroupsPermission` (boolean, optional) - If true, request 'read_api' scope to fetch group memberships. Defaults to false.

### `dex.GitHubConnector`

Manages a GitHub connector in Dex. This connector allows users to authenticate using their GitHub accounts and supports organization and team-based access control.

**Example:**
```typescript
const github = new dex.GitHubConnector("github", {
    connectorId: "github",
    name: "GitHub",
    clientId: "your-github-client-id",
    clientSecret: "your-github-client-secret",
    redirectUri: "https://dex.example.com/callback",
    orgs: [
        { name: "my-organization" },
        { 
            name: "my-organization-with-teams",
            teams: ["red-team", "blue-team"]
        }
    ],
    teamNameField: "slug", // Optional: "name", "slug", or "both" - default: "slug"
    useLoginAsID: false, // Optional: use username as ID
    // For GitHub Enterprise:
    // hostName: "git.example.com",
    // rootCA: "/etc/dex/ca.crt",
}, { provider });
```

**Inputs:**
- `connectorId` (string, required) - Unique identifier for the GitHub connector.
- `name` (string, required) - Human-readable name for the connector.
- `clientId` (string, required) - GitHub OAuth app client ID.
- `clientSecret` (string, required, secret) - GitHub OAuth app client secret.
- `redirectUri` (string, required) - Redirect URI registered in GitHub OAuth app. Must match Dex's callback URL.
- `orgs` (GitHubOrg[], optional) - List of GitHub organizations with optional team restrictions. Only users in these orgs/teams will be allowed to authenticate.
  - `name` (string, required) - GitHub organization name.
  - `teams` (string[], optional) - List of team names within the organization. If empty, all members of the organization can authenticate.
- `loadAllGroups` (boolean, optional) - If true, load all groups (teams) the user is a member of. Defaults to false.
- `teamNameField` (string, optional) - Field to use for team names in group claims. Valid values: 'name', 'slug', or 'both'. Defaults to 'slug'.
- `useLoginAsID` (boolean, optional) - If true, use GitHub login username as the user ID. Defaults to false.
- `preferredEmailDomain` (string, optional) - Preferred email domain. If set, users with emails in this domain will be preferred.
- `hostName` (string, optional) - GitHub Enterprise hostname (e.g., 'github.example.com'). Leave empty for github.com.
- `rootCA` (string, optional) - Root CA certificate for GitHub Enterprise (PEM format). Required if using self-signed certificates.

### `dex.GoogleConnector`

Manages a Google connector in Dex. This connector allows users to authenticate using their Google accounts and supports domain and group-based access control.

**Example:**
```typescript
const google = new dex.GoogleConnector("google", {
    connectorId: "google",
    name: "Google",
    clientId: "your-google-client-id",
    clientSecret: "your-google-client-secret",
    redirectUri: "https://dex.example.com/callback",
    promptType: "consent", // Optional: default is "consent"
    hostedDomains: ["example.com"], // Optional: domain whitelist for G Suite
    groups: ["admins@example.com"], // Optional: group whitelist for G Suite
    // For group fetching:
    // serviceAccountFilePath: "/path/to/googleAuth.json",
    // domainToAdminEmail: {
    //     "*": "super-user@example.com",
    //     "my-domain.com": "super-user@my-domain.com"
    // },
}, { provider });
```

**Inputs:**
- `connectorId` (string, required) - Unique identifier for the Google connector.
- `name` (string, required) - Human-readable name for the connector.
- `clientId` (string, required) - Google OAuth client ID.
- `clientSecret` (string, required, secret) - Google OAuth client secret.
- `redirectUri` (string, required) - Redirect URI registered in Google OAuth app. Must match Dex's callback URL.
- `promptType` (string, optional) - OAuth prompt type. Valid values: 'consent' (default) or 'select_account'.
- `hostedDomains` (string[], optional) - List of Google Workspace domains. Only users with email addresses in these domains will be allowed to authenticate.
- `groups` (string[], optional) - List of Google Groups. Only users in these groups will be allowed to authenticate.
- `serviceAccountFilePath` (string, optional) - Path to Google service account JSON file. Required for group-based access control.
- `domainToAdminEmail` (map[string]string, optional) - Map of domain names to admin email addresses. Used for group lookups in Google Workspace.

### `dex.LocalConnector`

Manages a local/builtin connector in Dex. The local connector provides username/password authentication stored in Dex's database. This is useful for testing or when you don't have an external identity provider.

**Example:**
```typescript
const local = new dex.LocalConnector("local", {
    connectorId: "local",
    name: "Local",
    enabled: true, // Optional: default is true
}, { provider });
```

**Inputs:**
- `connectorId` (string, required) - Unique identifier for the local connector.
- `name` (string, required) - Human-readable name for the connector.
- `enabled` (boolean, optional) - Whether the local connector is enabled. Defaults to true.

**Note:** The local connector requires `enablePasswordDB: true` in Dex configuration. User management is handled separately via Dex's static passwords or gRPC API.

## Configuration

The provider supports the following configuration options:

| Option | Type | Required | Description |
|--------|------|----------|-------------|
| `host` | string | Yes | Dex gRPC host:port (e.g., `dex.internal:5557`) |
| `caCert` | string | No | PEM-encoded CA certificate for validating Dex's TLS certificate |
| `clientCert` | string | No | PEM-encoded client certificate for mTLS |
| `clientKey` | string | No | PEM-encoded private key for the client certificate |
| `insecureSkipVerify` | boolean | No | Skip TLS verification (development only) |
| `timeoutSeconds` | number | No | Per-RPC timeout in seconds (default: 5) |

## Documentation

- [Full Documentation](https://github.com/kotaicode/pulumi-provider-dex#readme)
- [Dex Documentation](https://dexidp.io/docs/)
- [Pulumi Documentation](https://www.pulumi.com/docs/)

## License

MIT License - see [LICENSE](https://github.com/kotaicode/pulumi-provider-dex/blob/main/LICENSE) file for details.

## Support

- [GitHub Issues](https://github.com/kotaicode/pulumi-provider-dex/issues)
- [Pulumi Community Slack](https://slack.pulumi.com/)
