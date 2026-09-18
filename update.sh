#!/usr/bin/env bash
# Repin the vendored sip.js browser bundle.
#
# The Go binary embeds internal/web/assets/vendor/sip.min.js (built from
# the pinned npm tarball beside it). This script rebuilds both from a (new)
# version so the pin stays reproducible offline.
#
# Usage: ./update.sh [version]     (default: latest)
set -euo pipefail

version="${1:-}"
if [[ -z "$version" ]]; then
	version="$(
		python3 - <<'EOF'
import json, urllib.request
print(json.load(urllib.request.urlopen("https://registry.npmjs.org/sip.js"))["dist-tags"]["latest"])
EOF
	)"
fi

here="$(cd "$(dirname "$0")" && pwd)"
vendor="$here/internal/web/assets/vendor"
work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

echo "pinning sip.js $version"
python3 - "$version" "$work/sip.tgz" <<'EOF'
import sys, urllib.request
urllib.request.urlretrieve(f"https://registry.npmjs.org/sip.js/-/sip.js-{sys.argv[1]}.tgz", sys.argv[2])
EOF

tar -xzf "$work/sip.tgz" -C "$work"
esbuild "$work/package/lib/index.js" \
	--bundle --minify --format=iife --global-name=SIP \
	--outfile="$vendor/sip.min.js"
cp "$work/package/LICENSE.md" "$vendor/sip.min.js.LEGAL.txt"
mv "$work/sip.tgz" "$vendor/sip.js-$version.tgz"

# keep exactly one tarball pin
shopt -s nullglob
for tgz in "$vendor"/sip.js-*.tgz; do
	if [[ "$(basename "$tgz")" != "sip.js-$version.tgz" ]]; then
		rm -f "$tgz"
	fi
done

sha256sum "$vendor/sip.js-$version.tgz" "$vendor/sip.min.js"
echo "done — commit internal/web/assets/vendor/"
