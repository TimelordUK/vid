// Package oracle runs case files through nvim, the reference engine (DESIGN §8.2).
//
// One `nvim --embed --headless --clean` serves a whole test run. Each case: a Lua
// call resets the state a case can see and loads the text; the keys go in through
// nvim_input, as nvim's own functional tests feed them (feedkeys with "x" would sit
// in insert mode waiting for more); a second Lua call reads back the buffer,
// cursor, mode and registers, and is answered only once the keys are consumed. `--clean` means the
// owner's config and plugins never load: nvim's own defaults are the reference.
package oracle

import (
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/TimelordUK/vid/vidtest"
	"github.com/neovim/go-client/nvim"
)

// ErrNoNvim is returned by Start when nvim is not on PATH; tests skip on it.
var ErrNoNvim = errors.New("nvim not on PATH")

// Nvim is one embedded nvim.
type Nvim struct {
	v *nvim.Nvim
}

// Result is what nvim did with a case.
type Result struct {
	Buf  vidtest.Buf
	Mode string // in case-file words: normal, insert, …, or nvim:<mode> for others
	Regs map[rune]vidtest.Reg
	Err  string // an error the keys raised (E21 and the like), "" if none
}

// The state a case can see is reset before each one: a new buffer (fresh undo,
// changelist, marks), the registers, the last search and the last f/t/F/T.
// Not resettable this way: what `.` repeats and the `.` register. A case that
// needs those untouched would want a fresh nvim; none does yet.
const setup = `
set nomore noswapfile shortmess+=aAIcFW
lua << EOF
local resettable = '"0123456789abcdefghijklmnopqrstuvwxyz-/'
function _G.vid_load(lines, row, col)
  vim.cmd('silent! %bwipeout!')
  for i = 1, #resettable do vim.fn.setreg(resettable:sub(i, i), {}) end
  vim.fn.setcharsearch({ char = '' })
  vim.cmd('clearjumps')
  vim.api.nvim_buf_set_lines(0, 0, -1, true, lines)
  vim.api.nvim_win_set_cursor(0, { row, col })
end
function _G.vid_read(regnames)
  local regs = {}
  for _, r in ipairs(regnames) do
    regs[#regs + 1] = { r, vim.fn.getreg(r), vim.fn.getregtype(r) }
  end
  return {
    lines = vim.api.nvim_buf_get_lines(0, 0, -1, true),
    cursor = vim.api.nvim_win_get_cursor(0),
    mode = vim.api.nvim_get_mode().mode,
    errmsg = vim.v.errmsg,
    regs = regs,
  }
end
EOF
`

// Start launches nvim.
func Start() (*Nvim, error) {
	if _, err := exec.LookPath("nvim"); err != nil {
		return nil, ErrNoNvim
	}
	v, err := nvim.NewChildProcess(
		nvim.ChildProcessArgs("--embed", "--headless", "--clean", "-n", "-i", "NONE"),
		nvim.ChildProcessLogf(func(string, ...any) {}),
	)
	if err != nil {
		return nil, err
	}
	if _, err := v.Exec(setup, false); err != nil {
		v.Close()
		return nil, fmt.Errorf("oracle setup: %w", err)
	}
	return &Nvim{v: v}, nil
}

// Close stops nvim.
func (n *Nvim) Close() error { return n.v.Close() }

// Version is nvim's version line, for the test log.
func (n *Nvim) Version() string {
	var s string
	if err := n.v.ExecLua(`return tostring(vim.version())`, &s); err != nil {
		return "?"
	}
	return s
}

type luaResult struct {
	Lines  []string   `msgpack:"lines"`
	Cursor [2]int     `msgpack:"cursor"`
	Mode   string     `msgpack:"mode"`
	Err    string     `msgpack:"errmsg"`
	Regs   [][]string `msgpack:"regs"`
}

// Run feeds a case to nvim. regs names the registers to read back.
func (n *Nvim) Run(c *vidtest.Case, regs []rune) (*Result, error) {
	lines := strings.Split(c.Text.Text, "\n")
	row, col := rowCol(c.Text.Text, c.Text.Cursor)
	names := make([]string, len(regs))
	for i, r := range regs {
		names[i] = string(r)
	}
	if err := n.v.ExecLua(`vim.v.errmsg = '' vid_load(...)`, nil, lines, row+1, col); err != nil {
		return nil, err
	}
	if _, err := n.v.Input(c.Keys); err != nil {
		return nil, err
	}
	var lr luaResult
	if err := n.v.ExecLua(`return vid_read(...)`, &lr, names); err != nil {
		return nil, err
	}
	if lr.Mode != "n" {
		// back to normal mode for the next case
		if _, err := n.v.Input(`<C-\><C-n>`); err != nil {
			return nil, err
		}
	}
	text := strings.Join(lr.Lines, "\n")
	res := &Result{
		Buf:  vidtest.Buf{Text: text, Cursor: offset(lr.Lines, lr.Cursor[0]-1, lr.Cursor[1])},
		Mode: modeName(lr.Mode),
		Regs: map[rune]vidtest.Reg{},
		Err:  lr.Err,
	}
	for _, v := range lr.Regs {
		if len(v) != 3 || v[0] == "" {
			return nil, fmt.Errorf("oracle: bad register reply %q", v)
		}
		r := []rune(v[0])[0]
		res.Regs[r] = vidtest.Reg{Name: r, Type: regType(v[2]), Text: v[1]}
	}
	return res, nil
}

func rowCol(text string, off int) (row, col int) {
	before := text[:off]
	row = strings.Count(before, "\n")
	col = off - (strings.LastIndex(before, "\n") + 1)
	return row, col
}

func offset(lines []string, row, col int) int {
	off := 0
	for i := 0; i < row && i < len(lines); i++ {
		off += len(lines[i]) + 1
	}
	return off + col
}

func modeName(m string) string {
	switch {
	case m == "n":
		return "normal"
	case m == "i":
		return "insert"
	case m == "R":
		return "replace"
	case m == "v":
		return "visual"
	case m == "V":
		return "vline"
	case strings.HasPrefix(m, "no"):
		return "pending"
	}
	return "nvim:" + m
}

func regType(t string) byte {
	switch {
	case t == "V":
		return 'l'
	case strings.HasPrefix(t, "\x16"):
		return 'b'
	}
	return 'c'
}
