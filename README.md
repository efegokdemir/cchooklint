# cchooklint

[English](README.md) | [日本語](README.ja.md) | [中文](README.zh.md)

> **Unofficial.** Not affiliated with or endorsed by Anthropic.

`cchooklint` is a diagnostic CLI that statically analyzes Claude Code's `.claude/settings.json` hooks configuration on Windows, looking for two specific, verified bugs:

1. **Typos in `Bash`/`PowerShell` matcher tool names** — these silently disable the hook. No error, no warning from Claude Code itself.
2. **Safety-guard-style hooks whose `matcher` only covers `Bash` or `PowerShell`, not both** — the hook silently fails to fire when the other tool is used (a "fail open" gap).

## Why this exists

On Windows, Claude Code can run shell commands through two different tools: the Bash tool (Git Bash) and the PowerShell tool. A hook's `matcher` filters by tool name, so a hook configured with `matcher: "Bash"` only fires when the Bash tool is used — if Claude uses PowerShell instead, the hook is silently skipped.

This is especially dangerous for safety-guard hooks (e.g. ones that block dangerous commands): whoever configured it may not realize that half of their safety net is silently disabled. Both bugs described above were reproduced and verified on a real Windows machine before this project started.

`cchooklint` is read-only: it never modifies your settings files and never triggers hooks to test them at runtime. It only checks the specific `Bash`/`PowerShell` gap described above — it is not a general-purpose tool-name checker.

## Install

```sh
go install github.com/su-fu/cchooklint/cmd/cchooklint@latest
```

(Not yet published to a versioned release — build from source for now.)

## Usage

```sh
cchooklint
```

This scans, in order:

- `.claude/settings.json` (project)
- `.claude/settings.local.json` (project, local overrides)
- `~/.claude/settings.json` (user)

and prints a warning for every hook affected by either bug.

### Example

Given a `.claude/settings.json` with a typo'd matcher and a hook that only covers `Bash`:

```sh
$ cchooklint
[WARN] "Bahs" does not match any known tool name. Did you mean "Bash"?
[WARN] This hook only covers "Bash". It won't fire when the other tool is used. Consider changing it to "Bash|PowerShell".
```

For machine-readable output, use `--format=json`:

```json
{"version":1,"findings":[{"source_file":".claude/settings.json","event":"PreToolUse","severity":"WARN","code":"matcher_typo","args":["Bahs","Bash"]}]}
```

The JSON envelope is versioned and includes stable rule codes plus the raw
message arguments. `--lang` does not change JSON output.

### Language

Output language is selected in this order:

1. `--lang` flag (e.g. `cchooklint --lang=ja`)
2. `CCHOOKLINT_LANG` environment variable
3. `en` (default)

Supported: `en`, `ja`, `zh`. Unsupported values fall back to `en`.

### Exit codes

- `0` — no findings and no tool errors
- `1` — one or more hook findings were reported
- `2` — a settings discovery or loading error occurred

## License

[MIT](LICENSE)
