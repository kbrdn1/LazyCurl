package ui

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/kbrdn1/LazyCurl/internal/api"
	"github.com/kbrdn1/LazyCurl/internal/config"
	"github.com/kbrdn1/LazyCurl/internal/ui/components"
	"github.com/kbrdn1/LazyCurl/pkg/styles"
)

// EnvNodeType represents the type of environment tree node
type EnvNodeType int

const (
	EnvNode EnvNodeType = iota
	VarNode
)

// EnvTreeNode represents a node in the environment tree
type EnvTreeNode struct {
	Name     string
	Type     EnvNodeType
	Variable *api.EnvironmentVariable // For VarNode
	Expanded bool                     // Only for EnvNode
	Children []*EnvTreeNode
	Parent   *EnvTreeNode
	EnvFile  *api.EnvironmentFile // Reference to source environment
}

// EnvClipboard holds copied environment data
type EnvClipboard struct {
	Type    EnvNodeType
	Name    string
	EnvFile *api.EnvironmentFile     // For EnvNode
	VarData *api.EnvironmentVariable // For VarNode
}

// EnvironmentsView represents the environments panel
type EnvironmentsView struct {
	workspacePath    string
	environmentsPath string
	environments     []*api.EnvironmentFile
	tree             []*EnvTreeNode
	visible          []*EnvTreeNode
	cursor           int
	scrollOffset     int
	height           int
	activeEnvName    string // Currently active environment
	clipboard        *EnvClipboard

	// Search
	search      *components.SearchInput
	searchQuery string

	// Modals
	deleteModal *components.Modal
	newVarModal *components.Modal
	newEnvModal *components.Modal
	editModal   *components.Modal
	renameModal *components.Modal
	pendingNode *EnvTreeNode // Node being acted upon
}

// NewEnvironmentsView creates a new environments view
func NewEnvironmentsView(workspacePath string) *EnvironmentsView {
	ev := &EnvironmentsView{
		workspacePath:    workspacePath,
		environmentsPath: filepath.Join(workspacePath, ".lazycurl", "environments"),
		cursor:           0,
		scrollOffset:     0,
		activeEnvName:    "",
		search:           components.NewSearchInput(),
	}

	// Initialize modals
	ev.deleteModal = components.NewConfirmModal("Delete", "", "delete")
	ev.newVarModal = components.NewFormModal("New Variable", "new_var", []components.FormField{
		{Name: "name", Label: "Name", Type: "text", Placeholder: "variable_name"},
		{Name: "value", Label: "Value", Type: "text", Placeholder: "value"},
		{Name: "secret", Label: "Secret", Type: "checkbox", Value: "false"},
		{Name: "active", Label: "Active", Type: "checkbox", Value: "true"},
	})
	ev.newEnvModal = components.NewFormModal("New Environment", "new_env", []components.FormField{
		{Name: "name", Label: "Name", Type: "text", Placeholder: "environment_name"},
		{Name: "description", Label: "Description", Type: "text", Placeholder: "optional description"},
	})
	ev.editModal = components.NewFormModal("Edit Value", "edit", []components.FormField{
		{Name: "value", Label: "Value", Type: "text"},
		{Name: "secret", Label: "Secret", Type: "checkbox"},
		{Name: "active", Label: "Active", Type: "checkbox"},
	})
	ev.renameModal = components.NewInputModal("Rename", "New Name", "", "rename")

	ev.loadEnvironments()

	return ev
}

// loadEnvironments loads environments from the workspace path
func (e *EnvironmentsView) loadEnvironments() {
	envs, err := api.LoadAllEnvironments(e.environmentsPath)
	if err != nil {
		e.environments = []*api.EnvironmentFile{}
		e.tree = []*EnvTreeNode{}
		e.visible = []*EnvTreeNode{}
		return
	}

	e.environments = envs
	e.buildTree()
	e.refresh()

	// Set first environment as active by default
	if len(e.environments) > 0 && e.activeEnvName == "" {
		e.activeEnvName = e.environments[0].Name
	}
}

