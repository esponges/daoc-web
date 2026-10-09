package scenery

import (
	"math"
	"strings"
	"testing"
)

// NPineACL5.nif is five copies of NpineA, each at its listed offset: every
// copy's vertices are the tree's, shifted by that offset in x and y.
func TestTreeCluster(t *testing.T) {
	game := gamePath(t)
	dirs := ModelDirs(game, true)
	c := ClusterOf(dirs, "NPineACL5.nif")
	if c == nil {
		t.Skip("no NPineACL5 in the cluster table")
	}
	if !strings.EqualFold(c.Tree, "NpineA.NIF") || len(c.Offsets) != 5 {
		t.Fatalf("cluster is %s at %d offsets, want NpineA.NIF at 5", c.Tree, len(c.Offsets))
	}
	if c.Offsets[0] != [3]float32{125, -845, 0} {
		t.Errorf("first offset %v, want 125,-845,0", c.Offsets[0])
	}

	var tv, cv []Vertex
	var ti, ci []uint32
	tree, err := ConvertModel(dirs, "tree", c.Tree, &tv, &ti)
	if err != nil {
		t.Fatal(err)
	}
	clu, err := ConvertModel(dirs, "cluster", "NPineACL5.nif", &cv, &ci)
	if err != nil {
		t.Fatal(err)
	}
	if len(cv) != 5*len(tv) || len(ci) != 5*len(ti) {
		t.Fatalf("cluster has %d vertices and %d indices, want 5x the tree's %d and %d", len(cv), len(ci), len(tv), len(ti))
	}
	if len(clu.Groups) != len(tree.Groups) {
		t.Fatalf("%d draw groups, want the tree's %d", len(clu.Groups), len(tree.Groups))
	}
	for k, o := range c.Offsets {
		for i, v := range tv {
			w := cv[k*len(tv)+i]
			for a := 0; a < 3; a++ {
				if math.Abs(float64(w.P[a]-v.P[a]-o[a])) > 1e-3 {
					t.Fatalf("copy %d vertex %d at %v, want the tree's %v moved by %v", k, i, w.P, v.P, o)
				}
			}
		}
	}
	// Every index of every group points into the cluster's own vertices.
	for _, g := range clu.Groups {
		for _, i := range ci[g.First : g.First+g.Count] {
			if int(i) >= len(cv) {
				t.Fatalf("index %d past %d vertices", i, len(cv))
			}
		}
	}
	if clu.Max[0]-clu.Min[0] <= tree.Max[0]-tree.Min[0] {
		t.Errorf("cluster bounds %v..%v no wider than one tree's", clu.Min, clu.Max)
	}
}

// A row's unused slots are zeros and make no trees.
func TestClusterZeroSlots(t *testing.T) {
	game := gamePath(t)
	dirs := ModelDirs(game, true)
	c := ClusterOf(dirs, "HiberniaTall_WhitePine3CL10.nif")
	if c == nil {
		t.Skip("no HiberniaTall_WhitePine3CL10 in the cluster table")
	}
	if len(c.Offsets) != 10 {
		t.Errorf("CL10 cluster has %d trees, want 10", len(c.Offsets))
	}
	for _, name := range []string{"Elm1CL5.nif", "NPineACL5.nif"} {
		if c := ClusterOf(dirs, name); c == nil || len(c.Offsets) != 5 {
			t.Errorf("%s: want 5 trees from a row of 10 slots", name)
		}
	}
}

// zones/Nifs keeps nrs-hut1_normal and _frozen as empty scene roots; the
// huts themselves are in Newtowns/zones/Nifs.
func TestStandInFoundWhole(t *testing.T) {
	game := gamePath(t)
	dirs := ModelDirs(game, false)
	for _, file := range []string{"nrs-hut1_normal.nif", "nrs-hut1_frozen.nif", "nrshse1_normal.nif"} {
		var v []Vertex
		var i []uint32
		mo, err := ConvertModel(dirs, file, file, &v, &i)
		if err != nil {
			t.Errorf("%s: %v", file, err)
			continue
		}
		if len(v) < 1000 || len(mo.Groups) == 0 {
			t.Errorf("%s: %d vertices in %d groups, want a whole building", file, len(v), len(mo.Groups))
		}
	}
}
