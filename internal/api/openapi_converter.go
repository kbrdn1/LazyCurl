package api

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/pb33f/libopenapi/datamodel/high/base"
	v3 "github.com/pb33f/libopenapi/datamodel/high/v3"
)

// extractSecuritySchemes extracts security schemes from OpenAPI document
func extractSecuritySchemes(doc *v3.Document) map[string]*AuthConfig {
	schemes := make(map[string]*AuthConfig)

	if doc.Components == nil || doc.Components.SecuritySchemes == nil {
		return schemes
	}

	for pair := doc.Components.SecuritySchemes.First(); pair != nil; pair = pair.Next() {
		schemeName := pair.Key()
		scheme := pair.Value()

		if scheme == nil {
			continue
		}

		schemeType := strings.ToLower(scheme.Type)

		switch schemeType {
		case "http":
			httpScheme := strings.ToLower(scheme.Scheme)
			if httpScheme == "bearer" {
				schemes[schemeName] = &AuthConfig{
					Type:  "bearer",
					Token: "",
				}
			} else if httpScheme == "basic" {
				schemes[schemeName] = &AuthConfig{
					Type: "basic",
				}
			}
		case "apikey":
			schemes[schemeName] = &AuthConfig{
				Type:           "api_key",
				APIKeyName:     scheme.Name,
				APIKeyLocation: scheme.In,
			}
		case "oauth2":
			// OAuth2 not supported, skip silently
			continue
		}
	}

	return schemes
}

// getOperationSecurity determines the security for an operation
func getOperationSecurity(op *v3.Operation, globalSecurity []*base.SecurityRequirement, schemes map[string]*AuthConfig) *AuthConfig {
	// Check operation-level security first
	security := op.Security

	// If operation has explicit empty security, it's public
	if security != nil && len(security) == 0 {
		return nil
	}

	// If no operation-level security, use global security
	if security == nil {
		security = globalSecurity
	}

	// No security defined
	if len(security) == 0 {
		return nil
	}

	// Use first security requirement
	firstReq := security[0]
	if firstReq.Requirements == nil || firstReq.Requirements.Len() == 0 {
		return nil
	}

	// Get first scheme from the requirement
	firstPair := firstReq.Requirements.First()
	if firstPair == nil {
		return nil
	}

	schemeName := firstPair.Key()
	if authConfig, exists := schemes[schemeName]; exists {
		return authConfig
	}

	return nil
}

// convertPathsToFolders converts OpenAPI paths to folders and requests organized by tags
func convertPathsToFolders(paths *v3.Paths, doc *v3.Document, baseURL string, includeExamples bool) []Folder {
	// Extract security schemes from document
	schemes := extractSecuritySchemes(doc)

	// Get global security requirements
	var globalSecurity []*base.SecurityRequirement
	if doc.Security != nil {
		globalSecurity = doc.Security
	}

	// Map to organize requests by tag
	tagRequests := make(map[string][]CollectionRequest)
	var untaggedRequests []CollectionRequest

	// Iterate through all paths
	for pair := paths.PathItems.First(); pair != nil; pair = pair.Next() {
		path := pair.Key()
		pathItem := pair.Value()

		// Convert each operation
		requests := pathItemToRequests(path, pathItem, baseURL, includeExamples, globalSecurity, schemes)

		for _, req := range requests {
			// Determine tag for this request (stored in description temporarily during conversion)
			tag := extractTagFromRequest(&req)

			if tag == "" {
				untaggedRequests = append(untaggedRequests, req)
			} else {
				tagRequests[tag] = append(tagRequests[tag], req)
			}
		}
	}

	// Build folders from tags
	var folders []Folder

	for tag, requests := range tagRequests {
		folders = append(folders, Folder{
			Name:     tag,
			Requests: requests,
		})
	}

	// Add untagged folder if needed
	if len(untaggedRequests) > 0 {
		folders = append(folders, Folder{
			Name:     "Untagged",
			Requests: untaggedRequests,
		})
	}

	return folders
}

// pathItemToRequests converts a path's operations to requests
func pathItemToRequests(path string, item *v3.PathItem, baseURL string, includeExamples bool, globalSecurity []*base.SecurityRequirement, schemes map[string]*AuthConfig) []CollectionRequest {
	var requests []CollectionRequest

	methodOps := map[HTTPMethod]*v3.Operation{
		GET:     item.Get,
		POST:    item.Post,
		PUT:     item.Put,
		DELETE:  item.Delete,
		PATCH:   item.Patch,
		HEAD:    item.Head,
		OPTIONS: item.Options,
	}

	for method, op := range methodOps {
		if op != nil {
			req := operationToRequest(path, method, op, baseURL, includeExamples, globalSecurity, schemes)
			requests = append(requests, req)
		}
	}

	return requests
}

