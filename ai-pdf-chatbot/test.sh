#!/bin/bash

# AI PDF Chatbot Test Script
# This script demonstrates how to test the chatbot

echo "======================================"
echo "AI PDF Chatbot - Test Script"
echo "======================================"
echo ""

export OPENAI_MODEL=ernie-4.5-turbo-128k

# Check if OPENAI_API_KEY is set
if [ -z "$OPENAI_API_KEY" ]; then
    echo "❌ Error: OPENAI_API_KEY environment variable is not set"
    echo ""
    echo "Please set your OpenAI API key:"
    echo "  export OPENAI_API_KEY='sk-...'"
    echo ""
    exit 1
fi

echo "✅ OPENAI_API_KEY is set"
echo ""

# Change to the correct directory
cd "$(dirname "$0")/backend" || exit 1

# Display menu
echo "Choose an option:"
echo "  1) Ingest test document"
echo "  2) Start interactive chat"
echo "  3) Start web server"
echo "  4) Ingest and chat"
echo ""
read -p "Enter choice [1-4]: " choice

case $choice in
    1)
        echo ""
        echo "📂 Ingesting test_document.txt..."
        ./ai-pdf-chatbot -ingest ../test_document.txt
        ;;
    2)
        echo ""
        echo "💬 Starting interactive chat mode..."
        echo "Type 'exit' to quit"
        echo ""
        ./ai-pdf-chatbot -chat
        ;;
    3)
        echo ""
        echo "🚀 Starting web server on http://localhost:8080"
        echo "Press Ctrl+C to stop"
        echo ""
        ./ai-pdf-chatbot -server
        ;;
    4)
        echo ""
        echo "📂 Ingesting test_document.txt..."
        ./ai-pdf-chatbot -ingest ../test_document.txt
        echo ""
        echo "💬 Starting chat mode..."
        ./ai-pdf-chatbot -chat
        ;;
    *)
        echo "Invalid choice"
        exit 1
        ;;
esac
