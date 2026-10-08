// Command nifdump inspects a .nif and checks what was parsed is coherent.
//
// Beyond the footer landing at end-of-file, which Parse already enforces, the
// strongest available check is the bind pose: for every bone of a skinned
// shape, the bone's world transform composed with that bone's skin transform
// must reproduce the shape's one overall skin transform. That identity is what
// "bind pose" means, so if the hierarchy walk, the matrix layout or the skin
// transforms were read wrongly, it breaks. Run with -bind to see the residual.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"daocweb/internal/nif"
)

func main() {
	bind := flag.Bool("bind", false, "check the bind-pose identity for every skinned bone")
	tree := flag.Int("tree", 0, "print this many levels of the node hierarchy")
	shapes := flag.Bool("shapes", false, "list geometry")
	probeName := flag.String("probe", "", "dump all transforms bearing on this shape")
	tex := flag.Bool("tex", false, "list every shape's properties and texture slots")
	flag.Parse()
	if flag.NArg() < 1 {
		fmt.Fprintln(os.Stderr, "usage: nifdump [-bind] [-shapes] [-tree N] <file.nif>")
		os.Exit(2)
	}

	for _, path := range flag.Args() {
		raw, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", path, err)
			os.Exit(1)
		}
		f, err := nif.Parse(raw)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s: %v\n", filepath.Base(path), err)
			// A desync shows up as the first block whose layout is wrong, so
			// the blocks just before it are the useful context.
			if f != nil {
				last := 0
				for i, e := range f.Extents {
					if e.End != 0 {
						last = i
					}
				}
				fmt.Fprintln(os.Stderr, "  parsed up to:")
				for i := max(0, last-5); i <= last; i++ {
					e := f.Extents[i]
					fmt.Fprintf(os.Stderr, "    block %3d  %-22s %6d..%-6d (%d bytes)%s\n",
						i, e.Type, e.Start, e.End, e.End-e.Start, describe(f, i))
				}
			}
			os.Exit(1)
		}
		fmt.Printf("%s\n  %s  (%d bytes, %d blocks, roots %v)\n",
			filepath.Base(path), f.HeaderString, len(raw), len(f.Blocks), f.Roots)

		census := map[string]int{}
		for _, t := range f.TypeOf {
			census[t]++
		}
		keys := make([]string, 0, len(census))
		for k := range census {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool { return census[keys[i]] > census[keys[j]] })
		fmt.Print("  blocks:")
		for _, k := range keys {
			fmt.Printf(" %s=%d", k, census[k])
		}
		fmt.Println()

		if *shapes {
			listShapes(f)
		}
		if *tree > 0 {
			printTree(f, *tree)
		}
		if *tex {
			listTextures(f)
		}
		if *probeName != "" {
			probe(f, *probeName)
		}
		if *bind {
			checkBind(f)
			checkInverse(f)
		}
		fmt.Println()
	}
}

func listShapes(f *nif.File) {
	world := f.WorldTransforms()
	totV, totT := 0, 0
	fmt.Println("  geometry:")
	for i, b := range f.Blocks {
		s, ok := b.(*nif.TriShape)
		if !ok {
			continue
		}
		d, _ := f.Block(s.Data).(*nif.ShapeData)
		if d == nil {
			fmt.Printf("    %-28s (no data)\n", s.Name)
			continue
		}
		nb := 0
		if si, ok := f.Block(s.Skin).(*nif.SkinInstance); ok {
			nb = len(si.Bones)
		}
		uv := 0
		if len(d.UV) > 0 {
			uv = len(d.UV)
		}
		fmt.Printf("    %-28s v=%-5d tri=%-5d uvsets=%d colors=%-5v bones=%-3d y=%.0f\n",
			s.Name, len(d.Vertices), len(d.Triangles), uv, d.Colors != nil, nb,
			world[i].Trans[2])
		totV += len(d.Vertices)
		totT += len(d.Triangles)
	}
	fmt.Printf("  total: %d vertices, %d triangles\n", totV, totT)
}

func printTree(f *nif.File, depth int) {
	fmt.Println("  hierarchy:")
	var walk func(i int32, lvl int)
	walk = func(i int32, lvl int) {
		if lvl > depth {
			return
		}
		n := f.NodeAt(i)
		if n == nil {
			return
		}
		fmt.Printf("    %*s%s (%d children)\n", lvl*2, "", n.Name, len(n.Children))
		for _, c := range n.Children {
			walk(c, lvl+1)
		}
	}
	for _, r := range f.Roots {
		walk(r, 0)
	}
}

