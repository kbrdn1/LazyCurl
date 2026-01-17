#!/bin/bash

# =============================================================================
# Git Worktree Manager for LazyCurl (gwq-powered)
# =============================================================================
# Manages Git worktrees using gwq (https://github.com/d-kuro/gwq)
# Branch convention: <type>/#<issue-number>-<short-description>
# Worktree location: ~/cc-worktree/<project-name>/<type>-<issue>-<desc>
# =============================================================================

# Colors
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# Check if gwq is installed
check_gwq() {
    if ! command -v gwq &> /dev/null; then
        echo -e "${RED}Error: gwq is not installed!${NC}"
        echo -e "${YELLOW}Install with: brew install d-kuro/tap/gwq${NC}"
        echo -e "${YELLOW}Or: go install github.com/d-kuro/gwq/cmd/gwq@latest${NC}"
        exit 1
    fi
}

# Configuration
MAIN_REPO=$(git rev-parse --show-toplevel)
REPO_NAME=$(basename "$MAIN_REPO")
WORKTREE_BASE="$HOME/cc-worktree/${REPO_NAME}"

# Branch types following CONTRIBUTING.md convention
BRANCH_TYPES=(
    "feat:New feature implementation"
    "fix:Bug fix"
    "hotfix:Critical production bug fix"
    "docs:Documentation changes"
    "test:Test additions or modifications"
    "refactor:Code restructuring"
    "chore:Maintenance tasks"
    "perf:Performance improvements"
    "ci:CI/CD configuration"
    "build:Build system changes"
)

show_header() {
    echo -e "${CYAN}${BOLD}"
    echo "╔══════════════════════════════════════════════════════════════╗"
    echo "║           Git Worktree Manager (gwq) - LazyCurl              ║"
    echo "╠══════════════════════════════════════════════════════════════╣"
    echo "║  Location: ~/cc-worktree/${REPO_NAME}/<type>-<issue>-<desc>  ║"
    echo "╚══════════════════════════════════════════════════════════════╝"
    echo -e "${NC}"
}

show_menu() {
    echo -e "${YELLOW}${BOLD}Available actions:${NC}"
    echo -e "  ${GREEN}1${NC}) Create new worktree (guided)"
    echo -e "  ${GREEN}2${NC}) Create worktree (gwq fuzzy finder)"
    echo -e "  ${GREEN}3${NC}) List worktrees"
    echo -e "  ${GREEN}4${NC}) Navigate to worktree"
    echo -e "  ${GREEN}5${NC}) Remove worktree"
    echo -e "  ${GREEN}6${NC}) Monitor worktrees (status --watch)"
    echo -e "  ${GREEN}7${NC}) Cleanup (prune stale references)"
    echo -e "  ${GREEN}q${NC}) Quit"
    echo ""
}

