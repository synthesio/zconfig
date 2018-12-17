package zconfig

import (
	"os"
	"strings"
)

// A Provider that implements the repository.Provider interface.
type ArgsProvider struct {
	Args map[string]string
}

// Init fetch all keys available in the command-line and initialize the provider
// internal storage.
func (p *ArgsProvider) Init() {
	// Initialize the flags map.
	p.Args = make(map[string]string, len(os.Args))

	// For each argument, check if it starts with two dashes. If it does,
	// trim it, split around the first equal sign and set the flag value.
	// If there is no equal sign, and the next argument starts with a
	// double-dash, the flag is added without value, which allows to
	// differentiate between an empty and a non-existing flag.
	for i := 0; i < len(os.Args); i++ {
		arg := os.Args[i]

		if !strings.HasPrefix(arg, "--") {
			continue
		}

		arg = strings.TrimPrefix(arg, "--")
		parts := strings.SplitN(arg, "=", 2)

		if len(parts) == 1 && i+1 < len(os.Args) && !strings.HasPrefix(os.Args[i+1], "--") {
			parts = append(parts, os.Args[i+1])
			i += 1
		}

		parts = append(parts, "") // Avoid out-of-bound errors.
		p.Args[parts[0]] = parts[1]
	}
}

// Retrieve will return the value from the parsed command-line arguments.
// Arguments are parsed the first time the method is called. Arguments are
// expected to be in the form `--key=value` exclusively (for now).
func (p *ArgsProvider) Retrieve(key string) (value string, found bool, err error) {
	value, found = p.Args[key]
	return value, found, nil
}

// Name of the provider.
func (ArgsProvider) Name() string {
	return "args"
}

// Priority of the provider.
func (ArgsProvider) Priority() int {
	return 1
}

// A Provider that implements the repository.Provider interface.
type EnvProvider struct {
	Env map[string]string
}

// Init fetch all keys available in the command-line and initialize the provider
// internal storage.
func (p *EnvProvider) Init() {
	environ := os.Environ()
	// Initialize the flags map.
	p.Env = make(map[string]string, len(environ))

	// For each value, split around the first equal sign and set the
	// environment value.
	for _, value := range environ {
		parts := strings.SplitN(value, "=", 2)
		parts = append(parts, "") // Avoid out-of-bound errors.

		key := p.FormatKey(parts[0])
		p.Env[key] = parts[1]
	}
}

// Retrieve will return the value from the parsed environment variables.
// Variables are parsed the first time the method is called.
func (p *EnvProvider) Retrieve(key string) (value string, found bool, err error) {
	value, found = p.Env[p.FormatKey(key)]
	return value, found, nil
}

// Name of the provider.
func (EnvProvider) Name() string {
	return "env"
}

// Priority of the provider.
func (EnvProvider) Priority() int {
	return 2
}

func (EnvProvider) FormatKey(key string) (env string) {
	env = strings.ToUpper(key)
	env = strings.Replace(env, ".", "_", -1)
	return strings.Replace(env, "-", "_", -1)
}
