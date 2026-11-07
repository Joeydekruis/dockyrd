#!/bin/sh
# Dockyard uninstall (M0 skeleton) — stops platform service; keeps user data by default.
set -e

log() {
	printf '%s\n' "$*"
}

require_root() {
	if [ "$(id -u)" -ne 0 ]; then
		log "Run as root." >&2
		exit 1
	fi
}

main() {
	require_root
	if command -v systemctl >/dev/null 2>&1; then
		systemctl disable --now dockyard.service 2>/dev/null || true
		rm -f /etc/systemd/system/dockyard.service
		systemctl daemon-reload
	fi
	rm -f /usr/local/bin/dockyard /usr/local/bin/dy
	log "Dockyard platform removed. Data under /var/lib/dockyard and /etc/dockyard was kept."
	log "Re-run with DOCKYARD_PURGE_DATA=1 support in a later release to delete data."
}

main "$@"
