package components

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/kbrdn1/LazyCurl/internal/api"
	"github.com/kbrdn1/LazyCurl/pkg/styles"
)

// NodeType represents the type of tree node
type NodeType int

const (
	CollectionNode NodeType = iota
	FolderNode
	RequestNode
)

// TreeNode represents a single node in the tree
type TreeNode struct {
	ID         string      // Unique identifier
	Name       string      // Display name
	Type       NodeType    // Node type
	Children   []*TreeNode // Child nodes (empty for requests)
	Expanded   bool        // Whether folder is expanded
	HTTPMethod string      // HTTP method (only for RequestNode)
	URL        string      // Request URL (only for RequestNode)
	Depth      int         // Nesting level (0 = root)
	Parent     *TreeNode   // Reference to parent node
}

// Tree is the main tree view component
type Tree struct {
	Root         []*TreeNode  // Top-level nodes
	cursor       int          // Current cursor position in visible list
	visible      []*TreeNode  // Flattened visible nodes
	selected     *TreeNode    // Currently selected node
	height       int          // Available height for rendering
	scrollOffset int          // Scroll position for tall trees
	search       *SearchInput // Search input
	searchQuery  string       // Current search filter
}

// TreeSelectionMsg is sent when a request is selected
type TreeSelectionMsg struct {
	Node *TreeNode
}

// TreeExpandMsg is sent when a folder is expanded/collapsed
type TreeExpandMsg struct {
	Node     *TreeNode
	Expanded bool
}

// TreeActionMsg is sent for tree manipulation actions
type TreeActionMsg struct {
	Action string    // "rename", "delete", "yank", "paste", "new_request", "new_folder", "duplicate"
	Node   *TreeNode // Target node
}

// TreeRenameMsg is sent when renaming is initiated
type TreeRenameMsg struct {
	Node *TreeNode
}

// TreeDeleteMsg is sent when delete is requested
type TreeDeleteMsg struct {
	Node *TreeNode
}

// TreeYankMsg is sent when a node is yanked (copied)
type TreeYankMsg struct {
	Node *TreeNode
}

// TreePasteMsg is sent when paste is requested
type TreePasteMsg struct {
	TargetNode *TreeNode // Where to paste
}

// TreeNewRequestMsg is sent to create a new request
type TreeNewRequestMsg struct {
	ParentNode *TreeNode // nil for root level
}

// TreeNewFolderMsg is sent to create a new folder
type TreeNewFolderMsg struct {
	ParentNode *TreeNode // nil for root level
}

// TreeDuplicateMsg is sent to duplicate a node
type TreeDuplicateMsg struct {
	Node *TreeNode
}

// TreeEditRequestMsg is sent to edit a request
type TreeEditRequestMsg struct {
	Node *TreeNode
}

// NewTree creates a new tree from collections
func NewTree(collections []*api.CollectionFile) *Tree {
	t := &Tree{
		Root:    buildTree(collections),
		cursor:  0,
		visible: make([]*TreeNode, 0),
		search:  NewSearchInput(),
	}
	t.Refresh()
	return t
}

// buildTree converts collections to TreeNode hierarchy
func buildTree(collections []*api.CollectionFile) []*TreeNode {
	var nodes []*TreeNode
	for i, col := range collections {
		node := &TreeNode{
			ID:       fmt.Sprintf("col_%d", i),
			Name:     col.Name,
			Type:     CollectionNode,
			Expanded: true, // Collections start expanded
			Depth:    0,
		}
		node.Children = buildFolders(col.Folders, 1, node)
		node.Children = append(node.Children, buildRequests(col.Requests, 1, node)...)
		nodes = append(nodes, node)
	}
	return nodes
}

// buildFolders converts folder structures to tree nodes
func buildFolders(folders []api.Folder, depth int, parent *TreeNode) []*TreeNode {
	var nodes []*TreeNode
	for i, f := range folders {
		node := &TreeNode{
			ID:       fmt.Sprintf("%s_folder_%d", parent.ID, i),
			Name:     f.Name,
			Type:     FolderNode,
			Expanded: false, // Folders start collapsed
			Depth:    depth,
			Parent:   parent,
		}
		node.Children = buildFolders(f.Folders, depth+1, node)
		node.Children = append(node.Children, buildRequests(f.Requests, depth+1, node)...)
		nodes = append(nodes, node)
	}
	return nodes
}

