#!/bin/bash

# Post-entrypoint script to copy generated SVG to GitHub Actions summary
set -euo pipefail

# Default paths
TMP_SVG_PATH="tmp/${INPUT_OUTPUT_FILE_NAME}"

# Check if running in GitHub Actions environment
if [[ -z "$GITHUB_STEP_SUMMARY" ]]; then
  echo "Warning: Not running in GitHub Actions environment, skipping summary update"
  exit 0
fi

# Copy SVG content to GitHub Actions summary
echo "## Coding Metrics" >> "$GITHUB_STEP_SUMMARY"
echo "" >> "$GITHUB_STEP_SUMMARY"
echo "Generated coding metrics SVG:" >> "$GITHUB_STEP_SUMMARY"
echo "" >> "$GITHUB_STEP_SUMMARY"

# Embed SVG directly in the summary
echo '```svg' >> "$GITHUB_STEP_SUMMARY"
cat "$TMP_SVG_PATH" >> "$GITHUB_STEP_SUMMARY"
echo '```' >> "$GITHUB_STEP_SUMMARY"

echo "Successfully added SVG to GitHub Actions summary"
