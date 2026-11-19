# vid — design

> **vid** = **vi** e**d**itor (working name). A vi-mode buffer editor as a Go library: modes,
> motions, operators, text objects, registers, undo and dot-repeat over a buffer, with no
> terminal and no renderer of its own. Any Go program can host it; jqv is the first.

This is a living document. Settled decisions are stated plainly; open questions are marked
**OPEN** and tracked in [WISHLIST.md](WISHLIST.md). Everything below is a **proposal
(2026-11-19)** until the owner agrees it section by section, as in jqv and hit.

---

## 1. Why

The owner lives in vi mode: `Set-PSReadLineOption -EditMode vi` in pwsh, vi keys in jqv's
tree, zsh's vi mode on Linux. Each of those is someone else's vi, each a different subset,
and none can be reached into:

| Where | What's good | What hurts |
|---|---|---|
| PSReadLine vi mode | Built in, multi-line, `.` repeat, counts, `v` to an editor | Fixed set of bindings (a few text objects, no visual mode, no surround); not reusable outside pwsh; C# |
| zsh-vi-mode (jeffreytse) | Visual mode, surround, text objects, cursor shape per mode, `vv` to an editor | Pure zsh widgets on zle; tied to zsh |
| zsh's own vi mode | Always there | Bare: no surround, thin text objects, no visual line feedback |
| jqv's `internal/editor` | Ours, tested through the model, hosts plug in (Painter, Completer, Continues…) | Emacs keys only; one package inside jqv |
| A real nvim embedded | Everything | A process per editor, a huge surface, an opinion about everything |

**vid's promise:** *the vi you already have in your fingers, precise and fast, in any Go
program that has a buffer to edit — and nothing else.*

jqv's wishlist already names the seam (jqv Q-051): "the REPL defers to its own editor or to
another, first the owner's vi-mode buffer editor (a separate library, as zsh or PSReadLine do
vi mode … with no renderer of its own), which would carry over to the TUI's code pane
(T-029)". vid is that library, built as a project in its own right.

## 2. Principles

1. **A buffer and keys, nothing more.** vid's world is a text buffer, a cursor, and key
   events in. It owns no terminal, draws nothing, reads no files, spawns nothing. Its view of
   the world is a buffer; drivers bring keys, hosts show the result.
2. **No opinion.** The host decides what enter does, what `j` at the last line does, where
   history comes from, what a word is in its language, how it's coloured. vid offers
   defaults that match the shells, never requirements.
3. **Precision over breadth.** A smaller set of vi that behaves *exactly* as vim does (or
   deliberately and documentedly not) beats a wide set that is nearly right. Muscle memory
   punishes "nearly". Differences from vim are written down (§9) and tested as such.
4. **Ergonomics decided by use.** The owner dogfoods on pwsh 7 / Windows Terminal / Zellij
   daily; a binding that feels wrong in use is wrong, whatever vim does.
5. **Testing front and centre.** The engine is a pure function of (state, key) → (state,
   events). Every behaviour is a case file; nvim headless is the oracle (§8).
6. **Fast on the keystroke path.** No allocation storms, no goroutines, no timers, no I/O in
   the core. A key costs microseconds on any buffer a line editor or code pane holds.
7. **Plain Go, no cgo in the core.** `go install` works everywhere Go does, as jqv and hit
   do. Anything heavier (tree-sitter) is an optional, separate module.
8. **Not a vim port.** See §3. When in doubt, leave it out until use asks for it.

## 3. Scope

### 3.1 What vid is

- **A vi grammar engine:** `[count]["reg]operator[count]{motion|text-object}`, linewise and
  charwise, counts, registers, dot-repeat, undo/redo.
- **Modes:** insert, normal (command), operator-pending, visual (charwise and linewise),
  replace (`R`, `r`). The host is told on every change, so it can show the mode (prompt
  glyph, cursor shape).
- **Single and multi-line buffers.** A shell line that grew a second line and a 60-line jq
  program in a code pane are the same buffer.
- **Profiles** that capture where shells and editors disagree (§6.3): a *line* profile
  (shell-like: `j`/`k` reach history at the edges, `/` searches history, enter accepts) and
  a *buffer* profile (code-pane-like: `j`/`k` stop at the edges, `/` searches the buffer,
  enter is a newline in insert mode).
