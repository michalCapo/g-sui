SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c
.ONESHELL:
.DEFAULT_GOAL := help
# Use the shell variable in compound recipes so make -n never runs a release.
export MAKE

.PHONY: help run example check build test test-race test-browser tidy release

# Each invocation owns only this directory. Keep caches and dependencies intact.
define temporary_directory
workdir=$$(mktemp -d "$${TMPDIR:-/tmp}/gsui-$@.XXXXXX")
trap 'rm -rf -- "$$workdir"' EXIT
trap 'exit 130' INT
trap 'exit 143' TERM
trap 'exit 129' HUP
export TMPDIR="$$workdir"
endef

help:
	@printf '%s\n' \
	  'Usage: make <command>' \
	  '' \
	  '  help          Show this help (default)' \
	  '  run           Run the example with Air live reload' \
	  '  example       Run the example once without live reload' \
	  '  check         Run all non-mutating Go checks, build, and tests' \
	  '  build         Build all Go packages' \
	  '  test          Run Go tests (Node enables client runtime tests)' \
	  '  test-race     Run Go tests with the race detector' \
	  '  test-browser  Run Playwright checks against a running example' \
	  '  tidy          Update go.mod and go.sum' \
	  '  release       Create and push the next v1.MINOR.PATCH tag' \
	  '' \
	  'Environment: TMPDIR, Go toolchain variables, NODE_PATH,' \
	  '             GSUI_URL (default http://127.0.0.1:1424), GSUI_BROWSER.' \
	  'Tools: Go; Air for run; Node + Playwright for test-browser.' \
	  'check requires staticcheck and deadcode; gopls and golangci-lint are optional.'

run:
	@$(temporary_directory)
	# Air keeps its own logs/state under the project root.
	air_dir=$$(mktemp -d .gsui-air.XXXXXX)
	trap 'rm -rf -- "$$workdir" "$$air_dir"' EXIT
	printf -v build_cmd 'go build -o %q ./example' "$$workdir/example"
	air \
	  -build.cmd "$$build_cmd" \
	  -build.bin "$$workdir/example" \
	  -build.include_ext "go" \
	  -build.include_dir "ui,example,example/pages" \
	  -build.exclude_dir "tmp,.git,$$air_dir" \
	  -tmp_dir "$$air_dir"

example:
	@$(temporary_directory)
	go run ./example

build:
	@$(temporary_directory)
	go build ./...

test:
	@$(temporary_directory)
	go test ./...

test-race:
	@$(temporary_directory)
	go test -race ./...

test-browser:
	@$(temporary_directory)
	node ui/testdata/browser_runtime.cjs

tidy:
	@$(temporary_directory)
	go mod tidy

check:
	@$(temporary_directory)
	# Collect diagnostics from every check, even when an earlier check fails.
	set +e
	status=0
	GO_FILES=()
	while IFS= read -r -d '' file; do
	  GO_FILES+=("$$file")
	done < <(find . -name '*.go' -not -path './third_party/*' -print0)

	run() {
	  echo "==> $$*"
	  "$$@"
	  local code=$$?
	  if [[ $$code -ne 0 ]]; then
	    status=1
	    echo "!! failed: $$*" >&2
	  fi
	  echo
	}

	# Keep checks non-mutating and fail when go fix would change sources.
	run go fix -diff ./...

	echo "==> gofmt"
	unformatted="$$(gofmt -l "$${GO_FILES[@]}")" || status=1
	if [[ -n "$$unformatted" ]]; then
	  status=1
	  echo "The following Go files are not formatted:" >&2
	  echo "$$unformatted" >&2
	fi
	echo

	run go vet ./...
	run staticcheck ./...

	echo "==> gopls check"
	if command -v gopls >/dev/null 2>&1; then
	  gopls check -severity=hint "$${GO_FILES[@]}"
	  code=$$?
	  if [[ $$code -ne 0 ]]; then
	    status=1
	    echo "!! failed: gopls check" >&2
	  fi
	else
	  echo "gopls not found; skipping"
	fi
	echo

	echo "==> golangci-lint"
	if command -v golangci-lint >/dev/null 2>&1; then
	  golangci-lint run --max-issues-per-linter=0 --max-same-issues=0 ./...
	  code=$$?
	  if [[ $$code -ne 0 ]]; then
	    status=1
	    echo "!! failed: golangci-lint run" >&2
	  fi
	  echo

	  echo "==> golangci-lint revive lsp-style diagnostics"
	  revive_config="$$workdir/revive.yml"
	  printf '%s\n' \
	    'version: "2"' \
	    'linters:' \
	    '  default: none' \
	    '  enable:' \
	    '    - revive' \
	    '  settings:' \
	    '    revive:' \
	    '      rules:' \
	    '        - name: unused-parameter' \
	    '        - name: package-comments' > "$$revive_config"
	  golangci-lint run --config "$$revive_config" --max-issues-per-linter=0 --max-same-issues=0 ./...
	  code=$$?
	  rm -f "$$revive_config"
	  if [[ $$code -ne 0 ]]; then
	    status=1
	    echo "!! failed: golangci-lint revive lsp-style diagnostics" >&2
	  fi
	else
	  echo "golangci-lint not found; skipping"
	fi
	echo

	run deadcode ./...
	run "$$MAKE" --no-print-directory build
	run "$$MAKE" --no-print-directory test

	if [[ $$status -eq 0 ]]; then
	  echo "All checks passed."
	else
	  echo "Checks finished with errors/warnings." >&2
	fi
	exit "$$status"

