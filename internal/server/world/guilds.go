package world

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"daocweb/internal/server/protocol"
	"daocweb/internal/server/rules"
	"daocweb/internal/server/store"
)

const (
	KindGuildRegistrar = "guildregistrar"
	KindEmblem         = "emblem"

	guildAnswerLapse = 60 * time.Second
	EmblemFee        = 5000 // copper to buy or change an emblem. Provisional.
)

// guildBook is the world's guilds, kept in memory with the store behind them.
// Zones call in from their own goroutines, so it has a lock of its own; it may
// take the online lock while held, never the other way round.
type guildBook struct {
	mu      sync.Mutex
	byID    map[int64]*store.Guild
	of      map[int64]*store.GuildMember // by character
	forms   map[int64]*guildForm         // by the leader's character
	invites map[int64]*guildInvite       // by the invited character
	clock   func() time.Time
}

// guildForm is a guild waiting for every member of a group to agree.
type guildForm struct {
	name    string
	members []int64 // the leader first
	ok      map[int64]bool
	expires time.Time
}

// guildInvite is an invitation to a guild, or to an alliance when ally is set.
type guildInvite struct {
	guild   int64
	from    string
	ally    bool
	expires time.Time
}

func newGuildBook() *guildBook {
	return &guildBook{byID: map[int64]*store.Guild{}, of: map[int64]*store.GuildMember{}, forms: map[int64]*guildForm{},
		invites: map[int64]*guildInvite{}, clock: time.Now}
}

// loadGuilds reads every guild and membership from the store.
func (w *World) loadGuilds() {
	if w.Store == nil {
		return
	}
	gs, err := w.Store.Guilds()
	if err != nil {
		w.Log("guilds: %v", err)
		return
	}
	ms, _ := w.Store.GuildMembers()
	b := w.guilds
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, g := range gs {
		b.byID[g.ID] = g
	}
	for i := range ms {
		m := ms[i]
		b.of[m.CharID] = &m
	}
}

func canDo(g *store.Guild, rank, perm int) bool { return g.Perms[rank]&perm != 0 }

// guildOf is a character's guild and place in it, as copies.
func (w *World) guildOf(char int64) (g store.Guild, m store.GuildMember, ok bool) {
	b := w.guilds
	b.mu.Lock()
	defer b.mu.Unlock()
	mm := b.of[char]
	if mm == nil {
		return g, m, false
	}
	return *b.byID[mm.GuildID], *mm, true
}

// guildTag is the line under a player's name: "<Hammers of Odin>", or "".
func (w *World) guildTag(char int64) string {
	if g, _, ok := w.guildOf(char); ok {
		return "<" + g.Name + ">"
	}
	return ""
}

// setTag refreshes the guild line other players see for a character.
func (w *World) setTag(char int64) {
	if p := w.onlineChar(char); p != nil {
		p.guildTag.Store(w.guildTag(char))
	}
}

// memberChars lists a guild's members, best rank first.
func (b *guildBook) memberChars(gid int64) []int64 {
	var out []*store.GuildMember
	for _, m := range b.of {
		if m.GuildID == gid {
			out = append(out, m)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Rank != out[j].Rank {
			return out[i].Rank < out[j].Rank
		}
		return out[i].CharID < out[j].CharID
	})
	ids := make([]int64, len(out))
	for i, m := range out {
		ids[i] = m.CharID
	}
	return ids
}

