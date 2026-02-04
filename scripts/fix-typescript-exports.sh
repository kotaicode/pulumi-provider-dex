#!/bin/bash
# Post-process TypeScript SDK index.ts for isolatedModules and top-level exports.
# 1. Fix ProviderArgs to use "export type" (codegen emits "export { ProviderArgs }").
# 2. Add top-level resource re-exports so: import * as dex from "@kotaicode/pulumi-dex"; dex.Client works.

set -e

INDEX_FILE="sdk/typescript/nodejs/index.ts"

if [ ! -f "$INDEX_FILE" ]; then
    echo "Error: $INDEX_FILE not found"
    exit 1
fi

echo "Fixing TypeScript SDK exports (ProviderArgs + top-level resources)..."

# 1. Always fix ProviderArgs for isolatedModules (codegen overwrites this every time)
if grep -q 'export { ProviderArgs }' "$INDEX_FILE"; then
    sed -i.bak 's/export { ProviderArgs }/export type { ProviderArgs }/' "$INDEX_FILE"
    rm -f "${INDEX_FILE}.bak"
    echo "✓ Fixed ProviderArgs export for isolatedModules"
fi

# 2. Add top-level resource re-exports if not already present
if grep -q "Re-export resources at top level" "$INDEX_FILE"; then
    echo "Top-level resource exports already present, skipping insert."
else
    # Run Python script from stdin; pass INDEX_FILE as argv[1] ( - = read script from stdin)
    python3 - "$INDEX_FILE" << 'PYTHON_SCRIPT'
import sys

index_file = sys.argv[1]

with open(index_file, 'r') as f:
    content = f.read()

if "Re-export resources at top level" in content:
    print("Already present, skipping.")
    sys.exit(0)

insert_marker = "pulumi.runtime.registerResourcePackage"
if insert_marker not in content:
    print("Error: Could not find " + insert_marker + " in " + index_file, file=sys.stderr)
    sys.exit(1)

re_exports = """
// Re-export resources at top level for easier access.
// This allows:
//   import * as dex from "@kotaicode/pulumi-dex";
//   const c: dex.AzureOidcConnector = new dex.AzureOidcConnector(...);
//
// When isolatedModules is enabled, value and type exports must be separate.
export {
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

export type {
    AzureMicrosoftConnectorArgs,
    AzureOidcConnectorArgs,
    ClientArgs,
    CognitoOidcConnectorArgs,
    ConnectorArgs,
    GitHubConnectorArgs,
    GitLabConnectorArgs,
    GoogleConnectorArgs,
    LocalConnectorArgs,
} from "./resources";

export * as inputs from "./types/input";
export * as outputs from "./types/output";

"""

content = content.replace(insert_marker, re_exports + insert_marker)

with open(index_file, 'w') as f:
    f.write(content)

print("✓ Added top-level resource/type exports to TypeScript SDK")
PYTHON_SCRIPT
fi

echo "✓ TypeScript SDK exports fixed"