// operationToRequest converts a single operation to a request
func operationToRequest(path string, method HTTPMethod, op *v3.Operation, baseURL string, includeExamples bool, globalSecurity []*base.SecurityRequirement, schemes map[string]*AuthConfig) CollectionRequest {
	// Build URL from base URL and path
	url := buildURL(baseURL, path)

	// Determine request name
	name := op.Summary
	if name == "" && op.OperationId != "" {
		name = op.OperationId
	}
	if name == "" {
		name = string(method) + " " + path
	}

	// Extract parameters
	queryParams, headers, pathParams := parametersToKeyValues(op.Parameters)

	// Replace path parameters in URL
	url = replacePathParameters(url, pathParams)

	// Extract request body
	var body *BodyConfig
	if op.RequestBody != nil && includeExamples {
		body, headers = requestBodyToBodyConfig(op.RequestBody, headers)
	}

	// Determine security for this operation
	auth := getOperationSecurity(op, globalSecurity, schemes)

	req := CollectionRequest{
		ID:          GenerateID(),
		Name:        name,
		Description: op.Description,
		Method:      method,
		URL:         url,
		Params:      queryParams,
		Headers:     headers,
		Body:        body,
		Auth:        auth,
	}

	// Store tag for folder organization (will be extracted later)
	if len(op.Tags) > 0 {
		// Use a special marker in description that we'll extract later
		// This is a workaround since CollectionRequest doesn't have a Tags field
		req.Description = setTagMarker(req.Description, op.Tags[0])
	}

	return req
}

// buildURL constructs the full URL from base URL and path
func buildURL(baseURL, path string) string {
	if baseURL == "" {
		return path
	}

	// Remove trailing slash from base URL
	baseURL = strings.TrimSuffix(baseURL, "/")

	// Ensure path starts with /
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}

	return baseURL + path
}

// parametersToKeyValues converts OpenAPI parameters to LazyCurl format
func parametersToKeyValues(params []*v3.Parameter) (query, headers []KeyValueEntry, pathParams map[string]string) {
	pathParams = make(map[string]string)

	for _, param := range params {
		if param == nil {
			continue
		}

		entry := KeyValueEntry{
			Key:     param.Name,
			Value:   getParameterExample(param),
			Enabled: param.Required != nil && *param.Required,
		}

		switch param.In {
		case "query":
			query = append(query, entry)
		case "header":
			headers = append(headers, entry)
		case "path":
			pathParams[param.Name] = entry.Value
		}
	}

	return query, headers, pathParams
}

// getParameterExample extracts or generates an example value for a parameter
func getParameterExample(param *v3.Parameter) string {
	// Check for explicit example
	if param.Example != nil {
		return formatExample(param.Example)
	}

	// Check schema for example or default
	if param.Schema != nil {
		schema := param.Schema.Schema()
		if schema != nil {
			if schema.Example != nil {
				return formatExample(schema.Example)
			}
			if schema.Default != nil {
				return formatExample(schema.Default)
			}

			// Generate example from type
			return generateExampleFromType(schema)
		}
	}

	return ""
}

// formatExample converts an example value to string
func formatExample(example interface{}) string {
	if example == nil {
		return ""
	}

	switch v := example.(type) {
	case string:
		return v
	case bool:
		if v {
			return "true"
		}
		return "false"
	case float64:
		// Check if it's a whole number
		if v == float64(int64(v)) {
			return fmt.Sprintf("%d", int64(v))
		}
		// Use JSON marshaling for proper formatting
		b, _ := json.Marshal(v)
		return string(b)
	case int:
		return fmt.Sprintf("%d", v)
	case int64:
		return fmt.Sprintf("%d", v)
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		return string(b)
	}
}

// generateExampleFromType generates an example value based on schema type
func generateExampleFromType(schema *base.Schema) string {
	if schema == nil {
		return ""
	}

	// Handle OpenAPI 3.1 type arrays
	types := schema.Type
	if len(types) == 0 {
		return ""
	}

	schemaType := types[0]

	switch schemaType {
	case "string":
		return generateStringExample(schema.Format)
	case "integer":
		return "0"
	case "number":
		return "0.0"
	case "boolean":
		return "false"
	default:
		return ""
	}
}

