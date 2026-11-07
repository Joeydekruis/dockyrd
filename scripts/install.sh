#!/bin/sh
# dockyrd installer: paths, Docker (Debian/Ubuntu), release binary, init, systemd.
set -e

dockyrd_VERSION="${dockyrd_VERSION:-dev}"
dockyrd_INSTALL_REF="${dockyrd_INSTALL_REF:-development}"
INSTALL_BIN="${dockyrd_INSTALL_BIN:-}"
INSTALL_DY_BIN="${dockyrd_INSTALL_DY_BIN:-}"
CONFIG_DIR="${dockyrd_CONFIG_DIR:-/etc/dockyrd}"
DATA_DIR="${dockyrd_DATA_DIR:-/var/lib/dockyrd}"
BIN_DEST="${dockyrd_BIN_DEST:-/usr/local/bin/dockyrd}"
DY_BIN_DEST="${dockyrd_DY_BIN_DEST:-/usr/local/bin/dy}"
UNIT_SRC="${dockyrd_UNIT_SRC:-}"
SKIP_SYSTEMD="${dockyrd_SKIP_SYSTEMD:-0}"
SKIP_ROOT="${dockyrd_SKIP_ROOT:-0}"
dockyrd_SKIP_DOCKER_INSTALL="${dockyrd_SKIP_DOCKER_INSTALL:-0}"

log() {
	printf '%s\n' "$*"
}

require_root() {
	if [ "$SKIP_ROOT" = "1" ]; then
		return 0
	fi
	if [ "$(id -u)" -ne 0 ]; then
		log "This installer must run as root (or use dockyrd_* paths for local dev)." >&2
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
}

parse_args() {
	for arg in "$@"; do
		case "$arg" in
		--no-docker-install) dockyrd_SKIP_DOCKER_INSTALL=1 ;;
		esac
	done
}

docker_daemon_ready() {
	docker info >/dev/null 2>&1
}

ensure_docker_running() {
	if ! command -v systemctl >/dev/null 2>&1; then
		return 0
	fi
	if systemctl is-active --quiet docker 2>/dev/null; then
		return 0
	fi
	systemctl enable docker.service 2>/dev/null || true
	systemctl start docker.service 2>/dev/null || true
}

wait_for_docker() {
	i=0
	while ! docker_daemon_ready; do
		i=$((i + 1))
		if [ "$i" -gt 60 ]; then
			log "Docker daemon did not become ready." >&2
			exit 1
		fi
		sleep 1
	done
}

install_docker_via_apt() {
	if ! command -v apt-get >/dev/null 2>&1; then
		log "Docker is not installed and automatic install needs apt (Debian/Ubuntu)." >&2
		exit 1
	fi
	if [ ! -f /etc/os-release ]; then
		log "Cannot detect OS for Docker install (/etc/os-release missing)." >&2
		exit 1
	fi
	# shellcheck disable=SC1091
	. /etc/os-release
	case "${ID:-}" in
	ubuntu) distro=ubuntu ;;
	debian) distro=debian ;;
	*)
		log "Automatic Docker install supports Debian and Ubuntu only (detected: ${ID:-unknown})." >&2
		log "Install Docker manually, then re-run this script." >&2
		exit 1
		;;
	esac
	if [ -z "${VERSION_CODENAME:-}" ]; then
		log "Could not detect OS version codename for Docker apt repo." >&2
		exit 1
	fi
	log "Installing Docker Engine from download.docker.com (${ID} ${VERSION_CODENAME})…"
	install -m 0755 -d /etc/apt/keyrings
	curl -fsSL "https://download.docker.com/linux/${distro}/gpg" -o /etc/apt/keyrings/docker.asc
	chmod a+r /etc/apt/keyrings/docker.asc
	# shellcheck disable=SC2046
	echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/${distro} ${VERSION_CODENAME} stable" \
		>/etc/apt/sources.list.d/docker.list
	apt-get update -qq
	DEBIAN_FRONTEND=noninteractive apt-get install -y -qq \
		docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
}

