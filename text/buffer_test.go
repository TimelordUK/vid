package text

import (
	"reflect"
	"strings"
	"testing"

	"github.com/rivo/uniseg"
)

func TestLines(t *testing.T) {
	for _, tc := range []struct {
		s     string
		lines []string
	}{
		{"", []string{""}},
		{"abc", []string{"abc"}},
		{"a\nb", []string{"a", "b"}},
		{"a\n", []string{"a", ""}},
		{"\n\n", []string{"", "", ""}},
	} {
		b := New(tc.s)
		var got []string
		for n := range b.Lines() {
			got = append(got, b.Line(n))
		}
		if !reflect.DeepEqual(got, tc.lines) {
			t.Errorf("New(%q) lines %q, want %q", tc.s, got, tc.lines)
		}
	}
}

func TestLineOfAndPos(t *testing.T) {
	b := New("ab\ncd\n")
	for _, tc := range []struct{ off, line, col int }{
		{0, 0, 0}, {2, 0, 2}, // the newline belongs to the line it ends
		{3, 1, 0}, {5, 1, 2},
		{6, 2, 0}, // Len: the empty last line
	} {
		line, col := b.Pos(tc.off)
		if line != tc.line || col != tc.col {
			t.Errorf("Pos(%d) = %d,%d, want %d,%d", tc.off, line, col, tc.line, tc.col)
		}
		if got := b.Offset(line, col); got != tc.off {
			t.Errorf("Offset(%d,%d) = %d, want %d", line, col, got, tc.off)
		}
	}
	if got := b.Offset(0, 99); got != 2 {
		t.Errorf("Offset past the end of line 0 = %d, want 2 (its newline)", got)
	}
}

func TestApply(t *testing.T) {
	for _, tc := range []struct {
		s    string
		e    Edit
		want string
	}{
		{"foo bar", Edit{4, 7, "baz"}, "foo baz"},
		{"foo bar", Edit{3, 3, "\nx"}, "foo\nx bar"},
		{"a\nb\nc", Edit{1, 3, ""}, "a\nc"},
		{"a\nb\nc", Edit{1, 4, ""}, "ac"},
		{"a\nb\nc", Edit{0, 5, "z"}, "z"},
		{"", Edit{0, 0, "one\ntwo\n"}, "one\ntwo\n"},
		{"héllo", Edit{1, 3, "e"}, "hello"},
	} {
		b := New(tc.s)
		inv := b.Apply(tc.e)
		if b.String() != tc.want {
			t.Errorf("%q %+v = %q, want %q", tc.s, tc.e, b.String(), tc.want)
		}
		checkIndex(t, b)
		b.Apply(inv)
		if b.String() != tc.s {
			t.Errorf("%q %+v then its inverse = %q", tc.s, tc.e, b.String())
		}
		checkIndex(t, b)
	}
}

func TestCheck(t *testing.T) {
	b := New("héllo")
	for _, e := range []Edit{{-1, 0, ""}, {0, 9, ""}, {3, 2, ""}, {2, 2, "x"}, {0, 2, ""}} {
		if b.Check(e) == nil {
			t.Errorf("Check(%+v) passed", e)
		}
	}
	if b.Check(Edit{0, 0, "\x88"}) == nil {
		t.Error("Check passed text that is not UTF-8")
	}
	if got := New("a\x88b").String(); got != "a\uFFFDb" {
		t.Errorf("New kept invalid UTF-8: %q", got)
	}
	if err := b.Check(Edit{1, 3, ""}); err != nil {
		t.Errorf("Check over é: %v", err)
	}
}

func TestRuneOffsets(t *testing.T) {
	b := New("aé😀b")
	for runes, off := range []int{0, 1, 3, 7, 8} {
		if got := b.ByteOffset(runes); got != off {
			t.Errorf("ByteOffset(%d) = %d, want %d", runes, got, off)
		}
		if got := b.RuneOffset(off); got != runes {
			t.Errorf("RuneOffset(%d) = %d, want %d", off, got, runes)
		}
	}
}

// checkIndex holds the incrementally kept line index to a fresh one.
func checkIndex(t *testing.T, b *Buffer) {
	t.Helper()
	if fresh := New(b.String()); !reflect.DeepEqual(b.lines, fresh.lines) {
		t.Fatalf("line index %v for %q, want %v", b.lines, b.String(), fresh.lines)
	}
}