// buildRequests converts request structures to tree nodes
func buildRequests(requests []api.CollectionRequest, depth int, parent *TreeNode) []*TreeNode {
	var nodes []*TreeNode
	for _, r := range requests {
		nodes = append(nodes, &TreeNode{
			ID:         r.ID,
			Name:       r.Name,
			Type:       RequestNode,
			HTTPMethod: string(r.Method),
			URL:        r.URL,
			Depth:      depth,
			Parent:     parent,
		})
	}
	return nodes
}

// Refresh rebuilds visible list from current state
func (t *Tree) Refresh() {
	t.visible = make([]*TreeNode, 0)
	for _, node := range t.Root {
		t.flattenNode(node)
	}
	// Ensure cursor is within bounds
	if t.cursor >= len(t.visible) {
		t.cursor = len(t.visible) - 1
	}
	if t.cursor < 0 {
		t.cursor = 0
	}
	// Update selected node
	if len(t.visible) > 0 && t.cursor < len(t.visible) {
		t.selected = t.visible[t.cursor]
	} else {
		t.selected = nil
	}
}

// flattenNode recursively adds visible nodes to the list
func (t *Tree) flattenNode(node *TreeNode) {
	// If searching, check if this node or any descendant matches
	if t.searchQuery != "" {
		if !t.nodeMatchesSearch(node) {
			return
		}
	}

	t.visible = append(t.visible, node)
	if node.Expanded || t.searchQuery != "" {
		// When searching, show all matching descendants regardless of expanded state
		for _, child := range node.Children {
			t.flattenNode(child)
		}
	}
}

// nodeMatchesSearch checks if node or any descendant matches the search query
func (t *Tree) nodeMatchesSearch(node *TreeNode) bool {
	// Check if this node matches
	if MatchesQuery(node.Name, t.searchQuery) {
		return true
	}

	// Check if any child matches
	for _, child := range node.Children {
		if t.nodeMatchesSearch(child) {
			return true
		}
	}

	return false
}

// Selected returns the currently selected node
func (t *Tree) Selected() *TreeNode {
	return t.selected
}

// Up moves cursor up
func (t *Tree) Up() {
	if t.cursor > 0 {
		t.cursor--
		t.selected = t.visible[t.cursor]
		t.scrollIntoView()
	}
}

// Down moves cursor down
func (t *Tree) Down() {
	if t.cursor < len(t.visible)-1 {
		t.cursor++
		t.selected = t.visible[t.cursor]
		t.scrollIntoView()
	}
}

// Expand expands the selected folder
func (t *Tree) Expand() bool {
	if t.selected == nil {
		return false
	}
	if t.selected.Type == RequestNode {
		return false // Requests can't be expanded
	}
	if !t.selected.Expanded {
		t.selected.Expanded = true
		t.Refresh()
		return true
	}
	return false
}

// Collapse collapses the selected folder
func (t *Tree) Collapse() bool {
	if t.selected == nil {
		return false
	}
	if t.selected.Type == RequestNode {
		// For requests, go to parent
		if t.selected.Parent != nil {
			// Find parent in visible list
			for i, node := range t.visible {
				if node == t.selected.Parent {
					t.cursor = i
					t.selected = node
					t.scrollIntoView()
					return true
				}
			}
		}
		return false
	}
	if t.selected.Expanded {
		t.selected.Expanded = false
		t.Refresh()
		return true
	}
	// Already collapsed, go to parent
	if t.selected.Parent != nil {
		for i, node := range t.visible {
			if node == t.selected.Parent {
				t.cursor = i
				t.selected = node
				t.scrollIntoView()
				return true
			}
		}
	}
	return false
}

