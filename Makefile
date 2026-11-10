SHELL := /bin/bash
.SHELLFLAGS := -eu -o pipefail -c
.ONESHELL:
.DEFAULT_GOAL := help
# Use the shell variable in compound recipes so make -n never runs a release.
export MAKE

# Go module folders, relative to this file. Example: . app tools
GO_MODULES := .
# Target systems for vet, lint, dead code and build. Example: linux windows
# Tests run only when this machine's system is in the list.
GO_OSES := $(shell go env GOOS)
# app: dead code fails check. lib: dead code is a warning, exported API can look unused.
GO_KIND := lib
LINT_CONFIG := $(CURDIR)/.golangci.yml

.PHONY: help run example check check-extra check-security tools build test test-race test-browser tidy release

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
	  '  check         Run all checks without changing files' \
	  '  check-extra   Project-only checks, run by make check' \
	  '  check-security  Find known security problems in dependencies (needs internet)' \
	  '  tools         Install the check tools' \
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
	  'check requires staticcheck, deadcode and golangci-lint (make tools); gopls is optional.'

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

check: ## Run all checks without changing files
	@set +e
	status=0
	run() {
	  echo "==> $$*"
	  if ! "$$@"; then
	    status=1
	    echo "!! failed: $$*" >&2
	  fi
	  echo
	}

	for tool in go gofmt staticcheck deadcode golangci-lint; do
	  if ! command -v "$$tool" >/dev/null 2>&1; then
	    status=1
	    echo "!! missing tool: $$tool (run make tools)" >&2
	  fi
	done
	[[ $$status -eq 0 ]] || exit 1

	echo "==> gofmt -l"
	go_files=()
	while IFS= read -r -d '' file; do
	  go_files+=("$$file")
	done < <(find . -name '*.go' -not -path '*/vendor/*' -not -path '*/third_party/*' -not -path '*/node_modules/*' -not -path '*/testdata/*' -print0)
	if ! unformatted=$$(gofmt -l "$${go_files[@]}"); then
	  status=1
	  echo "!! failed: gofmt" >&2
	fi
	if [[ -n $$unformatted ]]; then
	  status=1
	  printf 'Not formatted:\n%s\n' "$$unformatted" >&2
	fi
	echo

	for mod in $(GO_MODULES); do
	  pushd "$$mod" >/dev/null
	  echo "### module $$mod"
	  echo
	  run go mod tidy -diff
	  run go fix -diff ./...

	  # Code is dead only when no target system uses it.
	  dead=""
	  first=1
	  for os in $(GO_OSES); do
	    run env GOOS="$$os" go vet ./...
	    run env GOOS="$$os" staticcheck ./...
	    run env GOOS="$$os" golangci-lint run --config "$(LINT_CONFIG)" ./...
	    run env GOOS="$$os" go build -o /dev/null ./...

	    echo "==> GOOS=$$os deadcode -test ./..."
	    if ! found=$$(GOOS="$$os" deadcode -test ./...); then
	      status=1
	      echo "!! failed: deadcode" >&2
	    fi
	    echo
	    if [[ $$first -eq 1 ]]; then
	      dead=$$found
	      first=0
	    else
	      dead=$$(comm -12 <(sort <<<"$$dead") <(sort <<<"$$found"))
	    fi
	  done
	  if [[ -n $$dead ]]; then
	    if [[ "$(GO_KIND)" == lib ]]; then
	      printf 'Warning, unused code (check before removing public API):\n%s\n\n' "$$dead"
	    else
	      status=1
	      printf 'Dead code:\n%s\n\n' "$$dead" >&2
	    fi
	  fi

	  if [[ " $(GO_OSES) " == *" $$(go env GOOS) "* ]]; then
	    run go test ./...
	  else
	    printf '==> go test skipped: %s is not in GO_OSES\n\n' "$$(go env GOOS)"
	  fi
	  popd >/dev/null
	done

	run "$(MAKE)" --no-print-directory check-extra

	if [[ $$status -eq 0 ]]; then
	  echo "All checks passed."
	else
	  echo "Checks finished with errors." >&2
	fi
	exit "$$status"

check-extra: ## Project-only checks, run by make check
	@echo "==> gopls check"
	if ! command -v gopls >/dev/null 2>&1; then
	  echo "gopls not found; skipping"
	  exit 0
	fi
	go_files=()
	while IFS= read -r -d '' file; do
	  go_files+=("$$file")
	done < <(find . -name '*.go' -not -path './third_party/*' -print0)
	gopls check -severity=hint "$${go_files[@]}"

check-security: ## Find known security problems in dependencies (needs internet)
	@set +e
	status=0
	for mod in $(GO_MODULES); do
	  for os in $(GO_OSES); do
	    echo "==> $$mod GOOS=$$os govulncheck ./..."
	    (cd "$$mod" && GOOS="$$os" govulncheck ./...) || status=1
	    echo
	  done
	done
	exit "$$status"

tools: ## Install the check tools
	go install honnef.co/go/tools/cmd/staticcheck@latest
	go install golang.org/x/tools/cmd/deadcode@latest
	go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest

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
