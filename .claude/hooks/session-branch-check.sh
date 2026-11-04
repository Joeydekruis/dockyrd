#!/usr/bin/env bash
# SessionStart: remind the agent of the one-branch policy when HEAD is not `development`.
set -euo pipefail
cd "${CLAUDE_PROJECT_DIR:-.}"
branch=$(git rev-parse --abbrev-ref HEAD 2>/dev/null || true)
if [ -n "$branch" ] && [ "$branch" != "development" ]; then
  msg="HARD FAIL (git-branch policy): HEAD is '$branch', not 'development'. Run 'git checkout development' (carrying uncommitted work) before any edit. Never create feature/* or fix/* branches."
  python3 -c 'import json,sys; print(json.dumps({"hookSpecificOutput":{"hookEventName":"SessionStart","additionalContext":sys.argv[1]}}))' "$msg"
fi
exit 0
