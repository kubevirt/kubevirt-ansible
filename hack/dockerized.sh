#!/bin/bash
set -ex

source $(dirname "$0")/common.sh

DOCKER_DIR=${KUBEVIRT_ANSIBLE_DIR}/hack/docker-test-builder

BUILDER=kubevirt-ansible

TEMPFILE=".rsynctemp"

SYNC_OUT=${SYNC_OUT:-true}
SYNC_VENDOR=${SYNC_VENDOR:-false}
SYNC_GENERATED=${SYNC_GENERATED:-false}

# Build the build container
(cd ${DOCKER_DIR} && docker build . -q -t ${BUILDER})

# Create the persistent docker volume
if [ -z "$(docker volume list | grep ${BUILDER})" ]; then
    docker volume create --name ${BUILDER}
fi

# Make sure that the output directory exists
docker run -v "${BUILDER}:/root:rw,z" --security-opt label:disable --rm ${BUILDER} mkdir -p /root/go/src/kubevirt.io/kubevirt-ansible/_out

# Make sure that the vendor directory exists
docker run -v "${BUILDER}:/root:rw,z" --security-opt label:disable --rm ${BUILDER} mkdir -p /root/go/src/kubevirt.io/kubevirt-ansible/vendor

# Start an rsyncd instance and make sure it gets stopped after the script exits
RSYNC_CID=$(docker run -d -v "${BUILDER}:/root:rw,z" --security-opt label:disable --expose 873 -P ${BUILDER} /usr/bin/rsync --no-detach --daemon --verbose)

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

# Copy kubevirt-ansible into the persistent docker volume
_rsync \
    --delete \
    --include 'hack/***' \
    --include 'vendor/***' \
    --include 'go.mod' \
    --include 'go.sum' \
    --include 'tests/***' \
    --include 'Makefile' \
    --exclude '*' \
    ${KUBEVIRT_ANSIBLE_DIR}/ \
    "rsync://root@127.0.0.1:${RSYNCD_PORT}/build"

# Run the command
test -t 1 && USE_TTY="-it"
docker run --rm -v "${BUILDER}:/root:rw,z" --security-opt label:disable ${USE_TTY} -w "/root/go/src/kubevirt.io/kubevirt-ansible" ${BUILDER} "$@"

if [ "$SYNC_VENDOR" = "true" ]; then
    _rsync --delete --include 'go.mod' --include 'go.sum' --exclude '*' --verbose "rsync://root@127.0.0.1:${RSYNCD_PORT}/build" ${KUBEVIRT_ANSIBLE_DIR}/
    _rsync --delete "rsync://root@127.0.0.1:${RSYNCD_PORT}/vendor" "${VENDOR_DIR}"
fi
# Copy the build output out of the container, make sure that _out exactly matches the build result
if [ "$SYNC_OUT" = "true" ]; then
    _rsync --delete "rsync://root@127.0.0.1:${RSYNCD_PORT}/out" ${OUT_DIR}
fi
# Copy generated sources
if [ "$SYNC_GENERATED" = "true" ]; then
    _rsync --delete "rsync://root@127.0.0.1:${RSYNCD_PORT}/build/tests/" ${KUBEVIRT_ANSIBLE_DIR}/tests
fi
