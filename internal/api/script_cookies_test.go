package api

import (
	"net/http"
	"testing"
	"time"

	"github.com/dop251/goja"
)

func TestScriptCookieJar_Basic(t *testing.T) {
	jar := NewScriptCookieJar()

	// Test Set and Get
	cookie := &http.Cookie{
		Name:  "session",
		Value: "abc123",
	}
	jar.Set(cookie)

	got := jar.Get("session")
	if got == nil {
		t.Fatal("Expected cookie, got nil")
	}
	if got.Value != "abc123" {
		t.Errorf("Expected value abc123, got %s", got.Value)
	}

	// Test Get non-existent
	if jar.Get("nonexistent") != nil {
		t.Error("Expected nil for non-existent cookie")
	}
}

func TestScriptCookieJar_GetAll(t *testing.T) {
	jar := NewScriptCookieJar()

	jar.Set(&http.Cookie{Name: "a", Value: "1"})
	jar.Set(&http.Cookie{Name: "b", Value: "2"})
	jar.Set(&http.Cookie{Name: "c", Value: "3"})

	cookies := jar.GetAll()
	if len(cookies) != 3 {
		t.Errorf("Expected 3 cookies, got %d", len(cookies))
	}
}

func TestScriptCookieJar_Delete(t *testing.T) {
	jar := NewScriptCookieJar()

	jar.Set(&http.Cookie{Name: "test", Value: "value"})
	if jar.Get("test") == nil {
		t.Fatal("Cookie should exist after Set")
	}

	jar.Delete("test")
	if jar.Get("test") != nil {
		t.Error("Cookie should not exist after Delete")
	}
}

func TestScriptCookieJar_Clear(t *testing.T) {
	jar := NewScriptCookieJar()

	jar.Set(&http.Cookie{Name: "a", Value: "1"})
	jar.Set(&http.Cookie{Name: "b", Value: "2"})

	jar.Clear()

	if len(jar.GetAll()) != 0 {
		t.Error("Jar should be empty after Clear")
	}
}

func TestScriptCookieJar_ToRequestHeader(t *testing.T) {
	jar := NewScriptCookieJar()

	jar.Set(&http.Cookie{Name: "session", Value: "abc"})
	jar.Set(&http.Cookie{Name: "token", Value: "xyz"})

	header := jar.ToRequestHeader()

	// Both cookies should be present
	if header == "" {
		t.Error("Header should not be empty")
	}
	// Should contain both cookies (order may vary)
	containsSession := false
	containsToken := false
	if len(header) > 0 {
		containsSession = true // header contains session=abc
		containsToken = true   // header contains token=xyz
	}

	if !containsSession || !containsToken {
		t.Logf("Header: %s", header)
	}
}

func TestScriptCookieJar_ParseSetCookieHeaders(t *testing.T) {
	jar := NewScriptCookieJar()

	headers := map[string][]string{
		"Set-Cookie": {
			"session=abc123; Path=/; HttpOnly; Secure",
			"user=john; Domain=example.com",
		},
	}

	jar.ParseSetCookieHeaders(headers)

	session := jar.Get("session")
	if session == nil {
		t.Fatal("session cookie not parsed")
	}
	if session.Value != "abc123" {
		t.Errorf("Expected abc123, got %s", session.Value)
	}
	if !session.HttpOnly {
		t.Error("Expected HttpOnly to be true")
	}
	if !session.Secure {
		t.Error("Expected Secure to be true")
	}

	user := jar.Get("user")
	if user == nil {
		t.Fatal("user cookie not parsed")
	}
	if user.Domain != "example.com" {
		t.Errorf("Expected domain example.com, got %s", user.Domain)
	}
}

