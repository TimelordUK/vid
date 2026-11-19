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
| C-002 | Positions in bytes, movement by grapheme, columns by display width with a host `Width` (DESIGN §5.3). **OPEN** until the owner agrees | 1 | planned |
| C-003 | `Editor`: `Feed(keys) → events`, `State()`, `SetText`, `Apply(edit)` for host rewrites (DESIGN §6.1) | 2 | planned |
| C-004 | Registers: unnamed, `0`, `1`–`9`, `-`, `a`–`z`/`A`–`Z`, `_`, `+`/`*` through the Clipboard hook; shareable between editors (DESIGN §6.4) | 2 | planned |
| C-005 | Undo grouped as vim's, cursor restored with the step; PSReadLine's `U` undo-all (DESIGN §6.5) | 2 | planned |
| C-006 | Dot-repeat from recorded keys, with a new count replacing the old (`3.`) | 2 | planned |
| C-007 | Hooks as small optional interfaces with `Func` adapters, matching jqv's editor hooks one for one (DESIGN §6.2) | 2 | planned |
| C-008 | Profiles: line and buffer (DESIGN §6.3); hosts can make their own | 3 | planned |
| C-009 | Undo tree (`g-` `g+`) | — | idea |
| C-010 | Multi-cursor | — | idea |

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
| V-009 | Surround: `cs ds ys`, `S` in visual (classic form) | 7 | idea |
| V-010 | `ctrl+a` / `ctrl+x` increment and decrement a number | 7 | idea |
| V-011 | `gu gU g~` case operators | 7 | idea |
| V-012 | `vv` hands the text to an external editor through the Editor hook | 8 | idea |
| V-013 | Macros (`q`, `@`) | — | idea |
| V-014 | Marks (`m`, `` ` ``, `'`) | — | idea |
| V-015 | A tiny host-routed `:` subset (`:s` on the buffer?) — only if use asks | — | idea |
| V-016 | Insert-mode escape mappings (`jk`), which need time in the engine (DESIGN §5.4) | — | idea |

## Hosts and adapters

| ID | What | Step | Status |
|---|---|---|---|
| H-001 | `cmd/vid run`: text, keys, print the final state; runs case files (DESIGN §7) | 1 | planned |
| H-002 | `cmd/vid` interactive smoke host: buffer, mode, pending command, registers, event log | 5 | planned |
| H-003 | `vid keys` (what the terminal sends and what it's bound to) and `vid bindings` (the keymap as a table) | 5 | planned |
| H-004 | `adapt/teav1`: Bubble Tea v1 keys to `key.Key`, an Editor as a `tea.Model` | 5 | planned |
| H-005 | `render`: the optional reference renderer, ANSI or markers | 5 | planned |
| H-006 | jqv: `[editor] mode = "vi"` in its REPL through the adapter, its F-keys as host actions, the mode in its prompt (jqv Q-051) | 6 | planned |
| H-007 | jqv's TUI code pane on vid's buffer profile (jqv T-029) | 6 | idea |
| H-008 | `vid edit`: stdin to stdout, behind a pwsh key and a zsh widget (RESEARCH §2.1) | 8 | idea |
| H-009 | Replacing PSReadLine through `PSConsoleHostReadLine`: spike options B and C (RESEARCH §2.2) | 8 | idea |
| H-010 | `adapt/teav2` when jqv moves to Bubble Tea v2 (jqv T-028) | — | idea |
| H-011 | Smoke host writes the session out as a case file, so a surprise becomes a test in one key | 5 | idea |

## Syntax

| ID | What | Step | Status |
|---|---|---|---|
| S-001 | `syntax.Tree` / `Node` interface: kind, range, parent, children, roles (DESIGN §10) | 7 | idea |
| S-002 | Structural text objects (`if af ia aa ic is`) and motions on the interface | 7 | idea |
| S-003 | Incremental selection: grow / shrink the visual selection by node | 7 | idea |
| S-004 | jqv provides a `syntax.Tree` over `jqsyntax` | 7 | idea |
| S-005 | Tree-sitter spike with `gotreesitter` (pure Go): jq, PowerShell, bash grammars against the same cases (RESEARCH §1) | 7 | idea |

## Quality

| ID | What | Step | Status |
|---|---|---|---|
| Q-001 | Case-file format and runner (`vidtest`), `-update` to re-bless (DESIGN §8.1) | 0 | planned |
| Q-002 | nvim headless oracle over the case files; skips without nvim, CI installs it (DESIGN §8.2) | 0 | planned |
| Q-003 | Engine fuzz: no panics, valid cursor, undo/redo and dot-repeat properties | 2 | planned |
| Q-004 | Key spike: every key Bubble Tea v1 sends on Windows Terminal and Zellij parses, and what esc costs | 0 | planned |
| Q-005 | Benchmarks and the per-key budget (DESIGN §8.5) | 2 | planned |
| Q-006 | CI on Windows and Linux; `scripts/test.ps1` runs the same | 1 | planned |
| Q-007 | Re-dump PSReadLine's bindings on each PSReadLine release and diff against PARITY.md | — | idea |
