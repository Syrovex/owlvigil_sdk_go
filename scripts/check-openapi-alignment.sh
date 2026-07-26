#!/bin/sh
set -eu

openapi_repo=${1:-../owlvigil_openapi}
dashboard_repo=${2:-../owlvigil_dashboard}
openapi_catalog="$openapi_repo/test/k6/openapi-management/catalog.js"
dashboard_catalog="$dashboard_repo/backend/internal/server/managementroute/catalog.go"
sdk_use_cases="management/all_operations_usecase_test.go"
sdk_live_use_cases="examples/openapi-smoke/main.go"
alignment_tmp=$(mktemp -d "${TMPDIR:-/tmp}/owlvigil-openapi-alignment.XXXXXX")
trap 'rm -rf "$alignment_tmp"' EXIT HUP INT TERM

if [ ! -f "$openapi_catalog" ]; then
    echo "Open API catalog not found: $openapi_catalog" >&2
    exit 2
fi

if [ ! -f "$dashboard_catalog" ]; then
    echo "Dashboard Management catalog not found: $dashboard_catalog" >&2
    exit 2
fi

extract_quoted_operations() {
    rg -o '"(DELETE|GET|PATCH|POST|PUT) /v1/[^"]+"' "$1" |
        sed 's/^"//; s/"$//' |
        sort -u
}

extract_manifest_operations() {
    rg '^(DELETE|GET|PATCH|POST|PUT) /v1/' "$1" |
        sort -u
}

extract_quoted_operations "$openapi_catalog" >"$alignment_tmp/openapi"
extract_manifest_operations "$dashboard_catalog" >"$alignment_tmp/dashboard"
extract_quoted_operations "$sdk_use_cases" >"$alignment_tmp/sdk"
extract_quoted_operations "$sdk_live_use_cases" >"$alignment_tmp/sdk-live"
openapi_count=$(wc -l <"$alignment_tmp/openapi" | tr -d ' ')
dashboard_count=$(wc -l <"$alignment_tmp/dashboard" | tr -d ' ')
sdk_count=$(wc -l <"$alignment_tmp/sdk" | tr -d ' ')
sdk_live_count=$(wc -l <"$alignment_tmp/sdk-live" | tr -d ' ')

if ! diff -u "$alignment_tmp/openapi" "$alignment_tmp/dashboard"; then
    echo "Dashboard Management routes do not match the Open API catalog" >&2
    exit 1
fi

if ! diff -u "$alignment_tmp/openapi" "$alignment_tmp/sdk"; then
    echo "Executable SDK Management use cases do not match the Open API catalog" >&2
    exit 1
fi

if ! diff -u "$alignment_tmp/openapi" "$alignment_tmp/sdk-live"; then
    echo "Live SDK Management use cases do not match the Open API catalog" >&2
    exit 1
fi

(
    cd "$openapi_repo"
    go test ./internal/facade -run '^TestManagementOperationsMatchK6Catalog$' -count=1
)
(
    cd "$dashboard_repo"
    go test ./backend/internal/server -run '^TestManagementRoutesMatchCatalog$' -count=1
)
go test ./management -run '^TestAllPublishedManagementOperationsHaveExecutableSDKUseCases$' -count=1
openapi_repo_abs=$(cd "$openapi_repo" && pwd)
OWLVIGIL_OPENAPI_REPO="$openapi_repo_abs" \
    go test ./management -run '^TestAllExecutableManagementUseCasesPassRefactoredOpenAPIFacade$' -count=1

echo "Open API alignment passed: openapi=$openapi_count dashboard=$dashboard_count sdk_executable=$sdk_count sdk_live=$sdk_live_count"
