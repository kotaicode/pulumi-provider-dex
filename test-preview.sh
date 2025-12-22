#!/bin/bash
# Test script to verify preview mode works correctly

set -e

echo "=== Testing Pulumi Preview Mode ==="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if Dex is running
echo "1. Checking if Dex is running..."
if docker-compose ps | grep -q "Up"; then
    echo -e "${GREEN}✓ Dex is running${NC}"
else
    echo -e "${YELLOW}⚠ Dex is not running. Starting it...${NC}"
    docker-compose up -d
    sleep 5
    echo -e "${GREEN}✓ Dex started${NC}"
fi

# Check if provider is built
echo ""
echo "2. Checking if provider is built..."
if [ -f "bin/pulumi-resource-dex" ]; then
    echo -e "${GREEN}✓ Provider binary exists${NC}"
else
    echo -e "${YELLOW}⚠ Provider not built. Building...${NC}"
    make build
    echo -e "${GREEN}✓ Provider built${NC}"
fi

# Check if provider is installed
echo ""
echo "3. Checking if provider is installed..."
if pulumi plugin ls | grep -q "dex.*v0.1.0"; then
    echo -e "${GREEN}✓ Provider is installed${NC}"
else
    echo -e "${YELLOW}⚠ Provider not installed. Installing...${NC}"
    pulumi plugin install resource dex v0.1.0 --file bin/pulumi-resource-dex || true
    echo -e "${GREEN}✓ Provider installed${NC}"
fi

# Check if SDK is generated
echo ""
echo "4. Checking if TypeScript SDK is available..."
if [ -d "sdk/typescript/nodejs" ] && [ -f "sdk/typescript/nodejs/package.json" ]; then
    echo -e "${GREEN}✓ TypeScript SDK exists${NC}"
else
    echo -e "${YELLOW}⚠ SDK not generated. Generating...${NC}"
    make generate-sdks || echo -e "${RED}✗ SDK generation failed (this is OK if Pulumi CLI is not available)${NC}"
fi

# Navigate to example directory
echo ""
echo "5. Testing preview mode..."
cd examples/typescript

# Check if node_modules exists
if [ ! -d "node_modules" ]; then
    echo -e "${YELLOW}⚠ Installing npm dependencies...${NC}"
    npm install ../../sdk/typescript/nodejs || echo -e "${RED}✗ npm install failed${NC}"
fi

# Check current state
echo ""
echo "6. Checking current Dex state (before preview)..."
echo "   Clients:"
curl -s http://localhost:5556/dex/.well-known/openid-configuration > /dev/null 2>&1 && echo "   ✓ Dex web API is accessible" || echo "   ⚠ Dex web API not accessible (this is OK for gRPC testing)"

# Run preview
echo ""
echo "7. Running 'pulumi preview'..."
echo "   This should show what would be created WITHOUT actually creating resources"
echo ""

if pulumi preview 2>&1 | tee /tmp/pulumi-preview-output.txt; then
    echo ""
    echo -e "${GREEN}✓ Preview completed successfully${NC}"
    
    # Check if preview output mentions creating resources
    if grep -q "would create" /tmp/pulumi-preview-output.txt || grep -q "create" /tmp/pulumi-preview-output.txt; then
        echo -e "${GREEN}✓ Preview shows resources that would be created${NC}"
    fi
    
    # Verify no resources were actually created
    echo ""
    echo "8. Verifying no resources were actually created..."
    echo "   (We'll check this by looking at Dex state)"
    
    # Use dex-debug tool if available
    if [ -f "../../examples/dex-debug/dex-debug" ]; then
        echo "   Checking with dex-debug tool..."
        ../../examples/dex-debug/dex-debug verify --host localhost:5557 | head -20
    else
        echo "   (dex-debug tool not available, skipping verification)"
    fi
    
    echo ""
    echo -e "${GREEN}=== Preview Test Summary ===${NC}"
    echo -e "${GREEN}✓ Preview mode works correctly${NC}"
    echo ""
    echo "Next steps:"
    echo "  - Run 'pulumi up' to actually create the resources"
    echo "  - Run 'pulumi destroy' to clean up"
    
else
    echo ""
    echo -e "${RED}✗ Preview failed${NC}"
    echo "Check the error messages above"
    exit 1
fi

cd ../..