// generateStringExample generates an example string based on format
func generateStringExample(format string) string {
	switch format {
	case "email":
		return "user@example.com"
	case "uri", "url":
		return "https://example.com"
	case "uuid":
		return "550e8400-e29b-41d4-a716-446655440000"
	case "date":
		return "2024-01-15"
	case "date-time":
		return "2024-01-15T10:30:00Z"
	case "password":
		return "********"
	case "byte":
		return "SGVsbG8gV29ybGQ="
	case "hostname":
		return "example.com"
	case "ipv4":
		return "192.168.1.1"
	case "ipv6":
		return "2001:0db8:85a3:0000:0000:8a2e:0370:7334"
	default:
		return "string"
	}
}

// replacePathParameters replaces {param} placeholders with example values
func replacePathParameters(url string, pathParams map[string]string) string {
	for param, value := range pathParams {
		placeholder := "{" + param + "}"
		if value == "" {
			value = "{{" + param + "}}"
		}
		url = strings.ReplaceAll(url, placeholder, value)
	}
	return url
}

// requestBodyToBodyConfig converts request body to BodyConfig
func requestBodyToBodyConfig(body *v3.RequestBody, headers []KeyValueEntry) (*BodyConfig, []KeyValueEntry) {
	if body == nil || body.Content == nil {
		return nil, headers
	}

	// Priority: application/json first
	mediaTypePriority := []string{
		"application/json",
		"application/x-www-form-urlencoded",
		"multipart/form-data",
		"text/plain",
		"application/xml",
		"text/xml",
	}

	var selectedMediaType string
	var selectedContent *v3.MediaType

	// Try to find by priority
	for _, mt := range mediaTypePriority {
		if content := body.Content.GetOrZero(mt); content != nil {
			selectedMediaType = mt
			selectedContent = content
			break
		}
	}

	// If no priority match, take first available
	if selectedContent == nil {
		for pair := body.Content.First(); pair != nil; pair = pair.Next() {
			selectedMediaType = pair.Key()
			selectedContent = pair.Value()
			break
		}
	}

	if selectedContent == nil {
		return nil, headers
	}

	// Add Content-Type header
	headers = addOrUpdateHeader(headers, "Content-Type", selectedMediaType)

	// Generate body content
	var bodyConfig *BodyConfig

	switch {
	case strings.Contains(selectedMediaType, "json"):
		bodyConfig = &BodyConfig{
			Type:    "json",
			Content: generateSchemaExample(selectedContent.Schema),
		}
	case strings.Contains(selectedMediaType, "form"):
		bodyConfig = &BodyConfig{
			Type:    "form-data",
			Content: generateFormDataExample(selectedContent.Schema),
		}
	default:
		bodyConfig = &BodyConfig{
			Type:    "raw",
			Content: generateSchemaExampleAsString(selectedContent.Schema),
		}
	}

	return bodyConfig, headers
}

// addOrUpdateHeader adds or updates a header in the list
func addOrUpdateHeader(headers []KeyValueEntry, key, value string) []KeyValueEntry {
	for i, h := range headers {
		if strings.EqualFold(h.Key, key) {
			headers[i].Value = value
			return headers
		}
	}
	return append(headers, KeyValueEntry{
		Key:     key,
		Value:   value,
		Enabled: true,
	})
}

// generateSchemaExample generates an example from a schema reference
func generateSchemaExample(schemaRef *base.SchemaProxy) interface{} {
	if schemaRef == nil {
		return nil
	}

	schema := schemaRef.Schema()
	if schema == nil {
		return nil
	}

	return schemaToExample(schema, 0)
}

// generateSchemaExampleAsString generates an example as a string
func generateSchemaExampleAsString(schemaRef *base.SchemaProxy) string {
	example := generateSchemaExample(schemaRef)
	if example == nil {
		return ""
	}

	b, err := json.MarshalIndent(example, "", "  ")
	if err != nil {
		return ""
	}
	return string(b)
}

// generateFormDataExample generates form data from schema
func generateFormDataExample(schemaRef *base.SchemaProxy) interface{} {
	if schemaRef == nil {
		return nil
	}

	schema := schemaRef.Schema()
	if schema == nil {
		return nil
	}

	// For form data, we return an object with field values
	if schema.Properties != nil {
		result := make(map[string]interface{})
		for pair := schema.Properties.First(); pair != nil; pair = pair.Next() {
			propSchema := pair.Value().Schema()
			if propSchema != nil {
				result[pair.Key()] = schemaToExample(propSchema, 0)
			}
		}
		return result
	}

	return nil
}

