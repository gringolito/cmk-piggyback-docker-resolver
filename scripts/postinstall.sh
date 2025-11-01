#!/bin/sh
set -euo pipefail
SERVICE_NAME=cmk-piggyback-docker-resolver

install() {
    systemctl daemon-reload ||:
    systemctl unmask "${SERVICE_NAME}@.service" ||:
    systemctl preset "${SERVICE_NAME}@.service" ||:
    printf "\n"
    printf "  ${SERVICE_NAME} service is not enabled to start on the next system boot and not set to start automatically upon installation.\n"
    printf "\n"
    printf "  Please create the site-specific configurations in:\n"
    printf "    /etc/cmk-piggyback-docker-resolver/\n"
    printf "\n"
    printf "  and start the service with:\n"
    printf "    systemctl enable --now ${SERVICE_NAME}@<SITE>.service\n"
    printf "\n"
}

upgrade() {
    printf "\033[32m  Reloading ${SERVICE_NAME} service\033[0m\n"
    systemctl daemon-reload ||:
    systemctl try-restart "${SERVICE_NAME}@*.service" ||:
}

# Step 2, check if this is a clean install or an upgrade
action="$1"
if  [ "$1" = "configure" ] && [ -z "$2" ]; then
    # Alpine linux does not pass args, and deb passes $1=configure
    action="install"
elif [ "$1" = "configure" ] && [ -n "$2" ]; then
    # deb passes $1=configure $2=<current version>
    action="upgrade"
fi

case "$action" in
  "1" | "install")
    install
    ;;
  "2" | "upgrade")
    upgrade
    ;;
  *)
    # Alpine -> $1 == version being installed
    install
    ;;
esac