// guildMsg is a member's view of their guild, or an empty one for no guild.
func (w *World) guildMsg(char int64) protocol.GuildInfo {
	g, me, ok := w.guildOf(char)
	if !ok {
		return protocol.GuildInfo{T: protocol.TGuild, Members: []protocol.GuildMember{}}
	}
	m := protocol.GuildInfo{T: protocol.TGuild, Name: g.Name, Rank: me.Rank, Title: g.Titles[me.Rank], Motd: g.Motd, Perms: g.Perms[me.Rank],
		Emblem: g.Emblem, Members: []protocol.GuildMember{}}
	if canDo(&g, me.Rank, rules.PermOfficerHear) {
		m.Omotd = g.OfficerM
	}
	b := w.guilds
	b.mu.Lock()
	if g.Alliance != 0 {
		if lead := b.byID[g.Alliance]; lead != nil {
			m.Alliance = lead.Name
		}
	}
	ids := b.memberChars(g.ID)
	notes := map[int64]store.GuildMember{}
	for _, id := range ids {
		notes[id] = *b.of[id]
	}
	b.mu.Unlock()
	for _, id := range ids {
		mm := notes[id]
		v := protocol.GuildMember{Name: w.Store.CharName(id), Rank: mm.Rank, Note: mm.Note}
		if o := w.onlineChar(id); o != nil {
			v.On, v.Level, v.Class = true, o.Char.Level, o.sh.class.Name
		}
		m.Members = append(m.Members, v)
	}
	return m
}

// refreshGuild sends every online member of a guild its current state and
// updates their name plates.
func (w *World) refreshGuild(gid int64) {
	b := w.guilds
	b.mu.Lock()
	ids := b.memberChars(gid)
	b.mu.Unlock()
	for _, id := range ids {
		if p := w.onlineChar(id); p != nil {
			w.setTag(id)
			p.send(w.guildMsg(id))
		}
	}
}

// sayGuild tells every online member of a guild.
func (w *World) sayGuild(gid int64, text string) {
	b := w.guilds
	b.mu.Lock()
	ids := b.memberChars(gid)
	b.mu.Unlock()
	for _, id := range ids {
		sys(w.onlineChar(id), text)
	}
}

// guildLogin sends a player their guild at login, with the message of the day.
func (w *World) guildLogin(p *Player) {
	w.setTag(p.Char.ID)
	m := w.guildMsg(p.Char.ID)
	p.send(m)
	if m.Name != "" && m.Motd != "" {
		p.send(protocol.ChatMsg{T: protocol.TChatMsg, Ch: "guild", From: "", Text: "Guild message: " + m.Motd})
	}
}

// guildCmd runs /gc <sub> ...
func (w *World) guildCmd(z *Zone, p *Player, f []string, line string) {
	if len(f) < 2 {
		sys(p, "Guild commands: form, invite, remove, leave, promote, demote, motd, omotd, emblem, ally, allyoff, note, info.")
		return
	}
	sub := strings.ToLower(f[1])
	arg := restAfter(line, 2)
	switch sub {
	case "form":
		w.guildForm(z, p, arg)
	case "invite":
		w.guildInvite(z, p, arg)
	case "leave":
		w.guildLeave(p)
	case "remove":
		w.guildRemove(p, arg)
	case "promote":
		w.guildRank(p, arg, +1)
	case "demote":
		w.guildRank(p, arg, -1)
	case "motd":
		w.guildMotd(p, arg, false)
	case "omotd":
		w.guildMotd(p, arg, true)
	case "note":
		w.guildNote(p, arg)
	case "emblem":
		w.guildEmblem(z, p, arg)
	case "ally":
		w.guildAlly(z, p)
	case "allyoff":
		w.guildAllyOff(p)
	case "info":
		g, me, ok := w.guildOf(p.Char.ID)
		if !ok {
			sys(p, "You are not in a guild.")
			return
		}
		sys(p, fmt.Sprintf("%s: you are %s (rank %d).", g.Name, g.Titles[me.Rank], me.Rank))
		w.sendGuildTo(p)
	default:
		sys(p, "Unknown guild command "+sub+".")
	}
}

func (w *World) sendGuildTo(p *Player) { p.send(w.guildMsg(p.Char.ID)) }

// nearNPC finds an NPC of a kind within reach of the player.
func (z *Zone) nearNPC(p *Player, kind string) *NPC {
	for _, n := range z.npcs {
		if n.Kind == kind && !n.dead() && !n.gone && n.inReach(p) {
			return n
		}
	}
	return nil
}