// GoToFirst jumps to the first item
func (t *Tree) GoToFirst() {
	if len(t.visible) > 0 {
		t.cursor = 0
		t.selected = t.visible[0]
		t.scrollIntoView()
	}
}

// GoToLast jumps to the last item
func (t *Tree) GoToLast() {
	if len(t.visible) > 0 {
		t.cursor = len(t.visible) - 1
		t.selected = t.visible[t.cursor]
		t.scrollIntoView()
	}
}

// scrollIntoView ensures cursor is visible
func (t *Tree) scrollIntoView() {
	if t.cursor < t.scrollOffset {
		t.scrollOffset = t.cursor
	}
	if t.height > 0 && t.cursor >= t.scrollOffset+t.height {
		t.scrollOffset = t.cursor - t.height + 1
	}
}

// VisibleRange returns the range of visible nodes
func (t *Tree) VisibleRange() (start, end int) {
	start = t.scrollOffset
	end = t.scrollOffset + t.height
	if end > len(t.visible) {
		end = len(t.visible)
	}
	return
}

// IsSearching returns true if search is active
func (t *Tree) IsSearching() bool {
	return t.search.IsVisible()
}

// moveToFirstMatch moves cursor to the first node that directly matches the search query
func (t *Tree) moveToFirstMatch() {
	if t.searchQuery == "" {
		return
	}
	for i, node := range t.visible {
		if MatchesQuery(node.Name, t.searchQuery) {
			t.cursor = i
			t.selected = node
			t.scrollIntoView()
			return
		}
	}
}

// nextMatch moves cursor to the next matching node
func (t *Tree) nextMatch() {
	if t.searchQuery == "" || len(t.visible) == 0 {
		return
	}
	// Start from cursor + 1, wrap around
	for i := 1; i <= len(t.visible); i++ {
		idx := (t.cursor + i) % len(t.visible)
		if MatchesQuery(t.visible[idx].Name, t.searchQuery) {
			t.cursor = idx
			t.selected = t.visible[idx]
			t.scrollIntoView()
			return
		}
	}
}

// prevMatch moves cursor to the previous matching node
func (t *Tree) prevMatch() {
	if t.searchQuery == "" || len(t.visible) == 0 {
		return
	}
	// Start from cursor - 1, wrap around
	for i := 1; i <= len(t.visible); i++ {
		idx := (t.cursor - i + len(t.visible)) % len(t.visible)
		if MatchesQuery(t.visible[idx].Name, t.searchQuery) {
			t.cursor = idx
			t.selected = t.visible[idx]
			t.scrollIntoView()
			return
		}
	}
}

// HasSearchQuery returns true if there's an active search query (not input visible)
func (t *Tree) HasSearchQuery() bool {
	return t.searchQuery != "" && !t.search.IsVisible()
}

