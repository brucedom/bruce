#!/bin/bash

# For example sake we will use the latest version of cfs
CURVER=$(curl --silent "https://api.github.com/repos/Nitecon/cfs/releases/latest" | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'|cut -c2-)

mkdir -p /opt/cfs/${CURVER}
cd /opt/cfs/${CURVER}
wget https://github.com/Nitecon/cfs/releases/download/v${CURVER}/cfs_${CURVER}_linux_amd64.tar.gz
tar xf cfs_${CURVER}_linux_amd64.tar.gz
ln -s /opt/cfs/${CURVER}/cfs /usr/bin/cfs
/usr/bin/cfs --config https://raw.githubusercontent.com/Nitecon/cfs/main/examples/nginx/install.yml > /var/log/cfs.log
