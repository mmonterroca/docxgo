#!/bin/bash
# Script to run all examples and generate .docx files for validation

set -e  # Exit on error

echo "🚀 Running all examples to generate .docx files..."
echo ""

# Change to examples directory
cd "$(dirname "$0")"

EXAMPLES=(
    "01_basic"
    "02_intermediate"
    "04_fields"
    "05_styles"
    "06_sections"
    "07_advanced"
    "08_images"
    "09_advanced_tables"
)

FAILED=()
PASSED=()
GENERATED_FILES=()

for example in "${EXAMPLES[@]}"; do
    echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    echo "📝 Running $example..."
    
    if cd "$example"; then
        # Build
        if go build -o ./example_binary 2>&1; then
            # Run the example
            if ./example_binary 2>&1; then
                echo "✅ $example executed successfully"
                PASSED+=("$example")
                
                # Find generated .docx files
                docx_files=$(find . -name "*.docx" -type f 2>/dev/null || true)
                if [ -n "$docx_files" ]; then
                    while IFS= read -r file; do
                        file_path="$example/$(basename "$file")"
                        GENERATED_FILES+=("$file_path")
                        echo "   📄 Generated: $(basename "$file")"
                    done <<< "$docx_files"
                fi
                
                # Clean up binary
                rm -f ./example_binary
            else
                echo "❌ $example execution failed"
                FAILED+=("$example")
            fi
        else
            echo "❌ $example build failed"
            FAILED+=("$example")
        fi
        cd ..
    else
        echo "❌ Could not enter directory $example"
        FAILED+=("$example")
    fi
    echo ""
done

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 SUMMARY"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Passed: ${#PASSED[@]}/${#EXAMPLES[@]}"
echo "❌ Failed: ${#FAILED[@]}/${#EXAMPLES[@]}"
echo "📄 Files generated: ${#GENERATED_FILES[@]}"

if [ ${#FAILED[@]} -gt 0 ]; then
    echo ""
    echo "Failed examples:"
    for example in "${FAILED[@]}"; do
        echo "  ❌ $example"
    done
fi

if [ ${#GENERATED_FILES[@]} -gt 0 ]; then
    echo ""
    echo "Generated .docx files:"
    for file in "${GENERATED_FILES[@]}"; do
        echo "  📄 $file"
    done
fi

echo ""
if [ ${#FAILED[@]} -gt 0 ]; then
    echo "❌ Some examples failed"
    exit 1
else
    echo "✅ All examples executed successfully!"
    echo "🎉 You can now open the generated .docx files to validate them"
    exit 0
fi
