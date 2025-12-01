package api

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CollectionRequest represents a saved request in a collection
type CollectionRequest struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Method      HTTPMethod        `json:"method"`
	URL         string            `json:"url"`
	Headers     map[string]string `json:"headers,omitempty"`
	Body        interface{}       `json:"body,omitempty"`
	Tests       []Test            `json:"tests,omitempty"`
}

// Folder represents a folder in a collection
type Folder struct {
	Name        string               `json:"name"`
	Description string               `json:"description,omitempty"`
	Folders     []Folder             `json:"folders,omitempty"`
	Requests    []CollectionRequest  `json:"requests,omitempty"`
}

// CollectionFile represents a collection file structure
type CollectionFile struct {
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Folders     []Folder            `json:"folders,omitempty"`
	Requests    []CollectionRequest `json:"requests,omitempty"`
	FilePath    string              `json:"-"` // Path to the file (not serialized)
}

// Test represents a test assertion for a request
type Test struct {
	Name   string `json:"name"`
	Assert string `json:"assert"`
}

// LoadCollection loads a collection from a JSON file
func LoadCollection(path string) (*CollectionFile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read collection file: %w", err)
	}

	var collection CollectionFile
	if err := json.Unmarshal(data, &collection); err != nil {
		return nil, fmt.Errorf("failed to parse collection JSON: %w", err)
	}

	collection.FilePath = path
	return &collection, nil
}

// SaveCollection saves a collection to a JSON file
func SaveCollection(collection *CollectionFile, path string) error {
	data, err := json.MarshalIndent(collection, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal collection: %w", err)
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write collection file: %w", err)
	}

	return nil
}

// LoadAllCollections loads all collections from a directory
func LoadAllCollections(dir string) ([]*CollectionFile, error) {
	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return []*CollectionFile{}, nil
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("failed to read collections directory: %w", err)
	}

	var collections []*CollectionFile
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".json") {
			continue
		}

		path := filepath.Join(dir, file.Name())
		collection, err := LoadCollection(path)
		if err != nil {
			// Log error but continue loading other collections
			fmt.Printf("Warning: failed to load collection %s: %v\n", file.Name(), err)
			continue
		}

		collections = append(collections, collection)
	}

	return collections, nil
}

// ToRequest converts a CollectionRequest to a Request
func (cr *CollectionRequest) ToRequest() *Request {
	return &Request{
		Method:  cr.Method,
		URL:     cr.URL,
		Headers: cr.Headers,
		Body:    cr.Body,
	}
}

// FromRequest creates a CollectionRequest from a Request
func FromRequest(req *Request, name string) *CollectionRequest {
	return &CollectionRequest{
		ID:      GenerateID(),
		Name:    name,
		Method:  req.Method,
		URL:     req.URL,
		Headers: req.Headers,
		Body:    req.Body,
	}
}

// GenerateID generates a unique ID for a request
func GenerateID() string {
	// Generate unique ID using timestamp and random number
	return fmt.Sprintf("req_%d_%d", time.Now().UnixNano(), os.Getpid()%1000)
}

// FindRequest searches for a request by ID in the collection
func (c *CollectionFile) FindRequest(id string) *CollectionRequest {
	// Search in top-level requests
	for i := range c.Requests {
		if c.Requests[i].ID == id {
			return &c.Requests[i]
		}
	}

	// Search in folders
	return c.findRequestInFolders(c.Folders, id)
}

// findRequestInFolders recursively searches for a request in folders
func (c *CollectionFile) findRequestInFolders(folders []Folder, id string) *CollectionRequest {
	for _, folder := range folders {
		// Search in folder requests
		for i := range folder.Requests {
			if folder.Requests[i].ID == id {
				return &folder.Requests[i]
			}
		}

		// Search in subfolders
		if result := c.findRequestInFolders(folder.Folders, id); result != nil {
			return result
		}
	}
	return nil
}

// AddRequest adds a request to the collection
func (c *CollectionFile) AddRequest(req *CollectionRequest) {
	if req.ID == "" {
		req.ID = GenerateID()
	}
	c.Requests = append(c.Requests, *req)
}

