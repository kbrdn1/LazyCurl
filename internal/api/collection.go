package api

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// KeyValueEntry represents a key-value pair with enabled state (for params, headers)
type KeyValueEntry struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Enabled bool   `json:"enabled"`
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	Type   string `json:"type"`             // "none", "bearer", "basic", "api_key"
	Token  string `json:"token,omitempty"`  // For bearer token
	Prefix string `json:"prefix,omitempty"` // For bearer prefix (default: "Bearer")
	// Basic auth
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
	// API Key
	APIKeyName     string `json:"api_key_name,omitempty"`
	APIKeyValue    string `json:"api_key_value,omitempty"`
	APIKeyLocation string `json:"api_key_location,omitempty"` // "header" or "query"
}

// BodyConfig represents request body configuration
type BodyConfig struct {
	Type    string      `json:"type"`              // "none", "json", "form-data", "raw", "binary"
	Content interface{} `json:"content,omitempty"` // JSON object, string, or form data
}

// ScriptConfig represents pre/post request scripts
type ScriptConfig struct {
	PreRequest  string `json:"pre_request,omitempty"`
	PostRequest string `json:"post_request,omitempty"`
}

// CollectionRequest represents a saved request in a collection
type CollectionRequest struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Method      HTTPMethod        `json:"method"`
	URL         string            `json:"url"`
	Params      []KeyValueEntry   `json:"params,omitempty"`      // Query parameters
	Headers     []KeyValueEntry   `json:"headers,omitempty"`     // Request headers (new format)
	HeadersMap  map[string]string `json:"headers_map,omitempty"` // Legacy headers format
	Auth        *AuthConfig       `json:"auth,omitempty"`        // Authentication config
	Body        *BodyConfig       `json:"body,omitempty"`        // Request body config
	Scripts     *ScriptConfig     `json:"scripts,omitempty"`     // Pre/post scripts
	Tests       []Test            `json:"tests,omitempty"`
}