// Update handles messages and returns updated tree
func (t *Tree) Update(msg tea.Msg, allowNavigation bool) (*Tree, tea.Cmd) {
	// Handle search messages first (they come from the search input component)
	switch msg := msg.(type) {
	case SearchUpdateMsg:
		t.searchQuery = msg.Query
		t.Refresh()
		// Move cursor to first matching node
		t.moveToFirstMatch()
		return t, nil

	case SearchCloseMsg:
		if msg.Canceled {
			t.searchQuery = ""
			t.Refresh()
		} else {
			// Search validated - cursor should already be on a match
			// Keep the searchQuery to maintain filter view
		}
		return t, nil
	}

	// Handle search input when visible
	if t.search.IsVisible() {
		var cmd tea.Cmd
		t.search, cmd = t.search.Update(msg)
		return t, cmd
	}

	if !allowNavigation {
		return t, nil
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			t.Down()
		case "k", "up":
			t.Up()
		case "l", "right":
			// Expand folder or select request
			if t.selected != nil {
				if t.selected.Type == RequestNode {
					// Send selection message
					return t, func() tea.Msg {
						return TreeSelectionMsg{Node: t.selected}
					}
				}
				t.Expand()
			}
		case "enter":
			// Enter always selects the current node
			if t.selected != nil {
				return t, func() tea.Msg {
					return TreeSelectionMsg{Node: t.selected}
				}
			}
		case "h", "left":
			t.Collapse()
		case "g":
			t.GoToFirst()
		case "G":
			t.GoToLast()
		case " ":
			// Space: toggle expansion for folders, open request for requests
			if t.selected != nil {
				if t.selected.Type == RequestNode {
					// Space opens the request
					return t, func() tea.Msg {
						return TreeSelectionMsg{Node: t.selected}
					}
				} else {
					// Toggle folder expansion
					if t.selected.Expanded {
						t.Collapse()
					} else {
						t.Expand()
					}
				}
			}

		// Action keys
		case "R":
			// Rename selected node
			if t.selected != nil {
				return t, func() tea.Msg {
					return TreeRenameMsg{Node: t.selected}
				}
			}
		case "d":
			// Delete selected node
			if t.selected != nil {
				return t, func() tea.Msg {
					return TreeDeleteMsg{Node: t.selected}
				}
			}
		case "y":
			// Yank (copy) selected node to clipboard
			if t.selected != nil {
				return t, func() tea.Msg {
					return TreeYankMsg{Node: t.selected}
				}
			}
		case "p":
			// Paste from clipboard to current location
			return t, func() tea.Msg {
				return TreePasteMsg{TargetNode: t.selected}
			}
		case "n":
			// In search mode: next match, otherwise: new request
			if t.HasSearchQuery() {
				t.nextMatch()
				return t, nil
			}
			return t, func() tea.Msg {
				return TreeNewRequestMsg{ParentNode: t.getParentFolder()}
			}
		case "N":
			// In search mode: previous match, otherwise: new folder
			if t.HasSearchQuery() {
				t.prevMatch()
				return t, nil
			}
			return t, func() tea.Msg {
				return TreeNewFolderMsg{ParentNode: t.getParentFolder()}
			}
		case "D":
			// Duplicate selected node
			if t.selected != nil {
				return t, func() tea.Msg {
					return TreeDuplicateMsg{Node: t.selected}
				}
			}
		case "c":
			// Edit request (only for RequestNode)
			if t.selected != nil && t.selected.Type == RequestNode {
				return t, func() tea.Msg {
					return TreeEditRequestMsg{Node: t.selected}
				}
			}
		case "i":
			// In search mode: reopen search input, otherwise: edit request
			if t.HasSearchQuery() {
				t.search.Show()
				return t, nil
			}
			if t.selected != nil && t.selected.Type == RequestNode {
				return t, func() tea.Msg {
					return TreeEditRequestMsg{Node: t.selected}
				}
			}
		case "/":
			// Open search
			t.search.Show()
			return t, nil
		case "esc":
			// Clear search filter if active
			if t.searchQuery != "" {
				t.searchQuery = ""
				t.Refresh()
				return t, nil
			}
		}
	}

	return t, nil
}

// getParentFolder returns the appropriate parent folder for new items
func (t *Tree) getParentFolder() *TreeNode {
	if t.selected == nil {
		return nil
	}
	// If selected is a folder/collection, use it as parent
	if t.selected.Type != RequestNode {
		return t.selected
	}
	// Otherwise use the parent of the selected request
	return t.selected.Parent
}

