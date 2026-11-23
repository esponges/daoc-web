package world

import (
	"fmt"
	"math"
	"strings"

	"daocweb/internal/server/protocol"
)

// Emotes: the classic set of the 1.65 client, each a slash command that plays
// an animation on everyone who can see the player and writes a line on the
// Emote channel. See the emotes capability. The lines are this project's.

// emote is one command. Clip is the name the client plays: the emote's column
// in anims.csv, converted by animconv as em_<clip>.
type emote struct {
	Cmd  string
	Clip string
	// You and They are the verb phrase in the second and third person
	// ("wave", "waves"); Prep goes between it and a target ("at", "to", or
	// nothing).
	You, They, Prep string
}

// emotes is the classic set, in the order of the client's own list.
var emotes = []emote{
	{"beckon", "beckon", "beckon", "beckons", "to"},
	{"blush", "blush", "blush", "blushes", "at"},
	{"bow", "bow", "bow", "bows", "to"},
	{"cheer", "cheer", "cheer", "cheers", "for"},
	{"clap", "clap", "clap", "claps", "for"},
	{"cry", "cry", "cry", "cries", "to"},
	{"curtsey", "curtsey", "curtsey", "curtseys", "to"},
	{"dance", "dance", "dance", "dances", "with"},
	{"flex", "flex", "flex your muscles", "flexes their muscles", "at"},
	{"kiss", "kiss", "blow a kiss", "blows a kiss", "to"},
	{"laugh", "laugh", "laugh", "laughs", "at"},
	{"point", "point", "point", "points", "at"},
	{"salute", "salute", "salute", "salutes", ""},
	{"bang", "bang", "bang on your shield", "bangs on a shield", "at"},
	{"victory", "victory", "celebrate victory", "celebrates victory", "with"},
	{"wave", "wave", "wave", "waves", "at"},
}

func emoteFor(cmd string) *emote {
	for i := range emotes {
		if emotes[i].Cmd == cmd {
			return &emotes[i]
		}
	}
	return nil
}

// lines are what the emoter, an onlooker and the target read. target is "" for
// an emote at no one.
func (e *emote) lines(who, target string) (you, others, toTarget string) {
	with := func(verb string) string {
		if e.Prep == "" {
			return verb
		}
		return verb + " " + e.Prep
	}
	if target == "" {
		return fmt.Sprintf("You %s.", e.You), fmt.Sprintf("%s %s.", who, e.They), ""
	}
	return fmt.Sprintf("You %s %s.", with(e.You), target),
		fmt.Sprintf("%s %s %s.", who, with(e.They), target),
		fmt.Sprintf("%s %s you.", who, with(e.They))
}

// TryEmote runs /<cmd> if it is an emote, and says whether it was one.
func (z *Zone) TryEmote(p *Player, cmd string) bool {
	e := emoteFor(strings.ToLower(cmd))
	if e == nil {
		return false
	}
	switch {
	case p.dead:
		p.errMsg("emote", "You cannot do that while dead.")
		return true
	case p.ride != nil:
		p.errMsg("emote", "You cannot do that while riding.")
		return true
	case z.now < p.combatUntil:
		p.errMsg("emote", "You cannot do that in combat.")
		return true
	case p.Body.Air:
		p.errMsg("emote", "You cannot do that in the air.")
		return true
	case swimming(z.G, p.Body.X, p.Body.Y):
		p.errMsg("emote", "You cannot do that while swimming.")
		return true
	}
	z.standUp(p)
	// Who it is at: the held target, if it is a player or a creature.
	var tname string
	var tplayer *Player
	if o := z.players[p.target]; o != nil && o != p {
		tname, tplayer = o.Char.Name, o
	} else if n := z.byID[p.target]; n != nil && !n.gone {
		tname = n.title()
	}
	you, others, toTarget := e.lines(p.Char.Name, tname)
	z.event(protocol.Event{K: "emote", From: p.ID, To: p.target, Text: e.Clip}, p.Body.X, p.Body.Y)
	send := func(to *Player, text string) {
		to.send(protocol.ChatMsg{T: protocol.TChatMsg, Ch: "emote", Text: text})
	}
	send(p, you)
	for _, o := range z.players {
		if o != p && o.ignoring(p) {
			continue
		}
		switch {
		case o == p:
		case o == tplayer:
			send(o, toTarget)
		case math.Hypot(o.Body.X-p.Body.X, o.Body.Y-p.Body.Y) <= SayRange:
			send(o, others)
		}
	}
	return true
}
