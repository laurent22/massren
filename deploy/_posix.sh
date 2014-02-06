#!/bin/bash

APPNAME=massren

# Get command line arguments
TARGET_OS=$1
TARGET_ARCH=$2
TARGET_FULL=$TARGET_OS
if [ -n "$TARGET_ARCH" ]; then
	TARGET_FULL=$TARGET_FULL-$TARGET_ARCH
fi

echo "Target = $TARGET_FULL"

# Get script current directory
SCRIPT_PWD="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_PWD/.."

# Build the executable
go build

mkdir -p "$SCRIPT_PWD/releases"
mv $APPNAME "$SCRIPT_PWD/releases"

# Create the archive filename
FILENAME=$APPNAME.$TARGET_FULL
FILENAME=$FILENAME.tar.gz

# Create the archive
echo "Creating $FILENAME..."
cd "$SCRIPT_PWD/releases"
tar czvf $FILENAME $APPNAME

# Get version number
VERSION=$(./$APPNAME --version)

INSTALL_FILE=$SCRIPT_PWD/../install/install.$TARGET_FULL.sh
echo "#!/usr/bin/env bash" > $INSTALL_FILE
echo "if [ -f \"$FILENAME\" ]; then mv -f \"$FILENAME\" \"$FILENAME.old\" ; fi" >> $INSTALL_FILE
echo "wget \"https://github.com/laurent22/massren/releases/download/v$VERSION/$FILENAME\"" >> $INSTALL_FILE
echo "tar xvzf $FILENAME" >> $INSTALL_FILE
echo "chmod 755 $APPNAME" >> $INSTALL_FILE
echo "mv $APPNAME /usr/bin" >> $INSTALL_FILE

chmod 755 $INSTALL_FILE

rm $APPNAME
