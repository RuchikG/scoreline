#!/bin/bash

# Scoreline highlights integration test
# Usage: ./test-scoreline-highlights.sh <match_id>

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

cd "$PROJECT_ROOT"

if [ "$1" = "--help" ] || [ "$1" = "-h" ] || [ -z "$1" ]; then
    echo "🔬 Scoreline Highlights Integration Test"
    echo ""
    echo "Usage: $0 <match_id>"
    echo ""
    echo "This tool tests the complete scoreline highlights pipeline:"
    echo "  1. Raw API response analysis"
    echo "  2. Scoreline FotMob client parsing"
    echo "  3. MatchDetails structure validation"
    echo "  4. UI display logic simulation"
    echo ""
    echo "Examples:"
    echo "  $0 4803233    Test highlights for match ID 4803233"
    echo ""
    exit 0
fi

echo "🔬 Testing scoreline highlights pipeline for match ID: $1"
echo ""

go run scripts/test_golazo_highlights.go "$1"
