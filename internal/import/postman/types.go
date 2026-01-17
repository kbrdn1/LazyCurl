package postman

import "encoding/json"

// Collection represents the root structure of a Postman Collection v2.1 file.
type Collection struct {
	Info     Info       `json:"info"`
	Item     []Item     `json:"item"`
	Variable []Variable `json:"variable,omitempty"`
	Auth     *Auth      `json:"auth,omitempty"`
}

// Info contains collection metadata.
type Info struct {
	PostmanID   string `json:"_postman_id,omitempty"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Schema      string `json:"schema"`
}

// Item represents either a request or a folder (item group).
// If Request is nil, it's a folder containing nested Items.
type Item struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	Request     *Request `json:"request,omitempty"`
	Item        []Item   `json:"item,omitempty"`
	Event       []Event  `json:"event,omitempty"`
}

// IsFolder returns true if this item is a folder (has no request but may have items).
func (i *Item) IsFolder() bool {
	return i.Request == nil
}

// Request contains the full request definition.
type Request struct {
	Method      string   `json:"method"`
	Header      []Header `json:"header,omitempty"`
	Body        *Body    `json:"body,omitempty"`
	URL         URL      `json:"url"`
	Auth        *Auth    `json:"auth,omitempty"`
	Description string   `json:"description,omitempty"`
}

// URL contains URL with parsed components.
type URL struct {
	Raw      string       `json:"raw"`
	Protocol string       `json:"protocol,omitempty"`
	Host     []string     `json:"host,omitempty"`
	Path     []string     `json:"path,omitempty"`
	Query    []QueryParam `json:"query,omitempty"`
}

// UnmarshalJSON handles URL being either a string or an object in Postman collections.
func (u *URL) UnmarshalJSON(data []byte) error {
	// Try string first (simple URL)
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		u.Raw = str
		return nil
	}

	// Fall back to object
	type urlAlias URL
	var alias urlAlias
	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}
	*u = URL(alias)
	return nil
}

// Header represents a request header.
type Header struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description,omitempty"`
	Disabled    bool   `json:"disabled,omitempty"`
}

// QueryParam represents a URL query parameter.
type QueryParam struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description,omitempty"`
	Disabled    bool   `json:"disabled,omitempty"`
}

// Body contains request body definition.
type Body struct {
	Mode       string            `json:"mode"`
	Raw        string            `json:"raw,omitempty"`
	URLEncoded []URLEncodedParam `json:"urlencoded,omitempty"`
	FormData   []FormDataParam   `json:"formdata,omitempty"`
	Options    *BodyOptions      `json:"options,omitempty"`
	File       *FileBody         `json:"file,omitempty"`
	GraphQL    *GraphQLBody      `json:"graphql,omitempty"`
}

// BodyOptions contains body format options.
type BodyOptions struct {
	Raw *RawOptions `json:"raw,omitempty"`
}

// RawOptions specifies the raw body language.
type RawOptions struct {
	Language string `json:"language"` // json, xml, text, javascript, html
}

// URLEncodedParam represents a URL-encoded body parameter.
type URLEncodedParam struct {
	Key         string `json:"key"`
	Value       string `json:"value"`
	Description string `json:"description,omitempty"`
	Disabled    bool   `json:"disabled,omitempty"`
}

// FormDataParam represents a form-data parameter.
type FormDataParam struct {
	Key         string `json:"key"`
	Value       string `json:"value,omitempty"`
	Type        string `json:"type"` // text or file
	Src         string `json:"src,omitempty"`
	Description string `json:"description,omitempty"`
	Disabled    bool   `json:"disabled,omitempty"`
}

// FileBody represents a file body.
type FileBody struct {
	Src     string `json:"src,omitempty"`
	Content string `json:"content,omitempty"`
}

// GraphQLBody represents a GraphQL body.
type GraphQLBody struct {
	Query     string `json:"query,omitempty"`
	Variables string `json:"variables,omitempty"`
}

// Auth contains authentication configuration.
type Auth struct {
	Type   string         `json:"type"`
	Bearer []AuthKeyValue `json:"bearer,omitempty"`
	Basic  []AuthKeyValue `json:"basic,omitempty"`
	APIKey []AuthKeyValue `json:"apikey,omitempty"`
}

// AuthKeyValue represents a key-value pair in auth configuration.
type AuthKeyValue struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Type  string `json:"type,omitempty"`
}

// Variable represents a collection-level variable.
type Variable struct {
	Key   string `json:"key"`
	Value string `json:"value"`
	Type  string `json:"type,omitempty"`
}

// Event represents a pre-request or test script.
type Event struct {
	Listen string `json:"listen"` // prerequest or test
	Script Script `json:"script"`
}

// Script contains script content.
type Script struct {
	Type string   `json:"type"` // text/javascript
	Exec []string `json:"exec"` // Script lines
}

// Environment represents a Postman environment export file.
type Environment struct {
	ID                   string             `json:"id,omitempty"`
	Name                 string             `json:"name"`
	Values               []EnvironmentValue `json:"values"`
	PostmanVariableScope string             `json:"_postman_variable_scope,omitempty"`
}

// EnvironmentValue represents an environment variable.
type EnvironmentValue struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Type    string `json:"type,omitempty"` // default or secret
	Enabled bool   `json:"enabled"`
}
