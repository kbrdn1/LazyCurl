package api

import (
	"errors"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dop251/goja"
)

// ScriptExecutor handles JavaScript script execution
type ScriptExecutor interface {
	// ExecutePreRequest runs a pre-request script
	// Returns modified request data and execution result
	ExecutePreRequest(script string, req *ScriptRequest, env *Environment) (*ScriptResult, error)

	// ExecutePostResponse runs a post-response script
	// Returns execution result with assertions and env changes
	ExecutePostResponse(script string, req *ScriptRequest, resp *ScriptResponse, env *Environment) (*ScriptResult, error)

	// SetTimeout configures the script execution timeout
	SetTimeout(timeout time.Duration)

	// GetTimeout returns the current timeout setting
	GetTimeout() time.Duration
}

// gojaExecutor implements ScriptExecutor using the Goja JavaScript runtime
type gojaExecutor struct {
	timeout time.Duration
}

// NewScriptExecutor creates a new script executor instance
func NewScriptExecutor() ScriptExecutor {
	return &gojaExecutor{
		timeout: 5 * time.Second, // Default timeout
	}
}

// SetTimeout configures the script execution timeout
func (e *gojaExecutor) SetTimeout(timeout time.Duration) {
	e.timeout = timeout
}

// GetTimeout returns the current timeout setting
func (e *gojaExecutor) GetTimeout() time.Duration {
	return e.timeout
}

// ExecutePreRequest runs a pre-request script
func (e *gojaExecutor) ExecutePreRequest(script string, req *ScriptRequest, env *Environment) (*ScriptResult, error) {
	if script == "" {
		return NewScriptResult(), nil
	}

	startTime := time.Now()
	result := NewScriptResult()

	// Create fresh runtime
	vm := goja.New()

	// Create console and environment wrappers
	console := NewScriptConsole()
	scriptEnv := NewScriptEnvironment(env)
	assertions := NewAssertionCollector()

	// Setup global objects
	if err := e.setupConsole(vm, console); err != nil {
		result.SetError(err)
		return result, err
	}

	if err := e.setupLCObject(vm, req, nil, scriptEnv, assertions, true); err != nil {
		result.SetError(err)
		return result, err
	}

	// Execute script with timeout
	err := e.executeWithTimeout(vm, script)

	// Collect results
	result.Duration = time.Since(startTime)
	result.ConsoleOutput = console.GetEntries()
	result.EnvChanges = scriptEnv.GetChanges()
	result.Assertions = assertions.GetResults()
	result.RequestModified = req.IsModified()

	if err != nil {
		scriptErr := e.extractScriptError(err)
		result.SetError(scriptErr)
		return result, scriptErr
	}

	result.Success = true
	return result, nil
}

// ExecutePostResponse runs a post-response script
func (e *gojaExecutor) ExecutePostResponse(script string, req *ScriptRequest, resp *ScriptResponse, env *Environment) (*ScriptResult, error) {
	if script == "" {
		return NewScriptResult(), nil
	}

	startTime := time.Now()
	result := NewScriptResult()

	// Create fresh runtime
	vm := goja.New()

	// Create console and environment wrappers
	console := NewScriptConsole()
	scriptEnv := NewScriptEnvironment(env)
	assertions := NewAssertionCollector()

	// Setup global objects
	if err := e.setupConsole(vm, console); err != nil {
		result.SetError(err)
		return result, err
	}

	if err := e.setupLCObject(vm, req, resp, scriptEnv, assertions, false); err != nil {
		result.SetError(err)
		return result, err
	}

	// Execute script with timeout
	err := e.executeWithTimeout(vm, script)

	// Collect results
	result.Duration = time.Since(startTime)
	result.ConsoleOutput = console.GetEntries()
	result.EnvChanges = scriptEnv.GetChanges()
	result.Assertions = assertions.GetResults()

	if err != nil {
		scriptErr := e.extractScriptError(err)
		result.SetError(scriptErr)
		return result, scriptErr
	}

	result.Success = true
	return result, nil
}

// executeWithTimeout runs the script with a timeout
func (e *gojaExecutor) executeWithTimeout(vm *goja.Runtime, script string) error {
	done := make(chan error, 1)

	timer := time.AfterFunc(e.timeout, func() {
		vm.Interrupt("script execution timeout")
	})
	defer timer.Stop()

	go func() {
		_, err := vm.RunString(script)
		done <- err
	}()

	err := <-done

	// Check if it was a timeout interrupt
	if err != nil {
		var interrupted *goja.InterruptedError
		if errors.As(err, &interrupted) {
			if strings.Contains(interrupted.String(), "timeout") {
				return &ScriptTimeoutError{Timeout: e.timeout}
			}
		}
	}

	return err
}