func checkBind(f *nif.File) {
	world := f.WorldTransforms()
	worst, worstWhere := float32(0), ""
	n, shapes := 0, 0
	for _, b := range f.Blocks {
		s, ok := b.(*nif.TriShape)
		if !ok {
			continue
		}
		si, _ := f.Block(s.Skin).(*nif.SkinInstance)
		if si == nil {
			continue
		}
		sd, _ := f.Block(si.Data).(*nif.SkinData)
		if sd == nil || len(sd.Bones) != len(si.Bones) {
			continue
		}
		shapes++
		// The invariant: for every bone, the bone's world transform composed
		// with that bone's skin transform reproduces the shape's own world
		// placement. Equivalently, each bone's skin transform is the inverse
		// bind matrix, which is exactly what GPU skinning needs.
		//
		// Note it is the shape's world transform, not NiSkinData's overall
		// skin transform: that one is the inverse of it.
		ref := world[shapeIdx(f, s)]
		for bi, boneRef := range si.Bones {
			if boneRef < 0 || int(boneRef) >= len(world) {
				continue
			}
			got := world[boneRef].Mul(sd.Bones[bi].Skin)
			d := got.MaxDiff(ref)
			n++
			if d > worst {
				worst = d
				bone := "?"
				if bn := f.NodeAt(boneRef); bn != nil {
					bone = bn.Name
				}
				worstWhere = fmt.Sprintf("%s / %s", s.Name, bone)
			}
		}
	}
	fmt.Printf("  bind pose: %d bone bindings across %d skinned shapes, worst residual %.6g",
		n, shapes, worst)
	if worstWhere != "" {
		fmt.Printf(" (%s)", worstWhere)
	}
	fmt.Println()
	// Tolerance is relative to the model's own size. The residual grows with
	// hierarchy depth because each level multiplies another float32 matrix;
	// the deepest bones here are fingers, ten joints from the root. A genuine
	// layout error desynchronises wholesale and lands orders of magnitude out,
	// not a fraction of a percent.
	const modelScale = 70 // a humanoid is about this tall in DAoC units
	switch {
	case n == 0:
		fmt.Println("    -> no skinning to check")
	case worst < 0.01*modelScale:
		fmt.Printf("    -> holds to %.3f%% of model height; hierarchy, matrix layout\n"+
			"       and skin transforms all read correctly\n", 100*worst/modelScale)
	default:
		fmt.Println("    -> BROKEN; a transform is being read wrongly")
	}
}

// describe adds whatever identifying detail a block type carries, which is
// usually a name and is what tells you where in the model a desync happened.
func describe(f *nif.File, i int) string {
	switch b := f.Block(int32(i)).(type) {
	case *nif.Node:
		return "  " + b.Name
	case *nif.TriShape:
		return fmt.Sprintf("  %s -> data %d skin %d", b.Name, b.Data, b.Skin)
	case *nif.ShapeData:
		return fmt.Sprintf("  %d verts, %d tris, %d uvsets", len(b.Vertices), len(b.Triangles), len(b.UV))
	case *nif.SkinInstance:
		return fmt.Sprintf("  %d bones, root %d", len(b.Bones), b.SkeletonRoot)
	case *nif.SkinData:
		return fmt.Sprintf("  %d bones", len(b.Bones))
	case *nif.SkinPartition:
		return fmt.Sprintf("  %d partitions", len(b.Parts))
	case *nif.Material:
		return "  " + b.Name
	}
	return ""
}

