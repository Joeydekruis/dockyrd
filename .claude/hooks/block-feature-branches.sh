#!/usr/bin/env bash
# PreToolUse(Bash): block commands that create or switch onto per-issue branches.
# Deleting leftover feature/fix branches is allowed. Port of .cursor/hooks/block-feature-branches.sh.
set -euo pipefail

input=$(cat)
command=$(printf '%s' "$input" | python3 -c 'import json,sys; print((json.load(sys.stdin).get("tool_input") or {}).get("command") or "")' 2>/dev/null || true)

deny() {
  python3 -c 'import json,sys; print(json.dumps({"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"deny","permissionDecisionReason":sys.argv[1]}}))' "$1"
  exit 0
}

# Create new per-issue branches
if printf '%s' "$command" | grep -Eqi 'git[[:space:]]+(checkout[[:space:]]+-b|switch[[:space:]]+-c)[[:space:]]+(feature/|fix/|jdekruis/)'; then
  deny "Blocked: do not create feature/fix branches. Stay on development (see .cursor/rules/git-branch.mdc)."
fi

# Switch onto an existing per-issue branch (allowed only if the target is development/main)
if printf '%s' "$command" | grep -Eqi 'git[[:space:]]+(checkout|switch)[[:space:]]+' \
  && printf '%s' "$command" | grep -Eqi '(feature/|fix/|jdekruis/)' \
  && ! printf '%s' "$command" | grep -Eqi 'git[[:space:]]+(checkout[[:space:]]+-b|switch[[:space:]]+-c)'; then
  if ! printf '%s' "$command" | grep -Eqi 'git[[:space:]]+(checkout|switch)[[:space:]]+([^[:space:]]+[[:space:]]+)*(development|main)([[:space:]]|$)'; then
    deny "Blocked: do not switch to feature/fix branches. Use: git checkout development"
  fi
fi

# git branch <name> (create), excluding delete/list/rename/upstream flags
if printf '%s' "$command" | grep -Eqi 'git[[:space:]]+branch[[:space:]]+' \
  && printf '%s' "$command" | grep -Eqi '(feature/|fix/|jdekruis/)' \
  && ! printf '%s' "$command" | grep -Eqi 'git[[:space:]]+branch[[:space:]]+(-[dD]|--[[:alnum:]-]*delete|--list|-v|-vv|-a|-r|-m|-M|--merged|--no-merged|-u|--set-upstream-to)'; then
  deny "Blocked: do not create feature/fix branches. Work only on development."
fi

exit 0