// guildForm starts forming a guild: the leader of a full group, at a registrar.
func (w *World) guildForm(z *Zone, p *Player, name string) {
	switch {
	case !rules.ValidGuildName(name):
		sys(p, "A guild name is 3 to 30 letters, spaces and apostrophes.")
		return
	case z.nearNPC(p, KindGuildRegistrar) == nil && !p.Admin:
		sys(p, "You must be near a Guild Registrar to form a guild.")
		return
	}
	if _, _, in := w.guildOf(p.Char.ID); in {
		sys(p, "You are already in a guild.")
		return
	}
	members := w.groupOf(p.Char.ID)
	if len(members) < rules.GuildFormMembers || members[0] != p.Char.ID {
		sys(p, fmt.Sprintf("You must lead a group of %d players to form a guild.", rules.GuildFormMembers))
		return
	}
	if p.Char.Money < rules.GuildFormFee {
		sys(p, "Forming a guild costs "+moneyText(rules.GuildFormFee)+".")
		return
	}
	for _, c := range members {
		if o := w.onlineChar(c); o == nil {
			sys(p, "Every member of your group must be in the game.")
			return
		} else if _, _, in := w.guildOf(c); in {
			sys(p, o.Char.Name+" is already in a guild.")
			return
		} else if o.Arch.Realm != p.Arch.Realm {
			sys(p, o.Char.Name+" is of another realm.")
			return
		}
	}
	if w.guildNameTaken(name) {
		sys(p, "That guild name is taken.")
		return
	}
	b := w.guilds
	b.mu.Lock()
	b.forms[p.Char.ID] = &guildForm{name: name, members: members, ok: map[int64]bool{p.Char.ID: true}, expires: b.clock().Add(guildAnswerLapse)}
	b.mu.Unlock()
	for _, c := range members[1:] {
		sys(w.onlineChar(c), p.Char.Name+" asks you to found the guild \""+name+"\". Type /accept to agree or /decline.")
	}
	sys(p, "You ask your group to found \""+name+"\". Every member must agree.")
}

func (w *World) guildNameTaken(name string) bool {
	b := w.guilds
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, g := range b.byID {
		if strings.EqualFold(g.Name, name) {
			return true
		}
	}
	return false
}

// answerGuild handles /accept and /decline for a guild's formation, invitation
// or alliance offer; it returns whether there was one.
func (w *World) answerGuild(p *Player, accept bool) bool {
	b := w.guilds
	b.mu.Lock()
	now := b.clock()
	// A founding the player has been asked to join.
	for leader, f := range b.forms {
		if leader == p.Char.ID {
			continue
		}
		in := false
		for _, c := range f.members {
			in = in || c == p.Char.ID
		}
		if !in || f.ok[p.Char.ID] {
			continue
		}
		if !accept || now.After(f.expires) {
			delete(b.forms, leader)
			b.mu.Unlock()
			sys(p, "You decline.")
			sys(w.onlineChar(leader), p.Char.Name+" does not agree: the guild is not formed.")
			return true
		}
		f.ok[p.Char.ID] = true
		done := len(f.ok) == len(f.members)
		b.mu.Unlock()
		sys(p, "You agree to found \""+f.name+"\".")
		if done {
			w.finishForm(leader, f)
		}
		return true
	}
	inv := b.invites[p.Char.ID]
	if inv == nil {
		b.mu.Unlock()
		return false
	}
	delete(b.invites, p.Char.ID)
	b.mu.Unlock()
	if !accept {
		sys(p, "You decline.")
		return true
	}
	if now.After(inv.expires) {
		sys(p, "The offer has lapsed.")
		return true
	}
	if inv.ally {
		w.joinAlliance(p, inv)
		return true
	}
	if _, _, in := w.guildOf(p.Char.ID); in {
		sys(p, "You are already in a guild.")
		return true
	}
	b.mu.Lock()
	g := b.byID[inv.guild]
	if g == nil {
		b.mu.Unlock()
		sys(p, "That guild is gone.")
		return true
	}
	m := &store.GuildMember{CharID: p.Char.ID, GuildID: g.ID, Rank: 9}
	b.of[p.Char.ID] = m
	b.mu.Unlock()
	w.Store.SetGuildMember(*m)
	w.sayGuild(g.ID, p.Char.Name+" has joined the guild.")
	w.refreshGuild(g.ID)
	return true
}

