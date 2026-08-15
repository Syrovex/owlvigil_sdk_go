#!/bin/sh

set -eu

go run ./scripts/check-docs-api

required_locale_docs='
README
01-quickstart
02-concepts
03-authentication-configuration
04-gateway
05-management
06-management-access-and-keys
07-management-operations
08-webhooks
09-errors-troubleshooting
10-reference-examples
'

for locale in en-US zh-CN; do
	for name in $required_locale_docs; do
		file="docs/$locale/$name.md"
		if [ ! -s "$file" ]; then
			echo "missing or empty $locale documentation file: $file" >&2
			exit 1
		fi
	done
done

for locale in en-US zh-CN; do
	if ! grep -Fq "docs/$locale/README.md" README.md; then
		echo "README is missing $locale documentation entry" >&2
		exit 1
	fi
done

domains='
billing
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
	if ! grep -Fq "management/$domain.go" docs/en-US/*.md; then
		echo "English guides are missing source domain: management/$domain.go" >&2
		exit 1
	fi
done

for source in README.md docs/en-US/*.md docs/zh-CN/*.md; do
	for target in $(grep -Eo '\]\([A-Za-z0-9._/-]+\.md(#[^)]*)?\)' "$source" | sed -e 's/^](/ /' -e 's/)$//' -e 's/#.*$//' | tr -d ' '); do
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

for source in management/*.go; do
	case "$source" in
	*_test.go | management/client.go) continue ;;
	esac
	for method in $(sed -n 's/^func (c \*Client) \([A-Z][A-Za-z0-9]*\).*/\1/p' "$source"); do
		if ! grep -Fq "\`$method\`" docs/en-US/10-reference-examples.md; then
			echo "English API reference is missing Management method: $method" >&2
			exit 1
		fi
		if ! grep -Fq "\`$method\`" docs/en-US/0[1-9]-*.md; then
			echo "English task guides are missing Management method: $method" >&2
			exit 1
		fi
		if ! grep -Fq "\`$method\`" docs/zh-CN/10-reference-examples.md; then
			echo "Chinese API reference is missing Management method: $method" >&2
			exit 1
		fi
		if ! grep -Fq "\`$method\`" docs/zh-CN/0[1-9]-*.md; then
			echo "Chinese task guides are missing Management method: $method" >&2
			exit 1
		fi
	done
done

echo "documentation checks passed"
