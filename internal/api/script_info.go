package api

import (
	"github.com/dop251/goja"
)

// ScriptInfo contains contextual information about the script execution
type ScriptInfo struct {
	ScriptType      string // "pre-request" or "post-response"
	RequestName     string // Name of the request being executed
	RequestID       string // ID of the request
	CollectionName  string // Name of the collection
	EnvironmentName string // Name of the active environment
	Iteration       int    // Current iteration (for collection runner)
}

// NewScriptInfo creates a new ScriptInfo with defaults
func NewScriptInfo() *ScriptInfo {
	return &ScriptInfo{
		Iteration: 1,
	}
}

// setupLCInfo creates the lc.info object for script context information
//
// #nosec G104 -- Goja Set returns error only for invalid types, safe here
//
//nolint:errcheck,unparam // Goja Set operations are safe in this context, error for interface consistency
func (e *gojaExecutor) setupLCInfo(vm *goja.Runtime, lc *goja.Object, info *ScriptInfo) error {
	infoObj := vm.NewObject()

	// lc.info.scriptType - "pre-request" or "post-response"
	infoObj.DefineAccessorProperty("scriptType", vm.ToValue(func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(info.ScriptType)
	}), nil, goja.FLAG_FALSE, goja.FLAG_TRUE)

	// lc.info.requestName - Name of the current request
	infoObj.DefineAccessorProperty("requestName", vm.ToValue(func(call goja.FunctionCall) goja.Value {
		if info.RequestName == "" {
			return goja.Undefined()
		}
		return vm.ToValue(info.RequestName)
	}), nil, goja.FLAG_FALSE, goja.FLAG_TRUE)

	// lc.info.requestId - ID of the current request
	infoObj.DefineAccessorProperty("requestId", vm.ToValue(func(call goja.FunctionCall) goja.Value {
		if info.RequestID == "" {
			return goja.Undefined()
		}
		return vm.ToValue(info.RequestID)
	}), nil, goja.FLAG_FALSE, goja.FLAG_TRUE)

	// lc.info.collectionName - Name of the collection
	infoObj.DefineAccessorProperty("collectionName", vm.ToValue(func(call goja.FunctionCall) goja.Value {
		if info.CollectionName == "" {
			return goja.Undefined()
		}
		return vm.ToValue(info.CollectionName)
	}), nil, goja.FLAG_FALSE, goja.FLAG_TRUE)

	// lc.info.environmentName - Name of the active environment
	infoObj.DefineAccessorProperty("environmentName", vm.ToValue(func(call goja.FunctionCall) goja.Value {
		if info.EnvironmentName == "" {
			return goja.Undefined()
		}
		return vm.ToValue(info.EnvironmentName)
	}), nil, goja.FLAG_FALSE, goja.FLAG_TRUE)

	// lc.info.iteration - Current iteration number (for collection runner)
	infoObj.DefineAccessorProperty("iteration", vm.ToValue(func(call goja.FunctionCall) goja.Value {
		return vm.ToValue(info.Iteration)
	}), nil, goja.FLAG_FALSE, goja.FLAG_TRUE)

	lc.Set("info", infoObj)
	return nil
}
