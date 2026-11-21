// Package vidtest reads and writes case files: the spec every engine is held to
// (DESIGN §8.1). nvim checks the spec (the oracle module); vid is checked against
// it; so vid agrees with nvim on every case that is not tagged `oracle: skip`.
//
// A case file is a series of cases, each starting with one or more `#` lines:
//
//	# dw from the middle of a word deletes to the start of the next
//	text:  foo b‸ar baz
//	keys:  dw
//	want:  foo b‸baz
//	reg:   " c ar␠
//
// Fields:
//
//	text:   the buffer before, with ‸ marking the cursor (required)
//	keys:   keys in vim notation, as nvim takes them: ihello<Esc><Left><Del> (required)
//	want:   the buffer after, with ‸ marking the cursor
//	mode:   the mode after: normal (the default, never written), insert, replace,
//	        visual, vline, pending
//	reg:    a register after: its name, its type (c charwise, l linewise, b blockwise)
//	        and its content; repeatable
//
// The cursor ‸ sits before the character it is on in normal mode, and at the
// insertion point in insert mode; both are the same byte offset.
//
// A buffer of several lines is a block of lines under the field, each indented by
// exactly four spaces, which are removed:
//
//	text:
//	    one
//	    t‸wo
//
// Inside a value, ␠ is a space (needed only at the start or end, where editors and
// eyes lose them), ⇥ a tab and ␤ a newline.
//
// Comment lines carry tags: `# oracle: skip — why` for a deliberate difference from
// nvim (DESIGN §9).
package vidtest

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// Cursor marks the cursor in text: and want:.
const Cursor = "‸"

// Buf is a buffer with a cursor, as a byte offset into Text.
type Buf struct {
	Text   string
	Cursor int
}

// Reg is a register's expected content.
type Reg struct {
	Name rune
	Type byte // 'c' charwise, 'l' linewise, 'b' blockwise
	Text string
}

// Case is one behaviour.
type Case struct {
	File string
	Line int // line of the case's first comment

	Comments   []string // the comment lines, without "# "; the first is the case's name
	OracleSkip string   // why nvim is not the reference here; "" when it is

	Text    Buf
	Keys    string
	Want    Buf
	HasWant bool
	Mode    string // "" means normal
	Regs    []Reg

	hasText bool
}

// Name is the first comment line.
func (c *Case) Name() string {
	if len(c.Comments) == 0 {
		return ""
	}
	return c.Comments[0]
}

// Where is file:line, for messages.
func (c *Case) Where() string { return fmt.Sprintf("%s:%d", c.File, c.Line) }

// ExpectedMode is the mode the case wants after its keys.
func (c *Case) ExpectedMode() string {
	if c.Mode == "" {
		return "normal"
	}
	return c.Mode
}

// Modes a case may name.
var Modes = map[string]bool{
	"normal": true, "insert": true, "replace": true,
	"visual": true, "vline": true, "pending": true,
}

// ParseFile reads a case file.
func ParseFile(path string) ([]*Case, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return Parse(f, path)
}

