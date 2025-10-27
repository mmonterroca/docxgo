#!/bin/bash
# Test script to verify all working examples compile

set -e  # Exit on error

echo "🧪 Testing all working examples..."
echo ""

EXAMPLES=(
    "01_basic"
    "02_intermediate"
    "04_fields"
    "05_styles"
    "06_sections"
    "07_advanced"
    "08_images"
    "09_advanced_tables"
    "basic"
)

FAILED=()
PASSED=()

for example in "${EXAMPLES[@]}"; do
    echo -n "Testing $example... "
    if cd "$example" && go build -o /dev/null 2>&1; then
        echo "✅ PASS"
        PASSED+=("$example")
        cd ..
    else
        echo "❌ FAIL"
        FAILED+=("$example")
        cd .. 2>/dev/null || true
    fi
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Results:"
echo "  Passed: ${#PASSED[@]}/${#EXAMPLES[@]}"
echo "  Failed: ${#FAILED[@]}/${#EXAMPLES[@]}"

if [ ${#FAILED[@]} -gt 0 ]; then
    echo ""
    echo "Failed examples:"
    for example in "${FAILED[@]}"; do
        echo "  - $example"
    done
    exit 1
else
    echo ""
    echo "✅ All working examples compiled successfully!"
    exit 0
fi
