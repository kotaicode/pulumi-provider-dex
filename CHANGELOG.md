# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- `dex.GitLabConnector` resource for GitLab.com and self-hosted GitLab instances
- `dex.GitHubConnector` resource for GitHub.com and GitHub Enterprise
- `dex.GoogleConnector` resource for Google Workspace and Google accounts
- `dex.LocalConnector` resource for local/builtin authentication
- GitHub Actions CI workflow for build, test, and lint
- Preview mode support (FR6.1) - simulate Dex calls without side effects during `pulumi preview`
- Improved error messages (FR6.2) - human-friendly error wrapping with context
- golangci-lint configuration for code quality
- Integration test infrastructure with Docker Compose

### Changed
- Error messages now include operation, resource type, and resource ID for better debugging

## [0.1.0] - 2025-01-XX

### Added
- Initial release of Pulumi provider for Dex
- `dex.Client` resource for managing OAuth2 clients
- `dex.Connector` resource for generic connector management
- `dex.AzureOidcConnector` resource for Azure AD/Entra ID (OIDC)
- `dex.AzureMicrosoftConnector` resource for Azure AD/Entra ID (Microsoft-specific)
- `dex.CognitoOidcConnector` resource for AWS Cognito user pools
- Provider configuration with TLS/mTLS support
- Secret generation for clients
- Local development environment with Docker Compose
- TypeScript and Go SDK generation
- Example Pulumi programs

