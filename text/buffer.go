// Package text is vid's buffer: a string, a line index, and one way to change it, an
// Edit (DESIGN §5.2). Positions are byte offsets into the UTF-8 text; movement steps by
// grapheme cluster; columns are display widths (DESIGN §5.3). The buffer holds "\n"
// only: hosts normalise "\r\n" on the way in and out.
package text

import (
	"fmt"
	"sort"
	"strings"
	"unicode/utf8"
)

// Edit replaces the bytes [From, To) with Text. Every change to a buffer is one.
type Edit struct {
	From, To int
	Text     string
}

// Buffer is the text and the start offset of each line. The zero value is an empty
// buffer.
type Buffer struct {
	s     string
	lines []int // lines[i] is the offset of line i's first byte; lines[0] == 0
}

// New makes a buffer holding s.
func New(s string) *Buffer {
	b := &Buffer{}
	b.Set(s)
	return b
}

// Set replaces the whole text. Invalid UTF-8 becomes U+FFFD.
func (b *Buffer) Set(s string) {
	s = strings.ToValidUTF8(s, "\uFFFD")
	b.s = s
	b.lines = append(b.lines[:0], 0)
	b.lines = indexLines(b.lines, s, 0)
}

// indexLines appends the start of every line that begins after a newline in s, where
// s starts at offset base.
func indexLines(lines []int, s string, base int) []int {
	for i := 0; ; {
		j := strings.IndexByte(s[i:], '\n')
		if j < 0 {
			return lines
		}
		i += j + 1
		lines = append(lines, base+i)
	}
}

// String is the whole text. It does not copy.
func (b *Buffer) String() string { return b.s }

// Len is the text's length in bytes.
func (b *Buffer) Len() int { return len(b.s) }

// Slice is the text of [from, to).
func (b *Buffer) Slice(from, to int) string { return b.s[from:to] }

// Check reports why e cannot apply to b, or nil if it can: the range must be inside
// the text, From <= To, both ends on rune boundaries, and Text valid UTF-8.
func (b *Buffer) Check(e Edit) error {
	switch {
	case e.From < 0 || e.To > len(b.s) || e.From > e.To:
		return fmt.Errorf("text: edit [%d,%d) outside [0,%d)", e.From, e.To, len(b.s))
	case !b.runeStart(e.From) || !b.runeStart(e.To):
		return fmt.Errorf("text: edit [%d,%d) splits a UTF-8 sequence", e.From, e.To)
	case !utf8.ValidString(e.Text):
		return fmt.Errorf("text: edit text %q is not valid UTF-8", e.Text)
	}
	return nil
}

func (b *Buffer) runeStart(off int) bool {
	return off == len(b.s) || utf8.RuneStart(b.s[off])
}

// Apply makes the edit and returns its inverse: applying that undoes this one.
// An edit that fails Check panics; hosts' edits are checked before they get here.
func (b *Buffer) Apply(e Edit) Edit {
	if err := b.Check(e); err != nil {
		panic(err)
	}
	inv := Edit{From: e.From, To: e.From + len(e.Text), Text: b.s[e.From:e.To]}
	if e.From == e.To && e.Text == "" {
		return inv
	}
	b.s = b.s[:e.From] + e.Text + b.s[e.To:]

	// Lines starting at or before From are unchanged; lines that started after To move
	// by the change in length; the new text brings its own.
	first := b.LineOf(e.From) + 1
	last := sort.SearchInts(b.lines, e.To+1) // first line starting after To
	tail := b.lines[last:]
	delta := len(e.Text) - (e.To - e.From)
	lines := append([]int(nil), b.lines[:first]...)
	lines = indexLines(lines, e.Text, e.From)
	for _, l := range tail {
		lines = append(lines, l+delta)
	}
	b.lines = lines
	return inv
}

// Lines is the number of lines: one more than the number of newlines.
func (b *Buffer) Lines() int { return len(b.lines) }

// LineStart is the offset of line n's first byte.
func (b *Buffer) LineStart(n int) int { return b.lines[n] }

// LineEnd is the offset of line n's newline, or Len for the last line.
func (b *Buffer) LineEnd(n int) int {
	if n+1 < len(b.lines) {
		return b.lines[n+1] - 1
	}
	return len(b.s)
}

// Line is line n's text, without its newline.
func (b *Buffer) Line(n int) string { return b.s[b.LineStart(n):b.LineEnd(n)] }

// LineOf is the line holding offset off. A newline belongs to the line it ends; Len
// belongs to the last line.
func (b *Buffer) LineOf(off int) int {
	return sort.SearchInts(b.lines, off+1) - 1
}

// Pos is the line and byte column of off.
func (b *Buffer) Pos(off int) (line, col int) {
	line = b.LineOf(off)
	return line, off - b.lines[line]
}

// Offset is the offset of byte column col on line, clamped to the line's end.
func (b *Buffer) Offset(line, col int) int {
	return min(b.lines[line]+col, b.LineEnd(line))
}

// RuneOffset converts a byte offset to a rune offset, for hosts that count in runes
// (jqv's editor does).
func (b *Buffer) RuneOffset(off int) int { return utf8.RuneCountInString(b.s[:off]) }

// ByteOffset converts a rune offset to a byte offset, clamped to Len.
func (b *Buffer) ByteOffset(runes int) int {
	off := 0
	for ; runes > 0 && off < len(b.s); runes-- {
		_, n := utf8.DecodeRuneInString(b.s[off:])
		off += n
	}
	return off
}
