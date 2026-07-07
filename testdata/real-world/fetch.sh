#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "${BASH_SOURCE[0]}")"

tail -n +2 sources.tsv | while IFS=$'\t' read -r project path url; do
	[ -z "$project" ] && continue
	dest="$project/$path"
	mkdir -p "$(dirname "$dest")"
	echo "fetching $dest"
	if ! curl -sS -f -m 30 -o "$dest" "$url"; then
		echo "  FAILED: $url" >&2
		rm -f "$dest"
	fi
done

echo "done"
