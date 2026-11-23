package rules

import "testing"

func TestGuildRules(t *testing.T) {
	const c = "player-guilds"
	vector(t, c, "Formed", func(t *testing.T) {
		if GuildFormMembers != 8 || GuildFormFee <= 0 {
			t.Error("a guild needs 8 players and costs something")
		}
		for _, n := range []string{"Hammers of Odin", "Sons of Tyr", "Ask"} {
			if !ValidGuildName(n) {
				t.Errorf("%q should pass", n)
			}
		}
		for _, n := range []string{"", "Ab", "1Guild", "Two  Spaces", "Trailing ", "Bad<Name>"} {
			if ValidGuildName(n) {
				t.Errorf("%q should fail", n)
			}
		}
	})
	vector(t, c, "Short group", func(t *testing.T) {
		if 5 >= GuildFormMembers {
			t.Error("5 players should be too few")
		}
	})
	vector(t, c, "Promotion limit", func(t *testing.T) {
		if MayPromote(3, 5, 2) {
			t.Error("rank 3 promoted someone to rank 2, above themselves")
		}
		if !MayPromote(3, 9, 4) || MayPromote(3, 3, 2) || MayPromote(3, 2, 1) {
			t.Error("promoting below oneself, and only members below, is the rule")
		}
		if !MayPromote(0, 9, 1) {
			t.Error("the Guildmaster may promote to rank 1")
		}
		if !MayDemote(1, 5) || MayDemote(1, 1) || MayDemote(1, 9) {
			t.Error("demotion limits")
		}
		if !MayRemove(2, 9) || MayRemove(2, 2) {
			t.Error("removal limits")
		}
		if DefaultPerms(0)&PermEmblem == 0 || DefaultPerms(2)&PermEmblem != 0 || DefaultPerms(9)&PermInvite != 0 {
			t.Error("default permissions")
		}
	})
}
