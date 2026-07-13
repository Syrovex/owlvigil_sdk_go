#!/bin/sh

set -eu

required_docs='
authentication
environments
gateway
management
access-control
model-routing
financial-governance
account
billing
teams
oauth2
streaming
webhooks
errors
pagination
examples
troubleshooting
quickstart
release-v0.1.0
'

for name in $required_docs; do
	file="docs/$name.md"
	if [ ! -s "$file" ]; then
		echo "missing or empty documentation file: $file" >&2
		exit 1
	fi
done

for name in $required_docs; do
	if ! grep -Fq "docs/$name.md" README.md; then
		echo "README is missing documentation link: docs/$name.md" >&2
		exit 1
	fi
done

domains='
billing
documentation
financial
gateway_keys
invitations
logs
members
models
orders
payment_methods
policies
providers
rbac
subscription
teams
topup
usage
user
webhooks
workspaces
'

for domain in $domains; do
	if ! grep -Fq "management/$domain.go" docs/management.md; then
		echo "management guide is missing source domain: management/$domain.go" >&2
		exit 1
	fi
done

for source in README.md docs/*.md; do
	for target in $(grep -Eo '\]\([A-Za-z0-9._/-]+\.md\)' "$source" | sed -e 's/^](/ /' -e 's/)$//' | tr -d ' '); do
		case "$target" in
		docs/*) candidate=$target ;;
		*) candidate=$(dirname "$source")/$target ;;
		esac
		if [ ! -f "$candidate" ]; then
			echo "broken Markdown link in $source: $target" >&2
			exit 1
		fi
	done
done

echo "documentation checks passed"