// AddRequestToFolder adds a request to a specific folder
func (c *CollectionFile) AddRequestToFolder(folderPath []string, req *CollectionRequest) error {
	if req.ID == "" {
		req.ID = GenerateID()
	}

	if len(folderPath) == 0 {
		c.Requests = append(c.Requests, *req)
		return nil
	}

	folder := c.findFolder(c.Folders, folderPath, 0)
	if folder == nil {
		return fmt.Errorf("folder not found: %s", strings.Join(folderPath, "/"))
	}

	folder.Requests = append(folder.Requests, *req)
	return nil
}

// findFolder recursively finds a folder by path
func (c *CollectionFile) findFolder(folders []Folder, path []string, depth int) *Folder {
	if depth >= len(path) {
		return nil
	}

	for i := range folders {
		if folders[i].Name == path[depth] {
			if depth == len(path)-1 {
				return &folders[i]
			}
			return c.findFolder(folders[i].Folders, path, depth+1)
		}
	}
	return nil
}

// CreateFolder creates a new folder in the collection
func (c *CollectionFile) CreateFolder(name string) {
	c.Folders = append(c.Folders, Folder{
		Name:     name,
		Requests: []CollectionRequest{},
		Folders:  []Folder{},
	})
}

// ValidateCollection validates a collection structure
func ValidateCollection(collection *CollectionFile) error {
	if collection.Name == "" {
		return fmt.Errorf("collection name is required")
	}

	// Validate requests
	for _, req := range collection.Requests {
		if err := validateRequest(&req); err != nil {
			return fmt.Errorf("invalid request '%s': %w", req.Name, err)
		}
	}

	// Validate folders
	for _, folder := range collection.Folders {
		if err := validateFolder(&folder); err != nil {
			return err
		}
	}

	return nil
}

// validateRequest validates a single request
func validateRequest(req *CollectionRequest) error {
	if req.Name == "" {
		return fmt.Errorf("request name is required")
	}
	if req.Method == "" {
		return fmt.Errorf("request method is required")
	}
	if req.URL == "" {
		return fmt.Errorf("request URL is required")
	}
	return nil
}

// validateFolder recursively validates folders
func validateFolder(folder *Folder) error {
	if folder.Name == "" {
		return fmt.Errorf("folder name is required")
	}

	for _, req := range folder.Requests {
		if err := validateRequest(&req); err != nil {
			return fmt.Errorf("invalid request in folder '%s': %w", folder.Name, err)
		}
	}

	for _, subfolder := range folder.Folders {
		if err := validateFolder(&subfolder); err != nil {
			return err
		}
	}

	return nil
}

// Save saves the collection to its file path
func (c *CollectionFile) Save() error {
	if c.FilePath == "" {
		return fmt.Errorf("collection has no file path")
	}
	return SaveCollection(c, c.FilePath)
}

// CreateFolderInPath creates a folder at the specified path
func (c *CollectionFile) CreateFolderInPath(folderPath []string, name string) error {
	newFolder := Folder{
		Name:     name,
		Requests: []CollectionRequest{},
		Folders:  []Folder{},
	}

	if len(folderPath) == 0 {
		c.Folders = append(c.Folders, newFolder)
		return nil
	}

	folder := c.findFolder(c.Folders, folderPath, 0)
	if folder == nil {
		return fmt.Errorf("folder not found: %s", strings.Join(folderPath, "/"))
	}

	folder.Folders = append(folder.Folders, newFolder)
	return nil
}

// DeleteRequest removes a request by ID from anywhere in the collection
func (c *CollectionFile) DeleteRequest(id string) bool {
	// Check top-level requests
	for i, req := range c.Requests {
		if req.ID == id {
			c.Requests = append(c.Requests[:i], c.Requests[i+1:]...)
			return true
		}
	}

	// Check in folders
	return c.deleteRequestFromFolders(&c.Folders, id)
}