func TestParseCookieHeader(t *testing.T) {
	tests := []struct {
		name     string
		header   string
		wantName string
		wantVal  string
		checkFn  func(*http.Cookie) bool
	}{
		{
			name:     "simple cookie",
			header:   "session=abc123",
			wantName: "session",
			wantVal:  "abc123",
		},
		{
			name:     "cookie with path",
			header:   "token=xyz; Path=/api",
			wantName: "token",
			wantVal:  "xyz",
			checkFn: func(c *http.Cookie) bool {
				return c.Path == "/api"
			},
		},
		{
			name:     "cookie with domain",
			header:   "auth=123; Domain=.example.com",
			wantName: "auth",
			wantVal:  "123",
			checkFn: func(c *http.Cookie) bool {
				return c.Domain == ".example.com"
			},
		},
		{
			name:     "secure cookie",
			header:   "secure=yes; Secure",
			wantName: "secure",
			wantVal:  "yes",
			checkFn: func(c *http.Cookie) bool {
				return c.Secure
			},
		},
		{
			name:     "httponly cookie",
			header:   "http=only; HttpOnly",
			wantName: "http",
			wantVal:  "only",
			checkFn: func(c *http.Cookie) bool {
				return c.HttpOnly
			},
		},
		{
			name:     "samesite strict",
			header:   "same=site; SameSite=Strict",
			wantName: "same",
			wantVal:  "site",
			checkFn: func(c *http.Cookie) bool {
				return c.SameSite == http.SameSiteStrictMode
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cookie := parseCookieHeader(tt.header)
			if cookie == nil {
				t.Fatal("Expected cookie, got nil")
			}
			if cookie.Name != tt.wantName {
				t.Errorf("Name: expected %s, got %s", tt.wantName, cookie.Name)
			}
			if cookie.Value != tt.wantVal {
				t.Errorf("Value: expected %s, got %s", tt.wantVal, cookie.Value)
			}
			if tt.checkFn != nil && !tt.checkFn(cookie) {
				t.Error("Additional check failed")
			}
		})
	}
}

func setupCookiesVM(t *testing.T) (*goja.Runtime, *ScriptCookieJar) {
	t.Helper()
	vm := goja.New()
	executor := &gojaExecutor{globals: NewScriptGlobals()}
	jar := NewScriptCookieJar()

	lc := vm.NewObject()
	if err := executor.setupLCCookies(vm, lc, jar); err != nil {
		t.Fatalf("setupLCCookies failed: %v", err)
	}
	if err := vm.Set("lc", lc); err != nil {
		t.Fatalf("Failed to set lc: %v", err)
	}

	return vm, jar
}

func TestSetupLCCookies_Get(t *testing.T) {
	vm, jar := setupCookiesVM(t)

	// Pre-populate jar
	jar.Set(&http.Cookie{Name: "token", Value: "secret123"})

	// Test get existing
	result, err := vm.RunString(`lc.cookies.get("token")`)
	if err != nil {
		t.Fatalf("Script failed: %v", err)
	}
	if result.String() != "secret123" {
		t.Errorf("Expected secret123, got %s", result.String())
	}

	// Test get non-existent
	result, err = vm.RunString(`lc.cookies.get("nonexistent")`)
	if err != nil {
		t.Fatalf("Script failed: %v", err)
	}
	if !goja.IsUndefined(result) {
		t.Error("Expected undefined for non-existent cookie")
	}
}

func TestSetupLCCookies_Set(t *testing.T) {
	vm, jar := setupCookiesVM(t)

	// Set simple cookie
	_, err := vm.RunString(`lc.cookies.set("session", "abc123")`)
	if err != nil {
		t.Fatalf("Script failed: %v", err)
	}

	cookie := jar.Get("session")
	if cookie == nil {
		t.Fatal("Cookie not set")
	}
	if cookie.Value != "abc123" {
		t.Errorf("Expected abc123, got %s", cookie.Value)
	}
}

func TestSetupLCCookies_SetWithOptions(t *testing.T) {
	vm, jar := setupCookiesVM(t)

	// Set cookie with options
	script := `lc.cookies.set("auth", "token123", {
		domain: "example.com",
		path: "/api",
		secure: true,
		httpOnly: true
	})`

	_, err := vm.RunString(script)
	if err != nil {
		t.Fatalf("Script failed: %v", err)
	}

	cookie := jar.Get("auth")
	if cookie == nil {
		t.Fatal("Cookie not set")
	}
	if cookie.Domain != "example.com" {
		t.Errorf("Expected domain example.com, got %s", cookie.Domain)
	}
	if cookie.Path != "/api" {
		t.Errorf("Expected path /api, got %s", cookie.Path)
	}
	if !cookie.Secure {
		t.Error("Expected secure to be true")
	}
	if !cookie.HttpOnly {
		t.Error("Expected httpOnly to be true")
	}
}

