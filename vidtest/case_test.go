package vidtest

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

const sample = `# dw from the middle of a word deletes to the start of the next
text:  foo b‸ar baz
keys:  dw
want:  foo b‸baz
reg:   " c ar␠

# dd on a middle line
# a second comment line
text:
    one
    t‸wo
    three
keys:  dd
want:
    one
    ‸three
reg:   " l two␤

# o opens a line in insert mode
# oracle: skip — an example of the tag
text:  ␠␠‸foo
keys:  obar
want:
    ␠␠foo
    bar‸
mode:  insert
`

func TestParse(t *testing.T) {
	cs, err := Parse(strings.NewReader(sample), "sample")
	if err != nil {
		t.Fatal(err)
	}
	if len(cs) != 3 {
		t.Fatalf("got %d cases, want 3", len(cs))
	}
	c := cs[0]
	if c.Text != (Buf{"foo bar baz", 5}) || c.Keys != "dw" || c.Want != (Buf{"foo bbaz", 5}) {
		t.Errorf("case 1: %+v", c)
	}
	if !reflect.DeepEqual(c.Regs, []Reg{{'"', 'c', "ar "}}) {
		t.Errorf("case 1 regs: %+v", c.Regs)
	}
	c = cs[1]
	if c.Text != (Buf{"one\ntwo\nthree", 5}) || c.Want != (Buf{"one\nthree", 4}) || c.Line != 7 {
		t.Errorf("case 2: %+v", c)
	}
	if !reflect.DeepEqual(c.Regs, []Reg{{'"', 'l', "two\n"}}) || len(c.Comments) != 2 {
		t.Errorf("case 2: %+v", c)
	}
	c = cs[2]
	if c.Text != (Buf{"  foo", 2}) || c.Want != (Buf{"  foo\nbar", 9}) || c.Mode != "insert" {
		t.Errorf("case 3: %+v", c)
	}
	if c.OracleSkip != "an example of the tag" {
		t.Errorf("case 3 skip: %q", c.OracleSkip)
	}
}

func TestFormatRoundTrip(t *testing.T) {
	cs, err := Parse(strings.NewReader(sample), "sample")
	if err != nil {
		t.Fatal(err)
	}
	out := Format(cs)
	if out != sample {
		t.Errorf("format is not canonical:\n%s", out)
	}
	again, err := Parse(strings.NewReader(out), "sample")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(cs, again) {
		t.Error("parse(format(cases)) differs")
	}
}

func TestEscapes(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"foo", "foo"},
		{" a ", "␠a␠"},
		{"   ", "␠␠␠"},
		{"a\tb", "a⇥b"},
		{"a b", "a b"},
	} {
		if got := escapeEnds(tc.in); got != tc.want {
			t.Errorf("escapeEnds(%q) = %q, want %q", tc.in, got, tc.want)
		}
		if got := unescape(escapeEnds(tc.in)); got != tc.in {
			t.Errorf("round trip %q gave %q", tc.in, got)
		}
	}
}

func TestErrors(t *testing.T) {
	for _, tc := range []struct{ name, in, want string }{
		{"no marker", "# x\ntext:  foo\nkeys:  x\n", "no cursor marker"},
		{"two markers", "# x\ntext:  f‸o‸o\nkeys:  x\n", "more than one"},
		{"no keys", "# x\ntext:  ‸foo\n", "no keys"},
		{"no text", "# x\nkeys:  x\n", "no text"},
		{"no name", "text:  ‸foo\n", "starts with a # line"},
		{"bad mode", "# x\ntext:  ‸a\nkeys:  x\nmode:  weird\n", "unknown mode"},
		{"bad field", "# x\ntext:  ‸a\nkeys:  x\nwhat:  y\n", "unknown field"},
		{"bad reg", "# x\ntext:  ‸a\nkeys:  x\nreg:   \" q a\n", "type is c, l or b"},
		{"skip without why", "# x\n# oracle: skip\ntext:  ‸a\nkeys:  x\n", "needs its reason"},
	} {
		_, err := Parse(strings.NewReader(tc.in), "t")
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: got %v, want an error containing %q", tc.name, err, tc.want)
		}
	}
}

// The spec itself must parse and be in canonical form, so -bless never rewrites
// what a person wrote except the fields it was asked to.
func TestSpecFilesAreCanonical(t *testing.T) {
	files, _ := filepath.Glob(filepath.Join("..", "testdata", "cases", "*.txt"))
	if len(files) == 0 {
		t.Skip("no case files")
	}
	for _, f := range files {
		cs, err := ParseFile(f)
		if err != nil {
			t.Error(err)
			continue
		}
		raw, _ := readFile(f)
		if got := Format(cs); got != raw {
			t.Errorf("%s is not in canonical form; the canonical form is:\n%s", f, got)
		}
	}
}

func readFile(path string) (string, error) {
	b, err := os.ReadFile(path)
	return strings.ReplaceAll(string(b), "\r\n", "\n"), err
}
