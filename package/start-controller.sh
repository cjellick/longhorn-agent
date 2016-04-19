#!/bin/bash
set -e -x

if [ ! -x "$(which curl)" ]; then
    apt-get install -y curl
fi

if [ ! -x "$(which modprobe)" ]; then
    apt-get install -y kmod
fi

if [ ! -d /sys/kernel/config ]; then
    modprobe target_core_user
fi

while ! ping -c 1 rancher-metadata; do
    echo Waiting for metadata
    sleep 1
done

mount --rbind /host/dev /dev
mount -t configfs none /sys/kernel/config

VOLUME=$(curl -s http://rancher-metadata/latest/self/stack/services/replica/metadata/longhorn/volume_name)
SIZE=$(curl -s http://rancher-metadata/latest/self/stack/services/replica/metadata/longhorn/volume_size)
DEV=/dev/longhorn/${VOLUME}

mkdir -p $(dirname $DEV)

cd /var/log
exec longhorn-agent --launch-controller --healthcheck --controller-port 5000 --tcmu-volume $VOLUME --tcmu-volume-size $SIZE --tcmu-volume-dev $DEV
