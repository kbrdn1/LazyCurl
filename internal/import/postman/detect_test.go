package postman

import (
	"path/filepath"
	"testing"
)

func TestDetectFileType_Collection(t *testing.T) {
	fileType, err := DetectFileType(filepath.Join("testdata", "simple_collection.json"))
	if err != nil {
		t.Fatalf("DetectFileType failed: %v", err)
	}

	if fileType != FileTypeCollection {
		t.Errorf("Expected FileTypeCollection, got %s", fileType)
	}
}

func TestDetectFileType_NestedCollection(t *testing.T) {
	fileType, err := DetectFileType(filepath.Join("testdata", "nested_collection.json"))
	if err != nil {
		t.Fatalf("DetectFileType failed: %v", err)
	}

	if fileType != FileTypeCollection {
		t.Errorf("Expected FileTypeCollection, got %s", fileType)
	}
}

func TestDetectFileType_Environment(t *testing.T) {
	fileType, err := DetectFileType(filepath.Join("testdata", "simple_environment.json"))
	if err != nil {
		t.Fatalf("DetectFileType failed: %v", err)
	}

	if fileType != FileTypeEnvironment {
		t.Errorf("Expected FileTypeEnvironment, got %s", fileType)
	}
}

func TestDetectFileType_EnvironmentByScope(t *testing.T) {
	// Environment detected by _postman_variable_scope field
	jsonData := []byte(`{
		"name": "Test Env",
		"values": [],
		"_postman_variable_scope": "environment"
	}`)

	fileType := DetectFileTypeFromBytes(jsonData)
	if fileType != FileTypeEnvironment {
		t.Errorf("Expected FileTypeEnvironment (by scope), got %s", fileType)
	}
}

func TestDetectFileType_EnvironmentByValues(t *testing.T) {
	// Environment detected by values array structure
	jsonData := []byte(`{
		"name": "Test Env",
		"values": [
			{"key": "test", "value": "value", "enabled": true}
		]
	}`)

	fileType := DetectFileTypeFromBytes(jsonData)
	if fileType != FileTypeEnvironment {
		t.Errorf("Expected FileTypeEnvironment (by values), got %s", fileType)
	}
}

func TestDetectFileType_Unknown(t *testing.T) {
	fileType, err := DetectFileType(filepath.Join("testdata", "not_postman.json"))
	if err != nil {
		t.Fatalf("DetectFileType failed: %v", err)
	}

	if fileType != FileTypeUnknown {
		t.Errorf("Expected FileTypeUnknown, got %s", fileType)
	}
}

func TestDetectFileType_InvalidJSON(t *testing.T) {
	jsonData := []byte(`{invalid}`)

	fileType := DetectFileTypeFromBytes(jsonData)
	if fileType != FileTypeUnknown {
		t.Errorf("Expected FileTypeUnknown for invalid JSON, got %s", fileType)
	}
}

func TestDetectFileType_FileNotFound(t *testing.T) {
	_, err := DetectFileType("nonexistent_file.json")
	if err == nil {
		t.Fatal("Expected error for missing file")
	}
}

func TestFileType_String(t *testing.T) {
	tests := []struct {
		fileType FileType
		expected string
	}{
		{FileTypeUnknown, "Unknown"},
		{FileTypeCollection, "Collection"},
		{FileTypeEnvironment, "Environment"},
	}

	for _, tt := range tests {
		if got := tt.fileType.String(); got != tt.expected {
			t.Errorf("FileType(%d).String() = %s, expected %s", tt.fileType, got, tt.expected)
		}
	}
}

func TestDetectFileType_EmptyEnvironmentWithName(t *testing.T) {
	// Edge case: empty values array with name field should be detected as environment
	jsonData := []byte(`{
		"name": "Empty Env",
		"values": []
	}`)

	fileType := DetectFileTypeFromBytes(jsonData)
	if fileType != FileTypeEnvironment {
		t.Errorf("Expected FileTypeEnvironment for empty env with name, got %s", fileType)
	}
}

func TestDetectFileType_CollectionBySchema(t *testing.T) {
	// Collection detected by schema URL
	jsonData := []byte(`{
		"info": {
			"name": "Test",
			"schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
		},
		"item": []
	}`)

	fileType := DetectFileTypeFromBytes(jsonData)
	if fileType != FileTypeCollection {
		t.Errorf("Expected FileTypeCollection (by schema), got %s", fileType)
	}
}

func TestDetectFileType_CollectionV20Schema(t *testing.T) {
	// Should also detect v2.0 schema
	jsonData := []byte(`{
		"info": {
			"name": "Test",
			"schema": "https://schema.getpostman.com/json/collection/v2.0.0/collection.json"
		},
		"item": []
	}`)

	fileType := DetectFileTypeFromBytes(jsonData)
	if fileType != FileTypeCollection {
		t.Errorf("Expected FileTypeCollection (v2.0 schema), got %s", fileType)
	}
}
