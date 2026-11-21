package oracle

import (
	"errors"
	"flag"
	"path/filepath"
	"testing"
	"time"

	"github.com/TimelordUK/vid/vidtest"
)

var bless = flag.Bool("bless", false,
	"write want:, mode: and the reg: contents from nvim into the case files; review the diff")

// TestSpecAgainstNvim holds every case in testdata/cases to nvim. The case files
// are the spec; this proves the spec says what nvim does.
func TestSpecAgainstNvim(t *testing.T) {
	nv, err := Start()
	if errors.Is(err, ErrNoNvim) {
		t.Skip("nvim not on PATH: the oracle is skipped")
	}
	if err != nil {
		t.Fatal(err)
	}
	defer nv.Close()
	t.Logf("oracle: nvim %s", nv.Version())

	files, _ := filepath.Glob(filepath.Join("..", "testdata", "cases", "*.txt"))
	if len(files) == 0 {
		t.Fatal("no case files in testdata/cases")
	}
	var n int
	var spent time.Duration
	for _, f := range files {
		cases, err := vidtest.ParseFile(f)
		if err != nil {
			t.Fatal(err)
		}
		changed := false
		t.Run(filepath.Base(f), func(t *testing.T) {
			for _, c := range cases {
				t.Run(c.Name(), func(t *testing.T) {
					if c.OracleSkip != "" {
						t.Skip("oracle: skip — " + c.OracleSkip)
					}
					start := time.Now()
					got, err := nv.Run(c, regNames(c))
					spent += time.Since(start)
					n++
					if err != nil {
						t.Fatalf("%s: %v", c.Where(), err)
					}
					if *bless {
						changed = blessCase(c, got) || changed
						return
					}
					check(t, c, got)
				})
			}
		})
		if changed {
			if err := vidtest.WriteFile(f, cases); err != nil {
				t.Fatal(err)
			}
			t.Logf("blessed %s from nvim: review the diff", f)
		}
	}
	if n > 0 {
		t.Logf("oracle: %d cases in %v, %v a case", n, spent.Round(time.Millisecond), (spent / time.Duration(n)).Round(time.Microsecond))
	}
}

func regNames(c *vidtest.Case) []rune {
	var rs []rune
	for _, r := range c.Regs {
		rs = append(rs, r.Name)
	}
	return rs
}

func check(t *testing.T, c *vidtest.Case, got *Result) {
	t.Helper()
	if got.Err != "" {
		t.Errorf("%s: nvim raised: %s", c.Where(), got.Err)
	}
	if !c.HasWant {
		t.Errorf("%s: no want:; nvim gives: %s  (go test -bless writes it)", c.Where(), vidtest.ShowBuf(got.Buf))
	} else if got.Buf != c.Want {
		t.Errorf("%s: %s\nkeys:  %s\nwant:  %s\nnvim:  %s", c.Where(), c.Name(), c.Keys,
			vidtest.ShowBuf(c.Want), vidtest.ShowBuf(got.Buf))
	}
	if got.Mode != c.ExpectedMode() {
		t.Errorf("%s: mode: want %s, nvim is in %s", c.Where(), c.ExpectedMode(), got.Mode)
	}
	for _, want := range c.Regs {
		if g := got.Regs[want.Name]; g != want {
			t.Errorf("%s: reg: want %s, nvim has %s", c.Where(), vidtest.FormatReg(want), vidtest.FormatReg(g))
		}
	}
}

// blessCase writes nvim's answer into the case and says whether anything changed.
func blessCase(c *vidtest.Case, got *Result) bool {
	changed := !c.HasWant || c.Want != got.Buf || c.ExpectedMode() != got.Mode
	c.Want, c.HasWant = got.Buf, true
	c.Mode = got.Mode
	if c.Mode == "normal" {
		c.Mode = ""
	}
	for i, r := range c.Regs {
		if g := got.Regs[r.Name]; g != r {
			c.Regs[i] = g
			changed = true
		}
	}
	return changed
}