// Folder represents a folder in a collection
type Folder struct {
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	Folders     []Folder            `json:"folders,omitempty"`
	Requests    []CollectionRequest `json:"requests,omitempty"`
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

// UnmarshalJSON implements custom unmarshaling to handle both old (map) and new (array) header/param formats
func (cr *CollectionRequest) UnmarshalJSON(data []byte) error {
	// Alias to avoid infinite recursion
	type Alias CollectionRequest

	// Temporary struct to handle both formats
	type TempRequest struct {
		Alias
		HeadersRaw json.RawMessage `json:"headers,omitempty"`
		BodyRaw    json.RawMessage `json:"body,omitempty"`
	}

	var temp TempRequest
	if err := json.Unmarshal(data, &temp); err != nil {
		return err
	}

	// Copy all the basic fields
	*cr = CollectionRequest(temp.Alias)

	// Handle headers - try array format first, then map format
	if len(temp.HeadersRaw) > 0 {
		// Try new array format first
		var headersArray []KeyValueEntry
		if err := json.Unmarshal(temp.HeadersRaw, &headersArray); err == nil {
			cr.Headers = headersArray
		} else {
			// Try old map format
			var headersMap map[string]string
			if err := json.Unmarshal(temp.HeadersRaw, &headersMap); err == nil {
				cr.Headers = make([]KeyValueEntry, 0, len(headersMap))
				for k, v := range headersMap {
					cr.Headers = append(cr.Headers, KeyValueEntry{Key: k, Value: v, Enabled: true})
				}
			}
		}
	}

	// Handle body - try new BodyConfig format first, then raw content
	if len(temp.BodyRaw) > 0 {
		// Try new BodyConfig format first
		var bodyConfig BodyConfig
		if err := json.Unmarshal(temp.BodyRaw, &bodyConfig); err == nil && bodyConfig.Type != "" {
			cr.Body = &bodyConfig
		} else {
			// Old format - body is raw content (string or object)
			var bodyContent interface{}
			if err := json.Unmarshal(temp.BodyRaw, &bodyContent); err == nil {
				// Determine type based on content
				switch v := bodyContent.(type) {
				case string:
					if v != "" {
						cr.Body = &BodyConfig{Type: "raw", Content: v}
					}
				case map[string]interface{}:
					cr.Body = &BodyConfig{Type: "json", Content: v}
				default:
					if v != nil {
						cr.Body = &BodyConfig{Type: "raw", Content: v}
					}
				}
			}
		}
	}

	return nil
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
	// Convert []KeyValueEntry to map[string]string for HTTP request
	headers := make(map[string]string)
	for _, h := range cr.Headers {
		if h.Enabled {
			headers[h.Key] = h.Value
		}
	}

	// Convert body config to interface{}
	var body interface{}
	if cr.Body != nil {
		body = cr.Body.Content
	}

	return &Request{
		Method:  cr.Method,
		URL:     cr.URL,
		Headers: headers,
		Body:    body,
	}
}

// FromRequest creates a CollectionRequest from a Request
func FromRequest(req *Request, name string) *CollectionRequest {
	// Convert map[string]string to []KeyValueEntry
	headers := make([]KeyValueEntry, 0, len(req.Headers))
	for k, v := range req.Headers {
		headers = append(headers, KeyValueEntry{Key: k, Value: v, Enabled: true})
	}

	// Convert body to BodyConfig
	var body *BodyConfig
	if req.Body != nil {
		body = &BodyConfig{
			Type:    "raw",
			Content: req.Body,
		}
	}

	return &CollectionRequest{
		ID:      GenerateID(),
		Name:    name,
		Method:  req.Method,
		URL:     req.URL,
		Headers: headers,
		Body:    body,
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

	// Search in folders (pass pointer to allow modifications)
	return c.findRequestInFolders(&c.Folders, id)
}

// findRequestInFolders recursively searches for a request in folders
// Uses pointer to slice to ensure modifications persist
func (c *CollectionFile) findRequestInFolders(folders *[]Folder, id string) *CollectionRequest {
	for fi := range *folders {
		// Search in folder requests
		for ri := range (*folders)[fi].Requests {
			if (*folders)[fi].Requests[ri].ID == id {
				return &(*folders)[fi].Requests[ri]
			}
		}

		// Search in subfolders
		if result := c.findRequestInFolders(&(*folders)[fi].Folders, id); result != nil {
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

// UpdateRequestURL updates only the URL of a request by ID
func (c *CollectionFile) UpdateRequestURL(id, url string) bool {
	req := c.FindRequest(id)
	if req != nil {
		req.URL = url
		return true
	}
	return false
}

// UpdateRequestBody updates the body of a request by ID
func (c *CollectionFile) UpdateRequestBody(id, bodyType, content string) bool {
	req := c.FindRequest(id)
	if req != nil {
		if bodyType == "none" || content == "" {
			req.Body = nil
		} else {
			// For JSON body, try to parse as JSON object
			if bodyType == "json" {
				var parsed interface{}
				if err := json.Unmarshal([]byte(content), &parsed); err == nil {
					req.Body = &BodyConfig{Type: bodyType, Content: parsed}
					return true
				}
			}
			// Fallback to raw string content
			req.Body = &BodyConfig{Type: bodyType, Content: content}
		}
		return true
	}
	return false
}

// UpdateRequestScripts updates the scripts of a request by ID
func (c *CollectionFile) UpdateRequestScripts(id, preRequest, postRequest string) bool {
	req := c.FindRequest(id)
	if req != nil {
		if preRequest == "" && postRequest == "" {
			req.Scripts = nil
		} else {
			req.Scripts = &ScriptConfig{
				PreRequest:  preRequest,
				PostRequest: postRequest,
			}
		}
		return true
	}
	return false
}

// UpdateRequestAuth updates the auth configuration of a request by ID
func (c *CollectionFile) UpdateRequestAuth(id string, auth *AuthConfig) bool {
	req := c.FindRequest(id)
	if req != nil {
		if auth == nil || auth.Type == "none" || auth.Type == "" {
			req.Auth = nil
		} else {
			req.Auth = auth
		}
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
		Params:      copyParams(original.Params),
		Headers:     copyHeaders(original.Headers),
		Auth:        copyAuthConfig(original.Auth),
		Body:        copyBodyConfig(original.Body),
		Scripts:     copyScriptConfig(original.Scripts),
	}

	// Add duplicate next to original - find where and add
	c.addRequestAfter(id, duplicate)
	return duplicate
}

// copyHeaders creates a copy of headers slice
func copyHeaders(h []KeyValueEntry) []KeyValueEntry {
	if h == nil {
		return nil
	}
	result := make([]KeyValueEntry, len(h))
	copy(result, h)
	return result
}

// copyParams creates a copy of params slice
func copyParams(p []KeyValueEntry) []KeyValueEntry {
	if p == nil {
		return nil
	}
	result := make([]KeyValueEntry, len(p))
	copy(result, p)
	return result
}

// copyBodyConfig creates a copy of body config
func copyBodyConfig(b *BodyConfig) *BodyConfig {
	if b == nil {
		return nil
	}
	return &BodyConfig{
		Type:    b.Type,
		Content: b.Content,
	}
}

// copyAuthConfig creates a copy of auth config
func copyAuthConfig(a *AuthConfig) *AuthConfig {
	if a == nil {
		return nil
	}
	return &AuthConfig{
		Type:           a.Type,
		Token:          a.Token,
		Prefix:         a.Prefix,
		Username:       a.Username,
		Password:       a.Password,
		APIKeyName:     a.APIKeyName,
		APIKeyValue:    a.APIKeyValue,
		APIKeyLocation: a.APIKeyLocation,
	}
}

// copyScriptConfig creates a copy of script config
func copyScriptConfig(s *ScriptConfig) *ScriptConfig {
	if s == nil {
		return nil
	}
	return &ScriptConfig{
		PreRequest:  s.PreRequest,
		PostRequest: s.PostRequest,
	}
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
			Params:      copyParams(req.Params),
			Headers:     copyHeaders(req.Headers),
			Auth:        copyAuthConfig(req.Auth),
			Body:        copyBodyConfig(req.Body),
			Scripts:     copyScriptConfig(req.Scripts),
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
		Params:      copyParams(original.Params),
		Headers:     copyHeaders(original.Headers),
		Auth:        copyAuthConfig(original.Auth),
		Body:        copyBodyConfig(original.Body),
		Scripts:     copyScriptConfig(original.Scripts),
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
