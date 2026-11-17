package world

// Songs and chants are spells with a pulse. Casting one turns it on: it lands
// at once and again every pulse interval on the caster's group, and casting it
// again turns it off. Only one is on at a time; the effect it leaves on a
// member lasts a little longer than the interval, so it lapses on one who
// leaves range or when the song stops. See the player-spellcasting spec.

// songState is the song a player has on and when it next pulses.
type songState struct {
	spell *SpellDef
	next  float64
}

// pulsing says whether a spell is a song or chant: it has a pulse and an
// interval to pulse at.
func (sp *SpellDef) pulsing() bool { return sp.Pulse != 0 && sp.Frequency > 0 }

// pulseEvery is the seconds between a song's pulses (the data gives tenths).
func (sp *SpellDef) pulseEvery() float64 { return float64(sp.Frequency) / 10 }

// effectSeconds is how long one application of a spell lasts: its duration, or
// for a song a pulse interval plus one second, so the next pulse lands before
// the last application ends.
func (sp *SpellDef) effectSeconds() float64 {
	if sp.pulsing() {
		return sp.pulseEvery() + 1
	}
	return float64(sp.Duration)
}

// songPlaying says whether the player has this very spell on.
func (p *Player) songPlaying(sp *SpellDef) bool {
	return sp.pulsing() && p.song != nil && p.song.spell.ID == sp.ID
}

// startSong turns a song on, ending any other: only one plays at a time. Its
// first application is the cast that just landed.
func (z *Zone) startSong(p *Player, sp *SpellDef) {
	if p.song != nil && p.song.spell.ID != sp.ID {
		z.stopSong(p, "You cancel your effect.")
	}
	p.song = &songState{spell: sp, next: z.now + sp.pulseEvery()}
	z.effectsChanged(p)
}

// stopSong turns the song off, with a line to the player if there is one. The
// effect it left on the group lapses by itself.
func (z *Zone) stopSong(p *Player, text string) {
	if p.song == nil {
		return
	}
	p.song = nil
	if text != "" {
		sys(p, text)
	}
	z.effectsChanged(p)
}

// tickSong pulses the player's song when its time has come: a caster who is
// held skips the pulse and keeps the song; one who cannot pay a pulse's power
// loses it.
func (z *Zone) tickSong(p *Player) {
	s := p.song
	if s == nil || z.now+1e-9 < s.next {
		return
	}
	s.next = z.now + s.spell.pulseEvery()
	if held(p) {
		return
	}
	if cost := float64(s.spell.PulsePower); cost > 0 {
		if p.power < cost {
			z.stopSong(p, "You do not have enough power and your song stops.")
			return
		}
		p.power -= cost
	}
	targets := []combatant{p}
	if s.spell.Target == "Group" {
		targets = z.castGroupTargets(p, s.spell)
	}
	for _, t := range targets {
		z.castSpell(p, t, s.spell)
	}
}