merge_docker_daemon_defaults() {
	if [ "$SKIP_ROOT" = "1" ] || [ "$dockyrd_SKIP_DOCKER_INSTALL" = "1" ]; then
		return 0
	fi
	if ! command -v python3 >/dev/null 2>&1; then
		return 0
	fi
	local path="/etc/docker/daemon.json"
	local changed
	changed="$(python3 - "$path" <<'PY'
import json, os, sys
path = sys.argv[1]
wanted = {
    "log-driver": "json-file",
    "log-opts": {"max-size": "10m", "max-file": "3"},
    "live-restore": True,
    "default-address-pools": [{"base": "10.200.0.0/16", "size": 27}],
}
data = {}
if os.path.isfile(path):
    try:
        with open(path) as f:
            data = json.load(f) or {}
    except (json.JSONDecodeError, OSError):
        data = {}
updated = dict(data)
for key, val in wanted.items():
    if key not in updated:
        updated[key] = val
if updated == data:
    print("no")
    sys.exit(0)
os.makedirs(os.path.dirname(path) or ".", exist_ok=True)
with open(path, "w") as f:
    json.dump(updated, f, indent=2)
    f.write("\n")
print("yes")
PY
)"
	if [ "$changed" = "yes" ]; then
		log "Merged Docker defaults into ${path} (log rotation, live-restore, address pools)."
		log "Restart docker when no workloads are running if networks were exhausted before."
	fi
}

ensure_docker() {
	if [ "$dockyrd_SKIP_DOCKER_INSTALL" = "1" ] || [ "$SKIP_ROOT" = "1" ]; then
		if ! command -v docker >/dev/null 2>&1; then
			log "Warning: Docker not installed (install skipped)." >&2
		fi
		return 0
	fi
	if command -v docker >/dev/null 2>&1; then
		ensure_docker_running
		wait_for_docker
		merge_docker_daemon_defaults
		log "✓ Docker is installed and running."
		return 0
	fi
	install_docker_via_apt
	ensure_docker_running
	wait_for_docker
	merge_docker_daemon_defaults
	log "✓ Docker installed and running."
}

ensure_paths() {
	mkdir -p "$CONFIG_DIR" "$DATA_DIR"
	chmod 700 "$CONFIG_DIR"
	chmod 750 "$DATA_DIR"
}

install_dy_cli() {
	dy_src="$1"
	if [ -z "$dy_src" ] || [ ! -f "$dy_src" ]; then
		log "Warning: dy API CLI not installed (binary missing from release or build)." >&2
		return 0
	fi
	mkdir -p "$(dirname "$DY_BIN_DEST")"
	install -m 755 "$dy_src" "$DY_BIN_DEST"
	log "✓ Installed API CLI: $DY_BIN_DEST"
}

install_binary() {
	mkdir -p "$(dirname "$BIN_DEST")"
	if [ -n "$INSTALL_BIN" ]; then
		install -m 755 "$INSTALL_BIN" "$BIN_DEST"
		if [ -n "$INSTALL_DY_BIN" ]; then
			install_dy_cli "$INSTALL_DY_BIN"
		else
			install_dy_cli "$(dirname "$INSTALL_BIN")/dy"
		fi
		return 0
	fi
	if download_release_binary; then
		return 0
	fi
	log "Could not download a release binary." >&2
	log "Build locally (make build) and re-run with:" >&2
	log "  dockyrd_INSTALL_BIN=/path/to/dockyrd dockyrd_INSTALL_DY_BIN=/path/to/dy $0" >&2
	exit 1
}

github_api_curl() {
	# shellcheck disable=SC2086
	token="${dockyrd_GITHUB_TOKEN:-${GITHUB_TOKEN:-}}"
	if [ -n "$token" ]; then
		curl -sf -H "Authorization: Bearer $token" \
			-H "Accept: application/vnd.github+json" \
			"$@"
	else
		curl -sf -H "Accept: application/vnd.github+json" "$@"
	fi
}

github_latest_tag() {
	repo="$1"
	github_api_curl "https://api.github.com/repos/$repo/releases/latest" \
		| sed -n 's/.*"tag_name": *"\([^"]*\)".*/\1/p' | head -n1
}

download_release_binary() {
	repo="${dockyrd_GITHUB_REPO:-Joeydekruis/dockyrd}"
	tag="${dockyrd_VERSION:-}"
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
	asset="dockyrd_${ver}_${os_name}_${arch}.tar.gz"
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
	tar -xzf "${tmp}/${asset}" -C "$tmp" dockyrd || return 1
	install -m 755 "${tmp}/dockyrd" "$BIN_DEST"
	if tar -xzf "${tmp}/${asset}" -C "$tmp" dy 2>/dev/null; then
		install_dy_cli "${tmp}/dy"
	else
		install_dy_cli ""
	fi
	return 0
}

