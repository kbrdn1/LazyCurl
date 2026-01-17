#!/bin/bash
# Conventional Commits validator with Gitmoji support
# Format: type[(scope)][!]: description [emoji]
#
# IMPORTANT: Emoji must be at the END for release-please compatibility
#
# Examples:
#   feat: add new feature
#   fix(auth): resolve login issue
#   feat: add new feature ✨
#   fix(auth): resolve login issue 🐛

COMMIT_MSG_FILE="$1"
COMMIT_MSG=$(cat "$COMMIT_MSG_FILE")

# Extract first line
FIRST_LINE=$(echo "$COMMIT_MSG" | head -n 1)

# Valid types
TYPES="feat|fix|docs|style|refactor|perf|test|build|ci|chore|revert"

# Regex pattern: type[(scope)][!]: description [optional emoji at end]
PATTERN="^($TYPES)(\([a-zA-Z0-9_-]+\))?!?:[[:space:]].+"

if [[ ! "$FIRST_LINE" =~ $PATTERN ]]; then
    echo ""
    echo -e "\033[1;31m[Bad Commit message] >>\033[0m $FIRST_LINE"
    echo ""
    echo -e "    \033[0;33mYour commit message does not follow Conventional Commits formatting"
    echo -e "    \033[0;34mhttps://www.conventionalcommits.org/\033[0;33m"
    echo ""
    echo -e "    Format: type(scope): description [emoji]"
    echo -e "    Note: Emoji must be at the END for release-please compatibility\033[0m"
    echo ""
    echo "        feat fix docs style refactor perf test build ci chore revert"
    echo ""
    echo -e "    \033[0;32mCorrect examples:\033[0m"
    echo ""
    echo "        feat: implement new API"
    echo "        feat: implement new API ✨"
    echo "        fix(auth): resolve login issue 🐛"
    echo ""
    echo -e "    \033[0;31mIncorrect (emoji at start):\033[0m"
    echo ""
    echo "        ✨ feat: implement new API"
    echo "        🐛 fix(auth): resolve login issue"
    echo ""
    exit 1
fi

exit 0