// FuzzApply: any valid edit keeps the line index exact, and its inverse restores the
// text.
func FuzzApply(f *testing.F) {
	f.Add("a\nb\nc", 1, 3, "x\ny")
	f.Add("", 0, 0, "\n")
	f.Add("héllo\nwörld", 2, 9, "")
	f.Fuzz(func(t *testing.T, s string, from, to int, ins string) {
		s = strings.ToValidUTF8(s, "\uFFFD") // as New does
		b := New(s)
		e := Edit{from, to, ins}
		if b.Check(e) != nil {
			return
		}
		inv := b.Apply(e)
		if want := s[:from] + ins + s[to:]; b.String() != want {
			t.Fatalf("got %q, want %q", b.String(), want)
		}
		checkIndex(t, b)
		b.Apply(inv)
		if b.String() != s {
			t.Fatalf("inverse gave %q, want %q", b.String(), s)
		}
		checkIndex(t, b)
	})
}

func TestGraphemes(t *testing.T) {
	// e + combining acute; thumbs up + skin tone; a wide CJK character; a newline
	s := "aéb👍🏽c中\nd"
	want := []int{0, 1, 4, 5, 13, 14, 17, 18, 19}
	b := New(s)
	var got []int
	for off := 0; ; off = b.NextGrapheme(off) {
		got = append(got, off)
		if off == len(s) {
			break
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("forward %v, want %v", got, want)
	}
	for i := len(want) - 1; i > 0; i-- {
		if p := b.PrevGrapheme(want[i]); p != want[i-1] {
			t.Errorf("PrevGrapheme(%d) = %d, want %d", want[i], p, want[i-1])
		}
	}
}

func TestColumn(t *testing.T) {
	b := New("a\tb中c\nx")
	var m Measure
	for _, tc := range []struct{ off, col int }{
		{0, 0}, {1, 1}, {2, 8}, {3, 9}, {6, 11}, {7, 12}, {8, 0},
	} {
		if got := b.Column(tc.off, m); got != tc.col {
			t.Errorf("Column(%d) = %d, want %d", tc.off, got, tc.col)
		}
	}
	for _, tc := range []struct{ col, off int }{
		{0, 0}, {1, 1}, {5, 1}, {7, 1}, {8, 2}, {9, 3}, {10, 3}, {11, 6}, {40, 7},
	} {
		if got := b.AtColumn(0, tc.col, m); got != tc.off {
			t.Errorf("AtColumn(0, %d) = %d, want %d", tc.col, got, tc.off)
		}
	}
	m = Measure{TabStop: 4, Width: func(string) int { return 1 }}
	if got := b.Column(6, m); got != 6 {
		t.Errorf("Column with tabstop 4 and narrow widths = %d, want 6", got)
	}
}

// FuzzGraphemes: stepping forward and back visit the same boundaries, the same ones
// uniseg finds over the whole text; and a column maps back to its grapheme.
func FuzzGraphemes(f *testing.F) {
	f.Add("aéb👍🏽c中\nd")
	f.Add("\t\n\r\n")
	f.Add("🇬🇧🇫🇷")
	f.Fuzz(func(t *testing.T, s string) {
		if !strings.Contains(s, "\r") { // the buffer holds \n only
			testGraphemes(t, s)
		}
	})
}

func testGraphemes(t *testing.T, s string) {
	s = strings.ToValidUTF8(s, "\uFFFD") // as New does
	b := New(s)
	var want []int
	for off, g := 0, uniseg.NewGraphemes(s); g.Next(); {
		want = append(want, off)
		off += len(g.Str())
	}
	var fwd []int
	for off := 0; off < len(s); off = b.NextGrapheme(off) {
		fwd = append(fwd, off)
	}
	if !reflect.DeepEqual(fwd, want) {
		t.Fatalf("forward %v, uniseg %v", fwd, want)
	}
	for i := len(want) - 1; i >= 0; i-- {
		end := len(s)
		if i+1 < len(want) {
			end = want[i+1]
		}
		if p := b.PrevGrapheme(end); p != want[i] {
			t.Fatalf("PrevGrapheme(%d) = %d, want %d", end, p, want[i])
		}
	}
	var m Measure
	for i, off := range want {
		end := len(s)
		if i+1 < len(want) {
			end = want[i+1]
		}
		if s[off] == '\n' || b.Column(end, m) == b.Column(off, m) {
			continue // newlines and zero-width clusters have no column of their own
		}
		line, _ := b.Pos(off)
		if got := b.AtColumn(line, b.Column(off, m), m); got != off {
			t.Fatalf("AtColumn(%d, Column(%d)) = %d", line, off, got)
		}
	}
}
