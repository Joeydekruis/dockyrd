#!/bin/sh
# Install the Railpack CLI when missing (default builder for app deploys).
set -e

log() {
	printf '%s\n' "$*"
}

if [ "${dockyrd_SKIP_RAILPACK:-0}" = "1" ]; then
	exit 0
fi

if command -v railpack >/dev/null 2>&1; then
	log "✓ Railpack CLI already installed"
	exit 0
fi

if ! command -v curl >/dev/null 2>&1; then
	log "curl is required to install Railpack." >&2
	exit 1
fi

bin_dir="/usr/local/bin"
if [ -n "${RAILPACK_BIN_DIR:-}" ]; then
	bin_dir="$RAILPACK_BIN_DIR"
elif [ -n "${dockyrd_BIN_DEST:-}" ]; then
	bin_dir=$(dirname "$dockyrd_BIN_DEST")
fi

export RAILPACK_BIN_DIR="$bin_dir"
log "Installing Railpack CLI to ${bin_dir}…"
curl -fsSL https://railpack.com/install.sh | sh -s -- -y
if ! command -v railpack >/dev/null 2>&1; then
	if [ -x "${bin_dir}/railpack" ]; then
		log "✓ Railpack CLI installed to ${bin_dir}/railpack"
		exit 0
	fi
	log "Railpack install finished but railpack is not on PATH." >&2
	exit 1
fi
log "✓ Railpack CLI installed"
