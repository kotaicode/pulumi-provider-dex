#!/bin/bash
# Quick local test before publishing: build TS SDK, link, run test-npm-package (assumes make test-npm-package = generate-schema + generate-sdks already ran).
set -e

REPO_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$REPO_ROOT"

echo "=== 1. Building TypeScript SDK (outDir=. so package is consumable) ==="
SDK_TS="$REPO_ROOT/sdk/typescript/nodejs"
cd "$SDK_TS"
# Build with outDir "." so index.js/index.d.ts end up at package root (needed for npm link / tsc resolution)
TS_CONFIG_BACKUP="$SDK_TS/tsconfig.json.bak"
cp tsconfig.json "$TS_CONFIG_BACKUP"
node -e "
const fs = require('fs');
const c = JSON.parse(fs.readFileSync('tsconfig.json', 'utf8'));
c.compilerOptions.outDir = '.';
fs.writeFileSync('tsconfig.json', JSON.stringify(c, null, 4));
"
npm install --silent
npm run build
mv "$TS_CONFIG_BACKUP" tsconfig.json

echo ""
echo "=== 2. Linking SDK and running test-npm-package ==="
npm link
cd "$REPO_ROOT/test-npm-package"
npm install --silent
npm link @kotaicode/pulumi-dex
echo ""
npm run test

echo ""
echo "=== Done: test-npm-package passed (isolatedModules + types) ==="
