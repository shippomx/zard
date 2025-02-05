#!/usr/bin/env bash

REPO_ROOT=$(realpath $(dirname "${BASH_SOURCE[0]}")/..)

source $REPO_ROOT/hack/log.sh

# tools::wait_for_port 等待端口启动
# @arg1: host
# @arg2: port
function tools::wait_for_port() {
	local start_time=$(date +%s)
	local TIMEOUT=120
    while ! nc -z "$1" "$2"; do
        log::info "Waiting for $1:$2 to be open..."
        sleep 1

		# Check if timeout has been reached
        local current_time=$(date +%s)
        if (( current_time - start_time > TIMEOUT )); then
            log::error "Timed out waiting for $1:$2 to be accessible."
            exit 1
        fi
    done
    log::info "$1:$2 is now open!"
}

function tools::install_goctl() {
    cd $REPO_ROOT/tools/goctl
    go build .
}
