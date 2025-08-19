#!/bin/bash
# Installation script for certwiz/cert

set -e

echo "Installing cert (certwiz)..."

# Build the binary
make build

# Install to /usr/local/bin
if [ -w /usr/local/bin ]; then
    cp cert /usr/local/bin/
    echo "✓ Installed cert to /usr/local/bin/"
else
    echo "Need sudo to install to /usr/local/bin"
    sudo cp cert /usr/local/bin/
    echo "✓ Installed cert to /usr/local/bin/"
fi

# Optional: Create certwiz symlink for backward compatibility
read -p "Create 'certwiz' symlink for backward compatibility? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    if [ -w /usr/local/bin ]; then
        ln -sf /usr/local/bin/cert /usr/local/bin/certwiz
    else
        sudo ln -sf /usr/local/bin/cert /usr/local/bin/certwiz
    fi
    echo "✓ Created certwiz symlink"
fi

echo ""
echo "Installation complete! You can now use:"
echo "  cert inspect google.com"
echo ""