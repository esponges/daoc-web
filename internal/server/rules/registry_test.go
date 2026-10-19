package rules

import (
	"flag"
	"fmt"
	"os"
	"sort"
	"sync"
	"testing"
)

// Every scenario a rule registers must be run by a test, through vector. After
// the package's tests run, TestMain checks none was left out (unless -run
// narrowed the run, when some are skipped on purpose).
var (
	ranMu sync.Mutex
	ran   = map[string]bool{}
)

// vector runs one rule book scenario as a subtest named after it.
func vector(t *testing.T, capability, scenario string, f func(t *testing.T)) {
	t.Helper()
	ranMu.Lock()
	ran[capability+"/"+scenario] = true
	ranMu.Unlock()
	t.Run(capability+"/"+scenario, f)
}

func TestMain(m *testing.M) {
	flag.Parse()
	code := m.Run()
	if code == 0 && flag.Lookup("test.run").Value.String() == "" {
		if missing := missingVectors(); len(missing) > 0 {
			for _, s := range missing {
				fmt.Fprintln(os.Stderr, "rule book scenario with no test:", s)
			}
			code = 1
		}
	}
	os.Exit(code)
}

func missingVectors() []string {
	var out []string
	for _, r := range Registry() {
		for _, s := range r.Scenarios {
			if !ran[r.Capability+"/"+s] {
				out = append(out, r.ID+": "+r.Capability+"/"+s)
			}
		}
	}
	sort.Strings(out)
	return out
}

func TestRegistryWellFormed(t *testing.T) {
	seen := map[string]bool{}
	for _, r := range Registry() {
		if r.ID == "" || r.Capability == "" || r.Requirement == "" {
			t.Errorf("rule %+v lacks an id, capability or requirement", r)
		}
		if seen[r.ID] {
			t.Errorf("rule %s registered twice", r.ID)
		}
		seen[r.ID] = true
		if len(r.Sources) == 0 {
			t.Errorf("rule %s names no source", r.ID)
		}
		for _, s := range r.Sources {
			if Sources[s] == "" {
				t.Errorf("rule %s names unknown source %q", r.ID, s)
			}
		}
	}
}

// near compares floats to a tolerance in the last shown digit of the specs.
func near(t *testing.T, what string, got, want, tol float64) {
	t.Helper()
	if d := got - want; d > tol || d < -tol {
		t.Errorf("%s = %v, want %v", what, got, want)
	}
}
