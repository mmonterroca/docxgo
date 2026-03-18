#!/bin/bash
# Run all Node.js examples and validate they produce valid .docx files
set -e

DIR="$(cd "$(dirname "$0")" && pwd)"
PASS=0
FAIL=0

run_example() {
  local name="$1" file="$2"
  echo "=== $name ==="
  if npx tsx "$DIR/$file" 2>&1; then
    PASS=$((PASS+1))
  else
    echo "  ✗ FAILED"
    FAIL=$((FAIL+1))
  fi
  echo ""
}

echo "========================================"
echo "  docxgo Node.js Examples"
echo "========================================"
echo ""

run_example "01: Basic Document" "01_basic.ts"
run_example "02: Tables and Sections" "02_tables_and_sections.ts"
run_example "03: Read, Modify, Save" "03_read_modify.ts"
run_example "04: Template / Mail Merge" "04_template.ts"
run_example "05: Patch and System" "05_patch_and_system.ts"

echo "========================================"
echo "  Output files:"
echo "========================================"
ls -la /tmp/node_0*.docx 2>/dev/null | awk '{print "  " $NF " (" $5 " bytes)"}'
echo ""
for f in /tmp/node_0*.docx; do
  [ -f "$f" ] && echo "  $(file "$f")"
done

echo ""
echo "========================================"
echo "  Results: $PASS passed, $FAIL failed"
echo "========================================"

[ "$FAIL" -eq 0 ] && exit 0 || exit 1