// setupConsole binds console object to the runtime
//
// #nosec G104 -- Goja Set returns error only for invalid types, safe here
//
//nolint:errcheck // Goja Set operations are safe in this context
func (e *gojaExecutor) setupConsole(vm *goja.Runtime, console *ScriptConsole) error {
	consoleObj := vm.NewObject()

	consoleObj.Set("log", func(call goja.FunctionCall) goja.Value {
		args := e.extractArgs(call)
		console.Log(args...)
		return goja.Undefined()
	})

	consoleObj.Set("info", func(call goja.FunctionCall) goja.Value {
		args := e.extractArgs(call)
		console.Info(args...)
		return goja.Undefined()
	})

	consoleObj.Set("warn", func(call goja.FunctionCall) goja.Value {
		args := e.extractArgs(call)
		console.Warn(args...)
		return goja.Undefined()
	})

	consoleObj.Set("error", func(call goja.FunctionCall) goja.Value {
		args := e.extractArgs(call)
		console.Error(args...)
		return goja.Undefined()
	})

	consoleObj.Set("debug", func(call goja.FunctionCall) goja.Value {
		args := e.extractArgs(call)
		console.Debug(args...)
		return goja.Undefined()
	})

	return vm.Set("console", consoleObj)
}

// setupLCObject creates the main `lc` global object
//
// #nosec G104 -- Goja Set returns error only for invalid types, safe here
//
//nolint:errcheck // Goja Set operations are safe in this context
func (e *gojaExecutor) setupLCObject(vm *goja.Runtime, req *ScriptRequest, resp *ScriptResponse, env *ScriptEnvironment, assertions *AssertionCollector, isPreRequest bool) error {
	lc := vm.NewObject()

	// Setup lc.request
	if err := e.setupLCRequest(vm, lc, req, isPreRequest); err != nil {
		return err
	}

	// Setup lc.response (only for post-response scripts)
	if !isPreRequest && resp != nil {
		if err := e.setupLCResponse(vm, lc, resp); err != nil {
			return err
		}
	}

	// Setup lc.environment
	if err := e.setupLCEnvironment(vm, lc, env); err != nil {
		return err
	}

	// Setup lc.test() and lc.expect()
	if err := e.setupLCTest(vm, lc, assertions); err != nil {
		return err
	}

	return vm.Set("lc", lc)
}

// setupLCRequest creates the lc.request object
//
// #nosec G104 -- Goja Set returns error only for invalid types, safe here
//
//nolint:errcheck // Goja Set operations are safe in this context
func (e *gojaExecutor) setupLCRequest(vm *goja.Runtime, lc *goja.Object, req *ScriptRequest, isMutable bool) error {
	reqObj := vm.NewObject()

	// lc.request.method (readonly)
	reqObj.DefineAccessorProperty("method", vm.ToValue(func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(req.Method())
	}), nil, goja.FLAG_FALSE, goja.FLAG_TRUE)

	// lc.request.url (getter/setter for pre-request, readonly for post-response)
	if isMutable {
		reqObj.DefineAccessorProperty("url",
			vm.ToValue(func(call goja.FunctionCall) goja.Value {
				return vm.ToValue(req.URL())
			}),
			vm.ToValue(func(call goja.FunctionCall) goja.Value {
				if len(call.Arguments) > 0 {
					req.SetURL(call.Arguments[0].String())
				}
				return goja.Undefined()
			}),
			goja.FLAG_FALSE, goja.FLAG_TRUE)
	} else {
		reqObj.DefineAccessorProperty("url", vm.ToValue(func(call goja.FunctionCall) goja.Value {
			return vm.ToValue(req.URL())
		}), nil, goja.FLAG_FALSE, goja.FLAG_TRUE)
	}

	// lc.request.headers
	headersObj := vm.NewObject()
	headersObj.Set("get", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return goja.Undefined()
		}
		name := call.Arguments[0].String()
		value := req.GetHeader(name)
		if value == "" {
			return goja.Undefined()
		}
		return vm.ToValue(value)
	})

	if isMutable {
		headersObj.Set("set", func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) >= 2 {
				name := call.Arguments[0].String()
				value := call.Arguments[1].String()
				req.SetHeader(name, value)
			}
			return goja.Undefined()
		})

		headersObj.Set("remove", func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) > 0 {
				name := call.Arguments[0].String()
				req.RemoveHeader(name)
			}
			return goja.Undefined()
		})
	}

	headersObj.Set("all", func(call goja.FunctionCall) goja.Value {
		headers := req.Headers()
		obj := vm.NewObject()
		for k, v := range headers {
			obj.Set(k, v)
		}
		return obj
	})

	reqObj.Set("headers", headersObj)

	// lc.request.body
	bodyObj := vm.NewObject()
	bodyObj.Set("raw", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(req.Body())
	})

	bodyObj.Set("json", func(call goja.FunctionCall) goja.Value {
		body := req.Body()
		if body == "" {
			return goja.Null()
		}
		// Parse JSON safely using JSON.parse to prevent code injection
		jsonParse, err := vm.RunString("JSON.parse")
		if err != nil {
			return goja.Null()
		}
		fn, ok := goja.AssertFunction(jsonParse)
		if !ok {
			return goja.Null()
		}
		result, err := fn(goja.Undefined(), vm.ToValue(body))
		if err != nil {
			return goja.Null()
		}
		return result
	})

	if isMutable {
		bodyObj.Set("set", func(call goja.FunctionCall) goja.Value {
			if len(call.Arguments) > 0 {
				content := call.Arguments[0].String()
				req.SetBody(content)
			}
			return goja.Undefined()
		})
	}

	reqObj.Set("body", bodyObj)

	lc.Set("request", reqObj)
	return nil
}

