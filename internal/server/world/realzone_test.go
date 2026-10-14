package world

import (
	"path/filepath"
	"testing"

	"daocweb/internal/server/store"
)

// Zone 100 with its imported population, ticked empty and then with a player.
func TestRealZone100Ticks(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	cat, err := LoadCatalog(filepath.Join(root, "server", "data", "world.db"))
	if err != nil {
		t.Skip("no world store: run go run ./cmd/dbimport")
	}
	st, _ := store.Open(":memory:")
	defer st.Close()
	w := New(Config{WebData: filepath.Join(root, "web", "data"), Cat: cat, Store: st, Log: t.Logf, Collision: OutdoorCollision})
	w.mu.Lock()
	z, err := w.zoneLocked(100)
	w.mu.Unlock()
	if err != nil {
		t.Skip("zone 100 is not converted:", err)
	}
	z.Stop() // we drive it by hand
	if len(z.npcs) < 1500 {
		t.Errorf("Mularn holds %d creatures, want its imported population", len(z.npcs))
	}
	for i := 0; i < 200; i++ {
		z.Tick() // with nobody there
	}
	a, _ := st.AccountFor("d:r", "r", false)
	ch := &store.Character{AccountID: a.ID, Name: "Realzone", Race: 6, Class: 22, Level: 5, Zone: 100, X: 40000, Y: 30000, HP: 500}
	st.CreateCharacter(ch)
	p := newPlayer(a, ch)
	z.join(p)
	for i := 0; i < 600; i++ {
		z.Tick()
	}
	if p.Body.Z < 2000 || p.Body.Z > 12000 {
		t.Errorf("player stands at z %v", p.Body.Z)
	}
	moved := 0
	for _, n := range z.npcs {
		if n.speedNow > 0 {
			moved++
		}
	}
	if moved == 0 {
		t.Error("in thirty seconds nobody in Mularn moved")
	}
}
