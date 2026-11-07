#!/bin/sh
# dockyrd uninstall (M0 skeleton) — stops platform service; keeps user data by default.
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
		systemctl disable --now dockyrd.service 2>/dev/null || true
		rm -f /etc/systemd/system/dockyrd.service
		systemctl daemon-reload
	fi
	rm -f /usr/local/bin/dockyrd /usr/local/bin/dy
	log "dockyrd platform removed. Data under /var/lib/dockyrd and /etc/dockyrd was kept."
	log "Re-run with dockyrd_PURGE_DATA=1 support in a later release to delete data."
}

main "$@"
