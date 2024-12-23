#!/bin/bash

# Variables
INSTALL_DIR="/tmp"
BINARY_URL="https://github.com/LazyBST/KMAgent/blob/main/bin/kmagent-linux-arm"
BINARY_NAME="kmagent-linux-arm"
CONFIG_FILE_PATH_TARGET="/tmp/otelcol.yaml"
CONFIG_SERVICE_ORIGIN_URL="http://localhost:3000"

curl -L "$BINARY_URL" -o "$INSTALL_DIR"

chmod +x "$INSTALL_DIR/$BINARY_NAME"

"$INSTALL_DIR/$BINARY_NAME" -configFilePath="$CONFIG_FILE_PATH_TARGET" -configServiceOriginUrl="$CONFIG_SERVICE_ORIGIN_URL"