- **Keymaps as data:** every binding is a table entry per mode. Hosts add their own named
  actions (`jqv.help`, `jqv.run-to-cursor`) and bind them in the same table.
- **Hooks for the host:** history, completion, accept/continue, clipboard, colour spans,
  syntax tree, word definition, notifications (§6.2).
- **A smoke host** (`cmd/vid`) that drives the engine interactively and from scripts, for
  testing and for trying bindings out (§7).
- **A test kit** (`vidtest`) that hosts reuse, so jqv's tests speak the same case format.

### 3.2 What vid is not (non-goals)

These are deliberate. Each can be revisited only through the wishlist, with a use that
asks for it.

| Not | Why |
|---|---|
| A vim / nvim clone or port | We replicate vi *editing*, not vim the program |
| Vimscript, Lua, any scripting or plugin language | The host is the extension mechanism; it's Go |
| Ex command language (`:s`, `:g`, ranges, `:normal`) | `:` belongs to the host (jqv has its own `:` commands). A tiny, host-routed subset may come later (V-wishlist) |
| Windows, splits, tabs, a buffer list | One buffer per editor; a host with several panes holds several editors |
| File I/O, encodings, `:w`, swap files | The host owns persistence |
| Rendering, scrolling, a terminal | The host draws. vid gives text, cursor, selection, mode and spans; an *optional* reference renderer exists for the smoke host |
| An LSP client | A host can be one and feed completion and diagnostics through hooks |
| Folding, multi-cursor, undo tree, macros (`q`) | Not at first; wishlist ideas, judged by use |
| Mouse | Not at first; a host can map mouse to cursor moves itself |
| Owning the event loop | vid is fed keys; it never reads input |

### 3.3 Out of the core, but in the project

- **Adapters:** Bubble Tea (v1 now, v2 when jqv moves, jqv T-028), and later whatever pwsh
  or zsh hosting needs. Each in its own package so the core imports none of them.
- **Syntax providers:** jqv's own `jqsyntax` (tolerant, already fuzzed against gojq) and a
  tree-sitter experiment, behind one interface (§10, [RESEARCH.md](RESEARCH.md)).

## 4. Where it's heading

```
                    ┌──────────────────────── hosts ────────────────────────┐
                    │ jqv REPL   jqv TUI code pane   cmd/vid smoke   pwsh?  │
                    │    │              │                 │           │     │
                    │    └────── adapter (Bubble Tea / raw keys) ─────┘     │
                    └──────────────┬──────────────────────────▲─────────────┘
                         keys in   │                          │ events, state
                                   ▼                          │
         ┌──────────────────────────── vid core ─────────────────────────────┐
         │ key  (notation, parsing)    keymap (tables per mode, profiles)    │
         │ engine (modes, grammar, counts, registers, undo, dot-repeat)      │
         │ motion · textobj  (pure functions over text)                      │
         │ text (buffer, positions, edits)       syntax (tree interface)     │
         └─────────────────────────────▲─────────────────────────────────────┘
                                       │ hooks (host-provided, optional)
              history · completion · accept/continue · clipboard · spans · syntax tree
```

Stages, each useful alone (detail in §11):

1. The engine and smoke host, at PSReadLine vi-mode parity for one line and many lines.
2. jqv's REPL with `[editor] mode = "vi"` (jqv Q-051), then the TUI code pane (jqv T-029).
3. What PSReadLine and zsh lack: visual mode, surround, more text objects; syntax-aware
   text objects and motions (jq first, through `jqsyntax`; then tree-sitter).
