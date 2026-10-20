package importer

// NotClassic says why an NPC class type is not part of the 1.65 world: the
// server still spawns and draws these NPCs, and they answer an interaction
// with a line saying their service is not part of this world.
type NotClassic struct {
	Reason string
	// Source is what dates the service after 1.65. Until one is recorded the
	// type stays on the list only because none of the adopted 1.65 sources
	// shows it; a type leaves the list when a source does (see the README).
	Source string
}

const noSource = "none yet: no adopted 1.65 source shows it"

// NotClassicKinds maps class type to its entry. Enchanters, rechargers and
// craft masters are not here: crafting is a later change, and whether
// enchanters were in 1.65 is not settled.
var NotClassicKinds = map[string]NotClassic{
	"DOL.GS.Scripts.LiveTeleporter":    {"a teleporter: Live-era travel the horse routes replace", noSource},
	"DOL.GS.Scripts.OFTeleporter":      {"a teleporter: Origin Fantasy convenience travel", noSource},
	"DOL.GS.Scripts.OFAssistant":       {"a teleporter's assistant", noSource},
	"DOL.GS.GameHastener":              {"a hastener: a speed boost on request", noSource},
	"DOL.GS.Keeps.FrontierHastener":    {"a hastener: a speed boost on request", noSource},
	"DOL.GS.AccountVaultKeeper":        {"an account vault: shared storage across characters", noSource},
	"DOL.GS.MarketExplorer":            {"the market explorer: a player market", noSource},
	"DOL.GS.GameSapphireSealsMerchant": {"a seals merchant: Atlas-era currency", noSource},
	"DOL.GS.GameEmeraldSealsMerchant":  {"a seals merchant: Atlas-era currency", noSource},
	"DOL.GS.GameDiamondSealsMerchant":  {"a seals merchant: Atlas-era currency", noSource},
	"DOL.GS.GameBountyMerchant":        {"a bounty merchant: a later reward shop", noSource},
}

// NotClassicOf looks a class type up in the not-classic list.
func NotClassicOf(classType string) (NotClassic, bool) {
	n, ok := NotClassicKinds[classType]
	return n, ok
}
