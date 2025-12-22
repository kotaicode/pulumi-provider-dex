# Dex Provider Examples

This directory contains example Pulumi programs demonstrating how to use the Dex provider.

## Prerequisites

1. **Dex running locally**: Use the provided `docker-compose.yml` in the root directory:
   ```bash
   docker-compose up -d
   ```

2. **Pulumi CLI installed**: See [Pulumi Installation](https://www.pulumi.com/docs/get-started/install/)

3. **Provider SDK generated**: Generate the SDKs first (see main README.md)

## TypeScript Example

The TypeScript example demonstrates:
- Creating an OAuth2 client
- Creating Azure/Entra ID connectors (both OIDC and Microsoft-specific)
- Creating AWS Cognito connectors
- Creating generic OIDC connectors

### Setup

```bash
cd examples/typescript
npm install
```

### Configuration

Before running, you'll need to:

1. **Update the example code** with your actual credentials:
   - Azure tenant ID and app credentials
   - Cognito user pool ID and app credentials
   - Or use the generic connector examples

2. **Configure the provider** in `index.ts`:
   - For local development: `host: "localhost:5557"` with `insecureSkipVerify: true`
   - For production: Configure TLS certificates

### Running

```bash
# Initialize a new Pulumi stack
pulumi stack init dev

# Preview changes
pulumi preview

# Apply changes
pulumi up

# View outputs
pulumi stack output

# Destroy resources
pulumi destroy
```

## Local Dex Setup

The `docker-compose.yml` in the root directory sets up a local Dex instance with:
- Web UI at http://localhost:5556
- gRPC API at localhost:5557 (insecure for development)
- Connector CRUD enabled via `DEX_API_CONNECTORS_CRUD=true`

### Testing the Connection

You can test the gRPC connection using the simple-client:

```bash
cd simple-client
go run . --host localhost:5557
```

## Production Considerations

For production use:

1. **Enable TLS/mTLS**: Configure proper TLS certificates in Dex and the provider
2. **Secure secrets**: Use Pulumi secrets management or external secret stores
3. **Network security**: Restrict access to Dex gRPC API
4. **Backup**: Ensure Dex storage is backed up (this example uses in-memory storage)

