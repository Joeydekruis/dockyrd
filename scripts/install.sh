#!/bin/sh
# Dockyard installer (M0 skeleton). Full Docker/release flow lands in later milestones.
set -e

DOCKYARD_VERSION="${DOCKYARD_VERSION:-dev}"
INSTALL_BIN="${DOCKYARD_INSTALL_BIN:-}"
CONFIG_DIR="${DOCKYARD_CONFIG_DIR:-/etc/dockyard}"
DATA_DIR="${DOCKYARD_DATA_DIR:-/var/lib/dockyard}"
BIN_DEST="${DOCKYARD_BIN_DEST:-/usr/local/bin/dockyard}"
UNIT_SRC="${DOCKYARD_UNIT_SRC:-}"
SKIP_SYSTEMD="${DOCKYARD_SKIP_SYSTEMD:-0}"
SKIP_ROOT="${DOCKYARD_SKIP_ROOT:-0}"

log() {
	printf '%s\n' "$*"
}

require_root() {
	if [ "$SKIP_ROOT" = "1" ]; then
		return 0
	fi
	if [ "$(id -u)" -ne 0 ]; then
		log "This installer must run as root (or use DOCKYARD_* paths for local dev)." >&2
		exit 1
	fi
}

preflight() {
	if ! command -v curl >/dev/null 2>&1; then
		log "curl is required." >&2
		exit 1
	fi
	arch=$(uname -m)
	case "$arch" in
	x86_64 | amd64 | aarch64 | arm64) ;;
	*)
		log "Unsupported architecture: $arch (need amd64 or arm64)." >&2
		exit 1
		;;
	esac
	if [ "$SKIP_ROOT" != "1" ] && ! command -v systemctl >/dev/null 2>&1; then
		log "systemd is required for service install." >&2
		exit 1
	fi
	if ! command -v docker >/dev/null 2>&1; then
		log "Warning: Docker not found — install Docker before deploying apps." >&2
	fi
}

ensure_paths() {
	mkdir -p "$CONFIG_DIR" "$DATA_DIR"
	chmod 700 "$CONFIG_DIR"
	chmod 750 "$DATA_DIR"
}

install_binary() {
	mkdir -p "$(dirname "$BIN_DEST")"
	if [ -n "$INSTALL_BIN" ]; then
		install -m 755 "$INSTALL_BIN" "$BIN_DEST"
		return 0
	fi
	if download_release_binary; then
		return 0
	fi
	log "Could not download a release binary." >&2
	log "Build locally (make build) and re-run with:" >&2
	log "  DOCKYARD_INSTALL_BIN=/path/to/dockyard $0" >&2
	exit 1
}

github_latest_tag() {
	repo="$1"
	curl -sf "https://api.github.com/repos/$repo/releases/latest" \
		| sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -n1
}

download_release_binary() {
	repo="${DOCKYARD_GITHUB_REPO:-Joeydekruis/dockyrd}"
	tag="${DOCKYARD_VERSION:-}"
	if [ -z "$tag" ] || [ "$tag" = "dev" ]; then
		tag=$(github_latest_tag "$repo") || return 1
	fi
	case "$tag" in
	v*) ;;
	*) tag="v$tag" ;;
	esac
	ver="${tag#v}"
	os_name=$(uname -s | tr '[:upper:]' '[:lower:]')
	arch=$(uname -m)
	case "$arch" in
	x86_64 | amd64) arch=amd64 ;;
	aarch64 | arm64) arch=arm64 ;;
	*) return 1 ;;
	esac
	asset="dockyard_${ver}_${os_name}_${arch}.tar.gz"
	base="https://github.com/${repo}/releases/download/${tag}"
	tmp=$(mktemp -d) || return 1
	trap 'rm -rf "$tmp"' EXIT
	if ! curl -sfL "${base}/${asset}" -o "${tmp}/${asset}"; then
		return 1
	fi
	if curl -sfL "${base}/checksums.txt" -o "${tmp}/checksums.txt"; then
		want=$(grep " ${asset}$" "${tmp}/checksums.txt" | awk '{print $1}')
		got=$(sha256sum "${tmp}/${asset}" | awk '{print $1}')
		if [ -n "$want" ] && [ "$want" != "$got" ]; then
			log "Checksum mismatch for ${asset}" >&2
			return 1
		fi
	fi
	tar -xzf "${tmp}/${asset}" -C "$tmp" dockyard || return 1
	install -m 755 "${tmp}/dockyard" "$BIN_DEST"
	return 0
}

install_systemd() {
	if [ "$SKIP_SYSTEMD" = "1" ]; then
		return 0
	fi
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
		"$BIN_DEST" init
}

main() {
	log "Dockyard install, version ${DOCKYARD_VERSION}"
	require_root
	preflight
	ensure_paths
	install_binary
	run_init
	install_systemd
	if [ "$SKIP_SYSTEMD" != "1" ] && command -v systemctl >/dev/null 2>&1; then
		systemctl restart dockyard.service || systemctl start dockyard.service
	fi
	if [ "$SKIP_SYSTEMD" = "1" ]; then
		log "✓ Dockyard installed (systemd skipped). Run: $BIN_DEST serve"
	else
		log "✓ Dockyard service installed. See init output above for the setup token."
	fi
}

main "$@"
