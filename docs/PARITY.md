# vid — parity matrix

What PSReadLine's vi mode binds today, what zsh-vi-mode adds, and when vid gets each.
**Step** is the order-of-work step in [DESIGN.md](DESIGN.md) §11. The PSReadLine column comes
from a dump on the owner's machine, not from memory:

```powershell
Import-Module PSReadLine; Set-PSReadLineOption -EditMode Vi
Get-PSReadLineKeyHandler -Bound    # PSReadLine 2.3.6, 2026-11-19
```

In the dump, `<k>` is a command-mode binding and a bare `k` an insert-mode one. PSReadLine
names a chord per operator and motion (`<d,w>` DeleteWord, `<c,w>` ViReplaceWord); vid has a
grammar instead, so one row below covers all of an operator's motions. Re-run the dump when
PSReadLine updates; `vid bindings --keymap psreadline` should print the same table from vid's
side (DESIGN §7).

## Normal mode: motions

| Keys | PSReadLine 2.3.6 | zsh-vi-mode | vid | Step |
|---|---|---|---|---|
| `h` `l` `←` `→` space backspace | yes | yes | yes | 1 |
| `w` `b` `e` | yes (`NextWordEnd`, `ViBackwardWord`, `ViNextWord`) | yes | yes | 1 |
| `W` `B` `E` (PSReadLine "glob") | yes | yes | yes | 1 |
| `ge` `gE` | no | yes | yes | 1 |
| `0` `^` `_` `$` home end | yes | yes | yes | 1 |
| `\|` (to column) | yes (`GotoColumn`) | yes | yes | 1 |
| `f` `F` `t` `T` `;` `,` | yes | yes | yes | 1 |
| `%` | yes (`ViGotoBrace`) | yes | yes | 1 |
| `gg` `G` | yes (first / last line) | yes | yes | 3 |
| `j` `k` `+` `-` | history (`NextHistory`, `PreviousHistory`) | line, then history | line profile: line, then history; buffer: line | 3 |
| `/` `?` `n` `N` | history search | history search | line profile: history; buffer: buffer search | 3 |
| `*` `#` (word search) | `#` is `PrependAndAccept` | — | buffer profile only | 4 |
| `{` `}` (paragraph) | no | — | buffer profile | 4 |
| ctrl+] (brace) | yes (`GotoBrace`, insert too) | — | yes | 1 |
| F3 / shift+F3 (char search) | yes | — | host keymap choice | 5 |

## Normal mode: operators and changes

| Keys | PSReadLine 2.3.6 | zsh-vi-mode | vid | Step |
|---|---|---|---|---|
| `d` + motion | yes, as fixed chords: `dw db dB dW de dE d0 d^ d$ dd d_ dG dg…` (`ViDGChord`) `dh dl dj dk df dF dt dT d%` | yes | grammar: any motion or object | 2 |
| `c` + motion | yes, fixed chords as `d` | yes | grammar | 2 |
| `y` + motion | yes, fixed chords (`yw yb y$ y0 y^ yy y% yh yl yW yB ye yE`) | yes | grammar | 2 |
| counts (`3dw`, `d3w`, `2d2w`) | `DigitArgument` | yes | yes, multiplied as vim | 2 |
| `x` `X` `s` `S` `D` `C` | yes | yes | yes | 2 |
| `Y` | no | yes | yes (vim's: line; nvim's `y$` is **OPEN**) | 2 |
| `p` `P` | yes | yes | yes, charwise and linewise | 2 |
| `r` `R` | yes (`ReplaceCharInPlace`, `ViReplaceUntilEsc`) | yes | yes | 2 |
| `~` | yes (`InvertCase`) | yes | yes | 2 |
| `g~` `gu` `gU` + motion | no | yes | yes | 7 |
| `J` | yes (`ViJoinLines`) | yes | yes | 3 |
| `o` `O` | yes (`ViAppendLine`, `ViInsertLine`) | yes | yes | 3 |
| `i` `a` `I` `A` | yes | yes | yes | 2 |
| `.` | yes (`RepeatLastCommand`) | yes | yes, from recorded keys | 2 |
| `u` | yes | yes | yes | 2 |
| `U` (undo all) | yes (`UndoAll`) | — | PSReadLine keymap only; vim's `U` (undo line) is not planned | 2 |
| ctrl+r (redo) | ctrl+r is reverse history search; redo is ctrl+y | redo | vim keymap: redo; PSReadLine keymap: as PSReadLine | 2 |
| ctrl+a / ctrl+x (increment) | no | yes | yes | 7 |
| registers (`"a`…) | no | partial (system clipboard) | yes, §6.4 | 2 |
| `#` (comment and accept) | yes (`PrependAndAccept`) | — | line profile, comment string from host | 3 |
| `v` | `ViEditVisually` ($EDITOR) | `v` visual, `vv` editor | `v` visual, `vv` editor (DESIGN §6.3) | 4 / 8 |
| ctrl+t (swap chars) | yes | — | PSReadLine keymap | 2 |