// setupLCResponse creates the lc.response object
//
// #nosec G104 -- Goja Set returns error only for invalid types, safe here
//
//nolint:errcheck // Goja Set operations are safe in this context
func (e *gojaExecutor) setupLCResponse(vm *goja.Runtime, lc *goja.Object, resp *ScriptResponse) error {
	respObj := vm.NewObject()

	// lc.response.status
	respObj.DefineAccessorProperty("status", vm.ToValue(func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(resp.Status())
	}), nil, goja.FLAG_FALSE, goja.FLAG_TRUE)

	// lc.response.statusText
	respObj.DefineAccessorProperty("statusText", vm.ToValue(func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(resp.StatusText())
	}), nil, goja.FLAG_FALSE, goja.FLAG_TRUE)

	// lc.response.time
	respObj.DefineAccessorProperty("time", vm.ToValue(func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(resp.Time())
	}), nil, goja.FLAG_FALSE, goja.FLAG_TRUE)

	// lc.response.headers
	headersObj := vm.NewObject()
	headersObj.Set("get", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return goja.Undefined()
		}
		name := call.Arguments[0].String()
		value := resp.GetHeader(name)
		if value == "" {
			return goja.Undefined()
		}
		return vm.ToValue(value)
	})

	headersObj.Set("all", func(call goja.FunctionCall) goja.Value {
		headers := resp.Headers()
		obj := vm.NewObject()
		for k, v := range headers {
			obj.Set(k, v)
		}
		return obj
	})

	respObj.Set("headers", headersObj)

	// lc.response.body
	bodyObj := vm.NewObject()
	bodyObj.Set("raw", func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(resp.Body())
	})

	bodyObj.Set("json", func(call goja.FunctionCall) goja.Value {
		body := resp.Body()
		if body == "" {
			return goja.Null()
		}
		// Parse JSON safely using JSON.parse to prevent code injection
		jsonParse, err := vm.RunString("JSON.parse")
		if err != nil {
			return goja.Null()
		}
		fn, ok := goja.AssertFunction(jsonParse)
		if !ok {
			return goja.Null()
		}
		result, err := fn(goja.Undefined(), vm.ToValue(body))
		if err != nil {
			return goja.Null()
		}
		return result
	})

	respObj.Set("body", bodyObj)

	lc.Set("response", respObj)
	return nil
}

// setupLCEnvironment creates the lc.environment object
//
// #nosec G104 -- Goja Set returns error only for invalid types, safe here
//
//nolint:errcheck // Goja Set operations are safe in this context
func (e *gojaExecutor) setupLCEnvironment(vm *goja.Runtime, lc *goja.Object, env *ScriptEnvironment) error {
	envObj := vm.NewObject()

	envObj.Set("get", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return goja.Undefined()
		}
		name := call.Arguments[0].String()
		value := env.Get(name)
		if value == "" {
			return goja.Undefined()
		}
		return vm.ToValue(value)
	})

	envObj.Set("set", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) >= 2 {
			name := call.Arguments[0].String()
			value := call.Arguments[1].String()
			env.Set(name, value)
		}
		return goja.Undefined()
	})

	envObj.Set("unset", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) > 0 {
			name := call.Arguments[0].String()
			env.Unset(name)
		}
		return goja.Undefined()
	})

	envObj.Set("has", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return vm.ToValue(false)
		}
		name := call.Arguments[0].String()
		return vm.ToValue(env.Has(name))
	})

	lc.Set("environment", envObj)
	return nil
}

