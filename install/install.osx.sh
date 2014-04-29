#!/usr/bin/env bash
if [ -f "massren.osx.tar.gz" ]; then mv -f "massren.osx.tar.gz" "massren.osx.tar.gz.old" ; fi
wget "https://github.com/laurent22/massren/releases/download/v1.2.1/massren.osx.tar.gz"
tar xvzf massren.osx.tar.gz
chmod 755 massren
mv massren /usr/bin
