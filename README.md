# deptest

Test your Go library changes against real dependents.

## What It Does

`deptest` helps Go library maintainers understand the impact of their changes by:
1. Discovering projects that depend on your module
2. Running their test suites against different versions of your code
3. Comparing results to show what broke or got fixed

## Installation

```bash
go build -o deptest ./cmd/deptest
```

Or install directly:

```bash
go install ./cmd/deptest
```

## Usage

### 1. Discover Dependents

Find projects that depend on your module:

```bash
./deptest discover github.com/sirupsen/logrus -limit 5 -o deps.json
```

This will:
- Search pkg.go.dev for dependents
- Limit results to top 5 projects
- Save to `deps.json`

### 2. Run Tests (Before Changes)

Test the current version:

```bash
./deptest test -i deps.json -o results-before.json
```

This will:
- Clone each dependent repository
- Run `go test ./...`
- Save results with pass/fail status

### 3. Make Your Changes

Modify your library code, then run tests again:

```bash
./deptest test -i deps.json -o results-after.json
```

### 4. Compare Results

See the impact of your changes:

```bash
./deptest compare results-before.json results-after.json
```

Output example:
```
Impact Analysis
Total projects analyzed: 5
Status changes: 2

NEWLY BROKEN (2):
  - github.com/example/project-a
  - github.com/example/project-b

Still passing: 3
Still failing: 0

This change breaks existing dependents!
```

## Options

### discover
- `-limit N`: Maximum number of dependents to fetch (default: 10)
- `-o file`: Output file (default: dependents.json)

### test
- `-i file`: Input dependents file (required)
- `-o file`: Output results file (default: results.json)
- `-timeout duration`: Timeout per project (default: 2m)

## Example Workflow

```bash
# Discover dependents of logrus
./deptest discover github.com/sirupsen/logrus -limit 5 -o deps.json

# Test current version
./deptest test -i deps.json -o before.json

# Make changes to your local copy of the library
# (You would typically use 'replace' in go.mod)

# Test with changes
./deptest test -i deps.json -o after.json

# Compare
./deptest compare before.json after.json
```

## Limitations (Prototype)

This is a prototype. Known limitations:

- **Security**: Tests run directly on host (no sandboxing)
- **Scale**: Limited to small number of dependents
- **Discovery**: Uses HTML parsing (unofficial API)
- **Performance**: Sequential testing only

## Future Enhancements

- Docker/container isolation for safe test execution
- Parallel test execution with concurrency control
- Better discovery using deps.dev API
- Caching of clone/test results
- Support for different Go versions
- Web UI for results visualization

## Architecture

```
deptest/
├── cmd/deptest/       # CLI entry point
├── pkg/
│   ├── discovery/     # Fetch dependents from pkg.go.dev
│   ├── runner/        # Clone repos and run tests
│   └── compare/       # Compare test results
```
