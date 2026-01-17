package postman

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/uuid"

	"github.com/kbrdn1/LazyCurl/internal/api"
)

const postmanSchemaV21 = "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"

// ExportCollection exports a LazyCurl collection to Postman Collection v2.1 format.
func ExportCollection(collection *api.CollectionFile, filePath string) error {
	data, err := ExportCollectionToBytes(collection)
	if err != nil {
		return err
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ExportCollectionToBytes exports a LazyCurl collection to Postman JSON bytes.
func ExportCollectionToBytes(collection *api.CollectionFile) ([]byte, error) {
	pc := convertToCollection(collection)
	data, err := json.MarshalIndent(pc, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal collection: %w", err)
	}
	return data, nil
}

// convertToCollection converts a LazyCurl CollectionFile to Collection.
func convertToCollection(collection *api.CollectionFile) *Collection {
	pc := &Collection{
		Info: Info{
			PostmanID:   uuid.New().String(),
			Name:        collection.Name,
			Description: collection.Description,
			Schema:      postmanSchemaV21,
		},
		Item: make([]Item, 0),
	}

	// Convert folders
	for _, folder := range collection.Folders {
		pc.Item = append(pc.Item, convertFolderToPostman(folder))
	}

	// Convert top-level requests
	for _, req := range collection.Requests {
		pc.Item = append(pc.Item, convertRequestToPostman(req))
	}

	return pc
}

// convertFolderToPostman converts a LazyCurl Folder to Item.
func convertFolderToPostman(folder api.Folder) Item {
	item := Item{
		Name:        folder.Name,
		Description: folder.Description,
		Item:        make([]Item, 0),
	}

	// Convert nested folders
	for _, subFolder := range folder.Folders {
		item.Item = append(item.Item, convertFolderToPostman(subFolder))
	}

	// Convert requests
	for _, req := range folder.Requests {
		item.Item = append(item.Item, convertRequestToPostman(req))
	}

	return item
}

// convertRequestToPostman converts a LazyCurl CollectionRequest to Item.
func convertRequestToPostman(req api.CollectionRequest) Item {
	postmanReq := Request{
		Method:      string(req.Method),
		Description: req.Description,
		URL:         convertURLToPostman(req.URL, req.Params),
		Header:      convertHeadersToPostman(req.Headers),
	}

	// Convert body
	if req.Body != nil {
		postmanReq.Body = convertBodyToPostman(req.Body)
	}

	// Convert auth
	if req.Auth != nil {
		postmanReq.Auth = convertAuthToPostman(req.Auth)
	}

	item := Item{
		Name:        req.Name,
		Description: req.Description,
		Request:     &postmanReq,
	}

	// Convert scripts
	if req.Scripts != nil {
		item.Event = convertScriptsToPostman(req.Scripts)
	}

	return item
}

// convertURLToPostman converts a URL string and params to URL.
func convertURLToPostman(urlStr string, params []api.KeyValueEntry) URL {
	postmanURL := URL{
		Raw: urlStr,
	}

	// Add query params
	for _, p := range params {
		postmanURL.Query = append(postmanURL.Query, QueryParam{
			Key:      p.Key,
			Value:    p.Value,
			Disabled: !p.Enabled,
		})
	}

	return postmanURL
}

// convertHeadersToPostman converts KeyValueEntry slice to Header slice.
func convertHeadersToPostman(headers []api.KeyValueEntry) []Header {
	if len(headers) == 0 {
		return nil
	}

	result := make([]Header, 0, len(headers))
	for _, h := range headers {
		result = append(result, Header{
			Key:      h.Key,
			Value:    h.Value,
			Disabled: !h.Enabled,
		})
	}
	return result
}

// convertBodyToPostman converts BodyConfig to Body.
func convertBodyToPostman(body *api.BodyConfig) *Body {
	if body == nil {
		return nil
	}

	switch body.Type {
	case "json":
		content := ""
		switch v := body.Content.(type) {
		case string:
			content = v
		default:
			if data, err := json.MarshalIndent(v, "", "  "); err == nil {
				content = string(data)
			}
		}
		return &Body{
			Mode: "raw",
			Raw:  content,
			Options: &BodyOptions{
				Raw: &RawOptions{Language: "json"},
			},
		}

	case "raw":
		content := ""
		if s, ok := body.Content.(string); ok {
			content = s
		}
		return &Body{
			Mode: "raw",
			Raw:  content,
			Options: &BodyOptions{
				Raw: &RawOptions{Language: "text"},
			},
		}

	case "form-data":
		var formData []FormDataParam
		switch v := body.Content.(type) {
		case []map[string]interface{}:
			for _, item := range v {
				param := FormDataParam{
					Key:  getString(item, "key"),
					Type: getString(item, "type"),
				}
				if param.Type == "" {
					param.Type = "text"
				}
				if param.Type == "file" {
					param.Src = getString(item, "src")
				} else {
					param.Value = getString(item, "value")
				}
				formData = append(formData, param)
			}
		case []interface{}:
			for _, i := range v {
				if item, ok := i.(map[string]interface{}); ok {
					param := FormDataParam{
						Key:  getString(item, "key"),
						Type: getString(item, "type"),
					}
					if param.Type == "" {
						param.Type = "text"
					}
					if param.Type == "file" {
						param.Src = getString(item, "src")
					} else {
						param.Value = getString(item, "value")
					}
					formData = append(formData, param)
				}
			}
		}
		return &Body{
			Mode:     "formdata",
			FormData: formData,
		}

	case "binary":
		src := ""
		if s, ok := body.Content.(string); ok {
			src = s
		}
		return &Body{
			Mode: "file",
			File: &FileBody{Src: src},
		}

	default:
		return nil
	}
}

// getString safely extracts a string from a map.
func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

// convertAuthToPostman converts AuthConfig to Auth.
func convertAuthToPostman(auth *api.AuthConfig) *Auth {
	if auth == nil || auth.Type == "" || auth.Type == "none" {
		return &Auth{Type: "noauth"}
	}

	switch auth.Type {
	case "bearer":
		return &Auth{
			Type: "bearer",
			Bearer: []AuthKeyValue{
				{Key: "token", Value: auth.Token, Type: "string"},
			},
		}

	case "basic":
		return &Auth{
			Type: "basic",
			Basic: []AuthKeyValue{
				{Key: "username", Value: auth.Username, Type: "string"},
				{Key: "password", Value: auth.Password, Type: "string"},
			},
		}

	case "api_key":
		location := auth.APIKeyLocation
		if location == "" {
			location = "header"
		}
		return &Auth{
			Type: "apikey",
			APIKey: []AuthKeyValue{
				{Key: "key", Value: auth.APIKeyName, Type: "string"},
				{Key: "value", Value: auth.APIKeyValue, Type: "string"},
				{Key: "in", Value: location, Type: "string"},
			},
		}

	default:
		return &Auth{Type: "noauth"}
	}
}

// convertScriptsToPostman converts ScriptConfig to Event slice.
func convertScriptsToPostman(scripts *api.ScriptConfig) []Event {
	if scripts == nil {
		return nil
	}

	var events []Event

	if scripts.PreRequest != "" {
		events = append(events, Event{
			Listen: "prerequest",
			Script: Script{
				Type: "text/javascript",
				Exec: strings.Split(scripts.PreRequest, "\n"),
			},
		})
	}

	if scripts.PostRequest != "" {
		events = append(events, Event{
			Listen: "test",
			Script: Script{
				Type: "text/javascript",
				Exec: strings.Split(scripts.PostRequest, "\n"),
			},
		})
	}

	return events
}

// ExportEnvironment exports a LazyCurl environment to Postman environment format.
func ExportEnvironment(env *api.EnvironmentFile, filePath string) error {
	data, err := ExportEnvironmentToBytes(env)
	if err != nil {
		return err
	}

	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

// ExportEnvironmentToBytes exports a LazyCurl environment to Postman JSON bytes.
func ExportEnvironmentToBytes(env *api.EnvironmentFile) ([]byte, error) {
	pe := convertToEnvironment(env)
	data, err := json.MarshalIndent(pe, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal environment: %w", err)
	}
	return data, nil
}

// convertToEnvironment converts a LazyCurl EnvironmentFile to Environment.
func convertToEnvironment(env *api.EnvironmentFile) *Environment {
	pe := &Environment{
		ID:                   uuid.New().String(),
		Name:                 env.Name,
		Values:               make([]EnvironmentValue, 0, len(env.Variables)),
		PostmanVariableScope: "environment",
	}

	// Sort keys for deterministic output
	keys := make([]string, 0, len(env.Variables))
	for key := range env.Variables {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		v := env.Variables[key]
		varType := "default"
		if v.Secret {
			varType = "secret"
		}

		pe.Values = append(pe.Values, EnvironmentValue{
			Key:     key,
			Value:   v.Value,
			Type:    varType,
			Enabled: v.Active,
		})
	}

	return pe
}
