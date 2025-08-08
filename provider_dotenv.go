package zconfig

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// A Provider that implements the repository.Provider interface for dotenv files.
type DotenvProvider struct {
	vars map[string]string
}

// NewDotenvProvider creates a provider that loads environment variables from a .env file
// in the current directory. If the file doesn't exist, the provider will be empty but still functional.
func NewDotenvProvider() *DotenvProvider {
	return NewDotenvProviderWithPath(".env")
}

// NewDotenvProviderWithPath creates a provider that loads environment variables from the specified .env file.
// If the file doesn't exist, the provider will be empty but still functional.
func NewDotenvProviderWithPath(path string) *DotenvProvider {
	p := &DotenvProvider{
		vars: make(map[string]string),
	}

	// Make path absolute to avoid issues with working directory changes
	if !filepath.IsAbs(path) {
		if abs, err := filepath.Abs(path); err == nil {
			path = abs
		}
	}

	p.loadFile(path)
	return p
}

// loadFile loads variables from the specified dotenv file.
func (p *DotenvProvider) loadFile(path string) {
	file, err := os.Open(path)
	if err != nil {
		// File doesn't exist or can't be opened, but that's okay
		return
	}
	defer func() {
		_ = file.Close()
	}()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse KEY=VALUE format
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}

		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)

		// Parse quoted values with basic unquoting
		value = p.unquoteValue(value)

		p.vars[key] = value
	}
}

// unquoteValue removes surrounding quotes and handles basic escape sequences.
// Note: This is a simplified implementation. For full shell-like quoting,
// consider using strconv.Unquote or a more sophisticated parser.
func (p *DotenvProvider) unquoteValue(value string) string {
	if len(value) < 2 {
		return value
	}

	// Handle double quotes
	if strings.HasPrefix(value, `"`) && strings.HasSuffix(value, `"`) {
		inner := value[1 : len(value)-1]
		// Use strings.Replacer for more efficient escape sequence handling
		replacer := strings.NewReplacer(
			`\"`, `"`,
			`\\`, `\`,
			`\n`, "\n",
			`\t`, "\t",
			`\r`, "\r",
		)
		return replacer.Replace(inner)
	}

	// Handle single quotes (no escape sequences in single quotes)
	if strings.HasPrefix(value, `'`) && strings.HasSuffix(value, `'`) {
		return value[1 : len(value)-1]
	}

	return value
}

// Retrieve will return the value from the loaded dotenv variables.
func (p *DotenvProvider) Retrieve(key string) (value interface{}, found bool, err error) {
	// Use the same key formatting as EnvProvider for consistency
	envKey := FormatEnvKey(key)
	value, found = p.vars[envKey]
	return value, found, nil
}

// Name of the provider.
func (p *DotenvProvider) Name() string {
	return "dotenv"
}

// Priority of the provider. Set to 3 so it comes after args (1) and env (2).
func (p *DotenvProvider) Priority() int {
	return 3
}
