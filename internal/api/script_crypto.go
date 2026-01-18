package api

import (
	"crypto/hmac"
	"crypto/md5"  //#nosec G501 -- MD5 provided for API compatibility, not security
	"crypto/sha1" //#nosec G505 -- SHA1 provided for API compatibility, not security
	"crypto/sha256"
	"crypto/sha512"
	"encoding/hex"

	"github.com/dop251/goja"
)

// setupLCCrypto creates the lc.crypto object for cryptographic hash operations
// Provides MD5, SHA1, SHA256, SHA512 hashes and HMAC-SHA256, HMAC-SHA512
//
// #nosec G104 -- Goja Set returns error only for invalid types, safe here
//
//nolint:errcheck,unparam // Goja Set operations are safe in this context, error for interface consistency
func (e *gojaExecutor) setupLCCrypto(vm *goja.Runtime, lc *goja.Object) error {
	cryptoObj := vm.NewObject()

	// lc.crypto.md5(data) - MD5 hash (hex encoded)
	// Note: MD5 is cryptographically weak, provided for API compatibility only
	cryptoObj.Set("md5", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return vm.ToValue("")
		}
		data := call.Arguments[0].String()
		hash := md5.Sum([]byte(data)) //#nosec G401 -- MD5 for API compat
		return vm.ToValue(hex.EncodeToString(hash[:]))
	})

	// lc.crypto.sha1(data) - SHA1 hash (hex encoded)
	// Note: SHA1 is cryptographically weak, provided for API compatibility only
	cryptoObj.Set("sha1", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return vm.ToValue("")
		}
		data := call.Arguments[0].String()
		hash := sha1.Sum([]byte(data)) //#nosec G401 -- SHA1 for API compat
		return vm.ToValue(hex.EncodeToString(hash[:]))
	})

	// lc.crypto.sha256(data) - SHA256 hash (hex encoded)
	cryptoObj.Set("sha256", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return vm.ToValue("")
		}
		data := call.Arguments[0].String()
		hash := sha256.Sum256([]byte(data))
		return vm.ToValue(hex.EncodeToString(hash[:]))
	})

	// lc.crypto.sha512(data) - SHA512 hash (hex encoded)
	cryptoObj.Set("sha512", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) == 0 {
			return vm.ToValue("")
		}
		data := call.Arguments[0].String()
		hash := sha512.Sum512([]byte(data))
		return vm.ToValue(hex.EncodeToString(hash[:]))
	})

	// lc.crypto.hmacSha256(data, secret) - HMAC-SHA256 (hex encoded)
	cryptoObj.Set("hmacSha256", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return vm.ToValue("")
		}
		data := call.Arguments[0].String()
		secret := call.Arguments[1].String()
		h := hmac.New(sha256.New, []byte(secret))
		h.Write([]byte(data))
		return vm.ToValue(hex.EncodeToString(h.Sum(nil)))
	})

	// lc.crypto.hmacSha512(data, secret) - HMAC-SHA512 (hex encoded)
	cryptoObj.Set("hmacSha512", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return vm.ToValue("")
		}
		data := call.Arguments[0].String()
		secret := call.Arguments[1].String()
		h := hmac.New(sha512.New, []byte(secret))
		h.Write([]byte(data))
		return vm.ToValue(hex.EncodeToString(h.Sum(nil)))
	})

	// lc.crypto.hmacSha1(data, secret) - HMAC-SHA1 (hex encoded)
	// Note: HMAC-SHA1 is provided for OAuth 1.0 compatibility
	cryptoObj.Set("hmacSha1", func(call goja.FunctionCall) goja.Value {
		if len(call.Arguments) < 2 {
			return vm.ToValue("")
		}
		data := call.Arguments[0].String()
		secret := call.Arguments[1].String()
		h := hmac.New(sha1.New, []byte(secret)) //#nosec G401 -- HMAC-SHA1 for OAuth compat
		h.Write([]byte(data))
		return vm.ToValue(hex.EncodeToString(h.Sum(nil)))
	})

	lc.Set("crypto", cryptoObj)
	return nil
}