// buildTree builds the tree structure from environments
func (e *EnvironmentsView) buildTree() {
	// Preserve expanded state from old tree
	expandedEnvs := make(map[string]bool)
	for _, node := range e.tree {
		if node.Type == EnvNode {
			expandedEnvs[node.EnvFile.FilePath] = node.Expanded
		}
	}

	e.tree = make([]*EnvTreeNode, 0, len(e.environments))

	for _, env := range e.environments {
		// Restore expanded state if it existed
		expanded := expandedEnvs[env.FilePath]

		envNode := &EnvTreeNode{
			Name:     env.Name,
			Type:     EnvNode,
			Expanded: expanded,
			EnvFile:  env,
			Children: make([]*EnvTreeNode, 0),
		}

		// Sort variable names for consistent display
		varNames := make([]string, 0, len(env.Variables))
		for name := range env.Variables {
			varNames = append(varNames, name)
		}
		sort.Strings(varNames)

		// Create child nodes for each variable
		for _, name := range varNames {
			variable := env.Variables[name]
			varNode := &EnvTreeNode{
				Name:     name,
				Type:     VarNode,
				Variable: variable,
				Parent:   envNode,
				EnvFile:  env,
			}
			envNode.Children = append(envNode.Children, varNode)
		}

		e.tree = append(e.tree, envNode)
	}
}

// refresh rebuilds the visible list
func (e *EnvironmentsView) refresh() {
	e.visible = make([]*EnvTreeNode, 0)

	for _, node := range e.tree {
		e.flattenNode(node)
	}

	// Ensure cursor is within bounds
	if e.cursor >= len(e.visible) {
		e.cursor = len(e.visible) - 1
	}
	if e.cursor < 0 {
		e.cursor = 0
	}
}

// flattenNode recursively adds visible nodes to the list
func (e *EnvironmentsView) flattenNode(node *EnvTreeNode) {
	// If searching, check if this node or any child matches
	if e.searchQuery != "" {
		if !e.nodeMatchesSearch(node) {
			return
		}
	}

	e.visible = append(e.visible, node)
	if (node.Expanded || e.searchQuery != "") && node.Type == EnvNode {
		// When searching, show all matching children regardless of expanded state
		for _, child := range node.Children {
			e.flattenNode(child)
		}
	}
}

// nodeMatchesSearch checks if node or any child matches the search query
func (e *EnvironmentsView) nodeMatchesSearch(node *EnvTreeNode) bool {
	// Check if this node matches
	if components.MatchesQuery(node.Name, e.searchQuery) {
		return true
	}

	// For EnvNode, check if any child variable matches
	if node.Type == EnvNode {
		for _, child := range node.Children {
			if e.nodeMatchesSearch(child) {
				return true
			}
		}
	}

	return false
}

// scrollIntoView ensures cursor is visible
func (e *EnvironmentsView) scrollIntoView() {
	if e.cursor < e.scrollOffset {
		e.scrollOffset = e.cursor
	}
	if e.height > 0 && e.cursor >= e.scrollOffset+e.height {
		e.scrollOffset = e.cursor - e.height + 1
	}
}

// getCurrentNode returns the currently selected node
func (e *EnvironmentsView) getCurrentNode() *EnvTreeNode {
	if e.cursor >= 0 && e.cursor < len(e.visible) {
		return e.visible[e.cursor]
	}
	return nil
}

// getEnvForNode returns the environment file for a node
func (e *EnvironmentsView) getEnvForNode(node *EnvTreeNode) *api.EnvironmentFile {
	if node == nil {
		return nil
	}
	return node.EnvFile
}

// envNameExists checks if an environment with the given name already exists
func (e *EnvironmentsView) envNameExists(name string) bool {
	for _, env := range e.environments {
		if env.Name == name {
			return true
		}
	}
	return false
}

// saveEnvironment saves an environment to disk
func (e *EnvironmentsView) saveEnvironment(env *api.EnvironmentFile) error {
	if env.FilePath == "" {
		env.FilePath = filepath.Join(e.environmentsPath, strings.ToLower(strings.ReplaceAll(env.Name, " ", "-"))+".json")
	}
	return api.SaveEnvironment(env, env.FilePath)
}

// hasActiveModal returns true if any modal is visible
func (e *EnvironmentsView) hasActiveModal() bool {
	return e.deleteModal.IsVisible() ||
		e.newVarModal.IsVisible() ||
		e.newEnvModal.IsVisible() ||
		e.editModal.IsVisible() ||
		e.renameModal.IsVisible()
}

// IsSearching returns true if search is active
func (e *EnvironmentsView) IsSearching() bool {
	return e.search.IsVisible()
}

