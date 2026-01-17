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

        # Get worktree path
        local worktree_path=$(gwq get "${SELECTED_TYPE}/#${issue_number}" 2>/dev/null || gwq get "${SELECTED_TYPE}" 2>/dev/null)

        if [ -n "$worktree_path" ] && [ -d "$worktree_path" ]; then
            setup_worktree "$worktree_path" "$issue_number" "$description"
            launch_claude "$worktree_path"
        else
            echo ""
            echo -e "${YELLOW}Next steps:${NC}"
            echo -e "  cd \$(gwq get ${SELECTED_TYPE})"
            echo -e "  make deps"
            echo -e "  claude"
        fi
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

# Setup worktree after creation (copy speckit files, create spec folder, etc.)
setup_worktree() {
    local worktree_path="$1"
    local issue_number="$2"
    local description="$3"

    echo -e "${BLUE}Setting up worktree...${NC}"

    # Copy speckit configuration files if they exist
    if [ -f "${MAIN_REPO}/.speckit.yaml" ]; then
        cp "${MAIN_REPO}/.speckit.yaml" "${worktree_path}/"
        echo -e "  ${GREEN}✓${NC} Copied .speckit.yaml"
    fi

    if [ -f "${MAIN_REPO}/.speckit.yml" ]; then
        cp "${MAIN_REPO}/.speckit.yml" "${worktree_path}/"
        echo -e "  ${GREEN}✓${NC} Copied .speckit.yml"
    fi

    # Copy speckit templates if they exist
    if [ -d "${MAIN_REPO}/.speckit" ]; then
        cp -r "${MAIN_REPO}/.speckit" "${worktree_path}/"
        echo -e "  ${GREEN}✓${NC} Copied .speckit/ templates"
    fi

    # Copy CLAUDE.md if it exists
    if [ -f "${MAIN_REPO}/CLAUDE.md" ]; then
        cp "${MAIN_REPO}/CLAUDE.md" "${worktree_path}/"
        echo -e "  ${GREEN}✓${NC} Copied CLAUDE.md"
    fi

    # Copy .claude/commands/ if it exists (speckit commands, etc.)
    if [ -d "${MAIN_REPO}/.claude/commands" ]; then
        mkdir -p "${worktree_path}/.claude"
        cp -r "${MAIN_REPO}/.claude/commands" "${worktree_path}/.claude/"
        echo -e "  ${GREEN}✓${NC} Copied .claude/commands/"
    fi

    # Copy .specify/ if it exists (speckit scripts and templates)
    if [ -d "${MAIN_REPO}/.specify" ]; then
        cp -r "${MAIN_REPO}/.specify" "${worktree_path}/"
        echo -e "  ${GREEN}✓${NC} Copied .specify/"
    fi

    # Copy .lazycurl/ if it exists (workspace: collections, environments, config)
    if [ -d "${MAIN_REPO}/.lazycurl" ]; then
        cp -r "${MAIN_REPO}/.lazycurl" "${worktree_path}/"
        echo -e "  ${GREEN}✓${NC} Copied .lazycurl/"
    fi

    # Create spec folder for this issue
    local spec_folder="${worktree_path}/specs/0${issue_number}-${description}"
    if [ ! -d "$spec_folder" ]; then
        mkdir -p "${spec_folder}/checklists"
        mkdir -p "${spec_folder}/contracts"
        echo -e "  ${GREEN}✓${NC} Created specs/0${issue_number}-${description}/"
    fi

    # Run make deps if Makefile exists
    if [ -f "${worktree_path}/Makefile" ]; then
        echo -e "${BLUE}Running make deps...${NC}"
        (cd "${worktree_path}" && make deps 2>/dev/null)
        echo -e "  ${GREEN}✓${NC} Dependencies installed"
    fi

    echo ""
}

# Launch Claude in worktree
launch_claude() {
    local worktree_path="$1"

    read -p "Launch Claude Code in worktree? [Y/n]: " launch
    if [[ ! "$launch" =~ ^[Nn]$ ]]; then
        echo -e "${CYAN}Launching Claude Code...${NC}"
        cd "${worktree_path}" && claude
    fi
}

# Quick create mode (non-interactive)
quick_create() {
    local type="$1"
    local issue="$2"
    local desc="$3"
    local auto_claude="${4:-}"

    if [[ -z "$type" ]] || [[ -z "$issue" ]] || [[ -z "$desc" ]]; then
        echo -e "${RED}Usage: $0 create <type> <issue-number> <description> [--claude]${NC}"
        echo -e "${YELLOW}Example: $0 create feat 123 user-authentication${NC}"
        echo -e "${YELLOW}Example: $0 create feat 123 user-authentication --claude${NC}"
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
    local clean_desc=$(echo "$desc" | tr '[:upper:]' '[:lower:]' | tr ' ' '-' | sed 's/[^a-z0-9-]//g')
    BRANCH_NAME="${type}/#${issue}-${clean_desc}"

    echo -e "${YELLOW}Creating worktree with gwq...${NC}"
    echo -e "  Branch: ${GREEN}${BRANCH_NAME}${NC}"

    gwq add -b "$BRANCH_NAME"

    if [ $? -eq 0 ]; then
        echo ""
        echo -e "${GREEN}${BOLD}Worktree created!${NC}"

        # Get worktree path
        local worktree_path=$(gwq get "${type}/#${issue}" 2>/dev/null || gwq get "${type}" 2>/dev/null)

        if [ -n "$worktree_path" ] && [ -d "$worktree_path" ]; then
            setup_worktree "$worktree_path" "$issue" "$clean_desc"

            if [ "$auto_claude" == "--claude" ]; then
                echo -e "${CYAN}Launching Claude Code...${NC}"
                cd "${worktree_path}" && claude
            else
                launch_claude "$worktree_path"
            fi
        else
            echo -e "${YELLOW}Next: cd \$(gwq get ${type}) && make deps && claude${NC}"
        fi
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
