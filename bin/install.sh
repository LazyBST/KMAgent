#!/bin/bash

# Variables
INSTALL_DIR="/tmp"
BINARY_URL="https://raw.githubusercontent.com/LazyBST/KMAgent/refs/heads/main/bin/kmagent-linux-arm64"
BINARY_NAME="kmagent-linux-arm64"
CONFIG_FILE_PATH_TARGET="/tmp/otelcol.yaml"
CONFIG_SERVICE_ORIGIN_URL="http://localhost:3000"

curl -L "$BINARY_URL" -o "$INSTALL_DIR/$BINARY_NAME"

chmod +x "$INSTALL_DIR/$BINARY_NAME"

"$INSTALL_DIR/$BINARY_NAME" -configFilePath="$CONFIG_FILE_PATH_TARGET" -configServiceOriginUrl="$CONFIG_SERVICE_ORIGIN_URL"
