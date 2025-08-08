# GitHub Actions Workflows

## CI Pipeline (`ci.yml`)

This workflow runs comprehensive checks on every pull request and push to main/master branches:

### Jobs

1. **Test** (`test`)
   - Runs on multiple Go versions (1.20-1.24)
   - Executes all tests with race detection (`-race`)

2. **Lint** (`lint`)
   - Uses golangci-lint with custom configuration
   - Checks code quality, style, and potential issues
   - Configuration in `.golangci.yml`

3. **Format Check** (`format`)
   - Validates code formatting with `goimports`
   - Ensures consistent code style
   - Fails if any files need formatting

### Triggers

- **Pull Requests**: All jobs run on PRs to main/master
- **Push**: All jobs run on direct pushes to main/master
- **Matrix Strategy**: Tests run on multiple Go versions to ensure compatibility

### Configuration Files

- `.golangci.yml`: Linting rules and configuration
- `go.mod`: Go version and dependencies
- Coverage reports: Uploaded to Codecov for tracking

This ensures code quality, compatibility, and proper formatting across all contributions.