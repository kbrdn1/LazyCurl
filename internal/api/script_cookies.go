package api

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dop251/goja"
)

// ScriptCookieJar manages cookies for script execution
// Cookies can be read from responses and set for future requests
type ScriptCookieJar struct {
	cookies map[string]*http.Cookie
	mu      sync.RWMutex
}

// NewScriptCookieJar creates a new cookie jar for scripts
func NewScriptCookieJar() *ScriptCookieJar {
	return &ScriptCookieJar{
		cookies: make(map[string]*http.Cookie),
	}
}

// Get retrieves a cookie by name
func (j *ScriptCookieJar) Get(name string) *http.Cookie {
	j.mu.RLock()
	defer j.mu.RUnlock()
	return j.cookies[name]
}

// GetAll returns all cookies
func (j *ScriptCookieJar) GetAll() []*http.Cookie {
	j.mu.RLock()
	defer j.mu.RUnlock()
	result := make([]*http.Cookie, 0, len(j.cookies))
	for _, c := range j.cookies {
		result = append(result, c)
	}
	return result
}

// Set adds or updates a cookie
func (j *ScriptCookieJar) Set(cookie *http.Cookie) {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.cookies[cookie.Name] = cookie
}

// Delete removes a cookie by name
func (j *ScriptCookieJar) Delete(name string) {
	j.mu.Lock()
	defer j.mu.Unlock()
	delete(j.cookies, name)
}

// Clear removes all cookies
func (j *ScriptCookieJar) Clear() {
	j.mu.Lock()
	defer j.mu.Unlock()
	j.cookies = make(map[string]*http.Cookie)
}

// ParseSetCookieHeaders parses Set-Cookie headers from response and adds to jar
func (j *ScriptCookieJar) ParseSetCookieHeaders(headers map[string][]string) {
	for key, values := range headers {
		if strings.EqualFold(key, "Set-Cookie") {
			for _, value := range values {
				if cookie := parseCookieHeader(value); cookie != nil {
					j.Set(cookie)
				}
			}
		}
	}
}

// ToRequestHeader returns cookie string for request header
func (j *ScriptCookieJar) ToRequestHeader() string {
	j.mu.RLock()
	defer j.mu.RUnlock()

	var parts []string
	for _, c := range j.cookies {
		parts = append(parts, c.Name+"="+c.Value)
	}
	return strings.Join(parts, "; ")
}

// parseCookieHeader parses a Set-Cookie header value
func parseCookieHeader(header string) *http.Cookie {
	parts := strings.Split(header, ";")
	if len(parts) == 0 {
		return nil
	}

	// Parse name=value
	nameValue := strings.SplitN(strings.TrimSpace(parts[0]), "=", 2)
	if len(nameValue) != 2 {
		return nil
	}

	cookie := &http.Cookie{
		Name:  nameValue[0],
		Value: nameValue[1],
	}

	// Parse attributes
	for _, part := range parts[1:] {
		attr := strings.TrimSpace(part)
		if attr == "" {
			continue
		}

		if strings.EqualFold(attr, "Secure") {
			cookie.Secure = true
			continue
		}
		if strings.EqualFold(attr, "HttpOnly") {
			cookie.HttpOnly = true
			continue
		}

		kv := strings.SplitN(attr, "=", 2)
		if len(kv) != 2 {
			continue
		}

		key := strings.ToLower(kv[0])
		value := kv[1]

		switch key {
		case "domain":
			cookie.Domain = value
		case "path":
			cookie.Path = value
		case "expires":
			if t, err := time.Parse(time.RFC1123, value); err == nil {
				cookie.Expires = t
			}
		case "max-age":
			// Parse max-age (not implemented for simplicity)
		case "samesite":
			switch strings.ToLower(value) {
			case "strict":
				cookie.SameSite = http.SameSiteStrictMode
			case "lax":
				cookie.SameSite = http.SameSiteLaxMode
			case "none":
				cookie.SameSite = http.SameSiteNoneMode
			}
		}
	}

	return cookie
}

// setupLCCookies creates the lc.cookies object for cookie management
//
// #nosec G104 -- Goja Set returns error only for invalid types, safe here
//
//nolint:errcheck,unparam // Goja Set operations are safe in this context, error for interface consistency
func (e *gojaExecutor) setupLCCookies(vm *goja.Runtime, lc *goja.Object, jar *ScriptCookieJar) error {
	cookiesObj := vm.NewObject()

	// lc.cookies.get(name) - Get cookie by name
	cookiesObj.Set("get", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return goja.Undefined()
		}
		name := call.Arguments[0].String()
		cookie := jar.Get(name)
		if cookie == nil {
			return goja.Undefined()
		}
		return vm.ToValue(cookie.Value)
	})

	// lc.cookies.getAll() - Get all cookies as array of objects
	cookiesObj.Set("getAll", func(call goja.FunctionCall) goja.Value {
		cookies := jar.GetAll()
		result := make([]map[string]interface{}, 0, len(cookies))
		for _, c := range cookies {
			cookieObj := map[string]interface{}{
				"name":     c.Name,
				"value":    c.Value,
				"domain":   c.Domain,
				"path":     c.Path,
				"secure":   c.Secure,
				"httpOnly": c.HttpOnly,
			}
			if !c.Expires.IsZero() {
				cookieObj["expires"] = c.Expires.Format(time.RFC1123)
			}
			result = append(result, cookieObj)
		}
		return vm.ToValue(result)
	})

	// lc.cookies.set(name, value, options?) - Set a cookie
	cookiesObj.Set("set", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return goja.Undefined()
		}

		name := call.Arguments[0].String()
		value := call.Arguments[1].String()

		cookie := &http.Cookie{
			Name:  name,
			Value: value,
		}

		// Parse options if provided
		if len(call.Arguments) >= 3 {
			if opts, ok := call.Arguments[2].Export().(map[string]interface{}); ok {
				if domain, ok := opts["domain"].(string); ok {
					cookie.Domain = domain
				}
				if path, ok := opts["path"].(string); ok {
					cookie.Path = path
				}
				if secure, ok := opts["secure"].(bool); ok {
					cookie.Secure = secure
				}
				if httpOnly, ok := opts["httpOnly"].(bool); ok {
					cookie.HttpOnly = httpOnly
				}
				if expires, ok := opts["expires"].(string); ok {
					if t, err := time.Parse(time.RFC1123, expires); err == nil {
						cookie.Expires = t
					}
				}
			}
		}

		jar.Set(cookie)
		return goja.Undefined()
	})

	// lc.cookies.delete(name) - Delete a cookie
	cookiesObj.Set("delete", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) > 0 {
			name := call.Arguments[0].String()
			jar.Delete(name)
		}
		return goja.Undefined()
	})

	// lc.cookies.clear() - Clear all cookies
	cookiesObj.Set("clear", func(call goja.FunctionCall) goja.Value {
		jar.Clear()
		return goja.Undefined()
	})

	// lc.cookies.has(name) - Check if cookie exists
	cookiesObj.Set("has", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return vm.ToValue(false)
		}
		name := call.Arguments[0].String()
		return vm.ToValue(jar.Get(name) != nil)
	})

	// lc.cookies.toHeader() - Get Cookie header string
	cookiesObj.Set("toHeader", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(jar.ToRequestHeader())
	})

	lc.Set("cookies", cookiesObj)
	return nil
}
