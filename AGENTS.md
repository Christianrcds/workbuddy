# Workbuddy - Product Development Guidelines

## Product Philosophy

Workbuddy is now a **product-focused CLI application**, not a learning exercise.

The goal is to make it:

- useful for real note-taking and task tracking
- reliable in day-to-day terminal use
- clear and predictable as a CLI interface
- maintainable as the feature set grows

Educational explanations are still welcome, but product quality, consistency, and delivery come first on this branch.

## How To Help

When assisting with this project:

1. **Prioritize usefulness**
   - Prefer features and changes that improve real user workflows.
   - Optimize for practical CLI ergonomics, reliability, and speed.
   - Treat confusing behavior or inconsistent commands as product bugs.

2. **Keep the architecture disciplined**
   - Preserve the current layering: `cmd/` -> `service` -> `repository/sqlc`.
   - Put business rules in the service layer.
   - Keep Cobra commands thin and focused on input/output handling.
   - Keep SQL explicit and typed through sqlc.

3. **Bias toward production-minded decisions**
   - Favor clear command semantics over clever shortcuts.
   - Keep migrations as the source of truth for schema behavior.
   - Prefer stable, testable interfaces over temporary hacks.
   - Improve performance when user-facing flows are affected.

4. **Document behavior changes**
   - Update README and relevant docs whenever public CLI behavior changes.
   - Keep examples current with the real command surface.
   - Record roadmap changes in `todo.md`.

## Code Change Expectations

- Explain what you are changing and why before editing.
- Prefer implementing complete behavior rather than partial scaffolding.
- Avoid adding duplicate command paths or overlapping UX without a strong reason.
- Keep public command names and flags consistent across docs, code, and tests.
- Add or update tests for behavior changes, especially around CLI flows and data access.

## Project Context

### What We’re Building

Workbuddy is a CLI application for:

- local-first note-taking
- task tracking
- tag-based organization
- fast terminal workflows backed by SQLite

### Current Architecture

```
cmd/           -> CLI commands and user-facing behavior
internal/note/ -> Business logic, repositories, sqlc types and queries
migrations/    -> Schema evolution
```

### Technologies In Use

- **Go 1.25.6**
- **SQLite** via `modernc.org/sqlite`
- **sqlc**
- **golang-migrate**
- **Cobra**
- **lipgloss**

## Product Priorities

Use `todo.md` as the main working backlog.

Current priorities:

1. Complete CRUD flows
2. Improve search and filtering
3. Improve reliability and validation
4. Improve performance of list/search data loading
5. Expand integration test coverage
6. Add user-facing polish without weakening the command model

## Review Focus

When reviewing changes, prioritize:

1. Command consistency
2. Data integrity
3. Error handling and user feedback
4. Test coverage of real workflows
5. Performance of common commands
6. Simplicity of the implementation

## Preferred Interaction Style

- Be direct and concrete.
- Explain tradeoffs when they matter.
- Point out risks, regressions, and unclear behavior.
- Avoid treating this branch like a tutorial project.
- Keep the focus on building a tool that people could actually use.
