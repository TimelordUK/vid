package text

import "github.com/rivo/uniseg"

// NextGrapheme is the offset just past the grapheme cluster starting at off, or Len at
// the end. "é" written as e + combining accent is one cluster; so is an emoji with a
// skin-tone modifier; "\n" is always one on its own.
func (b *Buffer) NextGrapheme(off int) int {
	if off >= len(b.s) {
		return len(b.s)
	}
	g, _, _, _ := uniseg.FirstGraphemeClusterInString(b.s[off:], -1)
	return off + len(g)
}

// PrevGrapheme is the offset of the grapheme cluster that ends at off, or 0 at the
// start. Clusters never span a newline, so it scans from the start of that line.
func (b *Buffer) PrevGrapheme(off int) int {
	if off <= 0 {
		return 0
	}
	at := b.LineStart(b.LineOf(off - 1))
	state := -1
	for rest := b.s[at:off]; ; {
		var g string
		g, rest, _, state = uniseg.FirstGraphemeClusterInString(rest, state)
		if rest == "" {
			return off - len(g)
		}
		at += len(g)
	}
}

// Measure says how wide text is on screen, for columns (`j`, `k`, `|`, virtual
// column). The zero value uses uniseg's widths (East Asian wide characters are two)
// and vim's tabstop of 8.
type Measure struct {
	Width   func(grapheme string) int // a host's own widths; nil for uniseg's
	TabStop int                       // 0 for 8
}

func (m Measure) tabStop() int {
	if m.TabStop > 0 {
		return m.TabStop
	}
	return 8
}

// advance is the column after grapheme g drawn at column col.
func (m Measure) advance(g string, w, col int) int {
	if g == "\t" {
		ts := m.tabStop()
		return col + ts - col%ts
	}
	if m.Width != nil {
		w = m.Width(g)
	}
	return col + w
}

// Column is the display column where off starts on its line, counting from 0.
func (b *Buffer) Column(off int, m Measure) int {
	col, state := 0, -1
	for rest := b.s[b.LineStart(b.LineOf(off)):off]; rest != ""; {
		var g string
		var w int
		g, rest, w, state = uniseg.FirstGraphemeClusterInString(rest, state)
		col = m.advance(g, w, col)
	}
	return col
}

// AtColumn is the offset of the grapheme on line that covers display column col: a
// wide character or a tab covers several. Past the end of the line it is the line's
// end; keeping the cursor off the newline in normal mode is the engine's business.
func (b *Buffer) AtColumn(line, col int, m Measure) int {
	at, end := b.LineStart(line), b.LineEnd(line)
	c, state := 0, -1
	for rest := b.s[at:end]; rest != ""; {
		var g string
		var w int
		g, rest, w, state = uniseg.FirstGraphemeClusterInString(rest, state)
		c = m.advance(g, w, c)
		if c > col {
			return at
		}
		at += len(g)
	}
	return end
}
