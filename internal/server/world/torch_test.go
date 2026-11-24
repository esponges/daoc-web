package world

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTorch(t *testing.T) {
	r := newRig(t)
	p := r.player("Bearer", 20, 30000, 30000, false)
	inbox(p)
	if !r.z.playerEntity(p).Torch {
		t.Fatal("a new player's torch is not lit")
	}
	r.z.Torch(p, "")
	if r.z.playerEntity(p).Torch || !strings.Contains(heard(p), "You put out your torch.") {
		t.Error("/torch did not put the torch out")
	}
	r.z.Torch(p, "")
	if !r.z.playerEntity(p).Torch || !strings.Contains(heard(p), "You light your torch.") {
		t.Error("/torch again did not light it")
	}
	r.z.Torch(p, "off")
	r.z.Torch(p, "off")
	if r.z.playerEntity(p).Torch {
		t.Error("/torch off left it lit")
	}
	r.z.Torch(p, "on")
	if !r.z.playerEntity(p).Torch {
		t.Error("/torch on left it out")
	}
	b, _ := json.Marshal(r.z.playerEntity(p))
	if !strings.Contains(string(b), `"tc":true`) {
		t.Errorf("a lit torch is not on the wire: %s", b)
	}
	p.dead = true
	if r.z.playerEntity(p).Torch {
		t.Error("a dead player carries a lit torch")
	}
	p.dead = false
	r.z.Torch(p, "bright")
	if !strings.Contains(heard(p), "Use /torch") {
		t.Error("a bad argument was not answered")
	}
}
