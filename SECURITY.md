# Security Policy

## Supported Versions

We actively support the following versions with security updates:

| Version | Supported          |
| ------- | ------------------ |
| 0.1.x   | :white_check_mark: |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security vulnerability, please follow these steps:

### 1. **Do NOT open a public issue**

Security vulnerabilities should be reported privately to allow us to address them before public disclosure.

### 2. **Report via GitHub Security Advisories**

The preferred method is to use GitHub's Security Advisories feature:

1. Go to the [Security tab](https://github.com/kotaicode/pulumi-provider-dex/security) in the repository
2. Click "Report a vulnerability"
3. Fill out the security advisory form with details about the vulnerability

### 3. **Alternative: Email**

If you prefer, you can email security concerns to the maintainers. Please include:

- Description of the vulnerability
- Steps to reproduce
- Potential impact
- Suggested fix (if any)

### 4. **What to Expect**

- **Acknowledgment:** We will acknowledge receipt within 48 hours
- **Initial Assessment:** We will provide an initial assessment within 7 days
- **Updates:** We will provide regular updates on the status of the vulnerability
- **Resolution:** We will work to resolve the issue as quickly as possible

### 5. **Disclosure Policy**

- We will disclose the vulnerability publicly after a fix is available
- We will credit you for reporting the vulnerability (unless you prefer to remain anonymous)
- We will coordinate with you on the disclosure timeline if requested

## Security Best Practices

When using this provider:

1. **Use TLS/mTLS:** Always use secure connections to Dex in production
2. **Protect Secrets:** Never commit secrets or credentials to version control
3. **Keep Updated:** Regularly update to the latest version
4. **Review Permissions:** Ensure Dex gRPC API access is properly restricted
5. **Monitor Logs:** Monitor Dex and provider logs for suspicious activity

## Known Security Considerations

### Current Limitations

- **Development Mode:** The provider currently uses insecure gRPC connections by default for development. **Do not use this in production.**
- **TLS/mTLS:** Full TLS/mTLS support is planned but not yet implemented. See the roadmap for details.

### Recommendations

- Use Dex's built-in TLS/mTLS features in production
- Restrict network access to Dex's gRPC API
- Use Pulumi secrets for sensitive configuration values
- Regularly audit Dex client and connector configurations

## Security Updates

Security updates will be:

- Released as patch versions (e.g., 0.1.1, 0.1.2)
- Documented in CHANGELOG.md
- Announced via GitHub releases
- Tagged with the `security` label

## Thank You

We appreciate your help in keeping this project secure. Thank you for responsibly disclosing vulnerabilities!

