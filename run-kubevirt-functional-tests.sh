#!/bin/bash -ex

usage() {
        echo "
Usage:

$0 [options]

This script will run Kubvirt's functional tests on an existing cluster.
The following are prerequisites for this script:

* You must have an up and running cluster.
* Kubevirt and it's testing resources must be deployed on the cluster.
  This can be achieved with 'kubevirt' and 'kubevirt_testing' roles.
* Docker must run on the local machine.
* Make should be installed on the local machine.

Optional arguments:
    -h,--help PATH
        Print this message and quite.

    -k,--kubeconfig PATH_TO_KUBECONFIG
        Kubeconfig which points to the cluster.
        The default path is '$HOME/.kube/config'.

    -c,--commit COMMIT
        Which commit to checkout before compiling the code.

    -o,--output OUTPUT_PATH
        A path for a directory for saving the test's log.
        If not specified the logs will not be saved.
"
}

verify_dependencies () {
    systemctl is-active docker > /dev/null || {
        echo "Please enable docker and re-run $0"
        exit 1
    }

    hash make > /dev/null || {
        echo "Please install make and re-run $0"
        exit 1
    }

    echo "Verified dependencies, we are ready to go..."
}

main() {
    local kubeconfig="$(realpath ${KUBECONFIG:-$HOME/.kube/config})"
    local commit="${COMMIT:-master}"
    local output_path="$OUTPUT_PATH"
    local kubevirt_repo_url="https://github.com/kubevirt/kubevirt.git"
    local kubevirt_repo_path="/tmp/kubevirt"
    local test_bin_path="${kubevirt_repo_path}/_out/tests/tests.test"

    [[ -f "$kubeconfig" ]] || {
        echo "kubeconfig $kubeconfig doesn't exist"
        exit 1
    }

    [[ -d "$kubevirt_repo_path" ]] || {
        echo "Cloning kubevirt's repo"
        git clone "$kubevirt_repo_url" "$kubevirt_repo_path"
    }

    pushd "$kubevirt_repo_path"
    echo "Checking out $commit"
    git checkout "$commit"

    echo "Compiling tests"
    make || {
        echo "Failed to compile tests. Cleaning cache and retrying..."
        make distclean
        make
    }
    popd

    echo "Running functional tests"
    if [[ -d "$output_path" ]]; then
        output_path="$(realpath "${output_path}/test.log")"
        echo "Logs will be written to $output_path"
        "$test_bin_path" -kubeconfig "$kubeconfig" | tee "$output_path"
    else
        "$test_bin_path" -kubeconfig "$kubeconfig"
    fi

    if [[ ${PIPESTATUS[0]} -ne 0 ]]; then
        echo "Some tests failed..."
    else
        echo "All tests passed !"
    fi

    exit "${PIPESTATUS[0]}"
}

options=$( \
    getopt \
        -o k:c:o:h \
        --long kubeconfig:,commit:,output:,help \
        -n 'run-kubevirt-functional-tests.sh' \
        -- "$@" \
)

if [[ "$?" != "0" ]]; then
    exit 1
fi
eval set -- "$options"

while true; do
    case $1 in
        -k|--kubeconfig)
            readonly KUBECONFIG="$2"
            shift 2
            ;;
        -c|--commit)
            readonly COMMIT="$2"
            shift 2
            ;;
        -o|--output)
            readonly OUTPUT_PATH="$2"
            shift 2
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        --)
            shift
            break
            ;;
        *)
            echo "Unkown option $1"
            usage
            exit 1
    esac
done

verify_dependencies
main