// deleteRequestFromFolders recursively searches and deletes a request
func (c *CollectionFile) deleteRequestFromFolders(folders *[]Folder, id string) bool {
	for i := range *folders {
		folder := &(*folders)[i]
		for j, req := range folder.Requests {
			if req.ID == id {
				folder.Requests = append(folder.Requests[:j], folder.Requests[j+1:]...)
				return true
			}
		}
		if c.deleteRequestFromFolders(&folder.Folders, id) {
			return true
		}
	}
	return false
}

// DeleteFolder removes a folder by name from the specified path
func (c *CollectionFile) DeleteFolder(folderPath []string, name string) bool {
	if len(folderPath) == 0 {
		// Delete from top-level
		for i, f := range c.Folders {
			if f.Name == name {
				c.Folders = append(c.Folders[:i], c.Folders[i+1:]...)
				return true
			}
		}
		return false
	}

	// Find parent folder
	parent := c.findFolder(c.Folders, folderPath, 0)
	if parent == nil {
		return false
	}

	for i, f := range parent.Folders {
		if f.Name == name {
			parent.Folders = append(parent.Folders[:i], parent.Folders[i+1:]...)
			return true
		}
	}
	return false
}

// RenameRequest renames a request by ID
func (c *CollectionFile) RenameRequest(id, newName string) bool {
	req := c.FindRequest(id)
	if req != nil {
		req.Name = newName
		return true
	}
	return false
}

// UpdateRequest updates a request's name, method, and URL by ID
func (c *CollectionFile) UpdateRequest(id, newName string, method HTTPMethod, url string) bool {
	req := c.FindRequest(id)
	if req != nil {
		req.Name = newName
		req.Method = method
		req.URL = url
		return true
	}
	return false
}

// RenameFolder renames a folder at the specified path
func (c *CollectionFile) RenameFolder(folderPath []string, oldName, newName string) bool {
	if len(folderPath) == 0 {
		// Rename top-level folder
		for i := range c.Folders {
			if c.Folders[i].Name == oldName {
				c.Folders[i].Name = newName
				return true
			}
		}
		return false
	}

	// Find parent folder
	parent := c.findFolder(c.Folders, folderPath, 0)
	if parent == nil {
		return false
	}

	for i := range parent.Folders {
		if parent.Folders[i].Name == oldName {
			parent.Folders[i].Name = newName
			return true
		}
	}
	return false
}

// DuplicateRequest duplicates a request by ID
func (c *CollectionFile) DuplicateRequest(id string) *CollectionRequest {
	original := c.FindRequest(id)
	if original == nil {
		return nil
	}

	duplicate := &CollectionRequest{
		ID:          GenerateID(),
		Name:        original.Name + " (copy)",
		Description: original.Description,
		Method:      original.Method,
		URL:         original.URL,
		Headers:     copyHeaders(original.Headers),
		Body:        original.Body,
	}

	// Add duplicate next to original - find where and add
	c.addRequestAfter(id, duplicate)
	return duplicate
}

// copyHeaders creates a copy of headers map
func copyHeaders(h map[string]string) map[string]string {
	if h == nil {
		return nil
	}
	copy := make(map[string]string)
	for k, v := range h {
		copy[k] = v
	}
	return copy
}

// addRequestAfter adds a request after another request with given ID
func (c *CollectionFile) addRequestAfter(afterID string, req *CollectionRequest) {
	// Check top-level
	for i, r := range c.Requests {
		if r.ID == afterID {
			// Insert after index i
			c.Requests = append(c.Requests[:i+1], append([]CollectionRequest{*req}, c.Requests[i+1:]...)...)
			return
		}
	}

	// Check in folders
	c.addRequestAfterInFolders(&c.Folders, afterID, req)
}

func (c *CollectionFile) addRequestAfterInFolders(folders *[]Folder, afterID string, req *CollectionRequest) bool {
	for i := range *folders {
		folder := &(*folders)[i]
		for j, r := range folder.Requests {
			if r.ID == afterID {
				folder.Requests = append(folder.Requests[:j+1], append([]CollectionRequest{*req}, folder.Requests[j+1:]...)...)
				return true
			}
		}
		if c.addRequestAfterInFolders(&folder.Folders, afterID, req) {
			return true
		}
	}
	return false
}

