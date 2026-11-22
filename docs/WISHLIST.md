# vid — wishlist

Ideas and work items. IDs are never renumbered; an item that's no longer wanted is marked
`dropped`, not deleted. Categories: `C-` core (text, engine, registers, undo, API), `V-` vi
grammar (motions, operators, text objects, modes, keymaps), `H-` hosts and adapters (smoke
host, Bubble Tea, jqv, shells), `S-` syntax (the tree interface, providers, tree-sitter),
`Q-` quality (tests, oracle, fuzz, benchmarks, CI).

Step = DESIGN §11. Status: `idea` · `planned` · `doing` · `done` · `dropped`.

## Core

| ID | What | Step | Status |
|---|---|---|---|
| C-001 | `text.Buffer`: a string and a line index; every change an `Edit{From, To, Text}` (DESIGN §5.2) | 1 | planned |
| C-002 | Positions in bytes, movement by grapheme, columns by display width with a host `Width` (DESIGN §5.3); agreed 2026-11-21 | 1 | planned |
| C-003 | `Editor`: `Feed(keys) → events`, `State()`, `SetText`, `Apply(edit)` for host rewrites (DESIGN §6.1) | 2 | planned |
| C-004 | Registers: unnamed, `0`, `1`–`9`, `-`, `a`–`z`/`A`–`Z`, `_`, `+`/`*` through the Clipboard hook; shareable between editors (DESIGN §6.4) | 2 | planned |
| C-005 | Undo grouped as vim's, cursor restored with the step; PSReadLine's `U` undo-all (DESIGN §6.5) | 2 | planned |
| C-006 | Dot-repeat from recorded keys, with a new count replacing the old (`3.`) | 2 | planned |
| C-007 | Hooks as small optional interfaces with `Func` adapters, matching jqv's editor hooks one for one (DESIGN §6.2) | 2 | planned |
| C-008 | Profiles: line and buffer (DESIGN §6.3); hosts can make their own | 3 | planned |
| C-009 | Undo tree (`g-` `g+`) | — | idea |
| C-010 | Multi-cursor | — | idea |
| C-011 | Viewport: `SetViewport(top, height)`; vid scrolls it and the host follows (DESIGN §3.1) | 3 | planned |
| C-012 | Extension API: `vid.Extension`, `Registry` (motions, text objects, operators, actions, default bindings, spans), continuations for extra keys; the core's text objects register through it; clashes an error at `New` (DESIGN §14) | 4 | planned |
| C-013 | Language-service hooks: `Formatter`, `Definitions`, `Symbols`, `Diagnostics` (DESIGN §15) | 6 | planned |
| C-014 | Jumplist: `ctrl+o` and its forward key; `gd`, symbol jumps, `G`, searches push (DESIGN §15) | 6 | planned |
| C-015 | State and events kept plain, serialisable data: no functions or channels (DESIGN §5.1) | 2 | planned |

## Vi grammar