4. Hosting in shells: first `vid edit` (edit the current line in vid and hand it back, as
   PSReadLine's `v` does with `$EDITOR`), then experiments in replacing the line editor
   itself (pwsh's `PSConsoleHostReadLine`) — research, not a promise ([RESEARCH.md](RESEARCH.md) §2).

## 5. Architecture

### 5.1 Packages (proposal)

Module `github.com/TimelordUK/vid`. Public packages, because the point is that others use
them; internals stay in `internal/`.

| Package | What | Depends on |
|---|---|---|
| `text` | `Buffer`: the text, positions, line index, `Edit{From, To, Text}`; grapheme and display-width helpers | stdlib, `rivo/uniseg` |
| `key` | `Key` (rune or named key + modifiers), vim notation (`<Esc>`, `<C-w>`, `d2w`) parse and format | stdlib |
| `motion` | Motions as pure functions: `(text, pos, count, opts) → Range` (`w`, `e`, `f`, `%`, `gg`…) | `text` |
| `textobj` | Text objects: `iw`, `aw`, `i"`, `a(`…, and syntax-backed ones | `text`, `syntax` |
| `syntax` | `Tree` and `Node` interfaces: kinds and spans, nothing language-specific | `text` |
| `keymap` | Tables per mode; the default maps (`PSReadLine`, `ZshViMode`, `Vim`); host actions | `key` |
| `vid` (root) | `Editor`: modes, the grammar, registers, undo, dot-repeat, hooks, events, state | all above |
| `render` | Optional reference renderer: state + spans → lines of text with a cursor, in ANSI or markers | `vid` |
| `adapt/teav1` | Bubble Tea v1: `tea.KeyMsg` → `key.Key`; a `tea.Model` wrapping an `Editor` | `vid`, bubbletea v1 |
| `vidtest` | The case-file runner and helpers, exported so hosts test with it | `vid` |
| `cmd/vid` | The smoke host (§7) | everything |

Only `adapt/*` and `cmd/vid` import a terminal library. The tree-sitter provider, if it
lands, lives in its own module (`vid/syntax/treesitter`, own `go.mod`) so the core's
dependency list never grows for those who don't want it.

### 5.2 Every change is an Edit

The buffer changes only through `Edit{From, To, Text}` (replace a range with text). This one
primitive gives, for free:

- **undo/redo**: an undo step is the inverse edits of a group;
- **dot-repeat**: replays the recorded keys of the last change (vim's model, not the edits);
- **the host's view of change**: `Changed` events carry the edits, so a host can update
  its own model, re-check, re-colour;
- **incremental parsing**: a tree-sitter `InputEdit` is derived from an `Edit` directly.

### 5.3 Positions

**OPEN, recommendation:** the core works in **byte offsets into UTF-8** (Go strings,
tree-sitter's unit, no conversion on every edit). Cursor movement steps by **grapheme
cluster** (`l` over `é` written as `e` + combining accent is one step; an emoji with a
skin-tone modifier is one step). Columns for `j`/`k` (vim's "virtual column") use display
width, from a host-overridable `Width(string) int` (uniseg by default; East Asian wide
characters are two). `text` offers rune ↔ byte conversion; jqv's editor works in runes
today, so its adapter converts at the seam.

The buffer holds `\n` only. Hosts normalise `\r\n` on the way in and out (pwsh's PSReadLine
buffer is already `\n`).

Storage: a plain string plus a line-start index. Buffers here are a shell line or a
program, kilobytes not megabytes; a rope or piece table is premature (benchmarks decide,
§8.5).

### 5.4 The engine is single-threaded and timeless

`Feed(keys...) → events`. No goroutines, no timers. Anything time-based (an `esc` that may
be the start of an alt sequence, a `jk` insert-mode escape mapping, the paste burst rule
jqv already has) is the adapter's or host's business, or the engine is told the time
(`Feed` takes a timestamp in a later version if mappings with timeouts earn a place,
**OPEN**). The same key sequence always gives the same result, which is what makes case
files and fuzzing work.

## 6. The API (sketch, not settled)

### 6.1 Driving it

```go
ed := vid.New(vid.Options{
    Profile: vid.LineProfile,       // or vid.BufferProfile
    Keymap:  keymap.PSReadLine(),   // data; hosts copy and change it
    Hooks:   myHooks,               // any of the optional interfaces below
})
ed.SetText("select(.age > 3)", 7)      // text and cursor
evs := ed.Feed(key.MustParse("ciwname<Esc>")...)
st := ed.State()
// st.Text, st.Cursor, st.Mode, st.Selection, st.Pending ("d2" while a command is half typed),
// st.Register (pending "a), st.Message (e.g. "E: nothing to repeat")
for _, ev := range evs {
    switch ev := ev.(type) {
    case vid.Changed:      // ev.Edits
    case vid.ModeChanged:  // ev.From, ev.To — the host changes prompt glyph / cursor shape
    case vid.Accepted:     // enter in the line profile: ev.Text is the finished entry
    case vid.Cancelled:    // ctrl+c
    case vid.EOF:          // ctrl+d on an empty buffer
    case vid.HostAction:   // a key bound to the host's own action: ev.Name, ev.Count
    case vid.Bell:         // a motion that failed, as vim beeps
    }
}
```

### 6.2 Hooks (all optional)

Small interfaces, detected with type assertions, each with a `Func` adapter, so a host
implements only what it has. They mirror what jqv's `internal/editor` already exposes, so
jqv moves over without losing anything:

| Hook | Shape | jqv today |
|---|---|---|
| `History` | `Prev(prefix) / Next() / Search(dir, query)`; the line profile's `k`/`j` at the edges, `/` `?` `n` `N`, up/down in insert | `History`, R-011 substring search |
| `Acceptor` | `Accept(text) (accept bool, why string)`: enter at the end of an unfinished entry continues | `Continues` |
| `Completer` | `Complete(text, pos) (from, to, []Candidate)`; Tab cycles | `Completer` |
| `Painter` | `Spans(text) []Span`: colour, by byte range; vid passes them through to state | `Painter` |
| `Hinter` | a line to show under the buffer | `Hinter` |
| `SyntaxProvider` | `Tree(text) syntax.Tree`, incremental if it can | (jqsyntax, not wired yet) |
| `Words` | what a word is (`$ENV`, `@csv` are words in jq; `-Path` in pwsh) | `WordAt`'s isWord |
| `Clipboard` | the `+` and `*` registers | (atotto/clipboard in jqv) |
| `Editor` | `vv`: hand the text to an external editor and take it back | — |

A host action (`jqv.help` on F1, `jqv.layout` on F8) is a keymap entry naming the action;
vid emits `HostAction`, and the host may hand back a rewrite (`ed.Apply(edit)`), which goes
through the same undo and change machinery as any other edit.

### 6.3 Profiles: where shells and editors disagree

| Behaviour | Line profile (shell, REPL) | Buffer profile (code pane) |
|---|---|---|
| enter in insert | accept (via `Acceptor`, which may continue) | newline |
| enter in normal | accept | `+` (down to first non-blank) |
| `j` / `k` in normal | move a line; at the last / first line, history (zsh's `down-line-or-history`; jqv already does this for up/down) | move a line; stop at the edges |
| `/` `?` `n` `N` | search history (PSReadLine, zsh) | search the buffer |
| `v` | visual mode; `vv` opens the external editor (zsh-vi-mode's choice; PSReadLine's `v` is the editor) | visual mode |
| `G` `gg` | last / first line of the buffer (PSReadLine) | same |
| ctrl+d | accept-or-exit on an empty line (PSReadLine `ViAcceptLineOrExit`) | half page down |
| `#` | comment the entry out and accept (PSReadLine `PrependAndAccept`), via the host's comment string | `*`/`#` word search backwards |

A profile is a set of these switches plus a default keymap; hosts can make their own.

### 6.4 Registers

Unnamed `""`, yank `"0`, the delete ring `"1`–`"9`, small delete `"-`, named `"a`–`"z`
(`"A`–`"Z` append), black hole `"_`, and `"+` / `"*` through the `Clipboard` hook (absent:
they behave as unnamed). Read-only registers (`".`, `"%`, `":`) are not planned. Registers
belong to the editor by default; a host can share one `Registers` value between editors (a
yank in jqv's REPL pasted in its code pane).

### 6.5 Undo

Linear, grouped as vim groups: an insert session from `i` to `<Esc>` is one step; each
normal-mode change is one step. `u`, `ctrl+r` (vim) and PSReadLine's `U` (undo all) and
`ctrl+y` / `ctrl+z` are keymap choices. Cursor position is restored with the step, as vim
does. An undo tree (`g-`, `g+`) is a non-goal for now.

## 7. The smoke host: `cmd/vid`

A host as thin as it can be, for testing and for feel. Like hit's `--print` and jqv's
`--print`, every interactive feature has a non-interactive twin.

```
vid                                    # interactive, inline, line profile, PSReadLine keymap
vid --profile buffer --keymap vim file.txt   # a file's text as a buffer; nothing is saved unless :w is wired
vid run --text 'foo |bar baz' --keys 'dw'     # prints the final state: text, cursor, mode, registers
vid run --case testdata/cases/word.txt        # runs case files, as the tests do
vid keys                               # print each key as the terminal sends it and what it's bound to (jqv's --show-keys)
vid bindings --keymap psreadline       # the keymap as a table: the parity matrix comes from this
vid oracle 'dw' 'foo |bar baz'         # run the same through nvim headless and diff (when nvim is on PATH)
vid edit                               # read text on stdin, edit it, write it on stdout (the shell hosting seed, §4)
```

The interactive host shows: the buffer, the mode, the pending command (`"a2d`), the
registers on a key, and an event log pane on demand, so a surprise can be turned straight
into a case file (`ctrl+s` in the smoke host writes the session out as one, **OPEN** key).

## 8. Testing

**Changes must never quietly undo behaviour we've already agreed on.** As in jqv and hit,
every layer is tested from day one and CI runs it all on Windows and Linux.

### 8.1 Case files: the spine

Every behaviour is a case: a starting text with a cursor, keys in vim notation, and what
should follow. Plain text, one file per area, readable without Go:

```
# dw from the middle of a word deletes to the start of the next
text:  foo b|ar baz
keys:  dw
want:  foo b|baz
reg:   "  ar␠

# a count on both sides multiplies
text:  |a b c d e f g
keys:  2d2w
want:  |e f g

# ciw on whitespace changes the whitespace (vim's rule)
text:  foo |   bar
keys:  ciwX<Esc>
want:  foo|Xbar
mode:  normal
```

- `|` marks the cursor: before the character it's on in normal mode, the insertion point in
  insert mode. Visual selections are `‹…›`. Multi-line text uses an indented block.
  **OPEN:** exact markers (they must never collide with test text; a `--marker` override
  per case).
- Optional lines: `mode:`, `reg:`, `msg:`, `events:`, `profile:`, `keymap:`.
- Tags: `# oracle: skip — why` for cases where vid deliberately differs from vim (all of
  them listed in §9), `# psreadline` for cases that pin a PSReadLine behaviour.
- `go test ./... -update` re-blesses `want:` lines from the engine; review the diff before
  committing, as in jqv.
- `vidtest` runs them, so jqv's tests can carry case files of their own with jq-specific
  word rules and hooks.

### 8.2 The oracle: nvim headless

jqv checks its filter stack against gojq *and* real jq when it's on PATH. vid does the same
with nvim: each case (not tagged `oracle: skip`) is also run through
`nvim --headless --clean -u NONE` with the text loaded, the keys fed, and the buffer,
cursor, mode and registers read back. A disagreement fails the test with both results.
Without nvim the oracle tests skip; CI installs it. This keeps "behaves exactly as vim"
honest without anyone remembering vim's corner cases (`cw` on a word acting as `ce`,
`dw` at the end of a line not joining, `yiw` cursor placement, exclusive motions turning
inclusive across lines…).

A PSReadLine oracle is harder (it wants a console); its behaviours are pinned by hand from
the binding dump in [PARITY.md](PARITY.md), checked by the owner in use.

### 8.3 Layers

| Layer | How | What it proves |
|---|---|---|
| `text` | table tests, fuzz | edits apply and invert; positions, graphemes, line index stay consistent |
| `key` | table tests, fuzz | notation round-trips (`parse(format(k)) == k`); every key name the adapters emit parses |
| `motion`, `textobj` | table tests over (text, pos, count) | each motion's range, inclusive/exclusive, linewise/charwise, on edge cases (empty buffer, one char, trailing spaces, unicode) |
| engine | case files (§8.1) + nvim oracle (§8.2) | the grammar, modes, registers, undo, dot-repeat |
| engine fuzz | random key streams | never panics; cursor always valid; `u` repeated returns the starting text; `ctrl+r` repeated returns to the end; `.` equals retyping the change |
| profiles | case files with `profile:` | history hooks called exactly at the edges; enter accepts or continues |
| adapters | key messages in, no terminal (as jqv's TUI tests) | every Bubble Tea key name maps; paste arrives as text, not commands |
| reference renderer | plain-text goldens, width invariants at 20/40/120 | cursor and spans drawn in the right place; nothing overflows |
| smoke host | `vid run` with fake stdin/stdout | CLI behaviour, case runner, error messages |
| benchmarks | `go test -bench` | per-key cost and allocations stay under budget (§8.5) |

### 8.4 Regression workflow

As in jqv and hit:

1. Something from use is wrong ("`cw` at the end of a word ate the space").
2. **First a failing case file**, in the smoke host's words if it came from there.
3. Fix, commit with the wishlist ID.
4. The cases only grow, so daily use becomes the test suite.

### 8.5 Performance budget

**OPEN** numbers, to be measured on the owner's work machine: a key in normal or insert mode
on a 200-line buffer well under 100 µs and a handful of allocations; `State()` without
copying the text. Benchmarks are tests, and CI fails a large regression.

### 8.6 CI

GitHub Actions on `windows-latest` and `ubuntu-latest`: gofmt check, `go vet`, `go test`
(with the nvim oracle), a short fuzz pass. `./scripts/test.ps1` runs what CI runs.

## 9. Deliberate differences from vim

Every one is tagged in its case file (`oracle: skip`), with the reason:

- Line profile `j`/`k`/`/` reach history (§6.3).
- Enter in the line profile accepts.
- `v` / `vv` (§6.3).
- *(to be grown, never silently)*

## 10. Syntax: an interface first, tree-sitter as one provider

vid knows no language. `syntax.Tree` is an interface: nodes with a kind, a byte range,
parent and children, and a few named roles a host's provider maps (`string`, `call`,
`argument`, `block`, `comment`). On it vid builds, language-free:

- **structural text objects**: `if`/`af` (function or call), `ia`/`aa` (argument), `ic`
  (comment), `is`/`as` (string, quote-aware where `i"` is not);
- **structural motions**: next/previous sibling, up to parent (as nvim's treesitter
  textobjects and incremental selection: `v` then a key to grow the selection node by node);
- **spans** for colour if the host wants vid to compute them from the tree.

Providers: jqv's `jqsyntax` first (it exists, it's tolerant of half-typed jq, and it's
fuzzed against gojq), wrapped by jqv. Then tree-sitter as an experiment, via a pure-Go
runtime so the "no cgo" rule holds. See [RESEARCH.md](RESEARCH.md) §1.

## 11. Order of work, each step useful alone

0. **These documents**, agreed with the owner. Then two spikes: the case-file format with
   the nvim oracle driving three cases end to end, and the `key` notation against every key
   Bubble Tea v1 sends on Windows Terminal and Zellij (jqv's key spike already found shift+
   enter and ctrl+enter arrive as `enter` there).
1. **`text`, `key`, motions** (`h l w b e W B E 0 ^ $ f F t T ; , %`), `vid run`, case
   files, oracle. No modes yet beyond normal.
2. **The grammar:** counts, `d c y`, `x X s S D C Y`, `p P`, `r ~ J`, registers, undo,
   redo, dot-repeat; insert mode with the shell insert keys (ctrl+w, ctrl+u, ctrl+h).
3. **Many lines and profiles:** `j k gg G o O + -`, linewise operators (`dd`, `dj`, `yy`),
   history at the edges, accept/continue.
4. **Visual mode** (`v`, `V`) and **text objects** (words, quotes, brackets, paragraphs).
5. **The interactive smoke host** and the Bubble Tea v1 adapter; `vid keys`, `vid bindings`.
   *PSReadLine parity reached here* ([PARITY.md](PARITY.md)).
6. **jqv integration** behind `[editor] mode = "vi"` (jqv Q-051): jqv's hooks through the
   adapter, its F-keys as host actions, the mode in its prompt.
7. **Beyond the shells:** surround (`cs ds ys`, zsh-vi-mode's), `ctrl+a`/`ctrl+x`
   increment, `gu gU g~`; then the syntax interface with `jqsyntax`, then tree-sitter.
8. **Shell hosting research:** `vid edit` behind a pwsh key; then the
   `PSConsoleHostReadLine` experiment ([RESEARCH.md](RESEARCH.md) §2).

## 12. Tech choices

- **Go** (1.25+, matching hit; jqv is on 1.27): one static binary for the smoke host,
  trivial cross-compile, the same toolchain as jqv and hit.
- **`rivo/uniseg`** for graphemes and widths (already in both sibling modules' graphs).
- **Bubble Tea** only in `adapt/` and `cmd/vid`.
- **No cgo** in anything the core module builds.
- **nvim** as a test-time oracle only.

## 13. Keys

Inside vid, keys are vi's; the clash list that matters is the host's. Defaults avoid what
Zellij, tmux and Windows Terminal take first (jqv DESIGN §5, hit DESIGN §14.1): in insert
mode no ctrl+b, ctrl+p/ctrl+n (Zellij modes; PSReadLine's vi insert mode doesn't bind them
either), alt+f, alt+arrows; ctrl+s (Zellij scroll) left alone. Normal mode is mostly plain
letters and is safe. `vid keys` shows what actually arrives, before anyone argues.