// moveToFirstMatch moves cursor to the first node that directly matches the search query
func (e *EnvironmentsView) moveToFirstMatch() {
	if e.searchQuery == "" {
		return
	}
	for i, node := range e.visible {
		if components.MatchesQuery(node.Name, e.searchQuery) {
			e.cursor = i
			e.scrollIntoView()
			return
		}
	}
}

// nextMatch moves cursor to the next matching node
func (e *EnvironmentsView) nextMatch() {
	if e.searchQuery == "" || len(e.visible) == 0 {
		return
	}
	for i := 1; i <= len(e.visible); i++ {
		idx := (e.cursor + i) % len(e.visible)
		if components.MatchesQuery(e.visible[idx].Name, e.searchQuery) {
			e.cursor = idx
			e.scrollIntoView()
			return
		}
	}
}

// prevMatch moves cursor to the previous matching node
func (e *EnvironmentsView) prevMatch() {
	if e.searchQuery == "" || len(e.visible) == 0 {
		return
	}
	for i := 1; i <= len(e.visible); i++ {
		idx := (e.cursor - i + len(e.visible)) % len(e.visible)
		if components.MatchesQuery(e.visible[idx].Name, e.searchQuery) {
			e.cursor = idx
			e.scrollIntoView()
			return
		}
	}
}

// HasSearchQuery returns true if there's an active search query (not input visible)
func (e *EnvironmentsView) HasSearchQuery() bool {
	return e.searchQuery != "" && !e.search.IsVisible()
}

