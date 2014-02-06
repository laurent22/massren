#!/usr/bin/env bash
wget "https://github.com/laurent22/massren/releases/download/v1.0.1/massren.linux-amd64.tar.gz"
tar xvzf massren.linux-amd64.tar.gz
chmod 755 massren
mv massren /usr/bin