| ID | What | Step | Status |
|---|---|---|---|
| V-001 | `key`: keys and vim notation, parse and format, round-trip fuzzed | 1 | planned |
| V-002 | Motions: `h l w b e ge W B E gE 0 ^ _ $ \| f F t T ; , %` and ctrl+] | 1 | planned |
| V-003 | Operators `d c y` with any motion, counts on both sides; `x X s S D C Y p P r R ~ J i a I A` | 2 | planned |
| V-004 | Insert mode with the shell's keys (ctrl+w, ctrl+u, ctrl+h, ctrl+delete, ctrl+arrows) | 2 | planned |
| V-005 | Lines: `j k + - gg G o O`, linewise operators | 3 | planned |
| V-006 | Visual mode `v` `V`, `o` to swap ends | 4 | planned |
| V-007 | Text objects: `iw aw iW aW`, quotes, brackets (`ib iB i[ i<`), `ip ap` | 4 | planned |
| V-008 | Keymaps as data: `keymap.PSReadLine()`, `keymap.ZshViMode()`, `keymap.Vim()`; host actions bound by name | 2 | planned |
| V-009 | Surround: `cs ds ys`, `S` in visual (classic form), as extension `ext/surround`, oracle against vim-surround pinned (DESIGN §14) | 7 | idea |
| V-010 | `ctrl+a` / `ctrl+x` increment and decrement a number | 7 | idea |
| V-011 | `gu gU g~` case operators | 7 | idea |
| V-012 | `vv` hands the text to an external editor through the Editor hook; in the pwsh host, the room key (H-015) may answer `vv` instead | 8 | idea |
| V-013 | Macros (`q`, `@`) | — | idea |
| V-014 | Marks (`m`, `` ` ``, `'`) | — | idea |
| V-015 | A tiny host-routed `:` subset (`:s` on the buffer?) — only if use asks | — | idea |
| V-016 | Insert-mode escape mappings (`jk`), which need time in the engine (DESIGN §5.4) | — | idea |
| V-017 | Jump labels, flash.nvim / leap style, as extension `ext/jump`: labels as spans over the viewport, usable as a motion (`d` + label) (DESIGN §14) | 7 | idea |
| V-018 | More text objects, mini.ai style, as an extension: arguments, function calls, without a tree (DESIGN §14) | 7 | idea |
| V-019 | `gc` / `gcc` comment toggle, core as in nvim 0.10+, the comment string from the host | 4 | idea |
| V-020 | `=` and `gq` format operators over the `Formatter` hook (DESIGN §15) | 6 | planned |
| V-021 | Buffer profile: `ctrl+d ctrl+u ctrl+e ctrl+y`, `zz zt zb`, `H M L`, scrolloff; buffer search `/ ? n N * #` | 6 | planned |
| V-022 | `gd`, `]d` `[d`, `K`; `[[` `]]` over symbols (**OPEN**, an `oracle: skip`) (DESIGN §15) | 6 | planned |

## Hosts and adapters

| ID | What | Step | Status |
|---|---|---|---|
| H-001 | `cmd/vid run`: text, keys, print the final state; runs case files (DESIGN §7) | 1 | planned |
| H-002 | `cmd/vid` interactive smoke host: buffer, mode, pending command, registers, event log | 5 | planned |
| H-003 | `vid keys` (what the terminal sends and what it's bound to) and `vid bindings` (the keymap as a table) | 5 | planned |
| H-004 | `adapt/teav1`: Bubble Tea v1 keys to `key.Key`, an Editor as a `tea.Model` | 5 | planned |
| H-005 | `render`: inline, multi-line, soft wrap, a scrolling region, resize; ANSI or markers; a real component for the pwsh host (DESIGN §5.1, §16) | 5 | planned |
| H-006 | jqv: `[editor] mode = "vi"` in its REPL through the adapter, its F-keys as host actions, the mode in its prompt (jqv Q-051) | 6 | planned |
| H-007 | jqv's code pane on vid's buffer profile (jqv Q-049 → T-029): viewport, search, language services — the "mini nvim" (DESIGN §15) | 6 | planned |
| H-008 | `vid edit`: stdin to stdout, behind a zsh widget or a pwsh key; no longer the pwsh story (DESIGN §16) | — | idea |
| H-009 | Replacing PSReadLine through `PSConsoleHostReadLine`: spike options B and C (RESEARCH §2.2). Superseded 2026-11-21: option B chosen, see H-012, H-013 | — | dropped |
| H-010 | `adapt/teav2` when jqv moves to Bubble Tea v2 (jqv T-028) | — | idea |
| H-011 | Smoke host writes the session out as a case file on F9, so a surprise becomes a test in one key | 5 | idea |
| H-012 | pwsh line host spike: a stub `vid serve` behind `PSConsoleHostReadLine`; latency, console handover, prompt position, resize, one Tab round trip, fallback to PSReadLine (DESIGN §16.4) | 0 | planned |
| H-013 | `vid serve` and the pwsh module: resident as `hit serve`, requests over the pipe (completion, completeness, history), PSReadLine history file, fallback to PSReadLine's ReadLine (DESIGN §16.2) | 8 | idea |
| H-014 | The multi-line feel in the smoke host with a fake `Acceptor`, for the owner to try before the pwsh host exists (DESIGN §16.3) | 5 | idea |
| H-015 | The room key: the inline region grows into a tall pane for the same buffer, collapses on accept; key **OPEN** (DESIGN §16.3) | 8 | idea |
| H-016 | hit's keys (Ctrl+R, Up search) reachable inside the pwsh host, probably host actions calling `hit serve` (DESIGN §16.4) | 8 | idea |
| H-017 | An LSP client behind the language-service hooks, as a host's adapter (DESIGN §15) | — | idea |

## Syntax

| ID | What | Step | Status |
|---|---|---|---|
| S-001 | `syntax.Tree` / `Node` interface: kind, range, parent, children, roles (DESIGN §10) | 7 | idea |
| S-002 | Structural text objects (`if af ia aa ic is`) and motions on the interface | 7 | idea |
| S-003 | Incremental selection: grow / shrink the visual selection by node, as an extension (DESIGN §14) | 7 | idea |
| S-004 | jqv provides a `syntax.Tree` over `jqsyntax` | 7 | idea |
| S-005 | Tree-sitter spike with `gotreesitter` (pure Go): jq, PowerShell, bash grammars against the same cases (RESEARCH §1) | 7 | idea |
| S-006 | A pwsh lexer in Go for colour in the pwsh host: hand-written and tolerant, or tree-sitter-powershell via S-005; per key, inside the key budget (DESIGN §16.2) | 8 | idea |

## Quality

| ID | What | Step | Status |
|---|---|---|---|
| Q-001 | Case-file format and runner (`vidtest`), `-update` to re-bless (DESIGN §8.1). Format, parser, canonical writer done 2026-11-21; the engine runner waits for an engine | 0 | doing |
| Q-002 | nvim oracle over the case files: one `nvim --embed --clean` per run over msgpack-RPC, in its own module; nvim's defaults are the reference; skips without nvim, CI installs it (DESIGN §8.2). Built 2026-11-21 with `-bless`; in CI with Q-006 | 0 | doing |
| Q-003 | Engine fuzz: no panics, valid cursor, undo/redo and dot-repeat properties | 2 | planned |
| Q-004 | Key spike: every key Bubble Tea v1 sends on Windows Terminal and Zellij parses, and what esc costs | 0 | planned |
| Q-005 | Benchmarks and the per-key budget (DESIGN §8.5) | 2 | planned |
| Q-006 | CI on Windows and Linux; `scripts/test.ps1` runs the same. Built 2026-11-22: both modules, nvim pinned at v0.12.6, `VID_REQUIRE_NVIM` makes a missing nvim fail instead of skip; done once green on GitHub | 1 | doing |
| Q-007 | Re-dump PSReadLine's bindings on each PSReadLine release and diff against PARITY.md | — | idea |
| Q-008 | Plugin oracle: `oracle/plugins.lock` pins source plugins at a commit, a script fetches them, a case names its plugin and the oracle loads only that one (DESIGN §14.4) | 7 | idea |
| Q-009 | Extension isolation: core cases with no extensions installed, a second run with all of them; each extension's own cases with only itself (DESIGN §14.4) | 4 | planned |
| Q-010 | A fresh nvim for a case that tags itself so (`.` with nothing to repeat, the `.` register), which the shared nvim cannot reset (DESIGN §8.2) | — | idea |
| Q-011 | Differential fuzz against nvim: random key strings over random texts through vid and nvim side by side; a disagreement is written out as a case file to judge (spec or bug) | 2 | idea |
