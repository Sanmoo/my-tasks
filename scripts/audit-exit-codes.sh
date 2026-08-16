#!/usr/bin/env bash
# Audit of the mt exit-code convention across every command: prints
# "<exit> <stdout-tail> | <stderr-tail>" per probe so inconsistencies
# in code or stream usage are visible at a glance.
# Usage: scripts/audit-exit-codes.sh <path-to-mt-binary>
set -u
MT=${1:-./bin/mt}
BASE=$(mktemp -d)
export XDG_CONFIG_HOME="$BASE/config"
export HOME="$BASE/home"
mkdir -p "$HOME"
export EDITOR="$BASE/fake-editor"
cat >"$EDITOR" <<'EOF'
#!/bin/sh
exit 0
EOF
chmod +x "$EDITOR"

V="$BASE/vault"
run() {
  local out err code
  out=$("$MT" "$@" 2>"$BASE/err")
  code=$?
  err=$(tr '\n' '|' <"$BASE/err" | cut -c1-110)
  out=$(printf '%s' "$out" | tr '\n' '|' | cut -c1-70)
  printf 'exit=%d  out=%-30s err=%s\n' "$code" "${out:0:30}" "$err"
}
label() { printf '\n== %s ==\n' "$1"; }

label "init & setup"
run init "$V" --prefix pkm
run init "$V" --prefix pkm
run init "$BASE/empty-name" --prefix ""
run init "$V/sub" --prefix sub --status review --status open

label "bookmark"
run bookmark
run bookmark list
run bookmark add pkm "$V"
run bookmark list
run bookmark add "@pkm" "$V"
run bookmark add bad/name "$V"
run bookmark rm nope
run bookmark rm pkm
run bookmark rm pkm

label "create/q (default bookmark)"
run bookmark add pkm "$V"
printf 'default: pkm\nbookmarks:\n  pkm: %s\n' "$V" >"$XDG_CONFIG_HOME/mt/config.yaml" 2>/dev/null || {
  mkdir -p "$XDG_CONFIG_HOME/mt"
  printf 'default: pkm\nbookmarks:\n  pkm: %s\n' "$V" >"$XDG_CONFIG_HOME/mt/config.yaml"
}
run create
run q
run create "first issue"
run q "quiet issue"
ID1=$("$MT" q "captured" | tr -d '[:space:]')

label "show/edit"
run show "$ID1"
run show nope
run show "a/b"
run show
run edit nope
run edit "$ID1"

label "status transitions"
run done "$ID1"
run close "$ID1"
run reopen "$ID1"
run status "$ID1" open
run status "$ID1" bogus
run status "$ID1"
run done
run done a b

label "defer"
run defer "$ID1" +2d
run defer "$ID1" "26-08-20 08:00"
run defer "$ID1" banana
run defer "$ID1"

label "undefer"
run defer "$ID1" 99-01-01 00:00
run undefer
run undefer "$ID1"
run undefer "$ID1"
run undefer "$ID1" extra
run undefer nope

label "dep"
ID2=$("$MT" q "second issue" | tr -d '[:space:]')
run dep
run dep nope
run dep add "$ID1" "$ID2"
run dep add "$ID1" "$ID2"
run dep add "$ID1" nope
run dep add "$ID1" "$ID1"
run dep add "$ID1"
run dep add "$ID1" "$ID2" extra
run dep rm "$ID1" "$ID2"
run dep rm "$ID1" "$ID2"
run dep rm "$ID1" nope
run dep rm "$ID1"
run dep rm "$ID1" "$ID2" extra

label "prioritize / rank"
run prioritize
run top "$ID1"
run bottom "$ID1"
run rank "$ID1" 1
run rank "$ID1" 0
run rank "$ID1" x
run unrank "$ID1"
run top

label "list family"
run list
run list --all
run list --status open
run list --label foo
run list extra
run ready
run overdue
run check
run check --fix
run check extra
run pick-next
run pick-next extra

label "comment"
run comment "$ID1" hello there
run comment "$ID1"
run comment

label "vault addressing"
run list @nope
run list @pkm
run list --vault "$V"
run list @pkm @pkm
run list --vault "$V" --vault "$V"

label "help"
run --help
run help
run help list
run help nope

echo
echo "cleanup: $BASE"
