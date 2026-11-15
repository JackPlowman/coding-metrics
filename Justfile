# ------------------------------------------------------------------------------
# General
# ------------------------------------------------------------------------------

export SRC_DIR := "./src"
export SRC_RECURSIVE := "./src/..."

build:
    go build -o coding-metrics ${SRC_DIR}

run:
    go run ${SRC_DIR}

alias fmt := lint-fix
alias fmt-check := lint

lint:
    golangci-lint run ${SRC_RECURSIVE}

lint-fix:
    golangci-lint run --fix ${SRC_RECURSIVE}

vulncheck:
    govulncheck ${SRC_RECURSIVE}

# ------------------------------------------------------------------------------
# Docker
# ------------------------------------------------------------------------------

docker-build:
    docker buildx build -t coding-metrics:latest .

# ------------------------------------------------------------------------------
# Prettier
# ------------------------------------------------------------------------------

# Check all files with prettier
prettier-check:
    prettier . --check

# Format all files with prettier
prettier-format:
    prettier . --check --write

# ------------------------------------------------------------------------------
# Justfile
# ------------------------------------------------------------------------------

# Format Justfile
format:
    just --fmt --unstable

# Check Justfile formatting
format-check:
    just --fmt --check --unstable

# ------------------------------------------------------------------------------
# Gitleaks
# ------------------------------------------------------------------------------

# Run gitleaks detection
gitleaks-detect:
    gitleaks detect --source .

# ------------------------------------------------------------------------------
# Prek
# ------------------------------------------------------------------------------

# Run prek checking on all pre-commit config files
prek-check:
    prek validate-config .pre-commit-config.yaml

# ------------------------------------------------------------------------------
# Zizmor
# ------------------------------------------------------------------------------

# Run zizmor checking
zizmor-check:
    uvx zizmor . --persona=auditor

# ------------------------------------------------------------------------------
# Actionlint
# ------------------------------------------------------------------------------

# Run actionlint checks
actionlint-check:
    actionlint

# ------------------------------------------------------------------------------
# Pinact
# ------------------------------------------------------------------------------

# Generate the GitHub Actions dependency lockfile
actions-lock:
    gh actions-lock

# Verify the workflow dependency lockfile against GitHub
actions-lock-check:
    gh actions-lock --verify

# Refresh workflow dependency pins
actions-lock-update:
    gh actions-lock --relock

# ------------------------------------------------------------------------------
# EditorConfig
# ------------------------------------------------------------------------------

# Check files format with EditorConfig
editorconfig-check:
    editorconfig-checker

# ------------------------------------------------------------------------------
# Git Hooks
# ------------------------------------------------------------------------------

# Install git hooks using prek
install-git-hooks:
    prek install

# ------------------------------------------------------------------------------
# Prek
# ------------------------------------------------------------------------------

# Update prek hooks and additional dependencies
prek-update:
    just prek-update-hooks
    just prek-update-additional-dependencies

# Prek update hooks
prek-update-hooks:
    prek update --freeze

prek-update-additional-dependencies:
    uv run --script https://raw.githubusercontent.com/JackPlowman/update-prek-additional-dependencies/refs/heads/main/update_prek_additional_dependencies.py

# ------------------------------------------------------------------------------
# Update All Tools
# ------------------------------------------------------------------------------

# Update all tools
update:
    just actions-lock-update
    just prek-update