// Parse reads cases; name is used in messages and kept in each case.
func Parse(r io.Reader, name string) ([]*Case, error) {
	p := parser{name: name}
	sc := bufio.NewScanner(r)
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		p.n++
		if err := p.line(strings.TrimRight(sc.Text(), "\r")); err != nil {
			return nil, err
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if err := p.flush(); err != nil {
		return nil, err
	}
	return p.cases, nil
}

type parser struct {
	name  string
	n     int
	cases []*Case
	cur   *Case
	// fields seen in the current case: a comment after one starts a new case
	inFields bool
	// an open block: the field it belongs to and its lines so far
	block      string
	blockLines []string
	blockLine  int
}

func (p *parser) errf(format string, a ...any) error {
	return fmt.Errorf("%s:%d: %s", p.name, p.n, fmt.Sprintf(format, a...))
}

func (p *parser) line(s string) error {
	if p.block != "" {
		if strings.HasPrefix(s, "    ") {
			p.blockLines = append(p.blockLines, s[4:])
			return nil
		}
		if strings.TrimSpace(s) == "" {
			p.blockLines = append(p.blockLines, "")
			return nil
		}
		if err := p.closeBlock(); err != nil {
			return err
		}
	}
	switch {
	case strings.TrimSpace(s) == "":
		return nil
	case strings.HasPrefix(s, "#"):
		if p.cur == nil || p.inFields {
			if err := p.flush(); err != nil {
				return err
			}
			p.cur = &Case{File: p.name, Line: p.n}
		}
		text := strings.TrimSpace(strings.TrimPrefix(s, "#"))
		p.cur.Comments = append(p.cur.Comments, text)
		if why, ok := cutTag(text, "oracle:"); ok {
			why = strings.TrimSpace(why)
			rest, ok := strings.CutPrefix(why, "skip")
			if !ok {
				return p.errf("oracle tag must be `oracle: skip — why`")
			}
			rest = strings.TrimLeft(rest, " —-:")
			if rest == "" {
				return p.errf("`oracle: skip` needs its reason")
			}
			p.cur.OracleSkip = rest
		}
		return nil
	}
	if p.cur == nil {
		return p.errf("a case starts with a # line naming it")
	}
	field, val, ok := strings.Cut(s, ":")
	if !ok || strings.ContainsAny(field, " \t") {
		return p.errf("want `field: value`, got %q", s)
	}
	p.inFields = true
	val = strings.TrimLeft(val, " ")
	switch field {
	case "text", "want":
		if val == "" {
			p.block, p.blockLines, p.blockLine = field, nil, p.n
			return nil
		}
		return p.setBuf(field, unescape(val))
	case "keys":
		if val == "" {
			return p.errf("keys: is empty")
		}
		p.cur.Keys = val
	case "mode":
		if !Modes[val] {
			return p.errf("unknown mode %q", val)
		}
		if val == "normal" {
			val = ""
		}
		p.cur.Mode = val
	case "reg":
		rg, err := parseReg(val)
		if err != nil {
			return p.errf("%v", err)
		}
		p.cur.Regs = append(p.cur.Regs, rg)
	default:
		return p.errf("unknown field %q", field)
	}
	return nil
}

func (p *parser) closeBlock() error {
	lines := p.blockLines
	for len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	for i := range lines {
		lines[i] = unescape(lines[i])
	}
	field := p.block
	p.block, p.blockLines = "", nil
	if len(lines) == 0 {
		return fmt.Errorf("%s:%d: %s: block is empty", p.name, p.blockLine, field)
	}
	return p.setBuf(field, strings.Join(lines, "\n"))
}

func (p *parser) setBuf(field, s string) error {
	b, err := ParseBuf(s)
	if err != nil {
		return p.errf("%s: %v", field, err)
	}
	if field == "text" {
		p.cur.Text, p.cur.hasText = b, true
	} else {
		p.cur.Want, p.cur.HasWant = b, true
	}
	return nil
}

func (p *parser) flush() error {
	if p.block != "" {
		if err := p.closeBlock(); err != nil {
			return err
		}
	}
	c := p.cur
	p.cur, p.inFields = nil, false
	if c == nil {
		return nil
	}
	if !c.hasText {
		return fmt.Errorf("%s: case has no text:", c.Where())
	}
	if c.Keys == "" {
		return fmt.Errorf("%s: case has no keys:", c.Where())
	}
	p.cases = append(p.cases, c)
	return nil
}

func cutTag(comment, tag string) (string, bool) {
	return strings.CutPrefix(comment, tag)
}

// ParseBuf takes text with exactly one cursor marker.
func ParseBuf(s string) (Buf, error) {
	i := strings.Index(s, Cursor)
	if i < 0 {
		return Buf{}, fmt.Errorf("no cursor marker %s", Cursor)
	}
	rest := s[i+len(Cursor):]
	if strings.Contains(rest, Cursor) {
		return Buf{}, fmt.Errorf("more than one cursor marker %s", Cursor)
	}
	return Buf{Text: s[:i] + rest, Cursor: i}, nil
}

func parseReg(s string) (Reg, error) {
	// `" c ar␠`: name, type, then the content after one space
	r := []rune(s)
	if len(r) < 3 || r[1] != ' ' {
		return Reg{}, fmt.Errorf("reg: want `name type content`, got %q", s)
	}
	t := r[2]
	if t != 'c' && t != 'l' && t != 'b' {
		return Reg{}, fmt.Errorf("reg: type is c, l or b, got %q", string(t))
	}
	content := ""
	if len(r) > 3 {
		if r[3] != ' ' {
			return Reg{}, fmt.Errorf("reg: want a space after the type in %q", s)
		}
		content = string(r[4:])
	}
	return Reg{Name: r[0], Type: byte(t), Text: unescape(content)}, nil
}

var unescaper = strings.NewReplacer("␠", " ", "⇥", "\t", "␤", "\n")

func unescape(s string) string { return unescaper.Replace(s) }