// finishForm creates the guild when every member has agreed, charging the leader.
func (w *World) finishForm(leader int64, f *guildForm) {
	lp := w.onlineChar(leader)
	b := w.guilds
	b.mu.Lock()
	delete(b.forms, leader)
	b.mu.Unlock()
	if lp == nil || lp.Char.Money < rules.GuildFormFee || w.guildNameTaken(f.name) {
		for _, c := range f.members {
			sys(w.onlineChar(c), "The guild could not be formed.")
		}
		return
	}
	if err := w.makeGuild(f.name, lp.Arch.Realm, f.members); err != nil {
		for _, c := range f.members {
			sys(w.onlineChar(c), "The guild could not be formed: "+err.Error())
		}
		return
	}
	lp.Char.Money -= rules.GuildFormFee
	w.Store.Tx(func(tx *store.Tx) error { return tx.SaveCharacter(lp.Char) })
	if lp.Zone != nil {
		lp.sendInv(w.Cat, lp.Char.Money)
	}
	b.mu.Lock()
	var gid int64
	for _, g := range b.byID {
		if g.Name == f.name {
			gid = g.ID
		}
	}
	b.mu.Unlock()
	w.sayGuild(gid, "The guild \""+f.name+"\" is founded!")
	w.refreshGuild(gid)
}

// makeGuild writes a new guild with default ranks and the members given, the
// first as Guildmaster.
func (w *World) makeGuild(name string, realm int, members []int64) error {
	g := &store.Guild{Name: name, Realm: realm}
	for r := 0; r < rules.GuildRanks; r++ {
		g.Perms[r], g.Titles[r] = rules.DefaultPerms(r), rules.DefaultTitle(r)
	}
	if err := w.Store.CreateGuild(g, members); err != nil {
		return err
	}
	b := w.guilds
	b.mu.Lock()
	b.byID[g.ID] = g
	for i, c := range members {
		rank := 9
		if i == 0 {
			rank = 0
		}
		b.of[c] = &store.GuildMember{CharID: c, GuildID: g.ID, Rank: rank}
	}
	b.mu.Unlock()
	return nil
}

// guildInvite invites a player (named, or the one targeted) into the guild.
func (w *World) guildInvite(z *Zone, p *Player, name string) {
	g, me, ok := w.guildOf(p.Char.ID)
	if !ok || !canDo(&g, me.Rank, rules.PermInvite) {
		sys(p, "You cannot invite anyone to a guild.")
		return
	}
	var to *Player
	if name != "" {
		to = w.onlineByName(name)
	} else {
		to = z.players[p.target]
	}
	switch {
	case to == nil || to == p:
		sys(p, "Name a player, or target one, to invite.")
		return
	case to.Arch.Realm != p.Arch.Realm:
		sys(p, "You cannot invite that player.")
		return
	}
	if _, _, in := w.guildOf(to.Char.ID); in {
		sys(p, to.Char.Name+" is already in a guild.")
		return
	}
	b := w.guilds
	if len(b.memberCharsLocked(g.ID)) >= rules.GuildMaxMembers {
		sys(p, "The guild is full.")
		return
	}
	b.mu.Lock()
	b.invites[to.Char.ID] = &guildInvite{guild: g.ID, from: p.Char.Name, expires: b.clock().Add(guildAnswerLapse)}
	b.mu.Unlock()
	sys(p, "You invite "+to.Char.Name+" to "+g.Name+".")
	sys(to, p.Char.Name+" invites you to the guild \""+g.Name+"\". Type /accept or /decline.")
}

func (b *guildBook) memberCharsLocked(gid int64) []int64 {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.memberChars(gid)
}

// guildLeave takes the player out of their guild. If they led it, the best
// ranked member takes over, or the guild ends when it was only them.
func (w *World) guildLeave(p *Player) {
	g, _, ok := w.guildOf(p.Char.ID)
	if !ok {
		sys(p, "You are not in a guild.")
		return
	}
	w.dropMember(g.ID, p.Char.ID, p.Char.Name+" has left the guild.")
	sys(p, "You leave the guild.")
}

