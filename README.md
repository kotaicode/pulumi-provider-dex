# Pulumi Provider for Dex

A Pulumi provider for managing Dex (https://dexidp.io/) resources via the Dex gRPC Admin API. This provider allows you to manage Dex OAuth2 clients and connectors (IdPs) as infrastructure-as-code.

## Features

- **OAuth2 Client Management**: Create, update, and delete Dex OAuth2 clients
- **Generic Connector Support**: Manage any Dex connector type (OIDC, LDAP, SAML, etc.)
- **OIDC Connector Support**: First-class support for OIDC connectors with typed configuration
- **Azure/Entra ID Integration**: 
  - `AzureOidcConnector` - Uses generic OIDC connector (type: `oidc`)
  - `AzureMicrosoftConnector` - Uses Dex's Microsoft-specific connector (type: `microsoft`)
- **AWS Cognito Integration**: `CognitoOidcConnector` for managing Cognito user pools as IdPs

## Installation

### Prerequisites

- [Pulumi CLI](https://www.pulumi.com/docs/get-started/install/) installed
- Go 1.24+ (for building the provider)
- Access to a Dex instance with gRPC API enabled

### Building the Provider

```bash
# Clone the repository
git clone https://github.com/kotaicode/pulumi-provider-dex.git
cd pulumi-provider-dex

# Build the provider binary
go build -o bin/pulumi-resource-dex ./cmd/pulumi-resource-dex

# Install the provider locally
pulumi plugin install resource dex v0.1.0 --file bin/pulumi-resource-dex
```

### Generating Language SDKs

After building the provider, generate SDKs for your preferred language:

```bash
# Generate TypeScript SDK
pulumi package gen-sdk bin/pulumi-resource-dex --language typescript --out sdk/typescript

# Generate Go SDK
pulumi package gen-sdk bin/pulumi-resource-dex --language go --out sdk/go

# Generate Python SDK (optional)
pulumi package gen-sdk bin/pulumi-resource-dex --language python --out sdk/python
```

## Configuration

The provider requires configuration to connect to your Dex gRPC API:

```typescript
import * as dex from "@kotaicode/pulumi-dex";

const provider = new dex.Provider("dex", {
    host: "dex.internal:5557", // Dex gRPC host:port
    // Optional: TLS configuration for mTLS
    caCert: fs.readFileSync("certs/ca.crt", "utf-8"),
    clientCert: fs.readFileSync("certs/client.crt", "utf-8"),
    clientKey: fs.readFileSync("certs/client.key", "utf-8"),
    // Or for development:
    // insecureSkipVerify: true,
});
```

### Environment Variables

You can also configure the provider using environment variables:

- `DEX_HOST` - Dex gRPC host:port
- `DEX_CA_CERT` - PEM-encoded CA certificate
- `DEX_CLIENT_CERT` - PEM-encoded client certificate
- `DEX_CLIENT_KEY` - PEM-encoded client private key
- `DEX_INSECURE_SKIP_VERIFY` - Skip TLS verification (development only)
- `DEX_TIMEOUT_SECONDS` - Per-RPC timeout in seconds

## Usage Examples

### Managing an OAuth2 Client

```typescript
import * as dex from "@kotaicode/pulumi-dex";

const webClient = new dex.Client("webClient", {
    clientId: "my-web-app",
    name: "My Web App",
    redirectUris: ["https://app.example.com/callback"],
    // secret is optional - will be auto-generated if omitted
}, { provider });

export const clientSecret = webClient.secret; // Pulumi secret
```

### Azure/Entra ID Connector (Generic OIDC)

```typescript
const azureConnector = new dex.AzureOidcConnector("azure-tenant-a", {
    connectorId: "azure-tenant-a",
    name: "Azure AD (Tenant A)",
    tenantId: "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
    clientId: "your-azure-app-client-id",
    clientSecret: "your-azure-app-client-secret", // Pulumi secret
    redirectUri: "https://dex.example.com/callback",
    scopes: ["openid", "profile", "email", "offline_access"],
    userNameSource: "preferred_username", // or "upn" or "email"
}, { provider });
```

### Azure/Entra ID Connector (Microsoft-Specific)

```typescript
const azureMsConnector = new dex.AzureMicrosoftConnector("azure-ms", {
    connectorId: "azure-ms",
    name: "Azure AD (Microsoft Connector)",
    tenant: "common", // or "organizations" or specific tenant ID
    clientId: "your-azure-app-client-id",
    clientSecret: "your-azure-app-client-secret",
    redirectUri: "https://dex.example.com/callback",
    groups: "groups", // Optional: group claim name
}, { provider });
```

### AWS Cognito Connector

```typescript
const cognitoConnector = new dex.CognitoOidcConnector("cognito-eu", {
    connectorId: "cognito-eu",
    name: "Cognito (EU)",
    region: "eu-central-1",
    userPoolId: "eu-central-1_XXXXXXX",
    clientId: "your-cognito-app-client-id",
    clientSecret: "your-cognito-app-client-secret",
    redirectUri: "https://dex.example.com/callback",
    userNameSource: "email", // or "sub"
}, { provider });
```

### Generic Connector (OIDC)

```typescript
const genericOidcConnector = new dex.Connector("github-oidc", {
    connectorId: "github-oidc",
    type: "oidc",
    name: "GitHub OIDC",
    oidcConfig: {
        issuer: "https://token.actions.githubusercontent.com",
        clientId: "your-github-oidc-client-id",
        clientSecret: "your-secret",
        redirectUri: "https://dex.example.com/callback",
        scopes: ["openid", "email", "profile"],
    },
}, { provider });
```

### Generic Connector (Raw JSON)

```typescript
const githubConnector = new dex.Connector("github", {
    connectorId: "github",
    type: "github",
    name: "GitHub",
    rawConfig: JSON.stringify({
        clientID: "your-github-client-id",
        clientSecret: "your-github-client-secret",
        redirectURI: "https://dex.example.com/callback",
        orgs: ["kotaicode"],
    }),
}, { provider });
```

## Resources

### `dex.Client`

Manages an OAuth2 client in Dex.

**Inputs:**
- `clientId` (string, required) - Unique identifier for the client
- `name` (string, required) - Display name
- `secret` (string, optional, secret) - Client secret (auto-generated if omitted)
- `redirectUris` (string[], required) - Allowed redirect URIs
- `trustedPeers` (string[], optional) - Trusted peer client IDs
- `public` (boolean, optional) - Public (non-confidential) client
- `logoUrl` (string, optional) - Logo image URL

**Outputs:**
- `id` - Resource ID (same as clientId)
- `clientId` - The client ID
- `secret` - The client secret (Pulumi secret)
- `createdAt` - Creation timestamp

### `dex.Connector`

Manages a generic connector in Dex.

**Inputs:**
- `connectorId` (string, required) - Unique identifier
- `type` (string, required) - Connector type (e.g., "oidc", "ldap", "saml", "github")
- `name` (string, required) - Display name
- `oidcConfig` (OIDCConfig, optional) - OIDC configuration (use when type="oidc")
- `rawConfig` (string, optional) - Raw JSON configuration (for non-OIDC connectors)

**Note:** Exactly one of `oidcConfig` or `rawConfig` must be provided.

### `dex.AzureOidcConnector`

Manages an Azure AD/Entra ID connector using generic OIDC.

**Inputs:**
- `connectorId` (string, required)
- `name` (string, required)
- `tenantId` (string, required) - Azure tenant ID (UUID)
- `clientId` (string, required) - Azure app client ID
- `clientSecret` (string, required, secret) - Azure app client secret
- `redirectUri` (string, required)
- `scopes` (string[], optional) - Defaults to `["openid", "profile", "email", "offline_access"]`
- `userNameSource` (string, optional) - "preferred_username" (default), "upn", or "email"
- `extraOidc` (map, optional) - Additional OIDC config fields

### `dex.AzureMicrosoftConnector`

Manages an Azure AD/Entra ID connector using Dex's Microsoft-specific connector.

**Inputs:**
- `connectorId` (string, required)
- `name` (string, required)
- `tenant` (string, required) - "common", "organizations", or tenant ID (UUID)
- `clientId` (string, required)
- `clientSecret` (string, required, secret)
- `redirectUri` (string, required)
- `groups` (string, optional) - Group claim name (requires admin consent)

### `dex.CognitoOidcConnector`

Manages an AWS Cognito user pool connector.

**Inputs:**
- `connectorId` (string, required)
- `name` (string, required)
- `region` (string, required) - AWS region (e.g., "eu-central-1")
- `userPoolId` (string, required) - Cognito user pool ID
- `clientId` (string, required) - Cognito app client ID
- `clientSecret` (string, required, secret) - Cognito app client secret
- `redirectUri` (string, required)
- `scopes` (string[], optional) - Defaults to `["openid", "email", "profile"]`
- `userNameSource` (string, optional) - "email" (default) or "sub"
- `extraOidc` (map, optional) - Additional OIDC config fields

## Local Development and Testing

### Running Dex Locally with Docker Compose

See `docker-compose.yml` for a local Dex setup with gRPC API enabled.

```bash
# Start Dex
docker-compose up -d

# Dex gRPC will be available at localhost:5557
# Dex web UI will be available at http://localhost:5556
```

### Example Pulumi Program

See the `examples/` directory for complete example programs.

## Dex Configuration Requirements

Your Dex instance must have the gRPC API enabled. Add this to your Dex configuration:

```yaml
grpc:
  addr: 127.0.0.1:5557
  tlsCert: /etc/dex/grpc.crt
  tlsKey: /etc/dex/grpc.key
  tlsClientCA: /etc/dex/client.crt
  reflection: true

# Enable connector CRUD (required for connector management)
enablePasswordDB: false
```

And set the environment variable:
```bash
export DEX_API_CONNECTORS_CRUD=true
```

## Security Considerations

- **Secrets**: All secrets (client secrets, TLS keys) are automatically marked as Pulumi secrets and encrypted in state
- **mTLS**: Strongly recommended for production use. Configure TLS certificates properly
- **Network**: Ensure Dex gRPC API is only accessible from trusted networks

## Contributing

Contributions are welcome! Please open an issue or submit a pull request.

## Dex Version Compatibility

This provider has been tested with:
- **Dex v2.4.0+** (with `DEX_API_CONNECTORS_CRUD=true`)

The provider requires:
- Dex gRPC API enabled
- `DEX_API_CONNECTORS_CRUD=true` environment variable set on Dex (required for connector CRUD operations)

For older Dex versions, connector management may not be available. Client management should work with any Dex version that exposes the gRPC API.

## Development

### Prerequisites
- Go 1.24.1+
- Pulumi CLI
- Docker and Docker Compose (for local testing)

### Building

```bash
make build
```

### Running Tests

```bash
# Unit tests
make test

# Integration tests (requires Dex running)
make dex-up
make test  # Run tests with integration tag
make dex-down
```

### Code Quality

```bash
# Run linter
golangci-lint run

# Format code
go fmt ./...
```

## Contributing

Contributions are welcome! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## License

[License TBD - Add MIT or Apache 2.0]

## Support

- **GitHub Issues**: https://github.com/kotaicode/pulumi-provider-dex/issues
- **Documentation**: https://github.com/kotaicode/pulumi-provider-dex#readme

