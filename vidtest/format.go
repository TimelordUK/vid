package vidtest

import (
	"fmt"
	"os"
	"strings"
)

// FormatBuf writes a buffer as a single-line value, or "" with ok false when it
// needs a block (several lines).
func FormatBuf(b Buf) (line string, ok bool) {
	s := b.Text[:b.Cursor] + Cursor + b.Text[b.Cursor:]
	if strings.Contains(s, "\n") {
		return "", false
	}
	return escapeEnds(s), true
}

// ShowBuf writes a buffer for a message: one line, or an indented block.
func ShowBuf(b Buf) string {
	if s, ok := FormatBuf(b); ok {
		return s
	}
	return "\n" + block(b)
}

func block(b Buf) string {
	s := b.Text[:b.Cursor] + Cursor + b.Text[b.Cursor:]
	var w strings.Builder
	for _, l := range strings.Split(s, "\n") {
		w.WriteString("    ")
		w.WriteString(escapeEnds(l))
		w.WriteString("\n")
	}
	return w.String()
}

// escapeEnds makes leading and trailing spaces visible, and every tab.
func escapeEnds(s string) string {
	s = strings.ReplaceAll(s, "\t", "⇥")
	lead := len(s) - len(strings.TrimLeft(s, " "))
	if lead == len(s) {
		return strings.Repeat("␠", lead)
	}
	trail := len(s) - len(strings.TrimRight(s, " "))
	return strings.Repeat("␠", lead) + s[lead:len(s)-trail] + strings.Repeat("␠", trail)
}

// FormatReg writes a register as its reg: value.
func FormatReg(r Reg) string {
	content := strings.ReplaceAll(r.Text, "\n", "␤")
	content = escapeEnds(content)
	if content == "" {
		return fmt.Sprintf("%c %c", r.Name, r.Type)
	}
	return fmt.Sprintf("%c %c %s", r.Name, r.Type, content)
}

// Format writes cases in the canonical layout. Parse(Format(cs)) gives cs back.
func Format(cases []*Case) string {
	var w strings.Builder
	for i, c := range cases {
		if i > 0 {
			w.WriteString("\n")
		}
		for _, cm := range c.Comments {
			w.WriteString("# ")
			w.WriteString(cm)
			w.WriteString("\n")
		}
		writeBuf(&w, "text", c.Text)
		keys := strings.ReplaceAll(c.Keys, "\n", "␤")
		w.WriteString("keys:  ")
		w.WriteString(escapeEnds(keys))
		w.WriteString("\n")
		if c.HasWant {
			writeBuf(&w, "want", c.Want)
		}
		if c.Mode != "" {
			w.WriteString("mode:  ")
			w.WriteString(c.Mode)
			w.WriteString("\n")
		}
		for _, r := range c.Regs {
			w.WriteString("reg:   ")
			w.WriteString(FormatReg(r))
			w.WriteString("\n")
		}
	}
	return w.String()
}

func writeBuf(w *strings.Builder, field string, b Buf) {
	if s, ok := FormatBuf(b); ok {
		w.WriteString(field + ":  " + s + "\n")
		return
	}
	w.WriteString(field + ":\n" + block(b))
}

// WriteFile writes cases back to their file, for -bless.
func WriteFile(path string, cases []*Case) error {
	return os.WriteFile(path, []byte(Format(cases)), 0o644)
}