// Update handles messages for the environments view
func (e EnvironmentsView) Update(msg tea.Msg, cfg *config.GlobalConfig) (EnvironmentsView, tea.Cmd) {
	// Handle search messages first (they come from the search input component)
	switch msg := msg.(type) {
	case components.SearchUpdateMsg:
		e.searchQuery = msg.Query
		e.refresh()
		// Move cursor to first matching node
		e.moveToFirstMatch()
		return e, nil

	case components.SearchCloseMsg:
		if msg.Canceled {
			e.searchQuery = ""
			e.refresh()
		}
		return e, nil
	}

	// Handle search input when visible
	if e.search.IsVisible() {
		var cmd tea.Cmd
		e.search, cmd = e.search.Update(msg)
		return e, cmd
	}

	// Handle modal updates first - capture commands to get ModalCloseMsg
	var cmd tea.Cmd
	if e.deleteModal.IsVisible() {
		e.deleteModal, cmd = e.deleteModal.Update(msg)
		if cmd != nil {
			// Execute the command to get ModalCloseMsg
			closeMsg := cmd()
			if closeMsg, ok := closeMsg.(components.ModalCloseMsg); ok {
				return e.handleModalClose(closeMsg)
			}
		}
	}
	if e.newVarModal.IsVisible() {
		e.newVarModal, cmd = e.newVarModal.Update(msg)
		if cmd != nil {
			closeMsg := cmd()
			if closeMsg, ok := closeMsg.(components.ModalCloseMsg); ok {
				return e.handleModalClose(closeMsg)
			}
		}
	}
	if e.newEnvModal.IsVisible() {
		e.newEnvModal, cmd = e.newEnvModal.Update(msg)
		if cmd != nil {
			closeMsg := cmd()
			if closeMsg, ok := closeMsg.(components.ModalCloseMsg); ok {
				return e.handleModalClose(closeMsg)
			}
		}
	}
	if e.editModal.IsVisible() {
		e.editModal, cmd = e.editModal.Update(msg)
		if cmd != nil {
			closeMsg := cmd()
			if closeMsg, ok := closeMsg.(components.ModalCloseMsg); ok {
				return e.handleModalClose(closeMsg)
			}
		}
	}
	if e.renameModal.IsVisible() {
		e.renameModal, cmd = e.renameModal.Update(msg)
		if cmd != nil {
			closeMsg := cmd()
			if closeMsg, ok := closeMsg.(components.ModalCloseMsg); ok {
				return e.handleModalClose(closeMsg)
			}
		}
	}

	switch msg := msg.(type) {
	case components.ModalCloseMsg:
		return e.handleModalClose(msg)

	case tea.KeyMsg:
		// If modal is active, don't process other keys
		if e.hasActiveModal() {
			return e, nil
		}

		switch msg.String() {
		case "j", "down":
			if e.cursor < len(e.visible)-1 {
				e.cursor++
				e.scrollIntoView()
			}
		case "k", "up":
			if e.cursor > 0 {
				e.cursor--
				e.scrollIntoView()
			}
		case "l", "right", " ":
			// Expand environment
			if node := e.getCurrentNode(); node != nil {
				if node.Type == EnvNode && !node.Expanded {
					node.Expanded = true
					e.refresh()
				}
			}
		case "h", "left":
			// Collapse environment or go to parent
			if node := e.getCurrentNode(); node != nil {
				if node.Type == EnvNode && node.Expanded {
					node.Expanded = false
					e.refresh()
				} else if node.Type == VarNode && node.Parent != nil {
					// Go to parent environment
					for i, n := range e.visible {
						if n == node.Parent {
							e.cursor = i
							e.scrollIntoView()
							break
						}
					}
				}
			}

		case "s":
			// Toggle secret
			if node := e.getCurrentNode(); node != nil && node.Type == VarNode {
				env := e.getEnvForNode(node)
				if env != nil {
					env.ToggleVariableSecret(node.Name)
					_ = e.saveEnvironment(env) // Error intentionally ignored for UI responsiveness
				}
			}

		case "a", "A":
			// Toggle active for variable, or select env
			if node := e.getCurrentNode(); node != nil {
				if node.Type == VarNode {
					env := e.getEnvForNode(node)
					if env != nil {
						env.ToggleVariableActive(node.Name)
						_ = e.saveEnvironment(env) // Error intentionally ignored for UI responsiveness
					}
				} else if node.Type == EnvNode {
					e.activeEnvName = node.Name
				}
			}

		case "S":
			// Select environment
			if node := e.getCurrentNode(); node != nil {
				if node.Type == EnvNode {
					e.activeEnvName = node.Name
				} else if node.Parent != nil {
					e.activeEnvName = node.Parent.Name
				}
			}

		case "enter":
			// Set as active environment
			if node := e.getCurrentNode(); node != nil {
				if node.Type == EnvNode {
					e.activeEnvName = node.Name
				} else if node.Parent != nil {
					e.activeEnvName = node.Parent.Name
				}
			}

		case "c", "i":
			// In search mode: "i" reopens search input
			if e.HasSearchQuery() {
				e.search.Show()
				return e, nil
			}
			// Edit value
			if node := e.getCurrentNode(); node != nil && node.Type == VarNode {
				e.pendingNode = node
				e.editModal.SetFieldValue("value", node.Variable.Value)
				if node.Variable.Secret {
					e.editModal.SetFieldValue("secret", "true")
				} else {
					e.editModal.SetFieldValue("secret", "false")
				}
				if node.Variable.Active {
					e.editModal.SetFieldValue("active", "true")
				} else {
					e.editModal.SetFieldValue("active", "false")
				}
				e.editModal.Title = "Edit: " + node.Name
				e.editModal.Show()
			}

		case "R":
			// Rename
			if node := e.getCurrentNode(); node != nil {
				e.pendingNode = node
				e.renameModal.SetFieldValue("input", node.Name)
				if node.Type == EnvNode {
					e.renameModal.Title = "Rename Environment"
				} else {
					e.renameModal.Title = "Rename Variable"
				}
				e.renameModal.Show()
			}

		case "d":
			// Delete
			if node := e.getCurrentNode(); node != nil {
				e.pendingNode = node
				if node.Type == EnvNode {
					e.deleteModal.Message = "Delete environment: " + node.Name + "?"
				} else {
					path := node.Parent.Name + "/" + node.Name
					e.deleteModal.Message = "Delete variable: " + path + "?"
				}
				e.deleteModal.Show()
			}

		case "D":
			// Duplicate
			if node := e.getCurrentNode(); node != nil {
				if node.Type == EnvNode {
					// Duplicate environment
					if node.EnvFile != nil {
						newEnv := node.EnvFile.Clone()
						newName := node.Name + "_copy"
						// Check for unique name
						counter := 1
						for e.envNameExists(newName) {
							counter++
							newName = node.Name + "_copy" + fmt.Sprintf("%d", counter)
						}
						newEnv.Name = newName
						// Generate new file path
						envDir := filepath.Join(e.workspacePath, ".lazycurl", "environments")
						newFilePath := filepath.Join(envDir, newName+".json")
						newEnv.FilePath = newFilePath
						if err := api.SaveEnvironment(newEnv, newFilePath); err == nil {
							e.loadEnvironments()
						}
					}
				} else if node.Type == VarNode && node.Parent != nil {
					// Duplicate variable
					targetEnv := e.getEnvForNode(node)
					if targetEnv != nil && node.Variable != nil {
						newName := node.Name + "_copy"
						counter := 1
						for {
							if _, exists := targetEnv.Variables[newName]; !exists {
								break
							}
							counter++
							newName = node.Name + "_copy" + fmt.Sprintf("%d", counter)
						}
						targetEnv.Variables[newName] = &api.EnvironmentVariable{
							Value:  node.Variable.Value,
							Secret: node.Variable.Secret,
							Active: node.Variable.Active,
						}
						if err := api.SaveEnvironment(targetEnv, targetEnv.FilePath); err == nil {
							e.loadEnvironments()
						}
					}
				}
			}

		case "y":
			// Yank (copy)
			if node := e.getCurrentNode(); node != nil {
				e.clipboard = &EnvClipboard{
					Type: node.Type,
					Name: node.Name,
				}
				if node.Type == EnvNode {
					e.clipboard.EnvFile = node.EnvFile.Clone()
				} else {
					e.clipboard.VarData = &api.EnvironmentVariable{
						Value:  node.Variable.Value,
						Secret: node.Variable.Secret,
						Active: node.Variable.Active,
					}
				}
			}

		case "p":
			// Paste
			if e.clipboard != nil {
				if node := e.getCurrentNode(); node != nil {
					targetEnv := e.getEnvForNode(node)
					if targetEnv == nil {
						break
					}

					if e.clipboard.Type == VarNode && e.clipboard.VarData != nil {
						// Paste variable into current env
						newName := e.clipboard.Name
						// Check for duplicates and add suffix
						i := 1
						for targetEnv.HasVariable(newName) {
							newName = e.clipboard.Name + "_copy"
							if i > 1 {
								newName = e.clipboard.Name + "_copy" + string(rune('0'+i))
							}
							i++
						}
						targetEnv.SetVariableFull(newName, &api.EnvironmentVariable{
							Value:  e.clipboard.VarData.Value,
							Secret: e.clipboard.VarData.Secret,
							Active: e.clipboard.VarData.Active,
						})
						_ = e.saveEnvironment(targetEnv) // Error intentionally ignored for UI responsiveness
						e.buildTree()
						e.refresh()
					}
				}
			}

		case "n":
			// In search mode: next match, otherwise: new variable
			if e.HasSearchQuery() {
				e.nextMatch()
				return e, nil
			}
			if node := e.getCurrentNode(); node != nil {
				e.pendingNode = node
				// Reset form
				e.newVarModal.SetFieldValue("name", "")
				e.newVarModal.SetFieldValue("value", "")
				e.newVarModal.SetFieldValue("secret", "false")
				e.newVarModal.SetFieldValue("active", "true")
				e.newVarModal.Show()
			}

		case "N":
			// In search mode: previous match, otherwise: new environment
			if e.HasSearchQuery() {
				e.prevMatch()
				return e, nil
			}
			e.pendingNode = nil
			e.newEnvModal.SetFieldValue("name", "")
			e.newEnvModal.SetFieldValue("description", "")
			e.newEnvModal.Show()

		case "g":
			e.cursor = 0
			e.scrollIntoView()
		case "G":
			if len(e.visible) > 0 {
				e.cursor = len(e.visible) - 1
				e.scrollIntoView()
			}
		case "/":
			// Open search
			e.search.Show()
			return e, nil
		case "esc":
			// Clear search filter if active
			if e.searchQuery != "" {
				e.searchQuery = ""
				e.refresh()
				return e, nil
			}
		}
	}

	return e, nil
}

