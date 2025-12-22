# dex-debug

A debugging and testing tool for Dex gRPC API, built with Go and Kong CLI framework.

## Structure

This project follows standard Go CLI application structure:

```
dex-debug/
├── cmd/
│   └── dex-debug/
│       └── main.go          # Entry point and CLI definition
├── internal/
│   ├── client/              # Dex client connection logic
│   │   └── client.go
│   └── commands/            # Command implementations
│       ├── base.go          # Base command struct
│       ├── verify.go
│       ├── cleanup.go
│       ├── test_delete.go
│       ├── test_verification.go
│       ├── test_delete_direct.go
│       ├── test_delete_mywebapp.go
│       └── common.go        # Shared helpers
├── go.mod
├── go.sum
└── README.md
```

## Building

```bash
go build -o dex-debug ./cmd/dex-debug
```

## Usage

```bash
# List all clients and connectors
./dex-debug verify
# or
./dex-debug list

# Clean up test resources
./dex-debug cleanup

# Test deleting a specific client
./dex-debug test-delete <client-id>

# Test delete with full verification cycle
./dex-debug test-delete-direct

# Test delete with "my-web-app" client
./dex-debug test-delete-my-web-app

# Test delete verification logic
./dex-debug test-verification

# Specify a different Dex host
./dex-debug -H localhost:5557 verify
```

## Commands

- `verify` (aliases: `list`) - List all clients and connectors in Dex
- `cleanup` - Clean up test clients and connectors (excluding static ones)
- `test-delete <client-id>` - Test deleting a specific client by ID
- `test-delete-direct` - Test DeleteClient API with a test client (creates, deletes, verifies)
- `test-delete-my-web-app` - Test DeleteClient API with 'my-web-app' client
- `test-verification` - Test delete verification logic with 'my-web-app' client

## Flags

- `-H, --host` - Dex gRPC host:port (default: "localhost:5557")
