#!/usr/bin/env bash
if [ -f "massren.linux-amd64.tar.gz" ]; then mv -f "massren.linux-amd64.tar.gz" "massren.linux-amd64.tar.gz.old" ; fi
wget "https://github.com/laurent22/massren/releases/download/v1.2.0/massren.linux-amd64.tar.gz"
tar xvzf massren.linux-amd64.tar.gz
chmod 755 massren
mv massren /usr/bin
