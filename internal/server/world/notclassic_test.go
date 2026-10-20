package world

import (
	"strings"
	"testing"

	"daocweb/internal/importer"
)

func TestNotClassicKinds(t *testing.T) {
	for _, c := range []string{"DOL.GS.Scripts.LiveTeleporter", "DOL.GS.Scripts.OFTeleporter", "DOL.GS.Scripts.OFAssistant", "DOL.GS.GameHastener",
		"DOL.GS.Keeps.FrontierHastener", "DOL.GS.AccountVaultKeeper", "DOL.GS.MarketExplorer", "DOL.GS.GameSapphireSealsMerchant",
		"DOL.GS.GameEmeraldSealsMerchant", "DOL.GS.GameDiamondSealsMerchant", "DOL.GS.GameBountyMerchant"} {
		if npcKind(c) != KindNotClassic {
			t.Errorf("npcKind(%q) = %q, want notclassic", c, npcKind(c))
		}
		if nc, _ := importer.NotClassicOf(c); nc.Reason == "" || nc.Source == "" {
			t.Errorf("%s has no reason or source", c)
		}
	}
	for _, c := range []string{"DOL.GS.GameVaultKeeper", "DOL.GS.Blacksmith", "DOL.GS.Enchanter", "DOL.GS.GameStableMaster"} {
		if npcKind(c) == KindNotClassic {
			t.Errorf("%s is wrongly marked not classic", c)
		}
	}
}

// A teleporter answers with a line and does not move the player; a guard is
// refused with a short line, not silence.
func TestNotClassicAnswersAndGuardRefuses(t *testing.T) {
	r := newRig(t,
		Spawn{ID: 700, Name: "Teleporter", Model: 56, X: 30000, Y: 30000, Z: 5000, Level: 10, Size: 50, Speed: 191, Realm: 2, ClassType: "DOL.GS.Scripts.LiveTeleporter"},
		Spawn{ID: 701, Name: "Guard", Model: 56, X: 30050, Y: 30000, Z: 5000, Level: 10, Size: 50, Speed: 191, Realm: 2, ClassType: "DOL.GS.GameGuard"})
	p := r.player("Alice", 5, 30080, 30000, false)
	x, y := p.Body.X, p.Body.Y
	r.z.Interact(p, 700)
	types, raws := sent(p)
	said := false
	for i, ty := range types {
		said = said || ty == "chatmsg" && strings.Contains(string(raws[i]), "not part of this world")
	}
	if !said || p.Body.X != x || p.Body.Y != y {
		t.Errorf("teleporter answer %v, moved %v,%v -> %v,%v; want the line and no move", said, x, y, p.Body.X, p.Body.Y)
	}
	r.z.Interact(p, 701)
	if lastErr(t, p) == "" {
		t.Error("a guard's interaction was met with silence")
	}
}
