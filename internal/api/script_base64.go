package api

import (
	"encoding/base64"

	"github.com/dop251/goja"
)

// setupLCBase64 creates the lc.base64 object and global btoa/atob functions
// for Base64 encoding/decoding operations
//
// #nosec G104 -- Goja Set returns error only for invalid types, safe here
//
//nolint:errcheck,unparam // Goja Set operations are safe in this context, error for interface consistency
func (e *gojaExecutor) setupLCBase64(vm *goja.Runtime, lc *goja.Object) error {
	base64Obj := vm.NewObject()

	// lc.base64.encode(data) - Encode string to Base64
	base64Obj.Set("encode", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return vm.ToValue("")
		}
		data := call.Arguments[0].String()
		encoded := base64.StdEncoding.EncodeToString([]byte(data))
		return vm.ToValue(encoded)
	})

	// lc.base64.decode(encoded) - Decode Base64 to string
	base64Obj.Set("decode", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return vm.ToValue("")
		}
		encoded := call.Arguments[0].String()
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			// Return empty string on decode error (matches JS behavior)
			return vm.ToValue("")
		}
		return vm.ToValue(string(decoded))
	})

	lc.Set("base64", base64Obj)

	// Global btoa() - Binary to ASCII (encode)
	vm.Set("btoa", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return vm.ToValue("")
		}
		data := call.Arguments[0].String()
		encoded := base64.StdEncoding.EncodeToString([]byte(data))
		return vm.ToValue(encoded)
	})

	// Global atob() - ASCII to Binary (decode)
	vm.Set("atob", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return vm.ToValue("")
		}
		encoded := call.Arguments[0].String()
		decoded, err := base64.StdEncoding.DecodeString(encoded)
		if err != nil {
			// Return empty string on decode error
			return vm.ToValue("")
		}
		return vm.ToValue(string(decoded))
	})

	return nil
}