resolve_unit_src() {
	if [ -n "$UNIT_SRC" ] && [ -f "$UNIT_SRC" ]; then
		printf '%s' "$UNIT_SRC"
		return 0
	fi
	script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
	local_path="$script_dir/../packaging/dockyrd.service"
	if [ -f "$local_path" ]; then
		printf '%s' "$local_path"
		return 0
	fi
	repo="${dockyrd_GITHUB_REPO:-Joeydekruis/dockyrd}"
	ref="${dockyrd_INSTALL_REF:-development}"
	tmp=$(mktemp) || return 1
	if ! curl -sfL "https://raw.githubusercontent.com/${repo}/${ref}/packaging/dockyrd.service" -o "$tmp"; then
		rm -f "$tmp"
		return 1
	fi
	printf '%s' "$tmp"
}

install_systemd() {
	if [ "$SKIP_SYSTEMD" = "1" ]; then
		return 0
	fi
	install_build_resilience
	unit_src=$(resolve_unit_src) || {
		log "Missing unit file (set dockyrd_UNIT_SRC or run from a git checkout)." >&2
		exit 1
	}
	install -m 644 "$unit_src" /etc/systemd/system/dockyrd.service
	case "$unit_src" in
	/tmp/* | /var/tmp/*) rm -f "$unit_src" ;;
	esac
	systemctl daemon-reload
	systemctl enable dockyrd.service
}

install_build_resilience() {
	if [ "$SKIP_ROOT" = "1" ]; then
		return 0
	fi
	script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
	for unit in dy.slice dy-system.slice dy-apps.slice dy-build.slice; do
		src="$script_dir/../packaging/$unit"
		if [ -f "$src" ]; then
			install -m 644 "$src" "/etc/systemd/system/$unit"
		fi
	done
	if [ -f /sys/fs/cgroup/cgroup.controllers ]; then
		log "✓ cgroup v2 detected"
	else
		log "Warning: cgroup v2 not detected; build isolation disabled until host is upgraded." >&2
	fi
	# Swap sizing when no swap configured (briefing build resilience)
	if ! swapon --show 2>/dev/null | grep -q .; then
		mem_kb=$(awk '/MemTotal:/ {print $2}' /proc/meminfo 2>/dev/null || echo 0)
		swap_gb=2
		if [ "$mem_kb" -gt 8388608 ] 2>/dev/null; then
			swap_gb=0
		fi
		if [ "$swap_gb" -gt 0 ]; then
			swapfile=/swapfile
			if [ ! -f "$swapfile" ]; then
				log "Creating ${swap_gb}G swapfile for build resilience…"
				fallocate -l "${swap_gb}G" "$swapfile" 2>/dev/null || dd if=/dev/zero of="$swapfile" bs=1M count=$((swap_gb * 1024)) status=progress
				chmod 600 "$swapfile"
				mkswap "$swapfile" >/dev/null
				swapon "$swapfile"
				grep -q "$swapfile" /etc/fstab 2>/dev/null || echo "$swapfile none swap sw 0 0" >>/etc/fstab
			fi
		fi
	fi
	sysctl -w vm.swappiness=10 >/dev/null 2>&1 || true
	grep -q '^vm.swappiness' /etc/sysctl.d/99-dockyrd.conf 2>/dev/null || echo 'vm.swappiness=10' >/etc/sysctl.d/99-dockyrd.conf
	mem_kb=$(awk '/MemTotal:/ {print $2}' /proc/meminfo 2>/dev/null || echo 0)
	if [ "$mem_kb" -lt 8388608 ] 2>/dev/null && [ -d /sys/module/zswap ]; then
		cat >/etc/tmpfiles.d/dockyrd-zswap.conf <<'EOF'
w /sys/module/zswap/parameters/enabled          - - - - Y
w /sys/module/zswap/parameters/compressor       - - - - zstd
w /sys/module/zswap/parameters/max_pool_percent - - - - 20
EOF
		systemd-tmpfiles --create /etc/tmpfiles.d/dockyrd-zswap.conf 2>/dev/null || true
	fi
	systemctl daemon-reload
}

run_init() {
	dockyrd_CONFIG_DIR="$CONFIG_DIR" dockyrd_DATA_DIR="$DATA_DIR" \
		"$BIN_DEST" init
}

main() {
	parse_args "$@"
	log "dockyrd install, version ${dockyrd_VERSION}"
	require_root
	preflight
	ensure_docker
	ensure_paths
	install_binary
	run_init
	install_systemd
	if [ "$SKIP_SYSTEMD" != "1" ] && command -v systemctl >/dev/null 2>&1; then
		systemctl restart dockyrd.service || systemctl start dockyrd.service
	fi
	if [ "$SKIP_SYSTEMD" = "1" ]; then
		log "✓ dockyrd installed (systemd skipped). Run: $BIN_DEST serve"
	else
		log "✓ dockyrd service installed. See init output above for the setup token."
	fi
}

main "$@"
