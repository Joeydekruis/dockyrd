#!/usr/bin/env bash
# Block shell commands that create or switch onto per-issue branches.
# Deleting leftover feature/fix branches is allowed.
set -euo pipefail

input=$(cat)
command=$(printf '%s' "$input" | python3 -c 'import json,sys; print(json.load(sys.stdin).get("command") or "")' 2>/dev/null || true)

deny() {
  local msg="$1"
  python3 -c 'import json,sys; print(json.dumps({"permission":"deny","user_message":sys.argv[1],"agent_message":sys.argv[1]}))' "$msg"
  exit 0
}

# Create new per-issue branches
if printf '%s' "$command" | grep -Eqi 'git[[:space:]]+(checkout[[:space:]]+-b|switch[[:space:]]+-c)[[:space:]]+(feature/|fix/|jdekruis/)'; then
  deny "Blocked: do not create feature/fix branches. Stay on development (see .cursor/rules/git-branch.mdc)."
fi

# Explicit checkout/switch onto an existing per-issue branch (not -b/-c already handled)
if printf '%s' "$command" | grep -Eqi 'git[[:space:]]+(checkout|switch)[[:space:]]+' \
  && printf '%s' "$command" | grep -Eqi '(feature/|fix/|jdekruis/)' \
  && ! printf '%s' "$command" | grep -Eqi 'git[[:space:]]+(checkout[[:space:]]+-b|switch[[:space:]]+-c)'; then
  # Allow if the target is clearly development or main
  if ! printf '%s' "$command" | grep -Eqi 'git[[:space:]]+(checkout|switch)[[:space:]]+([^[:space:]]+[[:space:]]+)*(development|main)([[:space:]]|$)'; then
    deny "Blocked: do not switch to feature/fix branches. Use: git checkout development"
  fi
fi

# git branch <name> create (no -d/-D/-m/-c flags that are cleanup)
if printf '%s' "$command" | grep -Eqi 'git[[:space:]]+branch[[:space:]]+' \
  && printf '%s' "$command" | grep -Eqi '(feature/|fix/|jdekruis/)' \
  && ! printf '%s' "$command" | grep -Eqi 'git[[:space:]]+branch[[:space:]]+(-[dD]|--[[:alnum:]-]*delete|--list|-v|-vv|-a|-r|-m|-M|--merged|--no-merged|-u|--set-upstream-to)'; then
  deny "Blocked: do not create feature/fix branches. Work only on development."
fi

printf '%s\n' '{"permission":"allow"}'
exit 0
