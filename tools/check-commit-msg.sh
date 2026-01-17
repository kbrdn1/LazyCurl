#!/bin/bash
# Conventional Commits validator with Gitmoji support
# Format: [emoji] type[(scope)]: description
# Examples:
#   feat: add new feature
#   fix(auth): resolve login issue
#   ✨ feat: add new feature
#   🐛 fix(auth): resolve login issue

COMMIT_MSG_FILE="$1"
COMMIT_MSG=$(cat "$COMMIT_MSG_FILE")

# Extract first line
FIRST_LINE=$(echo "$COMMIT_MSG" | head -n 1)

# Valid types
TYPES="feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert"

# Regex patterns
# Pattern with optional emoji at start: [emoji] type[(scope)]: description
PATTERN="^(.*[[:space:]])?($TYPES)(\([a-zA-Z0-9_-]+\))?!?:[[:space:]].+"

if [[ ! "$FIRST_LINE" =~ $PATTERN ]]; then
    echo ""
    echo -e "\033[1;31m[Bad Commit message] >>\033[0m $FIRST_LINE"
    echo ""
    echo -e "    \033[0;33mYour commit message does not follow Conventional Commits formatting"
    echo -e "    \033[0;34mhttps://www.conventionalcommits.org/\033[0;33m"
    echo ""
    echo -e "    Conventional Commits start with one of the below types, followed by a colon,"
    echo -e "    followed by the commit subject. Optionally prefixed with a gitmoji:\033[0m"
    echo ""
    echo "        feat fix docs style refactor perf test build ci chore revert"
    echo ""
    echo -e "    \033[0;33mExample commit message adding a feature:\033[0m"
    echo ""
    echo "        feat: implement new API"
    echo "        ✨ feat: implement new API"
    echo ""
    echo -e "    \033[0;33mExample commit message fixing an issue:\033[0m"
    echo ""
    echo "        fix: remove infinite loop"
    echo "        🐛 fix: remove infinite loop"
    echo ""
    echo -e "    \033[0;33mExample commit with scope in parentheses:\033[0m"
    echo ""
    echo "        fix(account): remove infinite loop"
    echo "        🐛 fix(account): remove infinite loop"
    echo ""
    exit 1
fi

exit 0
