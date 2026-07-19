#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"

if command -v curl >/dev/null 2>&1; then
	download() { curl -sS -f -m 30 -o "$1" "$2"; }
elif command -v wget >/dev/null 2>&1; then
	download() { wget -q -T 30 -O "$1" "$2"; }
else
	echo "curl or wget is required" >&2
	exit 1
fi

failed=0
while IFS=$'\t' read -r project path url; do
	[ -z "$project" ] && continue
	dest="$project/$path"
	mkdir -p "$(dirname "$dest")"
	echo "fetching $dest"
	if ! download "$dest" "$url"; then
		echo "  FAILED: $url" >&2
		rm -f "$dest"
		failed=1
	fi
done < <(tail -n +2 sources.tsv)

if [ "$failed" -ne 0 ]; then
	exit 1
fi

echo "done"