// setupLCTest creates lc.test() and lc.expect() functions
//
// #nosec G104 -- Goja Set returns error only for invalid types, safe here
//
//nolint:errcheck // Goja Set operations are safe in this context
func (e *gojaExecutor) setupLCTest(vm *goja.Runtime, lc *goja.Object, assertions *AssertionCollector) error {
	// lc.test(name, fn)
	lc.Set("test", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return goja.Undefined()
		}

		name := call.Arguments[0].String()
		fn, ok := goja.AssertFunction(call.Arguments[1])
		if !ok {
			return goja.Undefined()
		}

		// Execute test function
		_, err := fn(goja.Undefined())
		if err != nil {
			assertions.RegisterTest(name, false, nil, nil, err.Error())
		} else {
			assertions.RegisterTest(name, true, nil, nil, "")
		}

		return goja.Undefined()
	})

	// lc.expect(value) - returns expectation chain builder
	lc.Set("expect", func(call goja.FunctionCall) goja.Value {
		var actual interface{}
		if len(call.Arguments) > 0 {
			actual = call.Arguments[0].Export()
		}

		return e.createExpectation(vm, actual, assertions)
	})

	return nil
}

// createExpectation creates an expectation object with matcher methods
//
// #nosec G104 -- Goja Set returns error only for invalid types, safe here
//
//nolint:errcheck,unparam // Goja Set operations safe; assertions reserved for future use
func (e *gojaExecutor) createExpectation(vm *goja.Runtime, actual interface{}, assertions *AssertionCollector) goja.Value {
	exp := vm.NewObject()

	// toBe - strict equality (safe for non-comparable types)
	exp.Set("toBe", func(call goja.FunctionCall) goja.Value {
		var expected interface{}
		if len(call.Arguments) > 0 {
			expected = call.Arguments[0].Export()
		}
		// Check if types are comparable to avoid panic
		var passed bool
		if isComparable(actual) && isComparable(expected) {
			passed = actual == expected
		} else {
			// Fall back to deep equality for non-comparable types (maps, slices)
			passed = reflect.DeepEqual(actual, expected)
		}
		if !passed {
			// Throw an error to fail the test
			panic(vm.ToValue("Expected " + formatArg(actual) + " to be " + formatArg(expected)))
		}
		return exp
	})

	// toEqual - deep equality (simplified)
	exp.Set("toEqual", func(call goja.FunctionCall) goja.Value {
		var expected interface{}
		if len(call.Arguments) > 0 {
			expected = call.Arguments[0].Export()
		}
		passed := deepEqual(actual, expected)
		if !passed {
			panic(vm.ToValue("Expected " + formatArg(actual) + " to equal " + formatArg(expected)))
		}
		return exp
	})

	// toBeTruthy
	exp.Set("toBeTruthy", func(call goja.FunctionCall) goja.Value {
		passed := isTruthy(actual)
		if !passed {
			panic(vm.ToValue("Expected " + formatArg(actual) + " to be truthy"))
		}
		return exp
	})

	// toBeFalsy
	exp.Set("toBeFalsy", func(call goja.FunctionCall) goja.Value {
		passed := !isTruthy(actual)
		if !passed {
			panic(vm.ToValue("Expected " + formatArg(actual) + " to be falsy"))
		}
		return exp
	})

	// toContain
	exp.Set("toContain", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return exp
		}
		substring := call.Arguments[0].String()
		str, ok := actual.(string)
		if !ok {
			panic(vm.ToValue("Expected a string but got " + formatArg(actual)))
		}
		if !strings.Contains(str, substring) {
			panic(vm.ToValue("Expected " + formatArg(actual) + " to contain " + substring))
		}
		return exp
	})

	// toHaveProperty
	exp.Set("toHaveProperty", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return exp
		}
		propName := call.Arguments[0].String()
		obj, ok := actual.(map[string]interface{})
		if !ok {
			panic(vm.ToValue("Expected an object but got " + formatArg(actual)))
		}
		if _, exists := obj[propName]; !exists {
			panic(vm.ToValue("Expected object to have property " + propName))
		}
		return exp
	})

	// toBeGreaterThan
	exp.Set("toBeGreaterThan", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return exp
		}
		expected := call.Arguments[0].ToFloat()
		actualNum := toFloat(actual)
		if actualNum <= expected {
			panic(vm.ToValue("Expected " + formatArg(actual) + " to be greater than " + formatArg(expected)))
		}
		return exp
	})

	// toBeLessThan
	exp.Set("toBeLessThan", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return exp
		}
		expected := call.Arguments[0].ToFloat()
		actualNum := toFloat(actual)
		if actualNum >= expected {
			panic(vm.ToValue("Expected " + formatArg(actual) + " to be less than " + formatArg(expected)))
		}
		return exp
	})

	return exp
}