// View renders the tree to a string
func (t *Tree) View(width, height int, active bool) string {
	var output []string

	// Count matches for search display
	matchCount := 0
	totalCount := t.countAllNodes()
	if t.searchQuery != "" {
		matchCount = t.countDirectMatches()
	}

	// Render search box if visible
	if t.search.IsVisible() {
		searchBox := t.search.ViewCompact(width, matchCount, totalCount)
		output = append(output, searchBox)
		height -= lipgloss.Height(searchBox) + 1
	} else if t.searchQuery != "" {
		// Show compact filter indicator with count
		filterStyle := lipgloss.NewStyle().
			Foreground(styles.Yellow)
		countStyle := lipgloss.NewStyle().
			Foreground(styles.Subtext0)
		escStyle := lipgloss.NewStyle().
			Foreground(styles.Subtext0).
			Italic(true)
		filterText := filterStyle.Render("/"+t.searchQuery) + countStyle.Render(fmt.Sprintf(" %d/%d", matchCount, totalCount)) + escStyle.Render(" esc")
		output = append(output, filterText)
		height--
	}

	t.height = height

	if len(t.visible) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(styles.Subtext0).
			Width(width).
			Align(lipgloss.Center)
		if t.searchQuery != "" {
			output = append(output, emptyStyle.Render("No matches found"))
		} else {
			output = append(output, emptyStyle.Render("No collections found\n\nPress 'n' to create one\nor add files to:\n.lazycurl/collections/"))
		}
		return strings.Join(output, "\n")
	}

	var lines []string
	start, end := t.VisibleRange()

	for i := start; i < end && i < len(t.visible); i++ {
		node := t.visible[i]
		line := t.renderNode(node, width, i == t.cursor, active)
		lines = append(lines, line)
	}

	output = append(output, strings.Join(lines, "\n"))
	return strings.Join(output, "\n")
}

// countAllNodes counts total nodes in tree
func (t *Tree) countAllNodes() int {
	count := 0
	var countNodes func([]*TreeNode)
	countNodes = func(nodes []*TreeNode) {
		for _, node := range nodes {
			count++
			countNodes(node.Children)
		}
	}
	countNodes(t.Root)
	return count
}

// countDirectMatches counts nodes that directly match the search query
func (t *Tree) countDirectMatches() int {
	if t.searchQuery == "" {
		return 0
	}
	count := 0
	var countMatches func([]*TreeNode)
	countMatches = func(nodes []*TreeNode) {
		for _, node := range nodes {
			if MatchesQuery(node.Name, t.searchQuery) {
				count++
			}
			countMatches(node.Children)
		}
	}
	countMatches(t.Root)
	return count
}

// renderNode renders a single tree node
func (t *Tree) renderNode(node *TreeNode, width int, selected bool, panelActive bool) string {
	// Check if this node directly matches the search query
	isDirectMatch := t.searchQuery != "" && MatchesQuery(node.Name, t.searchQuery)
	isSearching := t.searchQuery != ""

	// Calculate indent with tree lines
	var prefix string
	prefixStyle := lipgloss.NewStyle()
	if isSearching && !isDirectMatch {
		prefixStyle = prefixStyle.Foreground(styles.SearchDimmed)
	} else {
		prefixStyle = prefixStyle.Foreground(styles.Subtext0)
	}

	if node.Depth > 0 {
		prefixChars := strings.Repeat("│ ", node.Depth-1)
		// Check if this is the last sibling
		if node.Parent != nil {
			isLast := true
			for i, sibling := range node.Parent.Children {
				if sibling == node && i < len(node.Parent.Children)-1 {
					isLast = false
					break
				}
			}
			if isLast {
				prefixChars += "└─"
			} else {
				prefixChars += "├─"
			}
		}
		prefix = prefixStyle.Render(prefixChars)
	}

	// Choose icon based on type and expansion state
	var icon string
	switch node.Type {
	case CollectionNode, FolderNode:
		if node.Expanded {
			icon = "▼ "
		} else {
			icon = "▶ "
		}
	case RequestNode:
		icon = ""
	}

	// Build the line content with search-aware styling
	var content string
	if node.Type == RequestNode {
		// Add method badge for requests
		methodBadge := t.renderMethodBadge(node.HTTPMethod, isSearching && !isDirectMatch)
		nameStyle := lipgloss.NewStyle()
		if isSearching {
			if isDirectMatch {
				nameStyle = nameStyle.Foreground(styles.SearchMatch).Bold(true)
			} else {
				nameStyle = nameStyle.Foreground(styles.SearchDimmed)
			}
		}
		// Calculate available width for name: width - prefix - method badge - spaces
		prefixLen := lipgloss.Width(prefix)
		methodLen := lipgloss.Width(methodBadge)
		availableNameWidth := width - prefixLen - methodLen - 2 // 2 spaces
		name := node.Name
		if availableNameWidth > 0 && len(name) > availableNameWidth {
			name = name[:availableNameWidth] // Truncate without ellipsis
		}
		content = fmt.Sprintf("%s %s %s", prefix, methodBadge, nameStyle.Render(name))
	} else {
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
		// Calculate available width for name: width - prefix - icon
		prefixLen := lipgloss.Width(prefix)
		iconLen := lipgloss.Width(icon)
		availableNameWidth := width - prefixLen - iconLen
		name := node.Name
		if availableNameWidth > 0 && len(name) > availableNameWidth {
			name = name[:availableNameWidth] // Truncate without ellipsis
		}
		content = fmt.Sprintf("%s%s%s", prefix, iconStyle.Render(icon), nameStyle.Render(name))
	}

	// Apply selection styling based on node type and selection state
	style := lipgloss.NewStyle().Width(width)
	if selected {
		if panelActive {
			// Active panel selection
			style = style.Background(styles.SelectedPanelBg).Foreground(styles.SelectedPanelFg).Bold(true)
		} else {
			// Inactive panel selection
			style = style.Background(styles.SelectedRequestBg).Foreground(styles.SelectedRequestFg)
		}
	}
	// Don't override foreground if not selected - content already has correct colors

	return style.Render(content)
}

