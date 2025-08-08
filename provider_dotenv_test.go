package zconfig

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewDotenvProvider(t *testing.T) {
	// Create a temporary .env file
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, ".env")

	content := `# This is a comment
DATABASE_URL=postgres://localhost/test
API_KEY="secret-key"
DEBUG=true

# Another comment
PORT=8080
EMPTY_VALUE=
`

	err := os.WriteFile(envFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}

	// Change to temp directory to test default behavior
	oldDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(oldDir); err != nil {
			t.Errorf("Failed to restore directory: %v", err)
		}
	}()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change to temp directory: %v", err)
	}

	provider := NewDotenvProvider()

	tests := []struct {
		key      string
		expected string
		found    bool
	}{
		{"database.url", "postgres://localhost/test", true},
		{"api.key", "secret-key", true},
		{"debug", "true", true},
		{"port", "8080", true},
		{"empty.value", "", true},
		{"nonexistent", "", false},
	}

	for _, test := range tests {
		value, found, err := provider.Retrieve(test.key)
		if err != nil {
			t.Errorf("Unexpected error for key %s: %v", test.key, err)
			continue
		}

		if found != test.found {
			t.Errorf("For key %s: expected found=%v, got found=%v", test.key, test.found, found)
			continue
		}

		if found && value != test.expected {
			t.Errorf("For key %s: expected value=%s, got value=%s", test.key, test.expected, value)
		}
	}
}

func TestNewDotenvProviderWithPath(t *testing.T) {
	// Create a temporary .env file
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, "custom.env")

	content := `CUSTOM_VAR=custom-value
ANOTHER_VAR='quoted-value'
`

	err := os.WriteFile(envFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}

	provider := NewDotenvProviderWithPath(envFile)

	value, found, err := provider.Retrieve("custom.var")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !found {
		t.Error("Expected to find custom.var")
	}
	if value != "custom-value" {
		t.Errorf("Expected 'custom-value', got '%s'", value)
	}

	value, found, err = provider.Retrieve("another.var")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if !found {
		t.Error("Expected to find another.var")
	}
	if value != "quoted-value" {
		t.Errorf("Expected 'quoted-value', got '%s'", value)
	}
}

func TestDotenvProviderNonexistentFile(t *testing.T) {
	provider := NewDotenvProviderWithPath("/nonexistent/path/.env")

	value, found, err := provider.Retrieve("any.key")
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if found {
		t.Error("Expected not to find any key in nonexistent file")
	}
	if value != "" {
		t.Errorf("Expected empty value, got '%s'", value)
	}
}

func TestFormatEnvKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"simple", "SIMPLE"},
		{"database.url", "DATABASE_URL"},
		{"api-key", "API_KEY"},
		{"complex.key-with-dashes", "COMPLEX_KEY_WITH_DASHES"},
	}

	for _, test := range tests {
		result := FormatEnvKey(test.input)
		if result != test.expected {
			t.Errorf("FormatEnvKey(%s): expected %s, got %s", test.input, test.expected, result)
		}
	}
}

func TestDotenvProviderPriority(t *testing.T) {
	provider := NewDotenvProvider()

	if provider.Priority() != 3 {
		t.Errorf("Expected priority 3, got %d", provider.Priority())
	}
}

func TestDotenvProviderName(t *testing.T) {
	provider := NewDotenvProvider()

	if provider.Name() != "dotenv" {
		t.Errorf("Expected name 'dotenv', got '%s'", provider.Name())
	}
}

func TestDotenvProviderQuoteParsing(t *testing.T) {
	// Create a temporary .env file with various quote scenarios
	tmpDir := t.TempDir()
	envFile := filepath.Join(tmpDir, "quotes.env")

	content := `SIMPLE=unquoted value
DOUBLE_QUOTED="quoted value"
SINGLE_QUOTED='single quoted'
ESCAPED_QUOTES="value with \"escaped\" quotes"
ESCAPED_BACKSLASH="value with \\ backslash"
NEWLINE="line1\nline2"
TAB="col1\tcol2"
EMPTY_QUOTES=""
MIXED_QUOTES="don't mix quotes"
EQUALS_IN_VALUE=key=value=another
NESTED_QUOTES="'foo'"
REVERSE_NESTED='"bar"'
CARRIAGE_RETURN="line1\rline2"`

	err := os.WriteFile(envFile, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to create test .env file: %v", err)
	}

	provider := NewDotenvProviderWithPath(envFile)

	tests := []struct {
		key      string
		expected string
		found    bool
	}{
		{"simple", "unquoted value", true},
		{"double.quoted", "quoted value", true},
		{"single.quoted", "single quoted", true},
		{"escaped.quotes", `value with "escaped" quotes`, true},
		{"escaped.backslash", `value with \ backslash`, true},
		{"newline", "line1\nline2", true},
		{"tab", "col1\tcol2", true},
		{"empty.quotes", "", true},
		{"mixed.quotes", "don't mix quotes", true},
		{"equals.in.value", "key=value=another", true},
		{"nested.quotes", "'foo'", true},
		{"reverse.nested", "\"bar\"", true},
		{"carriage.return", "line1\rline2", true},
	}

	for _, test := range tests {
		value, found, err := provider.Retrieve(test.key)
		if err != nil {
			t.Errorf("Unexpected error for key %s: %v", test.key, err)
			continue
		}

		if found != test.found {
			t.Errorf("For key %s: expected found=%v, got found=%v", test.key, test.found, found)
			continue
		}

		if found && value != test.expected {
			t.Errorf("For key %s: expected value=%q, got value=%q", test.key, test.expected, value)
		}
	}
}
