#!/bin/bash
# Post-process TypeScript SDK index.ts to add top-level exports for types
# This allows: import * as dex from "@kotaicode/pulumi-dex"; const connector: dex.AzureOidcConnector = ...

set -e

INDEX_FILE="sdk/typescript/nodejs/index.ts"

if [ ! -f "$INDEX_FILE" ]; then
    echo "Error: $INDEX_FILE not found"
    exit 1
fi

echo "Adding top-level type exports to TypeScript SDK..."

# Check if exports already exist
if grep -q "Re-export resources at top level" "$INDEX_FILE"; then
    echo "Top-level exports already exist, skipping..."
    exit 0
fi

# Add re-exports before the pulumi.runtime.registerResourcePackage call
# We'll insert after the sub-module exports
python3 << 'PYTHON_SCRIPT'
import re
import sys

index_file = sys.argv[1]

with open(index_file, 'r') as f:
    content = f.read()

# Check if already modified
if "Re-export resources at top level" in content:
    print("Already modified, skipping...")
    sys.exit(0)

# Find the position to insert (before pulumi.runtime.registerResourcePackage)
insert_marker = "pulumi.runtime.registerResourcePackage"
if insert_marker not in content:
    print(f"Error: Could not find {insert_marker} in {index_file}")
    sys.exit(1)

# Create the re-export block
re_exports = """
// Re-export resources at top level for easier access
// This allows: import * as dex from "@kotaicode/pulumi-dex"; const connector: dex.AzureOidcConnector = ...
export {
    AzureMicrosoftConnector,
    AzureMicrosoftConnectorArgs,
    AzureOidcConnector,
    AzureOidcConnectorArgs,
    Client,
    ClientArgs,
    CognitoOidcConnector,
    CognitoOidcConnectorArgs,
    Connector,
    ConnectorArgs,
    GitHubConnector,
    GitHubConnectorArgs,
    GitLabConnector,
    GitLabConnectorArgs,
    GoogleConnector,
    GoogleConnectorArgs,
    LocalConnector,
    LocalConnectorArgs,
} from "./resources";

// Re-export resource types at top level for type annotations
export type {
    AzureMicrosoftConnector,
    AzureOidcConnector,
    Client,
    CognitoOidcConnector,
    Connector,
    GitHubConnector,
    GitLabConnector,
    GoogleConnector,
    LocalConnector,
} from "./resources";

// Re-export input/output types for easier access
export * as inputs from "./types/input";
export * as outputs from "./types/output";

"""

# Insert before the registerResourcePackage call
content = content.replace(insert_marker, re_exports + insert_marker)

with open(index_file, 'w') as f:
    f.write(content)

print("✓ Added top-level type exports to TypeScript SDK")
PYTHON_SCRIPT
    "$INDEX_FILE"

echo "✓ TypeScript SDK exports fixed"