// renderMethodBadge returns a styled HTTP method badge
func (t *Tree) renderMethodBadge(method string, dimmed bool) string {
	var bg, fg lipgloss.Color

	if dimmed {
		// Dimmed style for non-matching items during search
		bg = styles.SearchDimmed
		fg = styles.Mantle
	} else {
		switch method {
		case "GET":
			bg, fg = styles.MethodGetBg, styles.MethodGetFg
		case "POST":
			bg, fg = styles.MethodPostBg, styles.MethodPostFg
		case "PUT":
			bg, fg = styles.MethodPutBg, styles.MethodPutFg
		case "DELETE":
			bg, fg = styles.MethodDeleteBg, styles.MethodDeleteFg
		case "PATCH":
			bg, fg = styles.MethodPatchBg, styles.MethodPatchFg
		case "HEAD":
			bg, fg = styles.MethodHeadBg, styles.MethodHeadFg
		case "OPTIONS":
			bg, fg = styles.MethodOptionsBg, styles.MethodOptionsFg
		default:
			bg, fg = styles.Surface1, styles.Text
		}
	}

	style := lipgloss.NewStyle().
		Background(bg).
		Foreground(fg).
		Padding(0, 1)

	return style.Render(method)
}

// SetHeight sets the available height for the tree
func (t *Tree) SetHeight(h int) {
	t.height = h
	t.scrollIntoView()
}

// TreeState stores the state of the tree for restoration
type TreeState struct {
	ExpandedNodes map[string]bool // Map of node IDs to expanded state
	SelectedID    string          // ID of selected node
	CursorPos     int             // Cursor position
	ScrollOffset  int             // Scroll offset
}

// SaveState captures the current state of the tree
func (t *Tree) SaveState() *TreeState {
	state := &TreeState{
		ExpandedNodes: make(map[string]bool),
		CursorPos:     t.cursor,
		ScrollOffset:  t.scrollOffset,
	}

	// Save expanded state for all nodes
	t.saveExpandedState(t.Root, state.ExpandedNodes)

	// Save selected node ID
	if t.selected != nil {
		state.SelectedID = t.selected.ID
	}

	return state
}

// saveExpandedState recursively saves expanded state
func (t *Tree) saveExpandedState(nodes []*TreeNode, expanded map[string]bool) {
	for _, node := range nodes {
		if node.Type != RequestNode {
			expanded[node.ID] = node.Expanded
		}
		if len(node.Children) > 0 {
			t.saveExpandedState(node.Children, expanded)
		}
	}
}

