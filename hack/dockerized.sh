#!/bin/bash
set -e

source $(dirname "$0")/common.sh

DOCKER_DIR=${KUBEVIRT_ANSIBLE_DIR}/hack/docker-test-builder

BUILDER=kubevirt-ansible

TEMPFILE=".rsynctemp"

# Build the build container
(cd ${DOCKER_DIR} && docker build . -q -t ${BUILDER})

# Create the persistent docker volume
if [ -z "$(docker volume list | grep ${BUILDER})" ]; then
    docker volume create --name ${BUILDER}
fi

# Make sure that the output directory exists
docker run -v "${BUILDER}:/root:rw,z" --rm ${BUILDER} mkdir -p /root/go/src/kubevirt.io/kubevirt-ansible/_out

# Start an rsyncd instance and make sure it gets stopped after the script exits
RSYNC_CID=$(docker run -d -v "${BUILDER}:/root:rw,z" --expose 873 -P ${BUILDER} /usr/bin/rsync --no-detach --daemon --verbose)

function finish() {
    docker stop ${RSYNC_CID} >/dev/null 2>&1 &
    docker rm -f ${RSYNC_CID} >/dev/null 2>&1 &
}
trap finish EXIT

RSYNCD_PORT=$(docker port $RSYNC_CID 873 | cut -d':' -f2)

rsynch_fail_count=0

while ! rsync ${KUBEVIRT_ANSIBLE_DIR}/${RSYNCTEMP} "rsync://root@127.0.0.1:${RSYNCD_PORT}/build/${RSYNCTEMP}" &>/dev/null; do
    if [[ "$rsynch_fail_count" -eq 0 ]]; then
        printf "Waiting for rsyncd to be ready"
        sleep .1
    elif [[ "$rsynch_fail_count" -lt 30 ]]; then
        printf "."
        sleep 1
    else
        printf "failed"
        break
    fi
    rsynch_fail_count=$((rsynch_fail_count + 1))
done

printf "\n"

rsynch_fail_count=0

_rsync() {
    rsync -al "$@"
}

# Copy kubevirt into the persistent docker volume
_rsync --delete --exclude _out ${KUBEVIRT_ANSIBLE_DIR}/ "rsync://root@127.0.0.1:${RSYNCD_PORT}/build"

# Run the command
test -t 1 && USE_TTY="-it"
docker run --rm -v "${BUILDER}:/root:rw,z" ${USE_TTY} -w "/root/go/src/kubevirt.io/kubevirt-ansible" ${BUILDER} "$@"

# Copy from container content of out directory
_rsync --delete "rsync://root@127.0.0.1:${RSYNCD_PORT}/out" ${OUT_DIR}
