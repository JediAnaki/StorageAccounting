# StorageAccounting Development Guidelines

Auto-generated from all feature plans. Last updated: 2026-03-13

## Active Technologies

Go

## Project Structure

```text
.
```

## Commands

go test ./...

## Code Style

Keep the codebase idiomatic for Go and prefer the standard library when possible.

## Recent Changes

Initialized Spec Kit and Codex prompt files.

<!-- MANUAL ADDITIONS START -->
## Speckit Command Router

When a user message starts with one of these commands, treat it as a Spec Kit command invocation rather than plain text:

- `/speckit`
- `/speckit.specify`
- `/speckit.clarify`
- `/speckit.plan`
- `/speckit.tasks`
- `/speckit.implement`
- `/speckit.analyze`
- `/speckit.checklist`
- `/speckit.constitution`
- `/speckit.taskstoissues`

Routing rules:

- Load and follow the matching prompt file from `.codex/prompts/`.
- `/speckit` routes to `.codex/prompts/speckit.md`.
- `/speckit.specify` routes to `.codex/prompts/speckit.specify.md`.
- `/speckit.clarify` routes to `.codex/prompts/speckit.clarify.md`.
- `/speckit.plan` routes to `.codex/prompts/speckit.plan.md`.
- `/speckit.tasks` routes to `.codex/prompts/speckit.tasks.md`.
- `/speckit.implement` routes to `.codex/prompts/speckit.implement.md`.
- `/speckit.analyze` routes to `.codex/prompts/speckit.analyze.md`.
- `/speckit.checklist` routes to `.codex/prompts/speckit.checklist.md`.
- `/speckit.constitution` routes to `.codex/prompts/speckit.constitution.md`.
- `/speckit.taskstoissues` routes to `.codex/prompts/speckit.taskstoissues.md`.

Argument handling:

- Everything after the command name is the command arguments.
- Treat those arguments as the `$ARGUMENTS` value expected by the prompt file.
- If `/speckit` is called without arguments, briefly list the available Speckit commands.
- If an unknown `/speckit...` variant is used, list the supported commands instead of ignoring it.
<!-- MANUAL ADDITIONS END -->
