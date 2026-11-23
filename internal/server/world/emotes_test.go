package world

import (
	"encoding/json"
	"strings"
	"testing"

	"daocweb/internal/server/protocol"
)

// emoteLines are the lines of the emote channel a player has been sent.
func emoteLines(p *Player) (out []string) {
	for {
		select {
		case b := <-p.out:
			var m protocol.ChatMsg
			if json.Unmarshal(b, &m) == nil && m.T == protocol.TChatMsg && m.Ch == "emote" {
				out = append(out, m.Text)
			}
			continue
		default:
		}
		return out
	}
}

func emoteRig(t *testing.T) (*rig, *Player, *Player) {
	r := newRig(t)
	thora := r.player("Thora", 20, 30000, 30000, false)
	brann := r.player("Brann", 20, 30200, 30000, false)
	emoteLines(thora)
	emoteLines(brann)
	return r, thora, brann
}

func TestAKnownEmoteIsAccepted(t *testing.T) {
	r, thora, _ := emoteRig(t)
	r.w.command(r.z, thora, "wave")
	if got := emoteLines(thora); len(got) != 1 || got[0] != "You wave." {
		t.Errorf("Thora read %v, want \"You wave.\"", got)
	}
	for _, e := range emotes {
		if emoteFor(e.Cmd) == nil {
			t.Errorf("emote %s is not found by its name", e.Cmd)
		}
	}
	if len(emotes) != 16 {
		t.Errorf("%d emotes, want the classic 16", len(emotes))
	}
}

func TestALaterEmoteIsUnknown(t *testing.T) {
	r, thora, _ := emoteRig(t)
	for _, cmd := range []string{"hug", "rofl", "yawn", "shrug"} {
		r.w.command(r.z, thora, cmd)
		if got := lastErr(t, thora); !strings.Contains(got, "Unknown command") {
			t.Errorf("/%s was answered %q, want Unknown command", cmd, got)
		}
	}
}

func TestEveryoneNearbySeesAnEmote(t *testing.T) {
	r, thora, brann := emoteRig(t)
	far := r.player("Far", 20, 50000, 50000, false)
	emoteLines(far)
	thora.target = brann.ID
	r.w.command(r.z, thora, "bow")
	r.ticks(1)
	var event *protocol.Event
	for _, s := range snaps(brann) {
		for i := range s.Ev {
			if s.Ev[i].K == "emote" {
				event = &s.Ev[i]
			}
		}
	}
	if event == nil || event.From != thora.ID || event.To != brann.ID || event.Text != "bow" {
		t.Errorf("Brann saw the emote event %+v, want Thora's bow at Brann", event)
	}
	if got := emoteLines(far); len(got) != 0 {
		t.Errorf("someone far away read %v", got)
	}
}

func TestEmoteLinesByAudience(t *testing.T) {
	r, thora, brann := emoteRig(t)
	onlooker := r.player("Sigrid", 20, 30300, 30000, false)
	emoteLines(onlooker)
	thora.target = brann.ID
	r.w.command(r.z, thora, "wave")
	for who, want := range map[*Player]string{
		thora:    "You wave at Brann.",
		brann:    "Thora waves at you.",
		onlooker: "Thora waves at Brann.",
	} {
		if got := emoteLines(who); len(got) != 1 || got[0] != want {
			t.Errorf("%s read %v, want %q", who.Char.Name, got, want)
		}
	}
}

func TestEmoteAtNoOneAndAtACreature(t *testing.T) {
	r := newRig(t, wolf(1, 30100, 30000, 3))
	thora := r.player("Thora", 20, 30000, 30000, false)
	brann := r.player("Brann", 20, 30200, 30000, false)
	emoteLines(thora)
	emoteLines(brann)
	r.w.command(r.z, thora, "wave")
	if got := emoteLines(brann); len(got) != 1 || got[0] != "Thora waves." {
		t.Errorf("Brann read %v, want \"Thora waves.\"", got)
	}
	emoteLines(thora)
	thora.target = r.z.npcs[0].ID
	r.w.command(r.z, thora, "dance")
	if got := emoteLines(thora); len(got) != 1 || !strings.HasPrefix(got[0], "You dance with ") {
		t.Errorf("at a creature Thora read %v", got)
	}
}

func TestEmotesAreRefused(t *testing.T) {
	cases := map[string]func(r *rig, p *Player){
		"in combat":      func(r *rig, p *Player) { p.combatUntil = r.z.now + 10 },
		"dead":           func(r *rig, p *Player) { p.dead = true },
		"in the air":     func(r *rig, p *Player) { p.Body.Air = true },
		"while swimming": func(r *rig, p *Player) { r.z.G = pond{} },
	}
	for name, set := range cases {
		t.Run(name, func(t *testing.T) {
			r, thora, brann := emoteRig(t)
			set(r, thora)
			r.w.command(r.z, thora, "dance")
			if got := emoteLines(brann); len(got) != 0 {
				t.Errorf("Brann read %v of a refused emote", got)
			}
			if lastErr(t, thora) == "" {
				t.Error("the refusal gave no reason")
			}
			r.ticks(1)
			for _, s := range snaps(brann) {
				for _, e := range s.Ev {
					if e.K == "emote" {
						t.Error("a refused emote was played")
					}
				}
			}
		})
	}
}

func TestEmotingStandsASeatedPlayerUp(t *testing.T) {
	r, thora, _ := emoteRig(t)
	r.z.Sit(thora)
	r.w.command(r.z, thora, "cheer")
	if thora.seated {
		t.Error("cheering left them seated")
	}
	if got := emoteLines(thora); len(got) != 1 {
		t.Errorf("Thora read %v, want the cheer line", got)
	}
}