// handleModalClose handles modal close events
func (e EnvironmentsView) handleModalClose(msg components.ModalCloseMsg) (EnvironmentsView, tea.Cmd) {
	if !msg.Result.Confirmed {
		e.pendingNode = nil
		return e, nil
	}

	switch msg.Tag {
	case "delete":
		if e.pendingNode != nil {
			if e.pendingNode.Type == EnvNode {
				// Delete environment file
				if e.pendingNode.EnvFile.FilePath != "" {
					// Would need os.Remove here
				}
				// Remove from list
				for i, env := range e.environments {
					if env == e.pendingNode.EnvFile {
						e.environments = append(e.environments[:i], e.environments[i+1:]...)
						break
					}
				}
			} else {
				// Delete variable
				env := e.getEnvForNode(e.pendingNode)
				if env != nil {
					env.DeleteVariable(e.pendingNode.Name)
					_ = e.saveEnvironment(env) // Error intentionally ignored for UI responsiveness
				}
			}
			e.buildTree()
			e.refresh()
		}

	case "edit":
		if e.pendingNode != nil && e.pendingNode.Type == VarNode {
			env := e.getEnvForNode(e.pendingNode)
			if env != nil {
				e.pendingNode.Variable.Value = msg.Result.Values["value"].(string)
				e.pendingNode.Variable.Secret = msg.Result.Values["secret"].(bool)
				e.pendingNode.Variable.Active = msg.Result.Values["active"].(bool)
				_ = e.saveEnvironment(env) // Error intentionally ignored for UI responsiveness
			}
		}

	case "rename":
		if e.pendingNode != nil {
			newName := msg.Result.Values["input"].(string)
			if newName != "" && newName != e.pendingNode.Name {
				env := e.getEnvForNode(e.pendingNode)
				if e.pendingNode.Type == EnvNode {
					env.Name = newName
					_ = e.saveEnvironment(env) // Error intentionally ignored for UI responsiveness
				} else if env != nil {
					// Rename variable
					v := env.Variables[e.pendingNode.Name]
					delete(env.Variables, e.pendingNode.Name)
					env.Variables[newName] = v
					_ = e.saveEnvironment(env) // Error intentionally ignored for UI responsiveness
				}
				e.buildTree()
				e.refresh()
			}
		}

	case "new_var":
		name := msg.Result.Values["name"].(string)
		value := msg.Result.Values["value"].(string)
		secret := msg.Result.Values["secret"].(bool)
		active := msg.Result.Values["active"].(bool)

		if name != "" && e.pendingNode != nil {
			var targetEnv *api.EnvironmentFile
			if e.pendingNode.Type == EnvNode {
				targetEnv = e.pendingNode.EnvFile
			} else {
				targetEnv = e.pendingNode.EnvFile
			}

			if targetEnv != nil {
				targetEnv.SetVariableFull(name, &api.EnvironmentVariable{
					Value:  value,
					Secret: secret,
					Active: active,
				})
				_ = e.saveEnvironment(targetEnv) // Error intentionally ignored for UI responsiveness
				e.buildTree()
				e.refresh()
			}
		}

	case "new_env":
		name := msg.Result.Values["name"].(string)
		desc := msg.Result.Values["description"].(string)

		if name != "" {
			newEnv := &api.EnvironmentFile{
				Name:        name,
				Description: desc,
				Variables:   make(map[string]*api.EnvironmentVariable),
			}
			e.environments = append(e.environments, newEnv)
			_ = e.saveEnvironment(newEnv) // Error intentionally ignored for UI responsiveness
			e.buildTree()
			e.refresh()
		}
	}

	e.pendingNode = nil
	return e, nil
}

