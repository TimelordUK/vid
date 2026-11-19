# vid

A vi-mode buffer editor as a Go library: modes, motions, operators, text objects, registers,
undo and dot-repeat over a buffer, with no terminal and no renderer of its own. Working name
(**vi** e**d**itor).

Think of the vi mode in PSReadLine or zsh-vi-mode, but as a library any Go program can host,
precise where those are approximate, and testable key by key. It is not a vim port: no
vimscript, no ex language, no windows or files. Its world is a buffer; the host brings the
keys and draws the result.

First host: [jqv](https://github.com/TimelordUK/jqv)'s REPL and, later, its code pane.

**Status:** design only. Nothing is built yet.

- [docs/DESIGN.md](docs/DESIGN.md): why, scope and non-goals, architecture, API sketch,
  testing, order of work
- [docs/PARITY.md](docs/PARITY.md): PSReadLine's and zsh-vi-mode's bindings against vid's plan
- [docs/RESEARCH.md](docs/RESEARCH.md): tree-sitter in Go, hosting vid in pwsh and zsh
- [docs/WISHLIST.md](docs/WISHLIST.md): every item, with IDs
