#!/bin/sh
# Dockyard installer (M0 skeleton). Full Docker/release flow lands in later milestones.
set -e

DOCKYARD_VERSION="${DOCKYARD_VERSION:-dev}"
INSTALL_BIN="${DOCKYARD_INSTALL_BIN:-}"
CONFIG_DIR="${DOCKYARD_CONFIG_DIR:-/etc/dockyard}"
DATA_DIR="${DOCKYARD_DATA_DIR:-/var/lib/dockyard}"
UNIT_SRC="${DOCKYARD_UNIT_SRC:-}"

log() {
	printf '%s\n' "$*"
}

require_root() {
	if [ "$(id -u)" -ne 0 ]; then
		log "This installer must run as root (or use DOCKYARD_* paths for local dev)." >&2
		exit 1
	fi
}

ensure_paths() {
	mkdir -p "$CONFIG_DIR" "$DATA_DIR"
	chmod 700 "$CONFIG_DIR"
	chmod 750 "$DATA_DIR"
}

install_binary() {
	if [ -n "$INSTALL_BIN" ]; then
		install -m 755 "$INSTALL_BIN" /usr/local/bin/dockyard
		return 0
	fi
	log "Release download is not wired in the M0 skeleton." >&2
	log "Build locally (make build) and re-run with:" >&2
	log "  DOCKYARD_INSTALL_BIN=/path/to/dockyard $0" >&2
	exit 1
}

install_systemd() {
	unit_src="$UNIT_SRC"
	if [ -z "$unit_src" ]; then
		script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
		unit_src="$script_dir/../packaging/dockyard.service"
	fi
	if [ ! -f "$unit_src" ]; then
		log "Missing unit file: $unit_src" >&2
		exit 1
	fi
	install -m 644 "$unit_src" /etc/systemd/system/dockyard.service
	systemctl daemon-reload
	systemctl enable dockyard.service
}

run_init() {
	DOCKYARD_CONFIG_DIR="$CONFIG_DIR" DOCKYARD_DATA_DIR="$DATA_DIR" \
		/usr/local/bin/dockyard init
}

main() {
	log "Dockyard install (skeleton), version ${DOCKYARD_VERSION}"
	require_root
	ensure_paths
	install_binary
	run_init
	install_systemd
	if command -v systemctl >/dev/null 2>&1; then
		systemctl restart dockyard.service || systemctl start dockyard.service
	fi
	log "✓ Dockyard service installed. See init output above for the setup token."
}

main "$@"