select_branch_type() {
    echo -e "${YELLOW}${BOLD}Select branch type:${NC}"
    local i=1
    for type_desc in "${BRANCH_TYPES[@]}"; do
        local type=$(echo "$type_desc" | cut -d: -f1)
        local desc=$(echo "$type_desc" | cut -d: -f2)
        echo -e "  ${GREEN}$i${NC}) ${BOLD}$type${NC} - $desc"
        ((i++))
    done
    echo ""
    read -p "Enter choice [1-${#BRANCH_TYPES[@]}]: " type_choice

    if [[ "$type_choice" =~ ^[0-9]+$ ]] && [ "$type_choice" -ge 1 ] && [ "$type_choice" -le ${#BRANCH_TYPES[@]} ]; then
        SELECTED_TYPE=$(echo "${BRANCH_TYPES[$((type_choice-1))]}" | cut -d: -f1)
        return 0
    else
        echo -e "${RED}Invalid selection${NC}"
        return 1
    fi
}

create_worktree_guided() {
    echo -e "${CYAN}${BOLD}Create New Worktree (Guided)${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

    # Select branch type
    if ! select_branch_type; then
        return 1
    fi

    # Get issue number
    echo ""
    read -p "Enter GitHub issue number: " issue_number
    if [[ ! "$issue_number" =~ ^[0-9]+$ ]]; then
        echo -e "${RED}Invalid issue number${NC}"
        return 1
    fi

    # Get description
    echo ""
    read -p "Enter short description (use-kebab-case): " description
    if [[ -z "$description" ]]; then
        echo -e "${RED}Description cannot be empty${NC}"
        return 1
    fi

    # Convert description to kebab-case
    description=$(echo "$description" | tr '[:upper:]' '[:lower:]' | tr ' ' '-' | sed 's/[^a-z0-9-]//g')

    # Build branch name following convention: type/#issue-description
    BRANCH_NAME="${SELECTED_TYPE}/#${issue_number}-${description}"

    echo ""
    echo -e "${YELLOW}Summary:${NC}"
    echo -e "  Branch: ${GREEN}${BRANCH_NAME}${NC}"
    echo -e "  Directory: ${GREEN}~/cc-worktree/${REPO_NAME}/${SELECTED_TYPE}-${issue_number}-${description}${NC}"
    echo ""

    read -p "Create worktree? [Y/n]: " confirm
    if [[ "$confirm" =~ ^[Nn]$ ]]; then
        echo -e "${YELLOW}Cancelled${NC}"
        return 0
    fi

    # Use gwq to create worktree
    echo -e "${BLUE}Creating worktree with gwq...${NC}"
    gwq add -b "$BRANCH_NAME"

    if [ $? -eq 0 ]; then
        echo ""
        echo -e "${GREEN}${BOLD}Worktree created successfully!${NC}"
        echo ""
        echo -e "${YELLOW}Next steps:${NC}"
        echo -e "  cd \$(gwq get ${SELECTED_TYPE})"
        echo -e "  make deps"
        echo -e "  claude"
    else
        echo -e "${RED}Failed to create worktree${NC}"
        return 1
    fi
}

create_worktree_fuzzy() {
    echo -e "${CYAN}${BOLD}Create Worktree (gwq fuzzy finder)${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${YELLOW}Tip: Type branch name like: feat/#123-my-feature${NC}"
    echo ""
    gwq add -i
}

list_worktrees() {
    echo -e "${CYAN}${BOLD}Current Worktrees${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    gwq list -v
}

navigate_worktree() {
    echo -e "${CYAN}${BOLD}Navigate to Worktree${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${YELLOW}Use fuzzy finder to select a worktree...${NC}"
    echo ""
    gwq cd
}

remove_worktree() {
    echo -e "${CYAN}${BOLD}Remove Worktree${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""

    # Show current worktrees
    gwq list -v
    echo ""

    read -p "Enter branch pattern to remove (or 'q' to cancel): " pattern
    if [[ "$pattern" == "q" ]]; then
        echo -e "${YELLOW}Cancelled${NC}"
        return 0
    fi

    read -p "Also delete the branch? [y/N]: " delete_branch

    if [[ "$delete_branch" =~ ^[Yy]$ ]]; then
        gwq remove -b "$pattern"
    else
        gwq remove "$pattern"
    fi
}

monitor_worktrees() {
    echo -e "${CYAN}${BOLD}Monitor Worktrees (Ctrl+C to stop)${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    gwq status --watch
}

cleanup_worktrees() {
    echo -e "${CYAN}${BOLD}Cleanup Worktrees${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo -e "${BLUE}Pruning stale worktree references...${NC}"
    gwq prune
    echo -e "${GREEN}Cleanup complete!${NC}"
}

# Quick create mode (non-interactive)
quick_create() {
    local type="$1"
    local issue="$2"
    local desc="$3"

    if [[ -z "$type" ]] || [[ -z "$issue" ]] || [[ -z "$desc" ]]; then
        echo -e "${RED}Usage: $0 create <type> <issue-number> <description>${NC}"
        echo -e "${YELLOW}Example: $0 create feat 123 user-authentication${NC}"
        echo ""
        echo -e "${YELLOW}Available types:${NC}"
        for type_desc in "${BRANCH_TYPES[@]}"; do
            local t=$(echo "$type_desc" | cut -d: -f1)
            echo -e "  - $t"
        done
        exit 1
    fi

    # Validate type
    local valid_type=false
    for type_desc in "${BRANCH_TYPES[@]}"; do
        local t=$(echo "$type_desc" | cut -d: -f1)
        if [ "$type" == "$t" ]; then
            valid_type=true
            break
        fi
    done

    if [ "$valid_type" = false ]; then
        echo -e "${RED}Invalid type: $type${NC}"
        exit 1
    fi

    # Build branch name
    desc=$(echo "$desc" | tr '[:upper:]' '[:lower:]' | tr ' ' '-' | sed 's/[^a-z0-9-]//g')
    BRANCH_NAME="${type}/#${issue}-${desc}"

    echo -e "${YELLOW}Creating worktree with gwq...${NC}"
    echo -e "  Branch: ${GREEN}${BRANCH_NAME}${NC}"

    gwq add -b "$BRANCH_NAME"

    if [ $? -eq 0 ]; then
        echo ""
        echo -e "${GREEN}${BOLD}Worktree created!${NC}"
        echo -e "${YELLOW}Next: cd \$(gwq get ${type}) && make deps && claude${NC}"
    fi
}

# Main
main() {
    # Check gwq is installed
    check_gwq

    # Check if we're in a git repository
    if ! git rev-parse --git-dir > /dev/null 2>&1; then
        echo -e "${RED}Error: Not in a git repository${NC}"
        exit 1
    fi

    # Handle command line arguments for quick mode
    case "$1" in
        create)
            quick_create "$2" "$3" "$4"
            exit 0
            ;;
        list)
            list_worktrees
            exit 0
            ;;
        cleanup)
            cleanup_worktrees
            exit 0
            ;;
    esac

    # Interactive mode
    show_header

    while true; do
        show_menu
        read -p "Select action: " action
        echo ""

        case "$action" in
            1) create_worktree_guided ;;
            2) create_worktree_fuzzy ;;
            3) list_worktrees ;;
            4) navigate_worktree ;;
            5) remove_worktree ;;
            6) monitor_worktrees ;;
            7) cleanup_worktrees ;;
            q|Q)
                echo -e "${GREEN}Goodbye!${NC}"
                exit 0
                ;;
            *)
                echo -e "${RED}Invalid option${NC}"
                ;;
        esac

        echo ""
        read -p "Press Enter to continue..."
        clear
        show_header
    done
}

main "$@"
