## Goal

Preserve the PostgreSQL password exactly when loading the repository’s `.env` file, and make configuration loading independent of an accidental working-directory mismatch.

## Approach

Use the existing `github.com/joho/godotenv` loader in the configuration adapter. Quote the connection string with single quotes so its literal `$` is not treated as dotenv variable interpolation. Optionally make the dotenv path explicit only if the application is routinely started outside the repository root; this preserves the current production behavior, where shell-injected variables take precedence.

## File Changes

- **Modify** `.env:5` — wrap `DATABASE_URL` in single quotes to preserve the literal password.
- **Modify (optional)** `internal/platform/config/config.go:21-30` — load a specifically configured development dotenv path if the service is intentionally run from another working directory.
- **Modify (recommended cleanup)** `internal/platform/config/config.go:31-32` — remove printing of the database URL, which exposes credentials in logs.

## Implementation Steps

### Task 1: Correct the dotenv value

1. In `.env:5`, set the value as `DATABASE_URL='postgres://user_metry:erw34$23Ert@localhost:5432/dbmetry?sslmode=disable'`. Single quotes prevent `godotenv` from expanding `$23E`.
2. Start the service from the repository root using `go run ./cmd/api`, which makes the existing `godotenv.Load()` resolve `.env` correctly.

### Task 2: Make loading diagnostics safe

1. In `internal/platform/config/config.go:31-32`, delete `fmt.Print(apiToken)` and its temporary variable so credentials cannot be emitted to stdout.
2. Retain the existing error warning from `godotenv.Load()`, or replace it with a non-sensitive, structured startup log.

### Task 3: Verify behavior

1. Confirm that startup reports successful dotenv loading and does not emit the connection string.
2. Confirm the PostgreSQL connection receives the password containing the literal `$` and can authenticate.
3. Run `go test ./...` after any Go-code cleanup.

## Acceptance Criteria

- The loaded `DATABASE_URL` includes the literal substring `erw34$23Ert`.
- The application started from the repository root finds `.env` without a missing-file warning.
- Startup output never contains `DATABASE_URL` or the database password.
- Existing tests pass after the optional cleanup.

## Verification Steps

- Run `go run ./cmd/api` from the repository root and inspect the dotenv-load message.
- Verify the database connection either succeeds or reports a database-side connectivity issue rather than a malformed/incorrect-password URL.
- Run `go test ./...` if `config.go` is modified.

## Risks & Mitigations

- **Running from another directory:** `godotenv.Load()` searches relative to the process working directory. Mitigate by always starting at the repository root or explicitly supplying the path in configuration.
- **Credential exposure:** the current debug print leaks the full database URL. Mitigate by removing it before sharing logs or deploying.