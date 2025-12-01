package ui

import (
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kbrdn1/LazyCurl/internal/api"
	"github.com/kbrdn1/LazyCurl/internal/config"
	"github.com/kbrdn1/LazyCurl/internal/ui/components"
)

// CollectionsView represents the collections panel
type CollectionsView struct {
	workspacePath   string
	collectionsPath string
	tree            *components.Tree
	collections     []*api.CollectionFile
	clipboard       *components.TreeNode // For yank/paste
}

// NewCollectionsView creates a new collections view
func NewCollectionsView(workspacePath string) *CollectionsView {
	cv := &CollectionsView{
		workspacePath:   workspacePath,
		collectionsPath: filepath.Join(workspacePath, ".lazycurl", "collections"),
	}

	// Load collections from workspace
	cv.loadCollections()

	return cv
}

// loadCollections loads collections from the workspace path
func (c *CollectionsView) loadCollections() {
	collections, err := api.LoadAllCollections(c.collectionsPath)
	if err != nil {
		// If no collections or error, create empty tree
		c.collections = []*api.CollectionFile{}
		c.tree = components.NewTree(c.collections)
		return
	}

	c.collections = collections
	c.tree = components.NewTree(collections)
}

// ReloadCollections reloads collections from disk while preserving tree state
func (c *CollectionsView) ReloadCollections() {
	// Save current tree state before reload
	var state *components.TreeState
	if c.tree != nil {
		state = c.tree.SaveState()
	}

	// Reload collections
	c.loadCollections()

	// Restore tree state after reload
	if state != nil && c.tree != nil {
		c.tree.RestoreState(state)
	}
}

// Update handles messages for the collections view
func (c CollectionsView) Update(msg tea.Msg, cfg *config.GlobalConfig) (CollectionsView, tea.Cmd) {
	// Forward all messages to tree component (including SearchUpdateMsg, SearchCloseMsg)
	allowNavigation := true
	tree, cmd := c.tree.Update(msg, allowNavigation)
	c.tree = tree
	return c, cmd
}

// View renders the collections view
func (c CollectionsView) View(width, height int, active bool) string {
	return c.tree.View(width, height, active)
}

// Selected returns the currently selected tree node
func (c CollectionsView) Selected() *components.TreeNode {
	return c.tree.Selected()
}

// GetTree returns the tree component for external access
func (c CollectionsView) GetTree() *components.Tree {
	return c.tree
}

// SetClipboard sets the clipboard node for copy/paste
func (c *CollectionsView) SetClipboard(node *components.TreeNode) {
	c.clipboard = node
}

// GetClipboard returns the clipboard node
func (c *CollectionsView) GetClipboard() *components.TreeNode {
	return c.clipboard
}

// GetCollectionsPath returns the path to collections directory
func (c *CollectionsView) GetCollectionsPath() string {
	return c.collectionsPath
}

// GetCollections returns the loaded collections
func (c *CollectionsView) GetCollections() []*api.CollectionFile {
	return c.collections
}

// FindCollectionByNode finds the collection that contains a tree node
func (c *CollectionsView) FindCollectionByNode(node *components.TreeNode) *api.CollectionFile {
	if node == nil {
		return nil
	}

	// Find the root collection node by walking up the parent chain
	root := node
	for root.Parent != nil {
		root = root.Parent
	}

	// Find the collection with matching name
	for _, col := range c.collections {
		if col.Name == root.Name {
			return col
		}
	}

	return nil
}

// GetFolderPath returns the folder path from a node to its collection
func (c *CollectionsView) GetFolderPath(node *components.TreeNode) []string {
	if node == nil {
		return nil
	}

	var path []string
	current := node

	// Walk up to collection (skip the collection itself)
	for current.Parent != nil {
		if current.Type == components.FolderNode {
			path = append([]string{current.Name}, path...)
		}
		current = current.Parent
	}

	return path
}

// AddRequestToCollection adds a new request to the appropriate collection
func (c *CollectionsView) AddRequestToCollection(name, method, url string, parentNode *components.TreeNode) error {
	col := c.FindCollectionByNode(parentNode)
	if col == nil {
		// No collection exists, create one
		return c.createDefaultCollectionWithRequest(name, method, url)
	}

	req := &api.CollectionRequest{
		ID:      api.GenerateID(),
		Name:    name,
		Method:  api.HTTPMethod(method),
		URL:     url,
		Headers: make(map[string]string),
	}

	// Get folder path
	folderPath := c.GetFolderPath(parentNode)

	// If parent is a folder, use its path; if it's a request, use its parent's path
	if parentNode != nil && parentNode.Type == components.RequestNode && parentNode.Parent != nil {
		folderPath = c.GetFolderPath(parentNode.Parent)
	}

	if err := col.AddRequestToFolder(folderPath, req); err != nil {
		return err
	}

	return col.Save()
}

// createDefaultCollectionWithRequest creates a new collection with a request
func (c *CollectionsView) createDefaultCollectionWithRequest(name, method, url string) error {
	col := &api.CollectionFile{
		Name:     "New Collection",
		Requests: []api.CollectionRequest{},
		Folders:  []api.Folder{},
		FilePath: filepath.Join(c.collectionsPath, "collection.json"),
	}

	req := &api.CollectionRequest{
		ID:      api.GenerateID(),
		Name:    name,
		Method:  api.HTTPMethod(method),
		URL:     url,
		Headers: make(map[string]string),
	}

	col.AddRequest(req)
	return col.Save()
}

