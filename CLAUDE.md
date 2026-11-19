# vid — notes for Claude

- Read `docs/DESIGN.md` before changing behaviour; keep it in sync when decisions change.
  Sections marked proposal stay proposals until the owner agrees them.
- New ideas go in `docs/WISHLIST.md` with the next free ID in their category (`C-` core,
  `V-` vi grammar, `H-` hosts/adapters, `S-` syntax, `Q-` quality). Never renumber; mark
  items `dropped` instead of deleting them.
- Reference wishlist IDs in commit messages (e.g. `V-002: w b e over graphemes`).
- vid is a library over a buffer: the core never touches a terminal, files, the clock or a
  process, and imports no terminal library (only `adapt/*` and `cmd/vid` do). No cgo in the
  core module.
- Not a vim port: check DESIGN §3.2 before adding anything; non-goals change only through
  the wishlist.
- Every behaviour is a case file; vim's behaviour is checked by the nvim oracle. A deliberate
  difference from vim is tagged `oracle: skip` with its reason and listed in DESIGN §9.
- Bug or behaviour change from use: a failing case file first, then the fix (DESIGN §8.4).
- Golden and case files: `go test ./... -update` re-blesses; review the diff before committing.
- The owner dogfoods on pwsh 7 / Windows Terminal / Zellij daily; ergonomics decided by use
  beat theory. Default keys avoid Zellij/tmux/Windows Terminal defaults (DESIGN §13).
- Sibling projects: jqv (`../jqv`, first host; its `internal/editor` is the seam, jqv Q-051)
  and hit (`../hit`, PSReadLine integration and the resident-process pattern).