// extractArgs converts Goja arguments to Go interface slice
func (e *gojaExecutor) extractArgs(call goja.FunctionCall) []interface{} {
	args := make([]interface{}, len(call.Arguments))
	for i, arg := range call.Arguments {
		args[i] = arg.Export()
	}
	return args
}

// extractScriptError converts a Goja error to a typed script error
func (e *gojaExecutor) extractScriptError(err error) error {
	if err == nil {
		return nil
	}

	// If it's already one of our error types, return as-is
	var timeoutErr *ScriptTimeoutError
	if errors.As(err, &timeoutErr) {
		return err
	}
	var syntaxErr *ScriptSyntaxError
	if errors.As(err, &syntaxErr) {
		return err
	}
	var execErr *ScriptExecutionError
	if errors.As(err, &execErr) {
		return err
	}

	// Check for syntax error (from Goja compile)
	errStr := err.Error()
	if strings.Contains(errStr, "SyntaxError") {
		line, col := extractLineColumn(errStr)
		return &ScriptSyntaxError{
			Message: errStr,
			Line:    line,
			Column:  col,
		}
	}

	// Handle Goja exceptions
	var exc *goja.Exception
	if errors.As(err, &exc) {
		scriptErr := &ScriptExecutionError{
			Message:    exc.Value().String(),
			StackTrace: exc.String(),
		}

		// Try to extract line number from stack trace
		line, col := extractLineColumn(exc.String())
		scriptErr.Line = line
		scriptErr.Column = col

		return scriptErr
	}

	return &ScriptExecutionError{
		Message: err.Error(),
		Cause:   err,
	}
}

// extractLineColumn extracts line and column from error string
func extractLineColumn(errStr string) (int, int) {
	// Try to match patterns like "at <script>:5:10" or "line 5, column 10"
	patterns := []string{
		`:(\d+):(\d+)`,
		`line (\d+).*column (\d+)`,
		`at.*:(\d+):(\d+)`,
	}

	for _, pattern := range patterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(errStr)
		if len(matches) >= 3 {
			line, _ := strconv.Atoi(matches[1])
			col, _ := strconv.Atoi(matches[2])
			return line, col
		}
	}

	return 0, 0
}

// deepEqual performs deep equality comparison using reflect.DeepEqual
// This provides stable comparison for all types including maps and slices
func deepEqual(a, b interface{}) bool {
	return reflect.DeepEqual(a, b)
}

// isComparable checks if a value can be safely compared with ==
// Returns false for maps, slices, and other non-comparable types
func isComparable(v interface{}) bool {
	if v == nil {
		return true
	}
	t := reflect.TypeOf(v)
	return t.Comparable()
}

// isTruthy returns true if the value is truthy in JavaScript sense
func isTruthy(v interface{}) bool {
	if v == nil {
		return false
	}
	switch val := v.(type) {
	case bool:
		return val
	case int:
		return val != 0
	case int64:
		return val != 0
	case int32:
		return val != 0
	case int16:
		return val != 0
	case int8:
		return val != 0
	case uint:
		return val != 0
	case uint64:
		return val != 0
	case uint32:
		return val != 0
	case uint16:
		return val != 0
	case uint8:
		return val != 0
	case float32:
		return val != 0
	case float64:
		return val != 0
	case string:
		return val != ""
	default:
		return true
	}
}

// toFloat converts an interface to float64
func toFloat(v interface{}) float64 {
	switch val := v.(type) {
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case int32:
		return float64(val)
	case float32:
		return float64(val)
	case float64:
		return val
	default:
		return 0
	}
}
