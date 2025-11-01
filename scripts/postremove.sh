#!/bin/sh
set -euo pipefail
SERVICE_NAME=cmk-piggyback-docker-resolver

remove() {
    systemctl disable "${SERVICE_NAME}@*.service" ||:
}

purge() {
    rm -rf /etc/cmk-piggyback-docker-resolver
}

upgrade() {
    true
}

action="$1"

case "$action" in
  "0" | "remove")
    remove
    ;;
  "1" | "upgrade")
    upgrade
    ;;
  "purge")
    purge
    ;;
  *)
    # Alpine
    remove
    ;;
esac
