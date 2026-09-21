# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## What this project is

`cchooklint` is a diagnostic CLI (unofficial, not affiliated with Anthropic) that statically analyzes Claude Code's `.claude/settings.json` hooks configuration on Windows, looking for two specific, verified bugs:

1. Typos in `Bash`/`PowerShell` matcher tool names (silently disables the hook, no error)
2. Safety-guard-style hooks whose `matcher` only covers `Bash` or `PowerShell`, not both — meaning the guard silently fails to fire when the other tool is used (a "fail open" gap)

Both bugs were personally reproduced and verified before this project started (see the investigation log referenced in `docs/design.md`). This is a personal side project whose goal is to (eventually) generate a small amount of revenue to offset the user's own Claude Code subscription cost — not a company product, no external deadline pressure.

**Read `docs/design.md` first** for the full background, v1 scope, architecture, rule definitions, i18n plan, naming rationale, and milestone breakdown. **Read `docs/progress.md`** for what has actually been done so far and what the next milestone is — treat its latest entry as the current state, not this file.

## How to collaborate with the user on this repo

- The user has only light Go experience and is deliberately using this project to learn Go. **Do not write full implementations for them.** Give hints, point at relevant standard-library packages/docs, explain the Go concept involved, and let them write the code. Only write code directly if they explicitly ask you to (e.g., for docs, config files, or something they've said is not the part they want to practice).
- Work in small increments tied to the milestones in `docs/design.md`. Don't jump ahead to a later milestone's design without checking in first.
- After a milestone (or a meaningful chunk of one) is completed, add a dated entry to `docs/progress.md` summarizing what was built and, importantly, what Go concept(s) the user practiced or learned — this is a deliberate learning log, not just a changelog.
- If a design decision made in `docs/design.md` needs to change based on what's learned during implementation, update `docs/design.md` directly (it's documented as a living document) rather than letting the doc drift out of sync with the code.
- Output strings for the CLI itself must go through the planned i18n layer (see `docs/design.md`), not be hardcoded inline, even in early scaffolding — retrofitting i18n later means rewriting every message call site.
