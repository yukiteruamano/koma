#!/bin/sh
# check-deps.sh - audit koma Go dependencies (versions + security).
#
# Advisory by default: reports problems but exits 0, so it can be run
# from a Makefile target or CI without failing builds.
#
# Options:
#   --update     bump outdated direct dependencies to latest (go get + tidy + vendor)
#   --vuln-only  only run the vulnerability scan
#   --quiet      suppress non-essential output
#   -h, --help   show usage

set -u

root="$(cd "$(dirname "$0")/.." && pwd)"
cd "$root" || exit 1

MODE="report"
QUIET=""

for arg in "$@"; do
	case "$arg" in
	--update) MODE="update" ;;
	--vuln-only) MODE="vuln" ;;
	--quiet) QUIET="1" ;;
	-h | --help)
		echo "usage: $0 [--update] [--vuln-only] [--quiet]"
		exit 0
		;;
	*)
		echo "unknown option: $arg" >&2
		exit 2
		;;
	esac
done

if [ -t 1 ]; then
	C_INFO="\033[1;36m"
	C_OK="\033[1;32m"
	C_WARN="\033[1;33m"
	C_ERR="\033[1;31m"
	C_RST="\033[0m"
else
	C_INFO=""
	C_OK=""
	C_WARN=""
	C_ERR=""
	C_RST=""
fi

info() { [ -z "$QUIET" ] && printf "${C_INFO}==>${C_RST} %s\n" "$1"; }
ok() { [ -z "$QUIET" ] && printf "${C_OK}   ok${C_RST} %s\n" "$1"; }
warn() { printf "${C_WARN}  warn${C_RST} %s\n" "$1"; }
err() { printf "${C_ERR}  err${C_RST} %s\n" "$1" >&2; }

command -v go >/dev/null 2>&1 || {
	err "Go toolchain not found in PATH. See https://go.dev/dl/"
	exit 1
}

gover="$(go version | sed -n 's/.*go\([0-9][0-9.]*\).*/\1/p')"
info "Go toolchain: $(go version | awk '{print $3}')"

required_go="$(sed -n 's/^go \([0-9][0-9.]*\).*/\1/p' go.mod | head -1)"
[ -z "$required_go" ] && required_go="1.26.0"

ver_ge() { # ver_ge <have> <need>  (compares major.minor)
	have_major="${1%%.*}"
	have_rest="${1#*.}"
	have_minor="${have_rest%%.*}"
	need_major="${2%%.*}"
	need_rest="${2#*.}"
	need_minor="${need_rest%%.*}"
	[ "$have_major" -gt "$need_major" ] && return 0
	[ "$have_major" -eq "$need_major" ] && [ "$have_minor" -ge "$need_minor" ] && return 0
	return 1
}

if ver_ge "$gover" "$required_go"; then
	ok "go $gover >= required $required_go (go.mod)"
else
	warn "go $gover is older than the $required_go required by go.mod"
fi

# Direct dependencies from go.mod (Path + Version), one per line.
direct="$(go mod edit -json | awk -F'"' '
	/"Path"/   { p = $4 }
	/"Version"/{ v = $4 }
	/"Indirect"/{ ind = 1 }
	/^[[:space:]]*},?[[:space:]]*$/ { if (p != "" && v != "" && ind != 1) print p " " v; p = ""; v = ""; ind = 0 }
')"

# Dependencies with no newer upstream release. They are frozen/abandoned and
# should be watched for security or functional issues.
frozen="github.com/darylhjd/mangodex github.com/ka-weihe/fast-levenshtein github.com/metafates/gache github.com/ivanpirog/coloredcobra github.com/dustin/go-humanize github.com/muesli/reflow"

# Up-to-date direct dependencies (no update available).
uptodate="$(printf '%s\n' "$direct" | awk '{print $1}' | sort -u)"

# Check frozen deps presence.
for dep in $frozen; do
	printf '%s\n' "$uptodate" | grep -qx "$dep" && warn "frozen/abandoned dep: $dep (no newer release)"
done

if [ "$MODE" != "vuln" ]; then
	info "Checking for outdated direct dependencies..."
	info "Running: go list -m -u -mod=mod all"

	updates="$(go list -m -u -mod=mod all 2>/dev/null | awk -v dirs="$direct" '
		BEGIN { n = split(dirs, arr, "\n"); for (i = 1; i <= n; i++) { split(arr[i], f, " "); if (f[1] != "") d[f[1]] = 1 } }
		/\[/ {
			pkg = $1; ver = $2; latest = $3
			sub(/^\[/, "", latest); sub(/\]$/, "", latest)
			if (d[pkg]) print pkg " " ver " " latest
		}
	')"

	if [ -z "$updates" ]; then
		ok "all direct dependencies are up to date"
	else
		printf '%s\n' "$updates" | while read -r pkg ver latest; do
			warn "outdated: $pkg $ver -> $latest"
		done
	fi

	if [ "$MODE" = "update" ]; then
		info "Updating outdated direct dependencies to latest..."
		[ -n "$updates" ] && {
			pkgs="$(printf '%s\n' "$updates" | awk '{print $1"@latest"}' | tr '\n' ' ')"
			# shellcheck disable=SC2086
			go get $pkgs
		}
		go mod tidy
		go mod vendor
		ok "dependencies updated; vendor/ regenerated"
	fi
fi

info "Scanning for known vulnerabilities with govulncheck..."
go run golang.org/x/vuln/cmd/govulncheck@latest ./...

exit 0
