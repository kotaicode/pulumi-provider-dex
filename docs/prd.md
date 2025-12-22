Product Requirements Document (PRD-style)
6.1 Overview

Problem

Kotaicode uses Dex as a central IdP, and needs to manage multiple upstream IdPs (Azure AD/Entra applications, Cognito user pools) in a reproducible, versioned, environment-aware way.

Currently, Dex configuration is either static (YAML) or manually managed via gRPC / Terraform, which is error-prone, not fully integrated with the rest of the infra code, and awkward for multi-tenant scenarios.

Solution

Build a Pulumi provider (dex) that lets us manage Dex clients and connectors (IdPs) as Pulumi resources, with high-level helpers for AzureAD and Cognito OIDC configuration.

6.2 Goals

Manage Dex OAuth2 clients as code from Pulumi.

Manage Dex connectors (IdPs) as code, with:

Generic connector support.

First-class OIDC support.

Opinionated helpers for Azure AD and Cognito user pools.

Fit into existing Pulumi workflows (per-env stacks, CI/CD).

Hide Dex gRPC complexity (including mTLS) from most users.

6.3 Non-goals (initially)

Managing all Dex internal objects (password DB, refresh tokens, etc.).

Providing full coverage of all connector types (LDAP, SAML, etc.)—beyond what’s needed for current use cases.

UI or web console on top—this is a code-only interface.

6.4 Users & personas

Infra/Platform engineers: define Dex IdP config beside Kubernetes clusters, apps, etc.

Tenant onboarding automation: internal tooling that creates new tenants and wants to add matching connectors in Dex.

6.5 User stories

As a platform engineer, I can define Dex clients for my apps in Pulumi so that every environment uses the same client IDs and redirect URIs.

As a platform engineer, I can add a new Azure tenant as an IdP to Dex by defining an AzureOidcConnector resource, and Pulumi will create/update the connector via the Dex API.

As a platform engineer, I can add a Cognito user pool as an IdP to Dex by defining a CognitoOidcConnector.

As a team, I can run pulumi up in CI and have changes to IdPs applied consistently to Dex, with drift detection and preview.

6.6 Functional requirements

FR1 – Provider configuration

FR1.1: Provider must accept Dex gRPC host and TLS configuration.

FR1.2: Provider must validate connectivity at first resource operation and fail early on misconfig.

FR1.3: Provider must support multiple Dex instances via separate provider instances in a Pulumi stack.

FR2 – Client management

FR2.1: dex.Client resource supports:

clientId, name, secret, redirectUris, public, trustedPeers.

FR2.2: Provider must create, update, delete clients via Dex gRPC.

FR2.3: secret must be treated as a Pulumi secret (encrypted in state).

FR2.4: If secret is omitted, provider must generate a strong secret and return it as output.

FR3 – Connector management (generic)

FR3.1: dex.Connector resource supports:

connectorId, type, name, oidcConfig, rawConfig.

FR3.2: Exactly one of oidcConfig or rawConfig must be supplied.

FR3.3: Provider must create, update, delete connectors via Dex gRPC.

FR3.4: For oidcConfig, provider must serialize to Dex OIDC connector JSON format.

FR3.5: On Read, provider should attempt to map the JSON back into oidcConfig where possible; otherwise, fall back to rawConfig.

FR4 – Azure OIDC convenience

FR4.1: AzureOidcConnector must accept tenantId, clientId, clientSecret, redirectUri, optional scopes & userNameSource.

FR4.2: Provider must derive the issuer from tenantId.

FR4.3: Provider must internally create/update a Dex OIDC connector with the right JSON.

FR4.4: Changing tenantId or connectorId should force a replace (delete + recreate).

FR5 – Cognito OIDC convenience

FR5.1: CognitoOidcConnector must accept region, userPoolId, clientId, clientSecret, redirectUri, scopes, userNameSource.

FR5.2: Provider must derive issuer from region and userPoolId.

FR5.3: Same CRUD semantics as Azure connector.

FR6 – Error handling & preview

FR6.1: On Pulumi preview, provider must simulate Dex calls (no side effects) and return diffs.

FR6.2: Provider should surface human-friendly errors (e.g. “Dex gRPC: connector already exists” with context).

6.7 Non-functional requirements

Security

TLS keys/secrets must be treated as Pulumi secrets.

No logging of secrets.

mTLS strongly recommended and clearly documented.

Performance

Cache gRPC connection per provider instance.

Reasonable timeouts; no unbounded hangs.

Compatibility

Support Dex ≥ version that exposes current gRPC API & connector CRUD feature flag.

Document tested Dex versions and behaviour if DEX_API_CONNECTORS_CRUD is not enabled.

Reliability

If Dex is unreachable, provider should fail cleanly and not leave partial state in Pulumi.

6.8 Open questions for design review

Terraform reuse: do we want to also provide a Terraform bridge variant (using pulumi-terraform-bridge or “Any Terraform Provider”) for backwards compatibility?

Other connectors: how soon do we need first-class resources for non-OIDC connectors (LDAP/SAML/GitHub)?

Import story: do we need pulumi import support for existing Dex configs in phase 1, or can that wait?