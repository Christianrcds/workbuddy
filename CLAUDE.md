# Workbuddy - Product Development Guidelines

## Product Philosophy

This branch treats Workbuddy as a real product.

The objective is no longer to optimize for teaching Go concepts first. The objective is to build a terminal app that is:

- useful
- consistent
- reliable
- easy to extend

Code should still be clean and idiomatic Go, but the main standard is product quality.

## Working Principles

### 1. Prefer complete user-facing behavior

- Implement features end-to-end when practical.
- Avoid leaving duplicate flows in place once a canonical flow exists.
- Reduce ambiguity in command behavior, flags, and output.

### 2. Respect the architecture

- `cmd/` handles CLI parsing and presentation.
- `internal/note/service.go` owns business rules and orchestration.
- `internal/note/repository.go` and sqlc own data access.
- `migrations/` remain the schema source of truth.

### 3. Optimize for maintenance

- Reuse existing service paths instead of duplicating logic in commands.
- Keep SQL explicit.
- Favor testable helpers and small units of behavior.
- Update generated code when sqlc queries change.

### 4. Keep docs and backlog current

- Update `README.md` when public behavior changes.
- Update `todo.md` when priorities are completed, added, or re-ordered.
- Keep examples aligned with the current CLI.

## Expectations For Changes

- Explain intended edits before making them.
- Add tests for behavior changes whenever practical.
- Prefer improving existing commands over adding new overlapping ones.
- Keep error messages actionable for CLI users.
- Preserve backward compatibility only when it materially benefits users.

## Current Product Goals

Focus development on:

1. CRUD completeness
2. Stronger search
3. Better validation and user feedback
4. Faster common workflows
5. More robust integration testing

## Review Criteria

When evaluating changes, pay most attention to:

1. Does this improve the product meaningfully?
2. Is the CLI behavior consistent and intuitive?
3. Is the change tested at the right level?
4. Does the implementation preserve layering and clarity?
5. Are docs and examples updated?
