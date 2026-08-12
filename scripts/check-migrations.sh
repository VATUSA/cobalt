#!/usr/bin/env bash
set -euo pipefail

BASE=""
HEAD=""
while [ $# -gt 0 ]; do
	case "$1" in
		--base)
			BASE="${2:-}"
			if [ -z "$BASE" ]; then
				echo "--base requires a git ref" >&2
				exit 2
			fi
			shift 2
			;;
		--head)
			HEAD="${2:-}"
			if [ -z "$HEAD" ]; then
				echo "--head requires a git ref" >&2
				exit 2
			fi
			shift 2
			;;
		*)
			echo "Usage: check-migrations.sh [--base <git-ref>] [--head <git-ref>]" >&2
			exit 2
			;;
	esac
done

cd "$(dirname "$0")/.."

version() {
	grep -oE '^[0-9]+' <<<"$(basename "$1")"
}

sha_for() {
	local path="$1" list="$2"
	printf '%s\n' "$list" | awk -F'\t' -v p="$path" '$1 == p { print $2; exit }'
}

if [ -n "$HEAD" ]; then
	if ! git rev-parse --verify "$HEAD" >/dev/null 2>&1; then
		echo "Head ref $HEAD not found in this checkout" >&2
		exit 1
	fi
	head_files=$(git ls-tree -r --name-only "$HEAD" -- sql/migrations | grep -E '\.sql$' || true)
else
	head_files=$(find sql/migrations -maxdepth 1 -name '*.sql' | sort)
fi

if [ -z "$head_files" ]; then
	echo "No migrations found in sql/migrations" >&2
	exit 1
fi

head_up=$(printf '%s\n' "$head_files" | grep -E '\.up\.sql$' || true)
if [ -z "$head_up" ]; then
	echo "No up migrations found in sql/migrations" >&2
	exit 1
fi

dups=$(while IFS= read -r f; do version "$f"; done <<<"$head_up" | sort | uniq -d)
if [ -n "$dups" ]; then
	echo "Duplicate migration version(s): $dups" >&2
	echo "Each version number must be unique across sql/migrations." >&2
	exit 1
fi

scheme_len=$(while IFS= read -r f; do v=$(version "$f"); echo "${#v}"; done <<<"$head_up" | sort -u | wc -l)
if [ "$scheme_len" -gt 1 ]; then
	echo "Mixed migration versioning detected." >&2
	echo "All versions must share a digit length: 6 for -seq, 14 for timestamps." >&2
	echo "The default migrate create (no -seq) writes timestamps and breaks -seq afterward." >&2
	echo "Use: migrate create -ext sql -dir sql/migrations -seq <name>" >&2
	exit 1
fi

while IFS= read -r f; do
	down="${f%.up.sql}.down.sql"
	if ! grep -qxF "$down" <<<"$head_files"; then
		echo "Missing down migration: $down" >&2
		exit 1
	fi
done <<<"$head_up"

if [ -n "$BASE" ]; then
	if ! git rev-parse --verify "$BASE" >/dev/null 2>&1; then
		echo "Base ref $BASE not found in this checkout" >&2
		exit 1
	fi
	base_files=$(git ls-tree -r --name-only "$BASE" -- sql/migrations | grep -E '\.sql$' || true)
	base_up=$(printf '%s\n' "$base_files" | grep -E '\.up\.sql$' || true)
	if [ -z "$base_up" ]; then
		echo "Base $BASE has no migrations; skipping ordering and completeness checks"
	else
		base_versions=$(while IFS= read -r f; do version "$f"; done <<<"$base_up")
		base_max=$(printf '%s\n' "$base_versions" | sort -n | tail -1)
		head_versions=$(while IFS= read -r f; do version "$f"; done <<<"$head_up")

		while IFS= read -r f; do
			v=$(version "$f")
			if ! grep -qxF "$v" <<<"$base_versions"; then
				if [ "$v" -le "$base_max" ]; then
					echo "Migration $f has version $v, not newer than base's max version $base_max" >&2
					echo "A migration version is consumed once merged to the release branch." >&2
					echo "Rename this migration's .up and .down files to a newer -seq number before merging." >&2
					exit 1
				fi
			fi
		done <<<"$head_up"

		while IFS= read -r v; do
			if ! grep -qxF "$v" <<<"$head_versions"; then
				echo "Base migration version $v is missing from head." >&2
				echo "Merged migrations cannot be deleted or renumbered; rebase if your branch predates it." >&2
				exit 1
			fi
		done <<<"$base_versions"
	fi

	if [ -n "$base_files" ]; then
		base_sha=$(git ls-tree -r "$BASE" -- sql/migrations | sed -nE 's#^[0-9]+ [a-z]+ ([0-9a-f]+)\t(sql/migrations/.*\.(up|down)\.sql)$#\2\t\1#p')
		if [ -n "$HEAD" ]; then
			head_sha=$(git ls-tree -r "$HEAD" -- sql/migrations | sed -nE 's#^[0-9]+ [a-z]+ ([0-9a-f]+)\t(sql/migrations/.*\.(up|down)\.sql)$#\2\t\1#p')
		else
			head_sha=""
			while IFS= read -r f; do
				head_sha+=$'\n'"$(printf '%s\t%s' "$f" "$(git hash-object "$f")")"
			done <<<"$base_files"
		fi
		while IFS= read -r path; do
			bs=$(sha_for "$path" "$base_sha")
			hs=$(sha_for "$path" "$head_sha")
			if [ -z "$hs" ]; then
				echo "Migration $path is missing from head (deleted or renamed)." >&2
				echo "Merged migrations are immutable; make changes in a new migration." >&2
				exit 1
			fi
			if [ "$hs" != "$bs" ]; then
				echo "Migration $path was modified relative to base." >&2
				echo "Merged migrations are immutable; make changes in a new migration." >&2
				exit 1
			fi
		done <<<"$(cut -f1 <<<"$base_sha")"
	fi
fi

echo "Migrations OK"