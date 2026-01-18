package api

import (
	"math/rand"
	"time"

	"github.com/dop251/goja"
	"github.com/google/uuid"
)

// Package-level name lists to avoid per-call allocations
var firstNames = []string{
	"Alice", "Bob", "Charlie", "Diana", "Edward", "Fiona",
	"George", "Hannah", "Ivan", "Julia", "Kevin", "Laura",
	"Michael", "Nancy", "Oscar", "Patricia", "Quinn", "Rachel",
	"Samuel", "Teresa", "Ulrich", "Victoria", "William", "Xena", "Yuri", "Zoe",
}

var lastNames = []string{
	"Anderson", "Brown", "Clark", "Davis", "Evans", "Foster",
	"Garcia", "Harris", "Ivanov", "Johnson", "King", "Lee",
	"Martinez", "Nelson", "O'Brien", "Patel", "Quinn", "Roberts",
	"Smith", "Taylor", "Underwood", "Vargas", "Wilson", "Xavier", "Young", "Zhang",
}

// setupLCVariables creates the lc.variables object for dynamic variable generation
// Provides UUID, timestamp, random numbers and strings for test data generation
//
//nolint:errcheck,unparam // Goja Set operations are safe in this context, error for interface consistency
func (e *gojaExecutor) setupLCVariables(vm *goja.Runtime, lc *goja.Object) error {
	varsObj := vm.NewObject()

	// lc.variables.uuid() - Generate a new UUID v4
	varsObj.Set("uuid", func(call goja.FunctionCall) goja.Value { // #nosec G104 -- Goja Set safe here
		return vm.ToValue(uuid.New().String())
	})

	// lc.variables.timestamp() - Current Unix timestamp in seconds
	varsObj.Set("timestamp", func(call goja.FunctionCall) goja.Value { // #nosec G104 -- Goja Set safe here
		return vm.ToValue(time.Now().Unix())
	})

	// lc.variables.timestampMs() - Current Unix timestamp in milliseconds
	varsObj.Set("timestampMs", func(call goja.FunctionCall) goja.Value { // #nosec G104 -- Goja Set safe here
		return vm.ToValue(time.Now().UnixMilli())
	})

	// lc.variables.isoTimestamp() - Current UTC time in ISO 8601 format
	varsObj.Set("isoTimestamp", func(call goja.FunctionCall) goja.Value { // #nosec G104 -- Goja Set safe here
		return vm.ToValue(time.Now().UTC().Format(time.RFC3339))
	})

	// lc.variables.randomInt(min, max) - Random integer in range [min, max]
	varsObj.Set("randomInt", func(call goja.FunctionCall) goja.Value { // #nosec G104 -- Goja Set safe here
		min := 0
		max := 100

		if len(call.Arguments) >= 1 {
			min = int(call.Arguments[0].ToInteger())
		}
		if len(call.Arguments) >= 2 {
			max = int(call.Arguments[1].ToInteger())
		}

		if min > max {
			min, max = max, min // Swap if reversed
		}

		if min == max {
			return vm.ToValue(min)
		}

		// #nosec G404 -- Random used for test data, not security
		result := rand.Intn(max-min+1) + min
		return vm.ToValue(result)
	})

	// lc.variables.randomFloat() - Random float in range [0, 1)
	varsObj.Set("randomFloat", func(call goja.FunctionCall) goja.Value { // #nosec G104 -- Goja Set safe here
		// #nosec G404 -- Random used for test data, not security
		return vm.ToValue(rand.Float64())
	})

	// lc.variables.randomString(length) - Random alphanumeric string
	varsObj.Set("randomString", func(call goja.FunctionCall) goja.Value { // #nosec G104 -- Goja Set safe here
		length := 16 // Default length
		if len(call.Arguments) >= 1 {
			length = int(call.Arguments[0].ToInteger())
			if length <= 0 {
				length = 1
			}
			if length > 1000 {
				length = 1000 // Reasonable max
			}
		}

		const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
		b := make([]byte, length)
		for i := range b {
			// #nosec G404 -- Random used for test data, not security
			b[i] = charset[rand.Intn(len(charset))]
		}
		return vm.ToValue(string(b))
	})

	// lc.variables.randomHex(length) - Random hex string
	varsObj.Set("randomHex", func(call goja.FunctionCall) goja.Value { // #nosec G104 -- Goja Set safe here
		length := 16 // Default length
		if len(call.Arguments) >= 1 {
			length = int(call.Arguments[0].ToInteger())
			if length <= 0 {
				length = 1
			}
			if length > 1000 {
				length = 1000
			}
		}

		const charset = "0123456789abcdef"
		b := make([]byte, length)
		for i := range b {
			// #nosec G404 -- Random used for test data, not security
			b[i] = charset[rand.Intn(len(charset))]
		}
		return vm.ToValue(string(b))
	})

	// lc.variables.randomEmail() - Generate random email address
	varsObj.Set("randomEmail", func(call goja.FunctionCall) goja.Value { // #nosec G104 -- Goja Set safe here
		const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
		domains := []string{"example.com", "test.com", "email.test", "mail.example"}

		// Generate username (8-12 chars)
		// #nosec G404 -- Random used for test data, not security
		usernameLen := rand.Intn(5) + 8
		username := make([]byte, usernameLen)
		for i := range username {
			// #nosec G404 -- Random used for test data, not security
			username[i] = charset[rand.Intn(len(charset))]
		}

		// #nosec G404 -- Random used for test data, not security
		domain := domains[rand.Intn(len(domains))]

		return vm.ToValue(string(username) + "@" + domain)
	})

	// lc.variables.randomFirstName() - Random first name from list
	varsObj.Set("randomFirstName", func(call goja.FunctionCall) goja.Value { // #nosec G104 -- Goja Set safe here
		// #nosec G404 -- Random used for test data, not security
		return vm.ToValue(firstNames[rand.Intn(len(firstNames))])
	})

	// lc.variables.randomLastName() - Random last name from list
	varsObj.Set("randomLastName", func(call goja.FunctionCall) goja.Value { // #nosec G104 -- Goja Set safe here
		// #nosec G404 -- Random used for test data, not security
		return vm.ToValue(lastNames[rand.Intn(len(lastNames))])
	})

	// lc.variables.randomBoolean() - Random boolean value
	varsObj.Set("randomBoolean", func(call goja.FunctionCall) goja.Value { // #nosec G104 -- Goja Set safe here
		// #nosec G404 -- Random used for test data, not security
		return vm.ToValue(rand.Intn(2) == 1)
	})

	lc.Set("variables", varsObj) // #nosec G104 -- Goja Set safe here
	return nil
}