// AddFolderToCollection adds a new folder to the appropriate collection
func (c *CollectionsView) AddFolderToCollection(name string, parentNode *components.TreeNode) error {
	col := c.FindCollectionByNode(parentNode)
	if col == nil {
		// Create a new collection with the folder
		return c.createDefaultCollectionWithFolder(name)
	}

	// Get folder path for parent
	folderPath := c.GetFolderPath(parentNode)

	if err := col.CreateFolderInPath(folderPath, name); err != nil {
		return err
	}

	return col.Save()
}

// createDefaultCollectionWithFolder creates a new collection with a folder
func (c *CollectionsView) createDefaultCollectionWithFolder(name string) error {
	col := &api.CollectionFile{
		Name:     "New Collection",
		Requests: []api.CollectionRequest{},
		Folders:  []api.Folder{},
		FilePath: filepath.Join(c.collectionsPath, "collection.json"),
	}

	col.CreateFolder(name)
	return col.Save()
}

// RenameNode renames a tree node (request or folder)
func (c *CollectionsView) RenameNode(node *components.TreeNode, newName string) error {
	if node == nil {
		return nil
	}

	col := c.FindCollectionByNode(node)
	if col == nil {
		return nil
	}

	switch node.Type {
	case components.CollectionNode:
		col.Name = newName
	case components.FolderNode:
		// Get parent path
		parentPath := c.GetFolderPath(node.Parent)
		col.RenameFolder(parentPath, node.Name, newName)
	case components.RequestNode:
		col.RenameRequest(node.ID, newName)
	}

	return col.Save()
}

// UpdateRequest updates a request node's name, method, and URL
func (c *CollectionsView) UpdateRequest(node *components.TreeNode, name, method, url string) error {
	if node == nil || node.Type != components.RequestNode {
		return nil
	}

	col := c.FindCollectionByNode(node)
	if col == nil {
		return nil
	}

	col.UpdateRequest(node.ID, name, api.HTTPMethod(method), url)
	return col.Save()
}

// DeleteNode deletes a tree node (request or folder)
func (c *CollectionsView) DeleteNode(node *components.TreeNode) error {
	if node == nil {
		return nil
	}

	col := c.FindCollectionByNode(node)
	if col == nil {
		return nil
	}

	switch node.Type {
	case components.CollectionNode:
		// Delete the entire collection file
		// Not implemented for safety - would need to delete the file
		return nil
	case components.FolderNode:
		parentPath := c.GetFolderPath(node.Parent)
		col.DeleteFolder(parentPath, node.Name)
	case components.RequestNode:
		col.DeleteRequest(node.ID)
	}

	return col.Save()
}

// DuplicateNode duplicates a tree node (request or folder)
func (c *CollectionsView) DuplicateNode(node *components.TreeNode) error {
	if node == nil {
		return nil
	}

	col := c.FindCollectionByNode(node)
	if col == nil {
		return nil
	}

	switch node.Type {
	case components.RequestNode:
		col.DuplicateRequest(node.ID)
	case components.FolderNode:
		parentPath := c.GetFolderPath(node.Parent)
		col.DuplicateFolder(parentPath, node.Name)
	case components.CollectionNode:
		// Cannot duplicate collection
		return nil
	}

	return col.Save()
}

// PasteNode pastes clipboard content to target location
// Target logic:
// - If target is a folder/collection: paste inside it
// - If target is a request: paste in same folder as the request
func (c *CollectionsView) PasteNode(clipboard *components.TreeNode, target *components.TreeNode) error {
	if clipboard == nil {
		return nil
	}

	// Find source collection
	sourceCol := c.FindCollectionByNode(clipboard)
	if sourceCol == nil {
		return nil
	}

	// Find target collection
	targetCol := c.FindCollectionByNode(target)
	if targetCol == nil {
		// If no target, use source collection root
		targetCol = sourceCol
	}

	// Determine target folder path based on cursor position
	var targetFolderPath []string
	if target != nil {
		switch target.Type {
		case components.CollectionNode:
			// Paste at collection root
			targetFolderPath = nil
		case components.FolderNode:
			// Paste inside the folder
			targetFolderPath = c.GetFolderPathIncluding(target)
		case components.RequestNode:
			// Paste in same folder as the request
			targetFolderPath = c.GetFolderPath(target.Parent)
		}
	}

	// Copy based on clipboard type
	switch clipboard.Type {
	case components.RequestNode:
		targetCol.CopyRequestToFolder(clipboard.ID, targetFolderPath)
	case components.FolderNode:
		sourcePath := c.GetFolderPath(clipboard.Parent)
		targetCol.CopyFolderToFolder(sourcePath, clipboard.Name, targetFolderPath)
	case components.CollectionNode:
		// Cannot paste collection
		return nil
	}

	return targetCol.Save()
}

// GetFolderPathIncluding returns the folder path including the node itself
func (c *CollectionsView) GetFolderPathIncluding(node *components.TreeNode) []string {
	if node == nil || node.Type != components.FolderNode {
		return nil
	}

	path := c.GetFolderPath(node.Parent)
	return append(path, node.Name)
}
