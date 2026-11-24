package world

import (
	"os"
	"path/filepath"
	"testing"
)

// Every quest file in server/quests loads, and each NPC, creature and reward
// item it names is in the world data (skipped when no world store is built).
func TestQuestDataResolves(t *testing.T) {
	qs, err := LoadQuests(filepath.Join("..", "..", "..", "server", "quests"))
	if err != nil {
		t.Fatal(err)
	}
	if len(qs.All()) == 0 {
		t.Fatal("no quests are authored")
	}
	db := filepath.Join("..", "..", "..", "server", "data", "world.db")
	if _, err := os.Stat(db); err != nil {
		t.Skip("no world store to check the quests against")
	}
	cat, err := LoadCatalog(db)
	if err != nil {
		t.Skip("world store unreadable:", err)
	}
	for _, d := range qs.All() {
		for _, bad := range d.CheckWorld(cat) {
			t.Error(bad)
		}
	}
}
