package postman

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/kbrdn1/LazyCurl/internal/api"
)

// ImportCollection imports a Postman Collection v2.1 file and converts it to LazyCurl format.
func ImportCollection(filePath string) (*ImportResult, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return ImportCollectionFromBytes(data)
}

// ImportCollectionFromBytes imports a Postman Collection from raw JSON bytes.
func ImportCollectionFromBytes(data []byte) (*ImportResult, error) {
	pc, err := parsePostmanCollection(data)
	if err != nil {
		return nil, err
	}

	if err := validatePostmanCollection(pc); err != nil {
		return nil, err
	}

	collection, summary := convertCollection(pc)
	return &ImportResult{
		Collection: collection,
		Summary:    *summary,
	}, nil
}

// parsePostmanCollection parses JSON bytes into a Collection struct.
func parsePostmanCollection(data []byte) (*Collection, error) {
	var pc Collection
	if err := json.Unmarshal(data, &pc); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}
	return &pc, nil
}

// validatePostmanCollection validates that the parsed data is a valid Postman Collection v2.1.
func validatePostmanCollection(pc *Collection) error {
	if pc.Info.Name == "" {
		return fmt.Errorf("invalid collection: name is required")
	}
	if !strings.Contains(pc.Info.Schema, "collection/v2") {
		return fmt.Errorf("not a valid Postman Collection v2.1 (missing or invalid schema)")
	}
	return nil
}

// convertCollection converts a Collection to a LazyCurl CollectionFile.
func convertCollection(pc *Collection) (*api.CollectionFile, *ImportSummary) {
	summary := &ImportSummary{
		CollectionName: pc.Info.Name,
	}

	collection := &api.CollectionFile{
		Name:        pc.Info.Name,
		Description: pc.Info.Description,
	}

	// Convert items (requests and folders)
	for _, item := range pc.Item {
		if item.IsFolder() {
			folder := convertFolder(item, summary)
			collection.Folders = append(collection.Folders, folder)
		} else {
			request := convertRequest(item, summary)
			collection.Requests = append(collection.Requests, request)
		}
	}

	return collection, summary
}

// convertFolder converts an Item folder to a LazyCurl Folder.
func convertFolder(item Item, summary *ImportSummary) api.Folder {
	summary.FoldersCount++

	folder := api.Folder{
		Name:        item.Name,
		Description: item.Description,
	}

	// Recursively convert nested items
	for _, subItem := range item.Item {
		if subItem.IsFolder() {
			subFolder := convertFolder(subItem, summary)
			folder.Folders = append(folder.Folders, subFolder)
		} else {
			request := convertRequest(subItem, summary)
			folder.Requests = append(folder.Requests, request)
		}
	}

	return folder
}

// convertRequest converts an Item request to a LazyCurl CollectionRequest.
func convertRequest(item Item, summary *ImportSummary) api.CollectionRequest {
	summary.RequestsCount++

	req := api.CollectionRequest{
		ID:     api.GenerateID(),
		Name:   item.Name,
		Method: api.HTTPMethod(strings.ToUpper(item.Request.Method)),
		URL:    convertURL(item.Request.URL),
	}

	// Convert description
	if item.Description != "" {
		req.Description = item.Description
	} else if item.Request.Description != "" {
		req.Description = item.Request.Description
	}

	// Convert headers
	req.Headers = convertHeaders(item.Request.Header)

	// Convert query params
	req.Params = convertQueryParams(item.Request.URL.Query)

	// Convert body
	if item.Request.Body != nil {
		req.Body = convertBody(item.Request.Body, summary, item.Name)
	}

	// Convert auth
	if item.Request.Auth != nil {
		req.Auth = convertAuth(item.Request.Auth, summary, item.Name)
	}

	// Handle scripts (store but warn)
	if len(item.Event) > 0 {
		req.Scripts = convertScripts(item.Event, summary, item.Name)
	}

	return req
}

// convertURL extracts the URL string from URL.
func convertURL(url URL) string {
	return url.Raw
}

// convertHeaders converts Header slice to KeyValueEntry slice.
func convertHeaders(headers []Header) []api.KeyValueEntry {
	if len(headers) == 0 {
		return nil
	}

	result := make([]api.KeyValueEntry, 0, len(headers))
	for _, h := range headers {
		result = append(result, api.KeyValueEntry{
			Key:     h.Key,
			Value:   h.Value,
			Enabled: !h.Disabled,
		})
	}
	return result
}

// convertQueryParams converts QueryParam slice to KeyValueEntry slice.
func convertQueryParams(params []QueryParam) []api.KeyValueEntry {
	if len(params) == 0 {
		return nil
	}

	result := make([]api.KeyValueEntry, 0, len(params))
	for _, p := range params {
		result = append(result, api.KeyValueEntry{
			Key:     p.Key,
			Value:   p.Value,
			Enabled: !p.Disabled,
		})
	}
	return result
}