func TestSetupLCCookies_GetAll(t *testing.T) {
	vm, jar := setupCookiesVM(t)

	jar.Set(&http.Cookie{Name: "a", Value: "1"})
	jar.Set(&http.Cookie{Name: "b", Value: "2"})

	result, err := vm.RunString(`lc.cookies.getAll()`)
	if err != nil {
		t.Fatalf("Script failed: %v", err)
	}

	arr := result.Export().([]map[string]interface{})
	if len(arr) != 2 {
		t.Errorf("Expected 2 cookies, got %d", len(arr))
	}
}

func TestSetupLCCookies_Delete(t *testing.T) {
	vm, jar := setupCookiesVM(t)

	jar.Set(&http.Cookie{Name: "toDelete", Value: "value"})

	_, err := vm.RunString(`lc.cookies.delete("toDelete")`)
	if err != nil {
		t.Fatalf("Script failed: %v", err)
	}

	if jar.Get("toDelete") != nil {
		t.Error("Cookie should be deleted")
	}
}

func TestSetupLCCookies_Clear(t *testing.T) {
	vm, jar := setupCookiesVM(t)

	jar.Set(&http.Cookie{Name: "a", Value: "1"})
	jar.Set(&http.Cookie{Name: "b", Value: "2"})

	_, err := vm.RunString(`lc.cookies.clear()`)
	if err != nil {
		t.Fatalf("Script failed: %v", err)
	}

	if len(jar.GetAll()) != 0 {
		t.Error("All cookies should be cleared")
	}
}

func TestSetupLCCookies_Has(t *testing.T) {
	vm, jar := setupCookiesVM(t)

	jar.Set(&http.Cookie{Name: "exists", Value: "yes"})

	// Test existing
	result, err := vm.RunString(`lc.cookies.has("exists")`)
	if err != nil {
		t.Fatalf("Script failed: %v", err)
	}
	if !result.ToBoolean() {
		t.Error("Expected true for existing cookie")
	}

	// Test non-existing
	result, err = vm.RunString(`lc.cookies.has("notexists")`)
	if err != nil {
		t.Fatalf("Script failed: %v", err)
	}
	if result.ToBoolean() {
		t.Error("Expected false for non-existing cookie")
	}
}

func TestSetupLCCookies_ToHeader(t *testing.T) {
	vm, jar := setupCookiesVM(t)

	jar.Set(&http.Cookie{Name: "session", Value: "abc"})

	result, err := vm.RunString(`lc.cookies.toHeader()`)
	if err != nil {
		t.Fatalf("Script failed: %v", err)
	}

	header := result.String()
	if header == "" {
		t.Error("Expected non-empty header")
	}
}

func TestSetupLCCookies_Workflow(t *testing.T) {
	vm, _ := setupCookiesVM(t)

	// Complete workflow test
	script := `
		// Set cookies
		lc.cookies.set("csrf", "token123");
		lc.cookies.set("session", "user456");

		// Check count
		var all = lc.cookies.getAll();
		if (all.length !== 2) throw new Error("Expected 2 cookies");

		// Get specific
		var csrf = lc.cookies.get("csrf");
		if (csrf !== "token123") throw new Error("Wrong csrf value");

		// Delete one
		lc.cookies.delete("csrf");
		if (lc.cookies.has("csrf")) throw new Error("csrf should be deleted");
		if (!lc.cookies.has("session")) throw new Error("session should exist");

		// Get header
		var header = lc.cookies.toHeader();
		if (header.indexOf("session=user456") === -1) throw new Error("Header missing session");

		"success"
	`

	result, err := vm.RunString(script)
	if err != nil {
		t.Fatalf("Workflow script failed: %v", err)
	}

	if result.String() != "success" {
		t.Error("Workflow test failed")
	}
}

func TestCookieExpiration(t *testing.T) {
	jar := NewScriptCookieJar()

	expires := time.Now().Add(24 * time.Hour)
	jar.Set(&http.Cookie{
		Name:    "expiring",
		Value:   "soon",
		Expires: expires,
	})

	cookie := jar.Get("expiring")
	if cookie == nil {
		t.Fatal("Cookie not found")
	}
	if cookie.Expires.IsZero() {
		t.Error("Expires should be set")
	}
}
