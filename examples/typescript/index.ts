import * as pulumi from "@pulumi/pulumi";
import * as dex from "@kotaicode/pulumi-dex";

// Configure the Dex provider for local development
// Note: The SDK currently requires all fields, but for local testing with insecure Dex,
// we can use empty strings for TLS fields since insecureSkipVerify is true
const dexProvider = new dex.Provider("dex", {
    host: "localhost:5557", // Dex gRPC endpoint
    // For development with insecure Dex, use empty strings for TLS fields
    // (These won't be used when insecureSkipVerify is true)
    caCert: "",
    clientCert: "",
    clientKey: "",
    insecureSkipVerify: true,
    timeoutSeconds: 5,
});

// Example 1: Create an OAuth2 client
// Use stable IDs so Pulumi can properly manage resource lifecycle (create/update/delete)
const webClient = new dex.Client("webClient", {
    clientId: "my-web-app", // Stable ID - Pulumi will update existing or create new
    name: "My Web App",
    redirectUris: ["http://localhost:3000/callback"],
    // secret is optional - will be auto-generated if omitted
}, { provider: dexProvider });

// Example 2: Create a generic OIDC connector (simplest example for testing)
// Use stable IDs so Pulumi can properly manage resource lifecycle (create/update/delete)
const genericOidcConnector = new dex.Connector("generic-oidc", {
    connectorId: "generic-oidc", // Stable ID - Pulumi will update existing or create new
    type: "oidc",
    name: "Generic OIDC Provider",
    oidcConfig: {
        issuer: "https://example.com",
        clientId: "test-client-id",
        clientSecret: "test-client-secret",
        redirectUri: "http://localhost:5556/dex/callback",
        scopes: ["openid", "email", "profile"],
    },
}, { provider: dexProvider });

// Export outputs
export const webClientId = webClient.clientId;
export const webClientSecret = webClient.secret; // This is a Pulumi secret
export const genericConnectorId = genericOidcConnector.connectorId;

// Uncomment these examples when you have actual Azure/Cognito credentials:

/*
// Example 3: Create an Azure/Entra ID connector using generic OIDC
const azureConnector = new dex.AzureOidcConnector("azure-tenant", {
    connectorId: "azure-tenant",
    name: "Azure AD",
    tenantId: "your-tenant-id-here", // Replace with your Azure tenant ID
    clientId: "your-azure-app-client-id",
    clientSecret: "your-azure-app-client-secret",
    redirectUri: "http://localhost:5556/dex/callback",
    userNameSource: "preferred_username",
}, { provider: dexProvider });

// Example 4: Create an AWS Cognito connector
const cognitoConnector = new dex.CognitoOidcConnector("cognito", {
    connectorId: "cognito",
    name: "AWS Cognito",
    region: "us-east-1",
    userPoolId: "us-east-1_XXXXXXXXX", // Replace with your Cognito pool ID
    clientId: "your-cognito-client-id",
    clientSecret: "your-cognito-client-secret",
    redirectUri: "http://localhost:5556/dex/callback",
    userNameSource: "email",
}, { provider: dexProvider });

export const azureConnectorId = azureConnector.connectorId;
export const cognitoConnectorId = cognitoConnector.connectorId;
*/
