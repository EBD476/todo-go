#!/bin/bash

echo "🚀 Building Todo App with Bubble Tea..."

# Install dependencies
echo "📦 Installing dependencies..."
go mod tidy

# Build basic version
echo "🔨 Building basic version..."
go build -o todo-basic main.go

# Build advanced version
echo "🔨 Building advanced version..."
go build -o todo-advanced advanced.go

echo "✅ Build complete!"
echo ""
echo "📝 Available versions:"
echo "  ./todo-basic     - Basic todo app with core features"
echo "  ./todo-advanced  - Advanced todo app with categories, priorities, and due dates"
echo ""
echo "🎯 Run either version to start the interactive todo app!"
