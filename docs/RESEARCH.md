# vid — research

Questions to answer with spikes before anything here becomes design. Each ends with what
would settle it. Findings go back into this file, dated; decisions move to
[DESIGN.md](DESIGN.md).

## 1. Tree-sitter: what it would give us, and how to have it in Go

### 1.1 What nvim gets from it, and which parts fit a buffer editor

| nvim feature | Fits vid? | Notes |
|---|---|---|
| Highlighting from `highlights.scm` queries | Maybe, as a span provider | jqv already colours from its own lexer; tree-sitter would matter for languages we haven't written a lexer for (pwsh, bash) |
| Textobjects (`nvim-treesitter-textobjects`: `af` `if` `aa` `ia` `ac`) | **Yes, the main prize** | Language-free text objects over named nodes; DESIGN §10's `syntax.Tree` is shaped for this |
| Incremental selection (grow / shrink by node) | **Yes** | One key in visual mode climbs the tree; great on jq's nested pipes |
| Structural motions (next sibling, parent) | Yes | `]f` `[f`-style; keys **OPEN** |
| Folding, indent queries | Indent maybe (`o` in a code pane); folding no | Folding is a renderer concern |
| Incremental reparse on edit | Yes, for free from DESIGN §5.2's `Edit` | A key mustn't re-parse a 200-line program from scratch, though at that size even that may be fine |
| Error recovery | Yes, and it matters | A line being typed is broken most of the time; jqv's `jqsyntax` was written tolerant for exactly this |

### 1.2 Getting it in Go without giving up "no cgo"

- **cgo bindings** (the official `tree-sitter/go-tree-sitter`, and the older
  `smacker/go-tree-sitter`): mature, but need a C toolchain per target, which breaks
  `go install` for anyone without one and complicates Windows builds. Ruled out for the core;
  acceptable only in an optional module.
- **A pure-Go runtime:** `odvcencio/gotreesitter` describes itself as a tree-sitter runtime
  in Go with no cgo: parser, lexer, query engine, incremental reparsing and external
  scanners in Go, grammar tables extracted from upstream `parser.c` and embedded, a registry
  of 200+ grammars, `wasip1` builds. Its benchmarks are self-reported and change by version.
  **To verify in a spike:** that the jq grammar is there (or can be generated), parse speed
  and allocations on jqv's `examples/jq` programs, error recovery on half-typed lines,
  binary size cost per grammar, licence, and how alive the project is.
- **Grammars compiled to WASM run under a pure-Go WASM runtime** (wazero): a possible
  fallback if the pure-Go runtime disappoints. Not researched yet.

### 1.3 Tree-sitter vs our own parser

For jq specifically, jqv already has `jqsyntax`: tolerant, fuzzed against gojq in both
directions, and it knows things a generic grammar doesn't (what an unfinished line leaves
open). So the experiment is not "replace jqsyntax", it's:

1. Define `syntax.Tree` (DESIGN §10) and implement it over `jqsyntax`'s tree in jqv.
2. Build the text objects and incremental selection on the interface.
3. Implement the same interface over gotreesitter with the jq grammar, and over the
   PowerShell and bash grammars, and compare on the same case files.

**Settles it:** case files for `if`/`ia`/incremental selection pass with both providers on
jqv's example programs, and the tree-sitter provider stays inside the key budget
(DESIGN §8.5) on a 200-line program.

## 2. Hosting vid as a shell's line editor

**Decided (2026-11-21):** option B below, hit's resident pattern, with keys never crossing
the pipe; the design is now DESIGN §16 and the spike H-012. `vid edit` (2.1) stays an idea
for zsh, no longer the pwsh story. The notes below are kept as the reasoning.

The ladder, cheapest and safest first. Each rung is useful alone.

### 2.1 `vid edit` behind a key (no risk)

PSReadLine's `v` already hands the line to `$EDITOR`; zsh has `edit-command-line`. A pwsh
key handler (built as hit's are, through its line editor adapter `GetBuffer` /
`ReplaceBuffer`) runs `vid edit`, gives it the buffer and cursor, and puts back what comes
out. The shell keeps PSReadLine for everything else.

Cost to watch: hit measured process start on a managed work machine at ~3.5 s for a binary
the endpoint agent hasn't seen and ~58 ms once it has (hit DESIGN §16). Fine for an
occasional "edit this properly" key; not for every prompt.

### 2.2 Replacing PSReadLine: `PSConsoleHostReadLine`

pwsh's console host calls a function named `PSConsoleHostReadLine` for each line of input
if one is defined; that is how PSReadLine itself plugs in. A module defining it owns line
editing. Options for what's behind it:

| Option | Shape | For | Against |
|---|---|---|---|
| A. Spawn per prompt | function runs a Go binary that reads the console and prints the line | Simplest | The process-start cost above, on every prompt |
| B. Resident process | hit's `serve` shape: a child of the shell, named pipe control channel, draws on the inherited console | hit has done the hard parts (lifetime, DACL, fallback) | Two processes sharing a console's input; latency per key must be measured |
| C. In-process Go library | Go built as a C shared library, loaded by a small C# binary module via P/Invoke | No process, no IPC | Needs cgo for the `c-shared` build (a host-side choice, not vid's core); per-OS builds; console I/O from Go inside a .NET process |
| D. Stay inside PSReadLine | Bind vid's actions as PSReadLine key handlers that call out | Keeps completion, predictions, history | Every key through a handler; not really vid editing |

What PSReadLine gives that a replacement must give back or lose: tab completion through
`TabExpansion2` / `CommandCompletion.CompleteInput`, predictions (`PSReadLine` predictor
plug-ins), syntax colouring from pwsh's own tokenizer, multi-line continuation from the
parser (`Parser.ParseInput` incomplete input), history saving, and its screen handling on
resize and wrap. Completion and parsing need pwsh itself, so any option except D has to call
back into the runspace for them.

**Settles it:** a spike of B or C that edits one line with vid in Windows Terminal and
Zellij, hands it back to pwsh, and gets Tab completion from pwsh, measured per key on the
work machine. Until then this is research, and `vid edit` (2.1) is the shell story.

### 2.3 zsh

zle owns the line; widgets can read and replace `BUFFER` and `CURSOR`. The equivalents:
`vid edit` as a widget (2.1) is straightforward; replacing zle wholesale is not on the table.
A middle way — a widget per vi key calling a resident vid — would be D's shape and inherits
its cost. Parked until there's a Linux user asking.

## 3. Smaller questions

- **Esc latency:** terminals send esc as the start of alt sequences; Bubble Tea waits a little
  to tell them apart. Measure what arrives on Windows Terminal and Zellij with `vid keys`;
  normal mode must feel instant after esc.
- **Cursor shape per mode** (DECSCUSR: bar in insert, block in normal): which terminals honour
  it (Windows Terminal does; Zellij passes it through?). A host concern, but the smoke host
  should show it.
- **The nvim oracle's cost:** decided (2026-11-21): one `nvim --embed` for a whole test run
  over msgpack-RPC, not a process per case (DESIGN §8.2, Q-002).
- **Kitty keyboard protocol** (Bubble Tea v2, jqv T-028): would give shift+enter and ctrl+i vs
  tab. Matters for insert-mode bindings, not for normal mode.