// dropMember removes a character from a guild with a line to the rest.
func (w *World) dropMember(gid, char int64, line string) {
	b := w.guilds
	b.mu.Lock()
	m := b.of[char]
	if m == nil {
		b.mu.Unlock()
		return
	}
	wasMaster := m.Rank == 0
	delete(b.of, char)
	left := b.memberChars(gid)
	b.mu.Unlock()
	w.Store.RemoveGuildMember(char)
	w.setTag(char)
	if p := w.onlineChar(char); p != nil {
		p.send(w.guildMsg(char))
	}
	if len(left) == 0 {
		b.mu.Lock()
		delete(b.byID, gid)
		b.mu.Unlock()
		w.Store.DeleteGuild(gid)
		return
	}
	if wasMaster {
		b.mu.Lock()
		b.of[left[0]].Rank = 0
		nm := *b.of[left[0]]
		b.mu.Unlock()
		w.Store.SetGuildMember(nm)
		line += " " + w.Store.CharName(left[0]) + " is the new Guildmaster."
	}
	w.sayGuild(gid, line)
	w.refreshGuild(gid)
}

// target finds a member of the player's guild by name.
func (w *World) guildMemberNamed(gid int64, name string) (int64, bool) {
	id, _ := w.Store.CharID(name)
	if id == 0 {
		return 0, false
	}
	g, _, ok := w.guildOf(id)
	return id, ok && g.ID == gid
}

func (w *World) guildRemove(p *Player, name string) {
	g, me, ok := w.guildOf(p.Char.ID)
	if !ok || !canDo(&g, me.Rank, rules.PermRemove) {
		sys(p, "You cannot remove members.")
		return
	}
	id, in := w.guildMemberNamed(g.ID, name)
	if !in || id == p.Char.ID {
		sys(p, name+" is not a member you can remove.")
		return
	}
	_, tm, _ := w.guildOf(id)
	if !rules.MayRemove(me.Rank, tm.Rank) {
		sys(p, "You can remove only members ranked below you.")
		return
	}
	w.dropMember(g.ID, id, w.Store.CharName(id)+" has been removed from the guild.")
	sys(w.onlineChar(id), "You have been removed from the guild.")
}

// guildRank promotes (dir +1) or demotes (-1) a member.
func (w *World) guildRank(p *Player, name string, dir int) {
	g, me, ok := w.guildOf(p.Char.ID)
	if !ok || !canDo(&g, me.Rank, rules.PermPromote) {
		sys(p, "You cannot change ranks.")
		return
	}
	id, in := w.guildMemberNamed(g.ID, name)
	if !in {
		sys(p, name+" is not in your guild.")
		return
	}
	_, tm, _ := w.guildOf(id)
	newRank := tm.Rank - dir
	if dir > 0 && !rules.MayPromote(me.Rank, tm.Rank, newRank) || dir < 0 && !rules.MayDemote(me.Rank, tm.Rank) {
		sys(p, "You may not do that: ranks are changed only below your own.")
		return
	}
	b := w.guilds
	b.mu.Lock()
	b.of[id].Rank = newRank
	nm := *b.of[id]
	b.mu.Unlock()
	w.Store.SetGuildMember(nm)
	w.sayGuild(g.ID, fmt.Sprintf("%s is now %s.", w.Store.CharName(id), g.Titles[newRank]))
	w.refreshGuild(g.ID)
}

func (w *World) guildMotd(p *Player, text string, officer bool) {
	g, me, ok := w.guildOf(p.Char.ID)
	if !ok || !canDo(&g, me.Rank, rules.PermMotd) {
		sys(p, "You cannot set the message.")
		return
	}
	b := w.guilds
	b.mu.Lock()
	gg := b.byID[g.ID]
	if officer {
		gg.OfficerM = text
	} else {
		gg.Motd = text
	}
	cp := *gg
	b.mu.Unlock()
	w.Store.SaveGuild(&cp)
	sys(p, "The message is set.")
	w.refreshGuild(g.ID)
}

