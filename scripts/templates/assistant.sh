#!/bin/sh
set -e
#
# This script is meant for quick & easy install via:
#   'curl -sSL https://github.com/leominov/self-monitoring/releases/download/TMP_VERSION/assistant.sh | sh'
# or:
#   'wget -qO- https://github.com/leominov/self-monitoring/releases/download/TMP_VERSION/assistant.sh | sh'
#

VERSION="TMP_VERSION"

PLATFORM=`uname -s`
ARCH=`uname -m`

if [ "$PLATFORM" = "Linux" ] && [ "$ARCH" = "i686" ]; then
    ARCH="i386"
fi

NAME="gomon-$PLATFORM-$ARCH.tar.gz"
URL="https://github.com/leominov/self-monitoring/releases/download/$VERSION/$NAME"

echo "--- Downloading $PLATFORM/$ARCH"
curl -OL $URL
tar -xf $NAME
rm $NAME

echo "--- Preparing for installation $VERSION"
if [ ! -f "config.json" ]; then
    echo "    Creating new configuration (config.json)"
    mv example.config.json config.json
fi
echo " !  Please do not forget to update your monitoring configuration and restart the monitoring:"
echo "    service gomon restart"

echo "--- Installing"
sudo ./install.sh
