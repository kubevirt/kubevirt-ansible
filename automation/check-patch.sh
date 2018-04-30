#!/bin/bash -xe

# This script is meant to be run within a mock environment, using
# mock_runner.sh or chrooter, from the root of the repository.

get_run_path() {
    # if above ram_threshold KBs are available in /dev/shm, run there
    local suffix="${1:-lago}"
    local ram_threshold=15000000
    local avail_shm=$(df --output=avail /dev/shm | sed 1d)

    [[ "$avail_shm" -ge "$ram_threshold" ]] && \
        mkdir -p "/dev/shm/ost" && \
        echo "/dev/shm/ost/deployment-$suffix" || \
        echo "$PWD/deployment-$suffix"
}

collect_logs() {
    local run_path="$1"
    local artifacts_dir="exported-artifacts"
    local vms_logs="${artifacts_dir}/vms_logs"

    mkdir -p "$vms_logs"

    lago \
        --workdir "$run_path" \
        collect \
        --output "$vms_logs" \
        || :

    find "$run_path" \
        -name lago.log \
        -exec cp {} "$artifacts_dir" \;

    cp ansible.log "$artifacts_dir"
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
    if [[ ! -c "/dev/kvm" ]]; then
        mknod /dev/kvm c 10 232
    fi
}

install_requirements() {
    ansible-galaxy install -r requirements.yml
}

is_code_changed() {
    git diff-tree --no-commit-id --name-only -r HEAD..HEAD^ \
    | grep -v -E -e '\.md$'

    return $?
}

main() {
    # cluster: Openshift or Kubernetes

    local cluster="${CLUSTER:-openshift}"
    local ansible_modules_version="${ANSIBLE_MODULES_VERSION:-openshift-ansible-3.7.29-1}"
    local kubevirt_openshift_version="${OPENSHIFT_VERSION:-3.7}"
    local openshift_playbook_path="${OPENSHIFT_PLAYBOOK_PATH:-playbooks/byo/config.yml}"
    local provider="${PROVIDER:-lago}"
    local run_path="$(get_run_path "$cluster")"
    local args=("prefix=$run_path")
    local inventory_file="$(realpath inventory)"
    local storage_role="${STORAGE_ROLE:-storage-none}"

    trap "cleanup $run_path" EXIT

    set_params
    install_requirements

    if [[ "$cluster" == "openshift" ]]; then
        [[ -e openshift-ansible ]] \
        || git clone https://github.com/openshift/openshift-ansible

        pushd openshift-ansible
        git fetch origin
        git checkout "$ansible_modules_version" || {
            echo "Ansible modules $ansible_modules_version wasn't found"
            exit 1
        }
        popd

        args+=("openshift_ansible_dir=$(realpath openshift-ansible)")

    elif ! [[ "$cluster" == "kubernetes" ]]; then
        echo "$cluster unkown cluster type"
        exit 1
    fi

    [[ -f "$STD_CI_YUMREPOS" ]] && {
        echo "Using std-ci yum repos:"
	    cat "$STD_CI_YUMREPOS"
        args+=("std_ci_yum_repos=$(realpath "$STD_CI_YUMREPOS")")
    }

    args+=(
        "provider=$provider"
        "inventory_file=$inventory_file"
        "cluster=$cluster"
        "ansible_modules_version=$ansible_modules_version"
        "kubevirt_openshift_version=$kubevirt_openshift_version"
        "openshift_playbook_path=$openshift_playbook_path"
	"storage_role=$storage_role"
    )
    ansible-playbook \
        -u root \
        -i "$inventory_file" \
        -v \
        -e "${args[*]}" \
        playbooks/automation/check-patch.yml

    # Run integration tests
    make test

    # Deprovision resources
    ansible-playbook \
        -u root \
        -i "$inventory_file" \
        -v \
        -e "apb_action=deprovision" \
        playbooks/automation/deprovision.yml
}

if [[ "${BASH_SOURCE[0]}" == "$0" ]]; then
    is_code_changed || {
        echo 'Code did not changed, skipping tests...'
        exit 0
    }
    main "$@"
fi