// convertBody converts Body to BodyConfig.
func convertBody(body *Body, summary *ImportSummary, reqName string) *api.BodyConfig {
	switch body.Mode {
	case "raw":
		bodyType := "raw"
		// Check if it's JSON
		if body.Options != nil && body.Options.Raw != nil {
			switch body.Options.Raw.Language {
			case "json":
				bodyType = "json"
			case "xml", "html", "text", "javascript":
				bodyType = "raw"
			}
		}
		return &api.BodyConfig{
			Type:    bodyType,
			Content: body.Raw,
		}

	case "urlencoded":
		// Convert to form data structure
		formData := make([]map[string]interface{}, 0, len(body.URLEncoded))
		for _, param := range body.URLEncoded {
			if !param.Disabled {
				formData = append(formData, map[string]interface{}{
					"key":   param.Key,
					"value": param.Value,
				})
			}
		}
		return &api.BodyConfig{
			Type:    "form-data",
			Content: formData,
		}

	case "formdata":
		formData := make([]map[string]interface{}, 0, len(body.FormData))
		for _, param := range body.FormData {
			if !param.Disabled {
				entry := map[string]interface{}{
					"key":   param.Key,
					"value": param.Value,
					"type":  param.Type,
				}
				if param.Src != "" {
					entry["src"] = param.Src
					summary.AddWarningf("Request '%s' has file upload (path preserved only)", reqName)
				}
				formData = append(formData, entry)
			}
		}
		return &api.BodyConfig{
			Type:    "form-data",
			Content: formData,
		}

	case "file":
		summary.AddWarningf("Request '%s' uses file body mode (limited support)", reqName)
		if body.File == nil {
			summary.AddWarningf("Request '%s' has file body mode but missing file info", reqName)
			return &api.BodyConfig{
				Type:    "binary",
				Content: "",
			}
		}
		return &api.BodyConfig{
			Type:    "binary",
			Content: body.File.Src,
		}

	case "graphql":
		summary.AddWarningf("Request '%s' uses GraphQL body (imported as raw JSON)", reqName)
		if body.GraphQL == nil {
			summary.AddWarningf("Request '%s' missing GraphQL payload", reqName)
			return nil
		}
		// Convert GraphQL to JSON format using proper marshaling
		payload := map[string]interface{}{"query": body.GraphQL.Query}
		vars := strings.TrimSpace(body.GraphQL.Variables)
		if vars != "" && vars != "null" {
			var raw json.RawMessage
			if err := json.Unmarshal([]byte(vars), &raw); err != nil {
				summary.AddWarningf("Request '%s' has invalid GraphQL variables JSON (omitted)", reqName)
			} else {
				payload["variables"] = raw
			}
		}
		graphqlBytes, err := json.Marshal(payload)
		if err != nil {
			summary.AddWarningf("Request '%s' GraphQL body could not be marshaled", reqName)
			return &api.BodyConfig{
				Type:    "json",
				Content: fmt.Sprintf(`{"query": %q}`, body.GraphQL.Query),
			}
		}
		return &api.BodyConfig{
			Type:    "json",
			Content: string(graphqlBytes),
		}

	default:
		return nil
	}
}

// convertAuth converts Auth to AuthConfig.
func convertAuth(auth *Auth, summary *ImportSummary, reqName string) *api.AuthConfig {
	switch auth.Type {
	case "bearer":
		token := getAuthValue(auth.Bearer, "token")
		return &api.AuthConfig{
			Type:  "bearer",
			Token: token,
		}

	case "basic":
		return &api.AuthConfig{
			Type:     "basic",
			Username: getAuthValue(auth.Basic, "username"),
			Password: getAuthValue(auth.Basic, "password"),
		}

	case "apikey":
		return &api.AuthConfig{
			Type:           "api_key",
			APIKeyName:     getAuthValue(auth.APIKey, "key"),
			APIKeyValue:    getAuthValue(auth.APIKey, "value"),
			APIKeyLocation: getAuthValue(auth.APIKey, "in"),
		}

	case "noauth":
		return &api.AuthConfig{
			Type: "none",
		}

	case "oauth2":
		summary.AddWarningf("Request '%s' uses OAuth 2.0 (not supported)", reqName)
		return nil

	default:
		if auth.Type != "" {
			summary.AddWarningf("Request '%s' uses unsupported auth type '%s'", reqName, auth.Type)
		}
		return nil
	}
}

// getAuthValue retrieves a value from auth key-value pairs.
func getAuthValue(kvs []AuthKeyValue, key string) string {
	for _, kv := range kvs {
		if kv.Key == key {
			return kv.Value
		}
	}
	return ""
}

// convertScripts converts Event slice to ScriptConfig.
func convertScripts(events []Event, summary *ImportSummary, reqName string) *api.ScriptConfig {
	scripts := &api.ScriptConfig{}

	for _, event := range events {
		scriptContent := strings.Join(event.Script.Exec, "\n")

		switch event.Listen {
		case "prerequest":
			scripts.PreRequest = scriptContent
			summary.AddWarningf("Request '%s' has pre-request script (not executed)", reqName)
		case "test":
			scripts.PostRequest = scriptContent
			summary.AddWarningf("Request '%s' has test script (not executed)", reqName)
		}
	}

	if scripts.PreRequest == "" && scripts.PostRequest == "" {
		return nil
	}
	return scripts
}