// View renders the environments view
func (e EnvironmentsView) View(width, height int, active bool) string {
	var output []string

	// Count matches for search display
	matchCount := 0
	totalCount := e.countAllNodes()
	if e.searchQuery != "" {
		matchCount = e.countDirectMatches()
	}

	// Render search box if visible
	if e.search.IsVisible() {
		searchBox := e.search.ViewCompact(width, matchCount, totalCount)
		output = append(output, searchBox)
		height -= lipgloss.Height(searchBox) + 1
	} else if e.searchQuery != "" {
		// Show compact filter indicator with count
		filterStyle := lipgloss.NewStyle().
			Foreground(styles.Yellow)
		countStyle := lipgloss.NewStyle().
			Foreground(styles.Subtext0)
		escStyle := lipgloss.NewStyle().
			Foreground(styles.Subtext0).
			Italic(true)
		filterText := filterStyle.Render("/"+e.searchQuery) + countStyle.Render(fmt.Sprintf(" %d/%d", matchCount, totalCount)) + escStyle.Render(" esc")
		output = append(output, filterText)
		height--
	}

	e.height = height

	if len(e.visible) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(styles.Subtext0).
			Width(width).
			Align(lipgloss.Center)
		if e.searchQuery != "" {
			output = append(output, emptyStyle.Render("No matches found"))
		} else {
			output = append(output, emptyStyle.Render("No environments found\n\nPress N to create one\n\n.lazycurl/environments/"))
		}
		return strings.Join(output, "\n")
	}

	var lines []string
	start := e.scrollOffset
	end := e.scrollOffset + height
	if end > len(e.visible) {
		end = len(e.visible)
	}

	for i := start; i < end && i < len(e.visible); i++ {
		node := e.visible[i]
		line := e.renderNode(node, width, i == e.cursor, active)
		lines = append(lines, line)
	}

	output = append(output, strings.Join(lines, "\n"))
	return strings.Join(output, "\n")
}

