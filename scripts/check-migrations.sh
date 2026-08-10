#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

up_files=$(find sql/migrations -maxdepth 1 -name '*.up.sql' | sort)
if [ -z "$up_files" ]; then
	echo "No migrations found in sql/migrations" >&2
	exit 1
fi

dups=$(printf '%s\n' "$up_files" | sed -E 's#.*/([0-9]+)_.*#\1#' | sort | uniq -d)
if [ -n "$dups" ]; then
	echo "Duplicate migration version(s): $dups" >&2
	echo "Each version number must be unique across sql/migrations." >&2
	exit 1
fi

for f in $up_files; do
	down="${f%.up.sql}.down.sql"
	if [ ! -f "$down" ]; then
		echo "Missing down migration: $down" >&2
		exit 1
	fi
done

echo "Migrations OK"
