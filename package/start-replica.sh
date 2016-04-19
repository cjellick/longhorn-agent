#!/bin/bash
set -e -x

if [ ! -x "$(which ping)" ]; then
    apt-get install -y iputils-ping
fi

if [ ! -x "$(which curl)" ]; then
    apt-get install -y curl
fi

while ! ping -c 1 rancher-metadata >/dev/null; do
    echo Waiting for metadata
    sleep 1
done

NAME=$(curl -s http://rancher-metadata/latest/self/container/uuid)
VOLUME_SIZE=$(curl -s http://rancher-metadata/latest/self/service/metadata/longhorn/volume_size)
SLAB_SIZE=$(curl -s http://rancher-metadata/latest/self/service/metadata/longhorn/slab_size)

mkdir -p /volumes/$NAME
cd /volumes/$NAME

exec longhorn-agent --healthcheck --launch-replica --replica-name ${NAME} --tcmu-volume-size ${VOLUME_SIZE} --replica-slab-size ${SLAB_SIZE} --replica-host localhost --replica-port 5000 --controller-host controller --controller-port 5000 --replica-reported-host $(curl -s http://rancher-metadata/latest/self/container/primary_ip)
