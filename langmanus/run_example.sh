#!/bin/bash

# LangManus Example Runner

echo "=== LangManus Test Runner ==="
echo ""

# Check if OPENAI_API_KEY is set
if [ -z "$OPENAI_API_KEY" ]; then
    echo "❌ ERROR: OPENAI_API_KEY not set"
    echo "Please set your API key:"
    echo "  export OPENAI_API_KEY='your-key-here'"
    exit 1
fi

# Check SEARCH_API_KEY
if [ -z "$SEARCH_API_KEY" ]; then
    echo "⚠️  WARNING: SEARCH_API_KEY not set"
    echo "Search functionality will be limited. Get a free key at:"
    echo "  https://tavily.com/"
    echo ""
fi

# Example query
QUERY="${1:-研究 2025 年机器学习趋势并创建摘要}"

echo "Running query: $QUERY"
echo ""

# Run with verbose output
export VERBOSE=true

./langmanus "$QUERY"
