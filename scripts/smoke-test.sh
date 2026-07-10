#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR=$(CDPATH= cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." && pwd)
WORK_DIR=$(mktemp -d)
trap 'rm -rf "$WORK_DIR"' EXIT

BIN="$WORK_DIR/cipr"

(
    cd "$ROOT_DIR"
    go build -o "$BIN" .
)

export HOME="$WORK_DIR/home"
export XDG_CACHE_HOME="$WORK_DIR/cache"
mkdir -p "$HOME" "$XDG_CACHE_HOME"

run_and_expect() {
    local name=$1
    local expected=$2
    shift 2

    local output="$WORK_DIR/${name}.out"
    "$BIN" "$@" >"$output"
    if ! grep -Fq -- "$expected" "$output"; then
        echo "Smoke test '$name' did not contain expected output: $expected" >&2
        return 1
    fi
}

run_and_expect help 'default "config"' --help
run_and_expect aws "44.192.140.112/28" aws \
    --source "$ROOT_DIR/internal/testdata/mock_ip_ranges_response.json" \
    --ipv4 --filter-region us-east-1 --filter-service EBS \
    --filter-network-border-group us-east-1
run_and_expect azure "13.66.143.220/30" azure \
    --source "$ROOT_DIR/internal/testdata/azure_servicetags_sample.json" --ipv4
run_and_expect gcp "34.80.0.0/15" gcp \
    --source "$ROOT_DIR/internal/testdata/gcp_cloud_sample.json" --ipv4 \
    --filter-scope asia-east1 --filter-service "Google Cloud"
test "$(wc -l < "$WORK_DIR/gcp.out")" -eq 1
run_and_expect cloudflare-v4 "173.245.48.0/20" cloudflare \
    --source "$ROOT_DIR/internal/testdata/cloudflare_ipv4.txt" --ipv4
run_and_expect cloudflare-v6 "2400:cb00::/32" cloudflare \
    --source "$ROOT_DIR/internal/testdata/cloudflare_ipv6.txt" --ipv6
run_and_expect digitalocean "5.101.96.0/21" do \
    --source "$ROOT_DIR/internal/testdata/do.csv" --ipv4 \
    --filter-country NL --filter-city Amsterdam
run_and_expect github "192.30.252.0/22" github \
    --source "$ROOT_DIR/internal/testdata/github_meta_sample.json" --ipv4 \
    --filter-service web
run_and_expect icloud "172.224.224.0/27" icloud \
    --source "$ROOT_DIR/internal/testdata/icloud.csv" --ipv4 \
    --filter-country GB --filter-city London

CONFIGURE_HOME="$WORK_DIR/configure-home"
mkdir -p "$CONFIGURE_HOME"
HOME="$CONFIGURE_HOME" "$BIN" configure aws \
    --local-file "$ROOT_DIR/internal/testdata/aws.json" \
    --cache-ttl 0s \
    --proxy http://proxy.example:8080 \
    --debug >"$WORK_DIR/configure.out" 2>"$WORK_DIR/configure.err"
grep -Fq 'aws_active_source = "local-file"' "$WORK_DIR/configure.out"
grep -Fq 'proxy = "http://proxy.example:8080"' "$WORK_DIR/configure.out"
grep -Fq '[debug] config: using' "$WORK_DIR/configure.err"
grep -Fq "aws_local_file = '$ROOT_DIR/internal/testdata/aws.json'" \
    "$CONFIGURE_HOME/.config/cipr/cipr.toml"
grep -Fq "aws_cache_ttl = '0s'" "$CONFIGURE_HOME/.config/cipr/cipr.toml"

HOME="$CONFIGURE_HOME" "$BIN" configure aws --local-file= >"$WORK_DIR/configure-clear.out"
grep -Fq 'aws_active_source = "endpoint"' "$WORK_DIR/configure-clear.out"

CUSTOM_CONFIG="$WORK_DIR/custom-config/cipr.toml"
HOME="$CONFIGURE_HOME" "$BIN" --config "$CUSTOM_CONFIG" configure github \
    --cache-ttl 0s >"$WORK_DIR/configure-custom.out"
test -f "$CUSTOM_CONFIG"
grep -Fq "github_cache_ttl = '0s'" "$CUSTOM_CONFIG"

UNINSTALL_HOME="$WORK_DIR/uninstall-home"
mkdir -p "$UNINSTALL_HOME/.cipr/bin" "$UNINSTALL_HOME/.config/cipr" "$UNINSTALL_HOME/.cache/cipr"
cp "$BIN" "$UNINSTALL_HOME/.cipr/bin/cipr"
touch "$UNINSTALL_HOME/.config/cipr/cipr.toml"
touch "$UNINSTALL_HOME/.cache/cipr/ranges.cache"
HOME="$UNINSTALL_HOME" "$ROOT_DIR/uninstall.sh" >"$WORK_DIR/uninstall.out"

test ! -e "$UNINSTALL_HOME/.cipr/bin/cipr"
test ! -e "$UNINSTALL_HOME/.config/cipr"
test -e "$UNINSTALL_HOME/.cache/cipr/ranges.cache"

CONFIG_ONLY_HOME="$WORK_DIR/config-only-home"
mkdir -p "$CONFIG_ONLY_HOME/.config/cipr"
touch "$CONFIG_ONLY_HOME/.config/cipr/cipr.toml"
HOME="$CONFIG_ONLY_HOME" "$ROOT_DIR/uninstall.sh" >"$WORK_DIR/uninstall-config-only.out"
test ! -e "$CONFIG_ONLY_HOME/.config/cipr"

echo "Smoke tests passed."
