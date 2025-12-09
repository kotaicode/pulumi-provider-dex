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

### 3. Create an Azure/Entra ID Connector

```typescript
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
  - `AzureOidcConnector` - Uses generic OIDC connector
  - `AzureMicrosoftConnector` - Uses Dex's Microsoft-specific connector
- **AWS Cognito Integration**: `CognitoOidcConnector` for managing Cognito user pools

## Resources

### `dex.Client`

Manages an OAuth2 client in Dex.

**Example:**
```typescript
const client = new dex.Client("myClient", {
    clientId: "my-app",
    name: "My Application",
    redirectUris: ["https://app.example.com/callback"],
    public: false,
}, { provider });
```

### `dex.Connector`

Manages a generic connector in Dex (supports OIDC, LDAP, SAML, etc.).

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

### `dex.AzureOidcConnector`

Opinionated resource for Azure/Entra ID using generic OIDC connector.

**Example:**
```typescript
const azure = new dex.AzureOidcConnector("azure", {
    connectorId: "azure",
    name: "Azure AD",
    tenantId: "your-tenant-id",
    clientId: "your-app-id",
    clientSecret: "your-secret",
    redirectUri: "https://dex.example.com/callback",
}, { provider });
```

### `dex.AzureMicrosoftConnector`

Opinionated resource for Azure/Entra ID using Dex's Microsoft-specific connector.

**Example:**
```typescript
const azureMs = new dex.AzureMicrosoftConnector("azure-ms", {
    connectorId: "azure-ms",
    name: "Azure AD (Microsoft)",
    tenant: "common", // or "organizations" or specific tenant ID
    clientId: "your-app-id",
    clientSecret: "your-secret",
    redirectUri: "https://dex.example.com/callback",
}, { provider });
```

### `dex.CognitoOidcConnector`

Opinionated resource for AWS Cognito user pools.

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
}, { provider });
```

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

