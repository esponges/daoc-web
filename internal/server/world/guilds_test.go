package world

import (
	"fmt"
	"testing"

	"daocweb/internal/server/rules"
)

// guildRig is a zone with n players who are online and loaded.
func guildRig(t *testing.T, n int) (*grouped, []*Player) {
	g := newGrouped(t)
	var ps []*Player
	for i := 0; i < n; i++ {
		p := g.join(fmt.Sprintf("Viking%c", 'A'+i), float64(30000+i*50))
		g.w.loadSocial(p)
		ps = append(ps, p)
	}
	return g, ps
}

// groupAll puts every player in the first one's group.
func groupAll(g *grouped, ps []*Player) {
	for _, p := range ps[1:] {
		g.cmd(ps[0], "invite "+p.Char.Name)
		g.cmd(p, "accept")
	}
}

func TestFormingAGuildNeedsAFullGroupAndEveryone(t *testing.T) {
	g, ps := guildRig(t, 8)
	lead := ps[0]
	lead.Admin = true // stands in for being at a registrar
	lead.Char.Money = 30000
	groupAll(g, ps[:5])
	drain(lead)
	g.cmd(lead, "gc form Hammers of Odin")
	if out, _, _ := inbox(lead); !has(out, "lead a group of 8") {
		t.Errorf("a short group was not refused: %v", out)
	}
	groupAll(g, ps)
	for _, p := range ps[5:] {
		_ = p
	}
	g.cmd(lead, "gc form Hammers of Odin")
	for _, p := range ps[1:] {
		if _, _, in := g.w.guildOf(p.Char.ID); in {
			t.Fatal("someone was in the guild before agreeing")
		}
		g.cmd(p, "accept")
	}
	gd, me, ok := g.w.guildOf(lead.Char.ID)
	if !ok || gd.Name != "Hammers of Odin" || me.Rank != 0 {
		t.Fatalf("the leader's place: %+v %+v %v", gd, me, ok)
	}
	if _, m2, ok := g.w.guildOf(ps[3].Char.ID); !ok || m2.Rank != 9 {
		t.Errorf("a member's place: %+v %v", m2, ok)
	}
	if lead.Char.Money != 30000-rules.GuildFormFee {
		t.Errorf("the fee: leader has %d", lead.Char.Money)
	}
	// It was written, and a fresh world reads it.
	g.w.guilds = newGuildBook()
	g.w.loadGuilds()
	if _, _, ok := g.w.guildOf(ps[7].Char.ID); !ok {
		t.Error("the guild was not saved")
	}
}

func TestOneRefusalStopsTheGuild(t *testing.T) {
	g, ps := guildRig(t, 8)
	ps[0].Admin = true
	ps[0].Char.Money = 30000
	groupAll(g, ps)
	g.cmd(ps[0], "gc form Sons of Tyr")
	for _, p := range ps[1:7] {
		g.cmd(p, "accept")
	}
	g.cmd(ps[7], "decline")
	if _, _, ok := g.w.guildOf(ps[0].Char.ID); ok {
		t.Error("a guild was formed over a refusal")
	}
	if ps[0].Char.Money != 30000 {
		t.Error("the leader was charged for a guild that was not formed")
	}
}

func TestRegistrarIsNeededToForm(t *testing.T) {
	g, ps := guildRig(t, 8)
	ps[0].Char.Money = 30000
	groupAll(g, ps)
	drain(ps[0])
	g.cmd(ps[0], "gc form Away From Home")
	if out, _, _ := inbox(ps[0]); !has(out, "Guild Registrar") {
		t.Errorf("forming away from a registrar: %v", out)
	}
}

// makeTestGuild founds a guild of the first n players, the first as master.
func makeTestGuild(t *testing.T, g *grouped, ps []*Player, name string) {
	t.Helper()
	ids := make([]int64, len(ps))
	for i, p := range ps {
		ids[i] = p.Char.ID
	}
	if err := g.w.makeGuild(name, ps[0].Arch.Realm, ids); err != nil {
		t.Fatal(err)
	}
	g.w.refreshGuild(func() int64 { gd, _, _ := g.w.guildOf(ps[0].Char.ID); return gd.ID }())
	for _, p := range ps {
		drain(p)
	}
}