// FindFolder finds a folder by name in the collection
func (c *CollectionFile) FindFolderByName(folderPath []string, name string) *Folder {
	var folders *[]Folder
	if len(folderPath) == 0 {
		folders = &c.Folders
	} else {
		parent := c.findFolder(c.Folders, folderPath, 0)
		if parent == nil {
			return nil
		}
		folders = &parent.Folders
	}

	for i := range *folders {
		if (*folders)[i].Name == name {
			return &(*folders)[i]
		}
	}
	return nil
}

// DuplicateFolder duplicates a folder by name at the specified path
func (c *CollectionFile) DuplicateFolder(folderPath []string, name string) *Folder {
	original := c.FindFolderByName(folderPath, name)
	if original == nil {
		return nil
	}

	duplicate := copyFolder(original)
	duplicate.Name = original.Name + " (copy)"

	// Add duplicate next to original
	c.addFolderAfter(folderPath, name, duplicate)
	return duplicate
}

// copyFolder creates a deep copy of a folder
func copyFolder(f *Folder) *Folder {
	if f == nil {
		return nil
	}

	duplicate := &Folder{
		Name:        f.Name,
		Description: f.Description,
		Requests:    make([]CollectionRequest, len(f.Requests)),
		Folders:     make([]Folder, len(f.Folders)),
	}

	// Copy requests with new IDs
	for i, req := range f.Requests {
		duplicate.Requests[i] = CollectionRequest{
			ID:          GenerateID(),
			Name:        req.Name,
			Description: req.Description,
			Method:      req.Method,
			URL:         req.URL,
			Headers:     copyHeaders(req.Headers),
			Body:        req.Body,
		}
	}

	// Recursively copy subfolders
	for i, subfolder := range f.Folders {
		copied := copyFolder(&subfolder)
		if copied != nil {
			duplicate.Folders[i] = *copied
		}
	}

	return duplicate
}

// addFolderAfter adds a folder after another folder with given name
func (c *CollectionFile) addFolderAfter(folderPath []string, afterName string, folder *Folder) {
	var folders *[]Folder
	if len(folderPath) == 0 {
		folders = &c.Folders
	} else {
		parent := c.findFolder(c.Folders, folderPath, 0)
		if parent == nil {
			return
		}
		folders = &parent.Folders
	}

	for i, f := range *folders {
		if f.Name == afterName {
			// Insert after index i
			*folders = append((*folders)[:i+1], append([]Folder{*folder}, (*folders)[i+1:]...)...)
			return
		}
	}
}

// CopyRequestToFolder copies a request to a target folder
func (c *CollectionFile) CopyRequestToFolder(requestID string, targetFolderPath []string) *CollectionRequest {
	original := c.FindRequest(requestID)
	if original == nil {
		return nil
	}

	duplicate := &CollectionRequest{
		ID:          GenerateID(),
		Name:        original.Name + " (copy)",
		Description: original.Description,
		Method:      original.Method,
		URL:         original.URL,
		Headers:     copyHeaders(original.Headers),
		Body:        original.Body,
	}

	// Add to target folder
	if len(targetFolderPath) == 0 {
		c.Requests = append(c.Requests, *duplicate)
	} else {
		folder := c.findFolder(c.Folders, targetFolderPath, 0)
		if folder != nil {
			folder.Requests = append(folder.Requests, *duplicate)
		}
	}

	return duplicate
}

// CopyFolderToFolder copies a folder to a target location
func (c *CollectionFile) CopyFolderToFolder(sourcePath []string, sourceName string, targetFolderPath []string) *Folder {
	original := c.FindFolderByName(sourcePath, sourceName)
	if original == nil {
		return nil
	}

	duplicate := copyFolder(original)
	duplicate.Name = original.Name + " (copy)"

	// Add to target folder
	if len(targetFolderPath) == 0 {
		c.Folders = append(c.Folders, *duplicate)
	} else {
		folder := c.findFolder(c.Folders, targetFolderPath, 0)
		if folder != nil {
			folder.Folders = append(folder.Folders, *duplicate)
		}
	}

	return duplicate
}
