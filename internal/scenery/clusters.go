package scenery

// Tree clusters: the frontier zones place many of their trees five or ten
// at a time. A placed model such as NPineACL5.nif is not a model file at
// all: zones/trees/tree_clusters.mpk holds tree_clusters.csv, which names
// each cluster's tree and up to ten offsets from the placement, in the
// model's ground plane (x, y; z is 0 throughout), with zero rows for the
// slots a cluster does not use:
//
//	name,tree,x,y,z,x,y,z,...
//	NPineACL5.nif,NpineA.NIF,125,-845,0,825,-355,0,...,0,0,0
//
// A cluster is baked as its tree copied to each offset, one model, so that
// it stays one placement drawn as one instance.

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"daocweb/internal/mpak"
)

// Cluster is one tree_clusters.csv row.
type Cluster struct {
	Tree    string
	Offsets [][3]float32
}

var clusterTables sync.Map // directory -> map[string]*Cluster, keyed lower case

// clusters reads the cluster table from the first of dirs that has one.
func clusters(dirs []string) map[string]*Cluster {
	for _, d := range dirs {
		if t, ok := clusterTables.Load(d); ok {
			if m := t.(map[string]*Cluster); m != nil {
				return m
			}
			continue
		}
		m, err := readClusters(filepath.Join(d, "tree_clusters.mpk"))
		if err != nil {
			m = nil
		}
		clusterTables.Store(d, m)
		if m != nil {
			return m
		}
	}
	return nil
}

func readClusters(path string) (map[string]*Cluster, error) {
	if _, err := os.Stat(path); err != nil {
		return nil, err
	}
	a, err := mpak.Open(path)
	if err != nil {
		return nil, err
	}
	raw, err := a.Read("tree_clusters.csv")
	if err != nil {
		return nil, err
	}
	rd := csv.NewReader(bytes.NewReader(raw))
	rd.FieldsPerRecord = -1
	rows, err := rd.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("tree_clusters.csv: %w", err)
	}
	out := map[string]*Cluster{}
	for _, r := range rows {
		if len(r) < 5 || strings.EqualFold(strings.TrimSpace(r[0]), "name") {
			continue
		}
		c := &Cluster{Tree: strings.TrimSpace(r[1])}
		for i := 2; i+2 < len(r); i += 3 {
			var o [3]float32
			for k := 0; k < 3; k++ {
				v, _ := strconv.ParseFloat(strings.TrimSpace(r[i+k]), 32)
				o[k] = float32(v)
			}
			if o != ([3]float32{}) {
				c.Offsets = append(c.Offsets, o)
			}
		}
		if c.Tree != "" && len(c.Offsets) > 0 {
			out[strings.ToLower(strings.TrimSpace(r[0]))] = c
		}
	}
	return out, nil
}

// ClusterOf is the cluster a placed model file names, or nil.
func ClusterOf(dirs []string, file string) *Cluster {
	return clusters(dirs)[strings.ToLower(file)]
}

// bakeCluster bakes c's tree once and copies it to each offset, appending
// to the shared buffers. Each draw group holds every copy's triangles.
func bakeCluster(dirs []string, c *Cluster, name, file string, verts *[]Vertex, indices *[]uint32) (*Model, error) {
	var lv []Vertex
	var li []uint32
	base, err := ConvertModel(dirs, name, c.Tree, &lv, &li)
	if err != nil {
		return nil, fmt.Errorf("cluster tree %s: %w", c.Tree, err)
	}
	mo := &Model{Name: name, File: file, Embedded: base.Embedded}
	first := make([]uint32, len(c.Offsets)) // where each copy's vertices start
	for k, o := range c.Offsets {
		first[k] = uint32(len(*verts))
		for _, v := range lv {
			v.P[0] += o[0]
			v.P[1] += o[1]
			v.P[2] += o[2]
			*verts = append(*verts, v)
		}
		for a := 0; a < 3; a++ {
			lo, hi := base.Min[a]+o[a], base.Max[a]+o[a]
			if k == 0 || lo < mo.Min[a] {
				mo.Min[a] = lo
			}
			if k == 0 || hi > mo.Max[a] {
				mo.Max[a] = hi
			}
		}
	}
	for _, g := range base.Groups {
		ng := g
		ng.First = len(*indices)
		for k := range c.Offsets {
			for _, i := range li[g.First : g.First+g.Count] {
				*indices = append(*indices, i+first[k])
			}
		}
		ng.Count = g.Count * len(c.Offsets)
		mo.Groups = append(mo.Groups, ng)
	}
	// Trees carry no emitters; were one to, it would stand at the first
	// tree only.
	mo.Particles = base.Particles
	return mo, nil
}