func TestInviteAndRankLimits(t *testing.T) {
	g, ps := guildRig(t, 4)
	makeTestGuild(t, g, ps[:3], "Hammers of Odin")
	master, officer, newbie, outsider := ps[0], ps[1], ps[2], ps[3]
	// The master makes the second member an officer (rank 1), then a rank 3 one.
	g.w.guilds.of[officer.Char.ID].Rank = 3
	g.cmd(officer, "gc invite "+outsider.Char.Name)
	g.cmd(outsider, "accept")
	if _, m, ok := g.w.guildOf(outsider.Char.ID); !ok || m.Rank != 9 {
		t.Fatalf("an invited player did not join at rank 9: %+v %v", m, ok)
	}
	// A rank 3 officer cannot promote a fellow rank 3 or the master.
	drain(officer)
	g.cmd(officer, "gc promote "+master.Char.Name)
	g.cmd(officer, "gc promote "+newbie.Char.Name) // 9 -> 8, allowed
	if _, m, _ := g.w.guildOf(newbie.Char.ID); m.Rank != 8 {
		t.Errorf("a rank 3 officer could not promote a rank 9 to 8: %d", m.Rank)
	}
	g.w.guilds.of[newbie.Char.ID].Rank = 3
	drain(officer)
	g.cmd(officer, "gc promote "+newbie.Char.Name)
	if out, _, _ := inbox(officer); !has(out, "may not") {
		t.Errorf("promoting an equal rank was not refused: %v", out)
	}
	if _, m, _ := g.w.guildOf(master.Char.ID); m.Rank != 0 {
		t.Error("the master's rank changed")
	}
	// Only members below can be removed.
	g.cmd(officer, "gc remove "+master.Char.Name)
	if _, _, ok := g.w.guildOf(master.Char.ID); !ok {
		t.Error("the master was removed by an officer")
	}
	g.cmd(officer, "gc remove "+outsider.Char.Name)
	if _, _, ok := g.w.guildOf(outsider.Char.ID); ok {
		t.Error("a rank 9 member was not removed")
	}
}

func TestGuildChannelsReachTheRightRanks(t *testing.T) {
	g, ps := guildRig(t, 3)
	makeTestGuild(t, g, ps, "Hammers of Odin")
	master, member := ps[0], ps[2]
	g.w.guilds.of[ps[1].Char.ID].Rank = 1 // an officer
	g.cmd(master, "o officers only")
	if out, _, _ := inbox(ps[1]); !has(out, "officers only") {
		t.Errorf("an officer missed officer chat: %v", out)
	}
	if out, _, _ := inbox(member); has(out, "officers only") {
		t.Error("a plain member heard officer chat")
	}
	g.cmd(master, "gu everyone")
	if out, _, _ := inbox(member); !has(out, "everyone") {
		t.Errorf("a member missed guild chat: %v", out)
	}
	drain(master)
	g.cmd(member, "o can I?")
	if out, _, _ := inbox(member); !has(out, "may not speak") {
		t.Errorf("a member spoke on officer chat: %v", out)
	}
}

func TestAlliancesShareAChannel(t *testing.T) {
	g, ps := guildRig(t, 4)
	makeTestGuild(t, g, ps[:2], "Hammers of Odin")
	makeTestGuild(t, g, ps[2:], "Sons of Tyr")
	g.z.SetTarget(ps[0], ps[2].ID)
	g.cmd(ps[0], "gc ally")
	g.cmd(ps[2], "accept")
	lead, _, _ := g.w.guildOf(ps[0].Char.ID)
	other, _, _ := g.w.guildOf(ps[2].Char.ID)
	if lead.Alliance != lead.ID || other.Alliance != lead.ID {
		t.Fatalf("alliance ids: %d and %d, leader %d", lead.Alliance, other.Alliance, lead.ID)
	}
	drain(ps[3])
	g.cmd(ps[1], "as hello allies")
	if out, _, _ := inbox(ps[3]); !has(out, "hello allies") {
		t.Errorf("an allied guild missed alliance chat: %v", out)
	}
	// Only the leading guild's master may end it for all.
	g.cmd(ps[0], "gc allyoff")
	if o, _, _ := g.w.guildOf(ps[2].Char.ID); o.Alliance != 0 {
		t.Error("the alliance survived its leader ending it")
	}
}

func TestLeavingHandsOverTheGuild(t *testing.T) {
	g, ps := guildRig(t, 3)
	makeTestGuild(t, g, ps, "Hammers of Odin")
	g.w.guilds.of[ps[1].Char.ID].Rank = 2
	g.cmd(ps[0], "gc leave")
	if _, m, ok := g.w.guildOf(ps[1].Char.ID); !ok || m.Rank != 0 {
		t.Errorf("the best ranked member did not take over: %+v %v", m, ok)
	}
	g.cmd(ps[1], "gc leave")
	g.cmd(ps[2], "gc leave")
	if gs, _ := g.w.Store.Guilds(); len(gs) != 0 {
		t.Errorf("an empty guild was kept: %d", len(gs))
	}
}

func TestNamePlateAndWho(t *testing.T) {
	g, ps := guildRig(t, 2)
	makeTestGuild(t, g, ps[:1], "Hammers of Odin")
	if got := g.z.playerEntity(ps[0]).Guild; got != "<Hammers of Odin>" {
		t.Errorf("name plate %q", got)
	}
	if got := g.z.playerEntity(ps[1]).Guild; got != "" {
		t.Errorf("a guildless player shows %q", got)
	}
	drain(ps[1])
	g.cmd(ps[1], "who")
	if out, _, _ := inbox(ps[1]); !has(out, "<Hammers of Odin>") {
		t.Errorf("/who does not show the guild: %v", out)
	}
}