// countAllNodes counts total nodes in tree
func (e *EnvironmentsView) countAllNodes() int {
	count := 0
	for _, node := range e.tree {
		count++                     // env node
		count += len(node.Children) // variable nodes
	}
	return count
}

// countDirectMatches counts nodes that directly match the search query
func (e *EnvironmentsView) countDirectMatches() int {
	if e.searchQuery == "" {
		return 0
	}
	count := 0
	for _, node := range e.tree {
		if components.MatchesQuery(node.Name, e.searchQuery) {
			count++
		}
		for _, child := range node.Children {
			if components.MatchesQuery(child.Name, e.searchQuery) {
				count++
			}
		}
	}
	return count
}

// renderNode renders a single tree node with worktree style
func (e *EnvironmentsView) renderNode(node *EnvTreeNode, width int, selected bool, panelActive bool) string {
	var content string

	// Check if this node directly matches the search query
	isDirectMatch := e.searchQuery != "" && components.MatchesQuery(node.Name, e.searchQuery)
	isSearching := e.searchQuery != ""

	switch node.Type {
	case EnvNode:
		// Environment node: ▶/▼ EnvName ●
		icon := "▶ "
		if node.Expanded {
			icon = "▼ "
		}

		// Active indicator
		activeIndicator := ""
		if node.Name == e.activeEnvName {
			activeIndicator = " ●"
		}

		// Apply search styling
		iconStyle := lipgloss.NewStyle()
		nameStyle := lipgloss.NewStyle()
		if isSearching {
			if isDirectMatch {
				iconStyle = iconStyle.Foreground(styles.SearchMatch)
				nameStyle = nameStyle.Foreground(styles.SearchMatch).Bold(true)
			} else {
				iconStyle = iconStyle.Foreground(styles.SearchDimmed)
				nameStyle = nameStyle.Foreground(styles.SearchDimmed)
			}
		}

		content = iconStyle.Render(icon) + nameStyle.Render(node.Name+activeIndicator)

	case VarNode:
		// Worktree style: > []  value_name   value
		// Checkbox for active state (Unicode squares)
		checkbox := "☐"
		checkStyle := lipgloss.NewStyle().Foreground(styles.CheckboxOff)
		if node.Variable.Active {
			checkbox = "☑"
			checkStyle = checkStyle.Foreground(styles.CheckboxOn)
		}

		// Key name
		key := node.Name
		keyStyle := lipgloss.NewStyle().Foreground(styles.Subtext1)

		// Value (masked if secret)
		value := node.Variable.Value
		valueStyle := lipgloss.NewStyle().Foreground(styles.Text)

		if node.Variable.Secret {
			valueStyle = valueStyle.Foreground(styles.SecretColor)
			if len(value) > 0 {
				value = strings.Repeat("*", min(len(value), 10))
			} else {
				value = "***"
			}
		}

		if !node.Variable.Active {
			keyStyle = keyStyle.Foreground(styles.InactiveColor)
			valueStyle = valueStyle.Foreground(styles.InactiveColor)
			checkStyle = checkStyle.Foreground(styles.InactiveColor)
		}

		// Apply search styling (overrides other styles)
		if isSearching {
			if isDirectMatch {
				keyStyle = lipgloss.NewStyle().Foreground(styles.SearchMatch).Bold(true)
				valueStyle = lipgloss.NewStyle().Foreground(styles.SearchMatch)
				checkStyle = lipgloss.NewStyle().Foreground(styles.SearchMatch)
			} else {
				keyStyle = lipgloss.NewStyle().Foreground(styles.SearchDimmed)
				valueStyle = lipgloss.NewStyle().Foreground(styles.SearchDimmed)
				checkStyle = lipgloss.NewStyle().Foreground(styles.SearchDimmed)
			}
		}

		// Build prefix: "> " for selected or "  " for others
		linePrefix := "  "
		if selected {
			linePrefix = "> "
		}

		// Calculate spacing for worktree format: > []  key   value
		checkboxWidth := 3  // "[] " with space
		prefixWidth := 2    // "> " or "  "
		separatorWidth := 3 // "   " between key and value
		availableWidth := width - prefixWidth - checkboxWidth - separatorWidth
		if availableWidth < 10 {
			availableWidth = 10
		}

		// Key width (max 20, min 5)
		keyWidth := availableWidth / 2
		if keyWidth > 20 {
			keyWidth = 20
		}
		if keyWidth < 5 {
			keyWidth = 5
		}

		// Truncate key to fit (no ellipsis - just cut)
		if len(key) > keyWidth {
			key = key[:keyWidth]
		}
		// Pad key to align values
		keyPadded := key + strings.Repeat(" ", keyWidth-len(key))

		// Calculate remaining width for value
		valueWidth := availableWidth - keyWidth
		if valueWidth < 3 {
			valueWidth = 3
		}
		// Truncate value to fit (no ellipsis - just cut)
		if len(value) > valueWidth {
			value = value[:valueWidth]
		}

		content = linePrefix + checkStyle.Render(checkbox) + " " + keyStyle.Render(keyPadded) + "   " + valueStyle.Render(value)
	}

	// Apply selection styling
	style := lipgloss.NewStyle().Width(width)
	if selected {
		if panelActive {
			style = style.Background(styles.SelectedPanelBg).Foreground(styles.SelectedPanelFg).Bold(true)
		} else {
			style = style.Background(styles.SelectedRequestBg).Foreground(styles.SelectedRequestFg)
		}
	}
	// Don't override foreground if not selected - content already has correct colors

	return style.Render(content)
}

