# Dockyrd proxybuild — Caddy + Coraza WAF + rate limiting (xcaddy).
# Published as ghcr.io/joeydekruis/dockyrd-proxy alongside dockyrd releases.

ARG CADDY_VERSION=2.8.4

FROM caddy:${CADDY_VERSION}-builder AS builder

RUN xcaddy build \
	--with github.com/corazawaf/coraza-caddy/v2 \
	--with github.com/mholt/caddy-ratelimit

FROM caddy:${CADDY_VERSION}-alpine

COPY --from=builder /usr/bin/caddy /usr/bin/caddy

LABEL org.opencontainers.image.source="https://github.com/Joeydekruis/dockyrd"
LABEL dev.dockyrd.kind="proxybuild"