func (w *World) guildNote(p *Player, text string) {
	_, _, ok := w.guildOf(p.Char.ID)
	if !ok {
		sys(p, "You are not in a guild.")
		return
	}
	b := w.guilds
	b.mu.Lock()
	b.of[p.Char.ID].Note = text
	nm := *b.of[p.Char.ID]
	b.mu.Unlock()
	w.Store.SetGuildMember(nm)
	w.refreshGuild(nm.GuildID)
}

// guildEmblem buys or changes the emblem at an Emblem NPC.
func (w *World) guildEmblem(z *Zone, p *Player, arg string) {
	g, me, ok := w.guildOf(p.Char.ID)
	if !ok || !canDo(&g, me.Rank, rules.PermEmblem) {
		sys(p, "You cannot choose the guild's emblem.")
		return
	}
	n, err := strconv.Atoi(arg)
	if err != nil || n < 1 || n > rules.GuildEmblems {
		sys(p, fmt.Sprintf("Choose an emblem from 1 to %d.", rules.GuildEmblems))
		return
	}
	if z.nearNPC(p, KindEmblem) == nil {
		sys(p, "You must be near an Emblem Maker.")
		return
	}
	if p.Char.Money < EmblemFee {
		sys(p, "An emblem costs "+moneyText(EmblemFee)+".")
		return
	}
	p.Char.Money -= EmblemFee
	w.Store.Tx(func(tx *store.Tx) error { return tx.SaveCharacter(p.Char) })
	p.sendInv(w.Cat, p.Char.Money)
	b := w.guilds
	b.mu.Lock()
	gg := b.byID[g.ID]
	gg.Emblem = n
	cp := *gg
	b.mu.Unlock()
	w.Store.SaveGuild(&cp)
	w.sayGuild(g.ID, fmt.Sprintf("The guild's emblem is now number %d.", n))
	w.refreshGuild(g.ID)
}

// guildAlly invites the targeted Guildmaster of another guild into an alliance.
func (w *World) guildAlly(z *Zone, p *Player) {
	g, me, ok := w.guildOf(p.Char.ID)
	if !ok || me.Rank != 0 {
		sys(p, "Only a Guildmaster can form an alliance.")
		return
	}
	if g.Alliance != 0 && g.Alliance != g.ID {
		sys(p, "Only the alliance's leading guild can invite others.")
		return
	}
	to := z.players[p.target]
	if to == nil || to == p {
		sys(p, "Target the Guildmaster you want to ally with.")
		return
	}
	og, om, in := w.guildOf(to.Char.ID)
	switch {
	case !in || om.Rank != 0:
		sys(p, to.Char.Name+" is not a Guildmaster.")
		return
	case og.Realm != g.Realm:
		sys(p, "You cannot ally with that guild.")
		return
	case og.Alliance != 0:
		sys(p, og.Name+" is already in an alliance.")
		return
	}
	b := w.guilds
	b.mu.Lock()
	b.invites[to.Char.ID] = &guildInvite{guild: g.ID, from: p.Char.Name, ally: true, expires: b.clock().Add(guildAnswerLapse)}
	b.mu.Unlock()
	sys(p, "You offer "+og.Name+" an alliance.")
	sys(to, p.Char.Name+" of "+g.Name+" offers your guild an alliance. Type /accept or /decline.")
}

func (w *World) joinAlliance(p *Player, inv *guildInvite) {
	og, om, in := w.guildOf(p.Char.ID)
	if !in || om.Rank != 0 || og.Alliance != 0 {
		sys(p, "You cannot join that alliance.")
		return
	}
	b := w.guilds
	b.mu.Lock()
	lead := b.byID[inv.guild]
	if lead == nil {
		b.mu.Unlock()
		sys(p, "That guild is gone.")
		return
	}
	lead.Alliance = lead.ID
	b.byID[og.ID].Alliance = lead.ID
	lc, oc := *lead, *b.byID[og.ID]
	b.mu.Unlock()
	w.Store.SaveGuild(&lc)
	w.Store.SaveGuild(&oc)
	w.sayGuild(lead.ID, oc.Name+" joins your alliance.")
	w.sayGuild(og.ID, "Your guild joins the alliance led by "+lc.Name+".")
	w.refreshGuild(lead.ID)
	w.refreshGuild(og.ID)
}

