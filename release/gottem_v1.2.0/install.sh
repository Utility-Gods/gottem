#!/bin/bash
mkdir -p ~/.config/gottem
cp gottem_$(uname -s | tr '[:upper:]' '[:lower:]')_amd64 /usr/local/bin/gottem
chmod +x /usr/local/bin/gottem
echo "Gottem v1.2.0 installed successfully!"
