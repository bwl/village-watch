#!/bin/bash

# Script to automatically update README.md with fresh village layout sample
# Usage: ./scripts/update-readme-sample.sh

set -e

echo "🏘️  Updating README with fresh village layout sample..."

# Generate fresh village layout
TEMP_OUTPUT=$(mktemp)
go run ./cmd/village-watch --test --path=. > "$TEMP_OUTPUT"

# Extract just the map portion (8 lines after "=== Village Map ===")
MAP_SAMPLE=$(grep -A 8 "=== Village Map ===" "$TEMP_OUTPUT" | tail -8)

# Create temporary files for processing
TEMP_README=$(mktemp)
TEMP_SAMPLE=$(mktemp)

# Write the new sample section to a temporary file
cat > "$TEMP_SAMPLE" <<'EOF'
```
=== Village Layout Test ===
Generated village with modular building designs:

EOF

# Add the map sample
echo "$MAP_SAMPLE" >> "$TEMP_SAMPLE"

# Add the features section
cat >> "$TEMP_SAMPLE" <<'EOF'

Features:
- Standardized # walls for consistent building structure
- Unicode interior symbols: 🏛 (libraries), ⌂ (cottages), ◦ (districts), ▲ (warehouses)
- Multiple designs per building type with size-based selection
- Modular system supporting easy addition of new building designs
- Deterministic layout based on file structure
```
EOF

# Use awk to replace the sample section in README.md
awk '
BEGIN { in_sample = 0; sample_file = "'"$TEMP_SAMPLE"'" }
/^Here'\''s what Village Watch generates for this project:$/ {
    print $0
    print ""
    # Read and print the new sample
    while ((getline line < sample_file) > 0) {
        print line
    }
    close(sample_file)
    in_sample = 1
    next
}
/^Test your own layout:/ && in_sample {
    print ""
    print $0
    in_sample = 0
    next
}
!in_sample { print }
' README.md > "$TEMP_README"

# Replace the original README
mv "$TEMP_README" README.md

# Clean up
rm "$TEMP_OUTPUT" "$TEMP_SAMPLE"

echo "✅ README.md updated with fresh village layout sample!"
echo ""
echo "📝 Changes made:"
echo "   - Generated new village layout using current codebase"
echo "   - Updated sample output in README.md"
echo "   - Preserved all other README content"
echo ""
echo "🚀 Don't forget to commit the changes:"
echo "   git add README.md"
echo "   git commit -m \"Update README with fresh village layout sample\""