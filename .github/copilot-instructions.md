# Copilot Instructions for coding-metrics

## Project Overview

This is a **GitHub Action** that generates SVG badges for GitHub coding metrics. It's a Docker-based action written in Go that fetches user stats via GitHub GraphQL/REST APIs and commits a generated SVG to a repository.

## Architecture

**Main Flow** (`src/main.go`):

1. Fetch GitHub user data via REST API (`github_queries.go`)
2. Fetch detailed stats via GraphQL API (`github_graphql.go`)
3. Generate SVG content with stats visualization (`svg_content.go`)
4. Create SVG file (`svg.go`)
5. Commit SVG to repository using GitHub API (`commit.go`)

**Key Components**:

- `github_queries.go` - REST API calls and GraphQL queries for user stats (commits, PRs, issues, etc.)
- `github_graphql.go` - GraphQL client implementation with error handling
- `svg_content.go` - SVG generation using `github.com/twpayne/go-svg` - creates profile section, stats rows, and contribution graphs
- `commit.go` - Uses `github.com/go-github/v61` to commit generated SVG to the target repo

## Development Workflow

**Build & Run**:

- Use `just build` (not `go build` directly) - compiles to `coding-metrics` binary
- Use `just run` for local execution
- Docker build: `just docker-build` (creates multi-stage Alpine image)

**Linting & Quality**:

- `just lint` - golangci-lint (required before commit)
- `just vulncheck` - govulncheck for security
- Pre-commit hooks via Lefthook (`lefthook.yml`) run 10+ checks including gitleaks, prettier, actionlint, zizmor, pinact

**Environment Variables**:
All inputs prefixed with `INPUT_` (GitHub Actions convention):

- `INPUT_GITHUB_TOKEN` - for fetching user data
- `INPUT_WORKFLOW_GITHUB_TOKEN` - for committing changes
- `INPUT_DEBUG` - enables zap development logger
- `INPUT_TEST_MODE` - skips actual commit in `commit.go`
- See `action.yml` for full list

## Go Conventions

**Logging**:

- Uses `go.uber.org/zap` globally (`zap.L()`)
- Fatal errors stop execution immediately (action failure)
- Debug mode controlled by `INPUT_DEBUG` env var

**Error Handling**:

- Most errors are fatal (`zap.L().Fatal()`) since this is a GitHub Action
- GraphQL errors logged with full context
- File operations include `defer` cleanup with error checks

**Security**:

- `#nosec` comments mark intentional security exceptions (e.g., file creation in temp dir)
- All API tokens from environment variables only
- No hardcoded credentials

## Docker Specifics

Two-stage build in `Dockerfile`:

1. **builder** - Go 1.25.3-alpine, uses build caches for `/go/pkg/mod` and `/root/.cache/go-build`
2. **runner** - Alpine 3.22, non-root user (appuser), minimal attack surface

Runs as UID 10001, includes HEALTHCHECK via `pidof`.

## GraphQL Usage Pattern

Queries in `github_queries.go` use pagination (`after: $after` cursor):

```go
QueryGitHubQLAPI(query, variables, &result)
```

Fetches user ID first, then uses it in subsequent queries for commits/PRs/issues across all repos.

## Common Pitfalls

- Don't use `go run ./src` in workflows - use `just run`
- SVG dimensions are hardcoded (1000x380) in `svg.go` - coordinate changes in `svg_content.go` must respect this
- Test mode (`INPUT_TEST_MODE=true`) prevents actual commits - required for local testing
- All Just recipes expect to be run from workspace root
