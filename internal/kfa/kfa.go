// Package kfa reads DAoC's recorded animations.
//
// The .kfa files in anims/ are not a separate format: they are NetImmerse
// 4.2.1.0 NIFs holding nothing but NiKeyframeController and NiKeyframeData
// blocks. The nodes are unnamed; the file's single real node carries two
// linked lists that run in lockstep, a chain of NiStringExtraData and a chain
// of NiKeyframeController. The n-th string belongs to the n-th controller, and
// its value is an index into figures/animnode.dat, a list of 3ds Max Biped
// bone names.
package kfa

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"daocweb/internal/nif"
)

// Track is one bone's keys.
type Track struct {
	Bone      string
	Rotations []nif.QuatKey
	Trans     []nif.VecKey
	Scales    []nif.FloatKey
	Start     float32
	Stop      float32
}

// Anim is a clip: its tracks, and the latest key time among them.
type Anim struct {
	Name     string
	Duration float32
	Tracks   []Track
}

// Nodes reads the bone registry. Line N is the name of bone index N.
func Nodes(game string) ([]string, error) {
	path := filepath.Join(game, "figures", "animnode.dat")
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open %s: %w", path, err)
	}
	defer f.Close()
	var out []string
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		out = append(out, strings.TrimSpace(sc.Text()))
	}
	return out, sc.Err()
}

// Read parses a .kfa, walks the two parallel chains and resolves every
// track to a bone.
func Read(path string, nodes []string) (*Anim, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	f, err := nif.Parse(raw)
	if err != nil {
		return nil, err
	}
	if len(f.Roots) == 0 {
		return nil, fmt.Errorf("no root block")
	}
	root, _ := f.Block(f.Roots[0]).(*nif.Node)
	if root == nil || len(root.ExtraData) == 0 {
		return nil, fmt.Errorf("root is not a node carrying the track chains")
	}
	a := &Anim{Name: strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))}
	s, c := root.ExtraData[0], root.Controller
	for s >= 0 && c >= 0 {
		se, _ := f.Block(s).(*nif.StringExtra)
		kc, _ := f.Block(c).(*nif.KeyframeController)
		if se == nil || kc == nil {
			break
		}
		idx, err := strconv.Atoi(strings.TrimSpace(se.Value))
		if err != nil {
			return nil, fmt.Errorf("track names bone %q, which is not an index", se.Value)
		}
		if idx < 0 || idx >= len(nodes) {
			return nil, fmt.Errorf("track names bone index %d, outside animnode.dat's %d slots", idx, len(nodes))
		}
		t := Track{Bone: nodes[idx], Start: kc.StartTime, Stop: kc.StopTime}
		if kd, _ := f.Block(kc.Data).(*nif.KeyframeData); kd != nil {
			t.Rotations, t.Trans, t.Scales = kd.Rotations, kd.Translations, kd.Scales
		}
		if t.Stop > a.Duration {
			a.Duration = t.Stop
		}
		a.Tracks = append(a.Tracks, t)
		s, c = se.Next, kc.Next
	}
	if len(a.Tracks) == 0 {
		return nil, fmt.Errorf("no tracks")
	}
	return a, nil
}