release:
	@$(temporary_directory)
	# Module path from go.mod
	MODULE_PATH="github.com/michalCapo/g-sui"

	# Get the latest tag matching v1.* pattern (new versioning)
	LATEST_TAG=$$(git tag -l "v1.*" | sort -V | tail -1)

	if [ -z "$$LATEST_TAG" ]; then
	    # No v1.* tags exist, start at v1.1.0
	    NEW_VERSION="v1.1.0"
	    echo "No v1.* tags found. Starting at version $$NEW_VERSION"
	else
	    echo "Latest version: $$LATEST_TAG"

	    # Parse version: vMAJOR.MINOR.PATCH
	    MINOR=$$(echo "$$LATEST_TAG" | cut -d. -f2)
	    PATCH=$$(echo "$$LATEST_TAG" | cut -d. -f3)

	    # Increment patch version
	    NEW_PATCH=$$((PATCH + 1))
	    NEW_VERSION="v1.$$MINOR.$$NEW_PATCH"

	    echo "New version: $$NEW_VERSION"
	fi

	# Check if working tree is clean
	if ! git diff-index --quiet HEAD --; then
	    echo "Error: Working tree has uncommitted changes. Please commit or stash them first."
	    exit 1
	fi

	# Ensure go.mod is tidy
	echo "Ensuring go.mod is tidy..."
	"$$MAKE" --no-print-directory tidy

	# Check if go.mod has uncommitted changes after tidy
	if ! git diff-index --quiet HEAD -- go.mod go.sum; then
	    echo "Warning: go.mod or go.sum has changes after 'go mod tidy'"
	    echo "Please commit these changes before releasing:"
	    git diff --stat go.mod go.sum
	    exit 1
	fi

	# Verify module path matches go.mod
	MODULE_IN_GOMOD=$$(grep "^module " go.mod | awk '{print $$2}')
	if [ "$$MODULE_IN_GOMOD" != "$$MODULE_PATH" ]; then
	    echo "Error: Module path in go.mod ($$MODULE_IN_GOMOD) doesn't match expected ($$MODULE_PATH)"
	    exit 1
	fi

	# Check if we're on the correct branch (optional check)
	CURRENT_BRANCH=$$(git branch --show-current)
	echo "Current branch: $$CURRENT_BRANCH"

	# Ensure we have the latest from origin
	echo "Fetching latest from origin..."
	git fetch origin --tags

	# Create annotated tag
	echo "Creating tag $$NEW_VERSION..."
	git tag -a "$$NEW_VERSION" -m "Release version $$NEW_VERSION"

	# Push tag to origin
	echo "Pushing tag $$NEW_VERSION to origin..."
	git push origin "$$NEW_VERSION"

	# Also ensure the current branch is pushed
	echo "Ensuring current branch is pushed..."
	git push origin "$$CURRENT_BRANCH" --follow-tags || true

	# Register version with Go module proxy so @latest resolves
	echo "Registering version with Go module proxy..."
	GOPROXY=proxy.golang.org go list -m "$$MODULE_PATH@$$NEW_VERSION" || echo "Warning: Failed to register with proxy (may take a few minutes to propagate)"

	echo ""
	echo "Successfully released version $$NEW_VERSION"
	echo "  Tag created and pushed to origin"
	echo ""
	echo "Module can now be used as a dependency:"
	echo "  go get $$MODULE_PATH@$$NEW_VERSION"
	echo ""
	echo "Or in go.mod:"
	echo "  $$MODULE_PATH $$NEW_VERSION"
	echo ""
	echo "Next steps:"
	echo "  - Verify the module can be fetched: go list -m -versions $$MODULE_PATH"
	echo "  - Create a GitHub release: https://github.com/michalCapo/g-sui/releases/new"
	echo "  - Select tag: $$NEW_VERSION"
