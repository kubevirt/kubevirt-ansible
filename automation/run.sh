#!/bin/bash -xe

# This script is meant to be run within a mock environment, using
# mock_runner.sh or chrooter, from the root of the repository:

# if above RAM_THRESHOLD KBs are available in /dev/shm, run there
RAM_THRESHOLD=15000000

get_run_path() {
    local avail_shm
    avail_shm=$(df --output=avail /dev/shm | sed 1d)
    [[ "$avail_shm" -ge "$RAM_THRESHOLD" ]] && \
        mkdir -p "/dev/shm/ost" && \
        echo "/dev/shm/ost/deployment-$CLUSTER_TYPE" || \
        echo "$PWD/deployment-$CLUSTER_TYPE"
}

collect_logs() {
    local run_path="$1"
    local artifacts_dir="exported-artifacts"

    [[ -d "$artifacts_dir" ]] || mkdir exported-artifacts
    find "$run_path" \
        -name lago.log \
        -exec mv {} "$artifacts_dir" \;

    find . \
        -name ansible.log \
        -exec mv {} "$artifacts_dir" \;
}

cleanup() {
    set +e
    local run_path="$1"
    collect_logs "$run_path"
    lago --workdir "$run_path" destroy --yes \
    || force_cleanup
}

force_cleanup() {
    echo "Cleaning with libvirt"

    local domains=($( \
        virsh -c qemu:///system list --all --name \
        | egrep -w "lago-master[0-9]*|lago-node[0-9]*"
    ))
    local nets=($( \
        virsh -c qemu:///system net-list --all \
        | egrep -w "[[:alnum:]]{4}-.*" \
        | egrep -v "vdsm-ovirtmgmt" \
        | awk '{print $1;}' \
    ))

    for domain in "${domains[@]}"; do
        virsh -c qemu:///system destroy "$domain"
    done
    for net in "${nets[@]}"; do
        virsh -c qemu:///system net-destroy "$net"
    done

    echo "Cleaning with libvirt Done"
}

set_params() {
    # needed to run lago inside chroot
    # TO-DO: use libvirt backend instead
    export LIBGUESTFS_BACKEND=direct
    # uncomment the next lines for extra verbose output
    #export LIBGUESTFS_DEBUG=1 LIBGUESTFS_TRACE=1

    # ensure /dev/kvm exists, otherwise it will still use
    # direct backend, but without KVM(much slower).
    ! [[ -c "/dev/kvm" ]] && mknod /dev/kvm c 10 232
}


trap 'cleanup "$run_path"' EXIT
set_params
CLUSTER_TYPE="${CLUSTER_TYPE:-openshift}"
run_path="$(get_run_path)"
args=("prefix=$run_path")
ansible-galaxy install -r requirements.yml

if [[ "$CLUSTER_TYPE" == "openshift" ]]; then
    [[ -e openshift-ansible ]] || \
    git clone -b release-3.6 https://github.com/openshift/openshift-ansible
    args+=(
        "openshift_ansible_dir=$(realpath openshift-ansible)"
        "cluster_type=openshift"
    )
elif [[ "$CLUSTER_TYPE" == "kubernetes" ]]; then
    args+=("cluster_type=kubernetes")
else
    echo "$CLUSTER_TYPE unkown cluster type"
    exit 1
fi

ansible-playbook \
    -u root \
    -i inventory \
    -v \
    -e "${args[*]}" \
    deploy-with-lago.yml