// guildAllyOff leaves an alliance, or ends it when the leading guild's
// Guildmaster does.
func (w *World) guildAllyOff(p *Player) {
	g, me, ok := w.guildOf(p.Char.ID)
	if !ok || me.Rank != 0 || g.Alliance == 0 {
		sys(p, "You are not the Guildmaster of an allied guild.")
		return
	}
	b := w.guilds
	b.mu.Lock()
	var touched []int64
	if g.Alliance == g.ID { // the leader ends it for all
		for _, o := range b.byID {
			if o.Alliance == g.ID {
				o.Alliance = 0
				cp := *o
				w.Store.SaveGuild(&cp)
				touched = append(touched, o.ID)
			}
		}
	} else {
		b.byID[g.ID].Alliance = 0
		cp := *b.byID[g.ID]
		w.Store.SaveGuild(&cp)
		touched = append(touched, g.ID)
	}
	b.mu.Unlock()
	for _, id := range touched {
		w.sayGuild(id, "The alliance has ended for your guild.")
		w.refreshGuild(id)
	}
}

// guildSay sends a line on the guild, officer or alliance channel: to the
// online members whose rank may hear it.
func (w *World) guildSay(p *Player, channel, text string) {
	g, me, ok := w.guildOf(p.Char.ID)
	if !ok {
		sys(p, "You are not in a guild.")
		return
	}
	speak, hear, ch := rules.PermGuildSpeak, rules.PermGuildHear, "guild"
	if channel == "officer" {
		speak, hear, ch = rules.PermOfficerSpeak, rules.PermOfficerHear, "officer"
	}
	if !canDo(&g, me.Rank, speak) {
		sys(p, "You may not speak on that channel.")
		return
	}
	b := w.guilds
	b.mu.Lock()
	gids := []int64{g.ID}
	if channel == "alliance" {
		ch = "alliance"
		if g.Alliance == 0 {
			b.mu.Unlock()
			sys(p, "Your guild is not in an alliance.")
			return
		}
		gids = gids[:0]
		for _, o := range b.byID {
			if o.Alliance == g.Alliance {
				gids = append(gids, o.ID)
			}
		}
	}
	var to []int64
	for _, gid := range gids {
		og := b.byID[gid]
		for _, id := range b.memberChars(gid) {
			if channel == "alliance" || canDo(og, b.of[id].Rank, hear) {
				to = append(to, id)
			}
		}
	}
	b.mu.Unlock()
	m := protocol.ChatMsg{T: protocol.TChatMsg, Ch: ch, From: p.Char.Name, Text: text}
	for _, id := range to {
		if o := w.onlineChar(id); o != nil && !o.ignoring(p) {
			o.send(m)
		}
	}
}

// adminGuild founds a guild with the administrator as Guildmaster, for testing
// and for starting one without eight players.
func (w *World) adminGuild(p *Player, name string) {
	if !rules.ValidGuildName(name) || w.guildNameTaken(name) {
		sys(p, "That is not a free guild name.")
		return
	}
	if _, _, in := w.guildOf(p.Char.ID); in {
		sys(p, "You are already in a guild.")
		return
	}
	if err := w.makeGuild(name, p.Arch.Realm, []int64{p.Char.ID}); err != nil {
		sys(p, err.Error())
		return
	}
	w.refreshGuild(func() int64 { g, _, _ := w.guildOf(p.Char.ID); return g.ID }())
	sys(p, "You found \""+name+"\".")
}

// guildLine is the guild line under a player's name for other players.
func guildLine(p *Player) string {
	s, _ := p.guildTag.Load().(string)
	return s
}
