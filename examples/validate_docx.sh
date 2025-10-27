#!/bin/bash
# Script to validate the integrity of all generated .docx files

echo "🔍 Validating .docx file integrity..."
echo ""

VALID=0
INVALID=0

for file in /Users/mmonterroca/code/go-docx/examples/*/*.docx; do
    # Skip temp files
    if [[ "$file" =~ ~\$ ]]; then
        continue
    fi
    
    filename=$(basename "$file")
    
    # Test ZIP integrity
    if unzip -t "$file" > /dev/null 2>&1; then
        # Check for required OOXML files
        has_content_types=$(unzip -l "$file" | grep -q "\[Content_Types\].xml" && echo "yes" || echo "no")
        has_rels=$(unzip -l "$file" | grep -q "_rels/.rels" && echo "yes" || echo "no")
        has_document=$(unzip -l "$file" | grep -q "word/document.xml" && echo "yes" || echo "no")
        has_styles=$(unzip -l "$file" | grep -q "word/styles.xml" && echo "yes" || echo "no")
        
        if [[ "$has_content_types" == "yes" && "$has_rels" == "yes" && "$has_document" == "yes" && "$has_styles" == "yes" ]]; then
            echo "✅ $filename - VALID"
            ((VALID++))
        else
            echo "⚠️  $filename - INCOMPLETE (missing required files)"
            echo "   [Content_Types].xml: $has_content_types"
            echo "   _rels/.rels: $has_rels"
            echo "   word/document.xml: $has_document"
            echo "   word/styles.xml: $has_styles"
            ((INVALID++))
        fi
    else
        echo "❌ $filename - CORRUPTED (invalid ZIP)"
        ((INVALID++))
    fi
done

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "📊 VALIDATION SUMMARY"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ Valid: $VALID"
echo "❌ Invalid: $INVALID"
echo ""

if [ $INVALID -gt 0 ]; then
    echo "❌ Some files are invalid or corrupted"
    exit 1
else
    echo "✅ All .docx files are valid OOXML documents!"
    exit 0
fi