// RenderModal renders any active modal
func (e *EnvironmentsView) RenderModal(screenWidth, screenHeight int) string {
	if e.deleteModal.IsVisible() {
		return e.deleteModal.View(screenWidth, screenHeight)
	}
	if e.newVarModal.IsVisible() {
		return e.newVarModal.View(screenWidth, screenHeight)
	}
	if e.newEnvModal.IsVisible() {
		return e.newEnvModal.View(screenWidth, screenHeight)
	}
	if e.editModal.IsVisible() {
		return e.editModal.View(screenWidth, screenHeight)
	}
	if e.renameModal.IsVisible() {
		return e.renameModal.View(screenWidth, screenHeight)
	}
	return ""
}

// HasActiveModal returns true if any modal is visible
func (e *EnvironmentsView) HasActiveModal() bool {
	return e.hasActiveModal()
}

// GetActiveEnvironment returns the currently active environment
func (e *EnvironmentsView) GetActiveEnvironment() *api.EnvironmentFile {
	for _, env := range e.environments {
		if env.Name == e.activeEnvName {
			return env
		}
	}
	return nil
}

// GetActiveEnvironmentName returns the name of the active environment
func (e *EnvironmentsView) GetActiveEnvironmentName() string {
	return e.activeEnvName
}

// GetActiveEnvironmentVariables returns the variables of the active environment
func (e *EnvironmentsView) GetActiveEnvironmentVariables() map[string]string {
	env := e.GetActiveEnvironment()
	if env == nil {
		return make(map[string]string)
	}
	// Convert active variables to map
	vars := make(map[string]string)
	for key, v := range env.Variables {
		if v.Active {
			vars[key] = v.Value
		}
	}
	return vars
}

// GetBreadcrumb returns the breadcrumb path for the current cursor position
func (e *EnvironmentsView) GetBreadcrumb() []string {
	node := e.getCurrentNode()
	if node == nil {
		return []string{}
	}

	if node.Type == EnvNode {
		return []string{node.Name}
	}

	// VarNode - show environment > variable
	if node.Parent != nil {
		return []string{node.Parent.Name, node.Name}
	}
	return []string{node.Name}
}

// ReloadEnvironments reloads environments from disk
func (e *EnvironmentsView) ReloadEnvironments() {
	e.loadEnvironments()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