// RestoreState restores the tree state after reload
func (t *Tree) RestoreState(state *TreeState) {
	if state == nil {
		return
	}

	// Restore expanded state
	t.restoreExpandedState(t.Root, state.ExpandedNodes)

	// Refresh visible list
	t.Refresh()

	// Try to restore cursor to same node by ID
	if state.SelectedID != "" {
		for i, node := range t.visible {
			if node.ID == state.SelectedID {
				t.cursor = i
				t.selected = node
				t.scrollOffset = state.ScrollOffset
				t.scrollIntoView()
				return
			}
		}
	}

	// Fallback: restore cursor position if possible
	if state.CursorPos < len(t.visible) {
		t.cursor = state.CursorPos
		if t.cursor >= 0 && t.cursor < len(t.visible) {
			t.selected = t.visible[t.cursor]
		}
	}
	t.scrollOffset = state.ScrollOffset
	t.scrollIntoView()
}

// restoreExpandedState recursively restores expanded state
func (t *Tree) restoreExpandedState(nodes []*TreeNode, expanded map[string]bool) {
	for _, node := range nodes {
		if exp, ok := expanded[node.ID]; ok {
			node.Expanded = exp
		}
		if len(node.Children) > 0 {
			t.restoreExpandedState(node.Children, expanded)
		}
	}
}

// GetExpandedFolders returns a list of expanded folder names/IDs for session persistence
func (t *Tree) GetExpandedFolders() []string {
	var expanded []string
	t.collectExpandedFolders(t.Root, &expanded)
	return expanded
}

// collectExpandedFolders recursively collects expanded folder names
func (t *Tree) collectExpandedFolders(nodes []*TreeNode, expanded *[]string) {
	for _, node := range nodes {
		if node.Expanded && (node.Type == FolderNode || node.Type == CollectionNode) {
			*expanded = append(*expanded, node.ID)
		}
		if len(node.Children) > 0 {
			t.collectExpandedFolders(node.Children, expanded)
		}
	}
}

// SetExpandedFolders expands folders matching the given IDs
func (t *Tree) SetExpandedFolders(folderIDs []string) {
	if len(folderIDs) == 0 {
		return
	}
	// Create a set for quick lookup
	expandSet := make(map[string]bool)
	for _, id := range folderIDs {
		expandSet[id] = true
	}
	t.applyExpandedFolders(t.Root, expandSet)
	t.Refresh()
}

// applyExpandedFolders recursively expands folders matching IDs
func (t *Tree) applyExpandedFolders(nodes []*TreeNode, expandSet map[string]bool) {
	for _, node := range nodes {
		if expandSet[node.ID] {
			node.Expanded = true
		}
		if len(node.Children) > 0 {
			t.applyExpandedFolders(node.Children, expandSet)
		}
	}
}

// GetScrollPosition returns the current scroll offset for session persistence
func (t *Tree) GetScrollPosition() int {
	return t.scrollOffset
}

// SetScrollPosition sets the scroll offset from session state
func (t *Tree) SetScrollPosition(offset int) {
	if offset >= 0 {
		t.scrollOffset = offset
	}
}

// GetSelectedIndex returns the current cursor position for session persistence
func (t *Tree) GetSelectedIndex() int {
	return t.cursor
}

// SetSelectedIndex sets the cursor position from session state
func (t *Tree) SetSelectedIndex(index int) {
	if index >= 0 && index < len(t.visible) {
		t.cursor = index
		t.selected = t.visible[t.cursor]
	} else if len(t.visible) > 0 {
		t.cursor = 0
		t.selected = t.visible[0]
	}
}

// FindNodeByID searches the tree recursively for a node with the given ID
func (t *Tree) FindNodeByID(id string) *TreeNode {
	return t.findNodeByIDRecursive(t.Root, id)
}

// findNodeByIDRecursive is the recursive helper for FindNodeByID
func (t *Tree) findNodeByIDRecursive(nodes []*TreeNode, id string) *TreeNode {
	for _, node := range nodes {
		if node.ID == id {
			return node
		}
		if len(node.Children) > 0 {
			if found := t.findNodeByIDRecursive(node.Children, id); found != nil {
				return found
			}
		}
	}
	return nil
}
