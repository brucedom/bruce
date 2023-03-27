#!/bin/bash

# For example sake we will use the latest version of bruce
CURVER=$(curl --silent "https://api.github.com/repos/Nitecon/bruce/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'|cut -c2-)

mkdir -p /opt/bruce/${CURVER}
cd /opt/bruce/${CURVER}
wget https://github.com/Nitecon/bruce/releases/download/v${CURVER}/bruce_${CURVER}_linux_amd64.tar.gz
tar xf bruce_${CURVER}_linux_amd64.tar.gz
ln -s /opt/bruce/${CURVER}/bruce /usr/bin/bruce
/usr/bin/bruce --config https://raw.githubusercontent.com/Nitecon/bruce/main/examples/nginx/install.yml > /var/log/bruce.log