## Text objects

| Keys | PSReadLine 2.3.6 | zsh-vi-mode | vid | Step |
|---|---|---|---|---|
| `iw` `aw` | partial: the dump shows `<d,i>` bound to `ChordViTextObject` (custom group) | yes | yes | 4 |
| `iW` `aW` | — | yes | yes | 4 |
| `i"` `a"` `i'` `a'` `` i` `` `` a` `` | — | yes | yes | 4 |
| `i(` `a(` `ib` `i{` `iB` `i[` `i<` | — | yes | yes | 4 |
| `ip` `ap` (paragraph) | — | — | buffer profile | 4 |
| `it` `at` (tags) | — | — | not planned | — |
| syntax objects (`if` `ia` `ic`…) | — | — | via `syntax.Tree` (DESIGN §10) | 7 |

## Visual mode

| Keys | PSReadLine 2.3.6 | zsh-vi-mode | vid | Step |
|---|---|---|---|---|
| `v` `V` + motions + operators | no | yes | yes | 4 |
| `o` (other end) | — | yes | yes | 4 |
| ctrl+v (block) | — (ctrl+v is paste) | — | not planned | — |
| incremental node selection | — | — | yes, with a syntax tree | 7 |

## Surround (zsh-vi-mode's, vim-surround's)

| Keys | PSReadLine 2.3.6 | zsh-vi-mode | vid | Step |
|---|---|---|---|---|
| `cs"'` `ds"` `ys{motion}"` `S"` in visual | — | yes (classic; `s`-prefix variant) | yes, classic | 7 |

## Insert mode

| Keys | PSReadLine 2.3.6 (vi insert) | vid | Step |
|---|---|---|---|
| esc | `ViCommandMode` | yes | 2 |
| enter / shift+enter | `AcceptLine` / `AddLine` | line profile: accept via `Acceptor`; shift+enter a newline where the terminal sends it (jqv's spike: often it doesn't) | 3 |
| backspace, delete | yes | yes | 2 |
| ctrl+w, ctrl+backspace | `BackwardDeleteWord` / `BackwardKillWord` | yes | 2 |
| ctrl+u, ctrl+home / ctrl+end | delete to start / end of input | yes | 2 |
| ctrl+delete | `KillWord` | yes | 2 |
| ctrl+← ctrl+→, home, end | yes | yes | 2 |
| up / down | history | line profile: line, then history | 3 |
| tab / shift+tab, ctrl+space | completion | `Completer` hook, cycling as jqv | 5 |
| ctrl+r / ctrl+s | history search | `History` hook | 5 |
| ctrl+l | clear screen | host action | 5 |
| ctrl+c / ctrl+d | cancel / accept-or-exit | events `Cancelled` / `EOF` | 2 |
| ctrl+v | paste | `Clipboard` hook | 5 |
| ctrl+z / ctrl+y | undo / redo | PSReadLine keymap | 2 |
| ctrl+g | abort | host choice (Zellij takes ctrl+g in normal mode) | — |
| F1, alt+h, F2, F4 | help, parameter help, prediction views | host actions | — |

## Not from either: ideas parked in the wishlist

Macros (`q`), marks (`m`, `` ` ``), `ctrl+o` in insert mode, `gv`, `:s`, multi-cursor,
undo tree. Each needs a use that asks for it.
