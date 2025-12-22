# Testing the Dex Provider

## Prerequisites

1. **Docker running**: Start Docker Desktop or Colima
2. **Pulumi CLI installed**: `pulumi version`
3. **Provider installed**: `pulumi plugin ls | grep dex`
4. **SDK installed**: `npm install` in this directory

## Step 1: Start Local Dex

```bash
# From the project root
cd /Users/andreas/dev/kotaicode/pulumi-provider-dex
docker-compose up -d

# Verify Dex is running
curl http://localhost:5556/healthz
# Should return: ok

# Check gRPC port
netstat -an | grep 5557
```

## Step 2: Initialize Pulumi Stack

```bash
cd examples/typescript

# Create a new stack
pulumi stack init dev

# Or use an existing stack
pulumi stack select dev
```

## Step 3: Preview Changes

```bash
# Preview what will be created
pulumi preview
```

You should see:
- A `dex:index:Client` resource (webClient)
- A `dex:index:Connector` resource (generic-oidc)

## Step 4: Apply Changes

```bash
# Create the resources
pulumi up

# You'll be prompted to confirm. Type 'yes' to proceed.
```

## Step 5: Verify Resources

After `pulumi up` succeeds:

1. **Check Dex Web UI**: http://localhost:5556
   - You should see the connectors listed
   - The client should be visible

2. **Check outputs**:
   ```bash
   pulumi stack output
   ```

3. **View specific output**:
   ```bash
   pulumi stack output webClientId
   pulumi stack output webClientSecret  # This is a secret
   ```

## Step 6: Test Updates

Edit `index.ts` to modify a resource, then:

```bash
pulumi preview  # See what will change
pulumi up       # Apply changes
```

## Step 7: Clean Up

```bash
# Destroy all resources
pulumi destroy

# Stop Dex
cd ../..
docker-compose down
```

## Troubleshooting

### "Provider not found" error

```bash
# Reinstall the provider
cd ../..
pulumi plugin install resource dex v0.1.0 --file bin/pulumi-resource-dex
```

### "Connection refused" error

- Ensure Dex is running: `docker-compose ps`
- Check Dex logs: `docker-compose logs dex`
- Verify port 5557 is accessible: `netstat -an | grep 5557`

### "Failed to create Dex client" error

- Check Dex gRPC is enabled in `examples/dex-config.yaml`
- Verify `DEX_API_CONNECTORS_CRUD=true` is set (handled by docker-compose)
- Check Dex logs for errors: `docker-compose logs dex`

### TypeScript compilation errors

```bash
# Reinstall dependencies
npm install

# Rebuild
npm run build
```

## Next Steps

Once basic testing works:

1. Test with real Azure credentials (uncomment Azure examples)
2. Test with real Cognito credentials (uncomment Cognito examples)
3. Test updates and deletes
4. Test error cases (duplicate IDs, invalid configs, etc.)