// probe prints every transform that bears on one shape's skinning, which is
// the only practical way to work out which convention the file uses.
func probe(f *nif.File, want string) {
	world := f.WorldTransforms()
	parents := f.Parents()
	for i, b := range f.Blocks {
		s, ok := b.(*nif.TriShape)
		if !ok || s.Name != want {
			continue
		}
		chain := []string{}
		for p := parents[i]; p >= 0; p = parents[p] {
			if n := f.NodeAt(p); n != nil {
				chain = append(chain, n.Name)
			}
		}
		fmt.Printf("  shape %q (block %d)\n", s.Name, i)
		fmt.Printf("    parents: %v\n", chain)
		fmt.Printf("    local        %v\n", s.Local)
		fmt.Printf("    world        %v\n", world[i])
		d, _ := f.Block(s.Data).(*nif.ShapeData)
		if d != nil && len(d.Vertices) > 0 {
			fmt.Printf("    vertex[0]    (%.3f %.3f %.3f)   center (%.2f %.2f %.2f) r=%.2f\n",
				d.Vertices[0][0], d.Vertices[0][1], d.Vertices[0][2],
				d.Center[0], d.Center[1], d.Center[2], d.Radius)
		}
		si, _ := f.Block(s.Skin).(*nif.SkinInstance)
		sd, _ := f.Block(si.Data).(*nif.SkinData)
		fmt.Printf("    skeletonRoot block %d\n", si.SkeletonRoot)
		fmt.Printf("    skinData.Skin %v\n", sd.Skin)
		for bi, ref := range si.Bones {
			name := "?"
			if n := f.NodeAt(ref); n != nil {
				name = n.Name
			}
			fmt.Printf("    bone %d %-18s\n", bi, name)
			fmt.Printf("        world    %v\n", world[ref])
			fmt.Printf("        boneSkin %v\n", sd.Bones[bi].Skin)
			fmt.Printf("        w*bs     %v\n", world[ref].Mul(sd.Bones[bi].Skin))
			fmt.Printf("        bs*w     %v\n", sd.Bones[bi].Skin.Mul(world[ref]))
		}
		return
	}
	fmt.Printf("  no shape named %q\n", want)
}

// shapeIdx finds a shape's block index by identity.
func shapeIdx(f *nif.File, want *nif.TriShape) int32 {
	for i, b := range f.Blocks {
		if b == any(want) {
			return int32(i)
		}
	}
	return 0
}

// checkInverse confirms NiSkinData's overall transform is the inverse of the
// shape's world placement, which is the other half of the bind-pose story.
func checkInverse(f *nif.File) {
	world := f.WorldTransforms()
	worst, n := float32(0), 0
	for i, b := range f.Blocks {
		s, ok := b.(*nif.TriShape)
		if !ok {
			continue
		}
		si, _ := f.Block(s.Skin).(*nif.SkinInstance)
		if si == nil {
			continue
		}
		sd, _ := f.Block(si.Data).(*nif.SkinData)
		if sd == nil {
			continue
		}
		if d := world[i].Inverse().MaxDiff(sd.Skin); d > worst {
			worst = d
		}
		n++
	}
	fmt.Printf("  skinData.Skin vs inverse(world[shape]): %d shapes, worst %.6g\n", n, worst)
}

// listTextures prints each shape's own properties, and every texture slot of
// any texturing property among them, with the file each slot names. A shape
// also inherits its ancestors' properties; those are listed under the nodes.
func listTextures(f *nif.File) {
	fmt.Println("  properties:")
	for i, b := range f.Blocks {
		var name string
		var props []int32
		switch o := b.(type) {
		case *nif.TriShape:
			name, props = "shape "+o.Name, o.Properties
		case *nif.Node:
			name, props = f.TypeOf[i]+" "+o.Name, o.Properties
		default:
			continue
		}
		if len(props) == 0 {
			continue
		}
		fmt.Printf("    %3d %s\n", i, name)
		for _, p := range props {
			fmt.Printf("          %3d %s", p, f.TypeOf[p])
			if ap, ok := f.Block(p).(*nif.AlphaProperty); ok {
				fmt.Printf("  flags=%#04x blend=%v src=%d dst=%d test=%v", ap.Flags, ap.Flags&1 != 0, (ap.Flags>>1)&15, (ap.Flags>>5)&15, ap.Flags&(1<<9) != 0)
			}
			if t, ok := f.Block(p).(*nif.Texturing); ok {
				slots := make([]string, 0, len(t.Slots))
				for s := range t.Slots {
					slots = append(slots, s)
				}
				sort.Strings(slots)
				for _, s := range slots {
					file := "?"
					if src, ok := f.Block(t.Slots[s].Source).(*nif.SourceTexture); ok {
						file = fmt.Sprintf("%q external=%d", src.FileName, src.UseExternal)
					}
					fmt.Printf("  %s=%s", s, file)
				}
			}
			fmt.Println()
		}
	}
}
