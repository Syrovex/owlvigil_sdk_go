#!/usr/bin/env sh
set -eu

if [ ! -f .env ]; then
	printf '%s\n' 'missing .env; copy .env.example and configure read-only credentials' >&2
	exit 1
fi

set -a
. ./.env
set +a

if [ -z "${OWLVIGIL_GATEWAY_KEY:-}" ] && [ -z "${OWLVIGIL_API_KEY:-}" ]; then
	printf '%s\n' 'configure OWLVIGIL_GATEWAY_KEY and/or OWLVIGIL_API_KEY in .env' >&2
	exit 1
fi

OWLVIGIL_LIVE_TEST=1 go test ./gateway ./management -run '^TestLive' -count=1 -v
