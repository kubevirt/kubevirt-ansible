#!/bin/bash -xe

# This script is meant to be run within a mock environment, using
# mock_runner.sh or chrooter, from the root of the repository.
readonly ARTIFACTS_PATH="exported-artifacts"
readonly VMS_LOGS_PATH="${ARTIFACTS_PATH}/vm_logs"

get_run_path() {
    # if above ram_threshold KBs are available in /dev/shm, run there
    local suffix="${1:-lago}"
    local ram_threshold=30000000
    local avail_shm=$(df --output=avail /dev/shm | sed 1d)

    [[ "$avail_shm" -ge "$ram_threshold" ]] && \
        mkdir -p "/dev/shm/ost" && \
        echo "/dev/shm/ost/deployment-$suffix" || \
        echo "$PWD/deployment-$suffix"
}

on_exit() {
    set +e

    local run_path="${1:?}"
    local skip_cleanup="${2:-false}"

    print_resources_usage "$run_path"
    collect_lago_log "$run_path" "$ARTIFACTS_PATH"
    collect_logs_from_vms "$run_path" "${VMS_LOGS_PATH}/on_exit"
    collect_ansible_log "$ARTIFACTS_PATH"
    kill_make_tests

    if "$skip_cleanup"; then
        echo "Skipping cleanup"
    else
        cleanup "$run_path"
    fi
}

print_resources_usage() {
    local run_path="${1:?}"

    echo "FS info:"
    df -h

    echo "RAM info:"
    free -h

    echo "Prefix size:"
    du -h -d 1 "${run_path}/default" || echo "Failed to get Prefix size"

    echo "Images Size:"
    ls -lhs "${run_path}/default/images" || echo "Failed to get image dir size"
}

collect_ansible_log() {
    local dest="${1:?}"

    cp ansible.log "$dest"
}

collect_lago_log() {
    local run_path="${1:?}"
    local dest="${2:?}"

    find "$run_path" \
        -name lago.log \
        -exec cp {} "$dest" \;
}

collect_logs_from_vms() {
    local run_path="${1:?}"
    local dest="${2:?}"

    lago \
        --workdir "$run_path" \
        collect \
        --output "$dest"
}

cleanup() {
    set +e
    local run_path="${1:?}"

    lago --workdir "$run_path" destroy --yes || force_cleanup
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
    get_virtctl
}

get_virtctl() {
    local version="$(sed -nE 's,^version:\s*(.*)$,\1,p' vars/all.yml)"
    local url="https://github.com/kubevirt/kubevirt/releases/download/v${version}/virtctl-v${version}-linux-amd64"
    local name="virtctl"
    local dest="/usr/bin/${name}"

    hash "$name" &> /dev/null || {
        curl -L -o "$dest" "$url"
        chmod 755 "$dest"
    }

    # virtctl version returns rc 1 if the server isn't available
    "$name" version || :
}

is_code_changed() {
    git diff-tree --no-commit-id --name-only -r HEAD..HEAD^ \
    | grep -v -E -e '\.md$'

    return $?
}

kill_make_tests() {
    local my_pid="$$"
    local make_tests_pid="$MAKE_TESTS_PID"
    local make_tests_ppid

    [[ "$make_tests_pid" ]] || return
    make_tests_ppid="$(ps -o ppid= "$make_tests_pid" || echo -1)"

    [[ "$my_pid" -eq "$make_tests_ppid" ]] && kill -9 "$make_tests_pid"
}

run() {
    local run_path="${1:?}"
    local cluster="${2:?}"

    local openshift_ansible_url="${OPENSHIFT_ANSIBLE_URL:-https://github.com/openshift/openshift-ansible}"
    local ansible_modules_version="${ANSIBLE_MODULES_VERSION:-openshift-ansible-3.7.29-1}"
    local kubevirt_openshift_version="${OPENSHIFT_VERSION:-3.7}"
    local openshift_playbook_path="${OPENSHIFT_PLAYBOOK_PATH:-playbooks/byo/config.yml}"
    local provider="${PROVIDER:-lago}"
    local args=("prefix=$run_path")
    local inventory_file="$(realpath inventory)"
    local storage_role="${STORAGE_ROLE:-storage-glusterfs}"

    set_params
    install_requirements
    ansible --version

    if [[ "$cluster" == "openshift" ]]; then
        rm -rf openshift-ansible
        git clone "$openshift_ansible_url"

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
        "cluster=$cluster"
        "platform=$cluster"
    )

    timeout \
        --kill-after 5m \
        20m \
        make build-tests &> "${ARTIFACTS_PATH}/generate-tests.log" &
    readonly MAKE_TESTS_PID="$!"

    ansible-playbook \
        -u root \
        -i "$inventory_file" \
        -v \
        -e@vars/all.yml \
        -e "${args[*]}" \
        playbooks/automation/check-patch.yml

    # Run integration tests
    wait "$MAKE_TESTS_PID" || {
        local ret="$?"
        echo "Error: Failed to compile tests"
        exit "$ret"
    }
    http_proxy="" make test

    collect_logs_from_vms "$run_path" "${VMS_LOGS_PATH}/post_tests"

    # Deprovision resources
    ansible-playbook \
        -u root \
        -i "$inventory_file" \
        -v \
        -e@vars/all.yml \
        -e "apb_action=deprovision" \
        -e "${args[*]}" \
        playbooks/automation/deprovision.yml
}

usage() {
    echo "

Deploy a cluster, deploy Kubevirt, and run tests

Optional arguments:
    -h,--help
        Show this message and quite

    --only-cleanup
        Destroy the environment and quite

    --skip-cleanup
        Don't destroy the environment on exit
"
}

main() {
    echo "$@"
    local options
    local only_cleanup=false
    local skip_cleanup=false
    local cluster="${CLUSTER:-openshift}" # Openshift or Kubernetes
    local run_path="${PWD}/deployment-${cluster}"

    options=$( \
        getopt \
            -o h \
            --long help,only-cleanup,skip-cleanup \
            -n 'check-patch.sh' \
            -- "$@" \
    )

    if [[ "$?" != "0" ]]; then
        echo "Failed to parse cmd line arguments" && exit 1
    fi

    eval set -- "$options"

    while true; do
        case $1 in
            -h|--help)
                usage
                exit 0
                ;;
            --only-cleanup)
                only_cleanup=true
                shift
                ;;
            --skip-cleanup)
                skip_cleanup=true
                shift
                ;;
            --)
                shift
                break
            ;;
            *)
                echo "Unknown flag $1"
                exit 1
                ;;
        esac
    done

    "$only_cleanup" && {
        echo "Only running cleanup"
        cleanup "$run_path"
        exit $?
    }

    is_code_changed || {
        echo 'Code did not changed, skipping tests...'
        exit 0
    }

    # When cleanup is skipped, the run path should be deterministic
    "$skip_cleanup" || run_path="$(get_run_path "$cluster")"
    trap "on_exit $run_path $skip_cleanup" EXIT

    mkdir -p "$ARTIFACTS_PATH"
    mkdir -p "$VMS_LOGS_PATH"

    make check

    run "$run_path" "$cluster"
}

[[ "${BASH_SOURCE[0]}" == "$0" ]] && main "$@"
