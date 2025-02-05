run:
  timeout: 5m
  modules-download-mode: readonly

linters:
  disable-all: true
  fast: false
  enable:
    - bodyclose # Checks whether HTTP response body is closed.
    - contextcheck # Check whether the function uses a non-inherited context.
    - durationcheck # Checks for common mistakes when working with time.Duration.
    - errcheck # Check that error return values are used.
    - exportloopref # Checks for references to the loop iterator variable from within the loop body.
    - godot # Check if comments end in a period.
    - goconst # Finds repeated strings that could be replaced by a constant.
    - gocyclo # Computes the cyclomatic complexity of functions.
    - gofmt # Checks if the code is gofmted.
    - gofumpt # Checks if the code is gofumpted.
    - goimports # Checks if the code is goimported.
    - gosec # Inspects source code for security problems by scanning the Go AST.
    - gosimple # Linter for Go source code that specializes in simplifying code.
    - govet # Vet examines Go source code and reports suspicious constructs. It is roughly the same as 'go vet' and uses its passes.
    - ineffassign # Detects when assignments to existing variables are not used.
    - mnd # Check for magic numbers.
    - misspell # Finds commonly misspelled English words in comments.
    - nilerr # Check for errors compared against nil.
    - nilnil
    - prealloc # Finds slice declarations that could potentially be preallocated.
    - predeclared # Check for shadowing of predeclared identifiers.
    - promlinter # Check Prometheus metrics naming via promlint.
    - revive # Drop-in replacement of golint.
    - rowserrcheck # Check for rows.Err() != nil without Next() call.
    - sqlclosecheck # Check for proper defer of sql.Rows.Close.
    - staticcheck # Statically detect bugs, both obvious and subtle ones.
    # - tagliatelle # Check for tagliatelle.
    - unused # Check for unused code.
    - unconvert # Remove unnecessary type conversions.
    - wastedassign # Check for useless assignments.
    # - whitespace # Check for leading and trailing white space.

linters-settings:
  whitespace:
    multi-func: true
  goconst:
    ignore-tests: true
  staticcheck:
    # skip custom json tag
    checks: ["-SA5008"]