// schemaToExample generates an example value from a JSON schema (with depth limit for circular refs)
func schemaToExample(schema *base.Schema, depth int) interface{} {
	if schema == nil || depth > 5 {
		return nil
	}

	// Use explicit example if available
	if schema.Example != nil {
		return schema.Example
	}

	// Use first item from examples if available
	if len(schema.Examples) > 0 {
		return schema.Examples[0]
	}

	// Get type (handle OpenAPI 3.1 type arrays)
	var schemaType string
	if len(schema.Type) > 0 {
		schemaType = schema.Type[0]
	}

	// Generate from type
	switch {
	case schemaType == "object" || slices.Contains(schema.Type, "object"):
		return generateObjectExample(schema, depth)

	case schemaType == "array" || slices.Contains(schema.Type, "array"):
		return generateArrayExample(schema, depth)

	case schemaType == "string" || slices.Contains(schema.Type, "string"):
		return generateStringExample(schema.Format)

	case schemaType == "integer" || slices.Contains(schema.Type, "integer"):
		if schema.Default != nil {
			return schema.Default
		}
		return 0

	case schemaType == "number" || slices.Contains(schema.Type, "number"):
		if schema.Default != nil {
			return schema.Default
		}
		return 0.0

	case schemaType == "boolean" || slices.Contains(schema.Type, "boolean"):
		if schema.Default != nil {
			return schema.Default
		}
		return false

	case schemaType == "null" || slices.Contains(schema.Type, "null"):
		return nil

	default:
		// Handle allOf, oneOf, anyOf
		if len(schema.AllOf) > 0 {
			return generateAllOfExample(schema.AllOf, depth)
		}
		if len(schema.OneOf) > 0 {
			return schemaToExample(schema.OneOf[0].Schema(), depth+1)
		}
		if len(schema.AnyOf) > 0 {
			return schemaToExample(schema.AnyOf[0].Schema(), depth+1)
		}
		return nil
	}
}

// generateObjectExample generates an example object
func generateObjectExample(schema *base.Schema, depth int) map[string]interface{} {
	result := make(map[string]interface{})

	if schema.Properties == nil {
		return result
	}

	for pair := schema.Properties.First(); pair != nil; pair = pair.Next() {
		propName := pair.Key()
		propSchemaProxy := pair.Value()

		if propSchemaProxy == nil {
			continue
		}

		propSchema := propSchemaProxy.Schema()
		if propSchema == nil {
			continue
		}

		result[propName] = schemaToExample(propSchema, depth+1)
	}

	return result
}

// generateArrayExample generates an example array
func generateArrayExample(schema *base.Schema, depth int) []interface{} {
	if schema.Items == nil {
		return []interface{}{}
	}

	itemSchema := schema.Items.A
	if itemSchema == nil {
		return []interface{}{}
	}

	item := schemaToExample(itemSchema.Schema(), depth+1)
	if item != nil {
		return []interface{}{item}
	}

	return []interface{}{}
}

// generateAllOfExample merges all schemas in allOf
func generateAllOfExample(allOf []*base.SchemaProxy, depth int) map[string]interface{} {
	result := make(map[string]interface{})

	for _, schemaProxy := range allOf {
		if schemaProxy == nil {
			continue
		}

		schema := schemaProxy.Schema()
		if schema == nil {
			continue
		}

		example := schemaToExample(schema, depth+1)
		if obj, ok := example.(map[string]interface{}); ok {
			for k, v := range obj {
				result[k] = v
			}
		}
	}

	return result
}

// Tag marker helpers for folder organization
const tagMarkerPrefix = "[[TAG:"
const tagMarkerSuffix = "]]"

// setTagMarker adds a hidden tag marker to the description
func setTagMarker(description, tag string) string {
	marker := tagMarkerPrefix + tag + tagMarkerSuffix
	if description == "" {
		return marker
	}
	return marker + " " + description
}

// extractTagFromRequest extracts and removes the tag marker from request description
func extractTagFromRequest(req *CollectionRequest) string {
	if !strings.HasPrefix(req.Description, tagMarkerPrefix) {
		return ""
	}

	endIdx := strings.Index(req.Description, tagMarkerSuffix)
	if endIdx == -1 {
		return ""
	}

	tag := req.Description[len(tagMarkerPrefix):endIdx]

	// Remove marker from description
	remaining := req.Description[endIdx+len(tagMarkerSuffix):]
	req.Description = strings.TrimPrefix(remaining, " ")

	return tag
}
