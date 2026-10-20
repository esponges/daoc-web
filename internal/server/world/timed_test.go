package world

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"daocweb/internal/server/protocol"
)

// atGame is a real time at which the game clock reads gs, a multiple of 24 so it
// is exact.
func atGame(gs int) time.Time { return time.Unix(int64(gs/24)+3600*480000, 0) }

const (
	t1758 = 17*3600 + 58*60 + 48 // 17:58:48
	t1801 = 18*3600 + 60 + 12    // 18:01:12, past any creature's minute
	t0601 = 6*3600 + 60 + 12
	t1200 = 12 * 3600
)

// timedRig is a zone with a night-only ghoul and a day-only bird on a clock the
// test moves.
func timedRig(t *testing.T) (*rig, *time.Time) {
	t.Helper()
	ghoul := wolf(1, 30000, 30000, 5)
	ghoul.Name, ghoul.ClassType, ghoul.AggroLevel, ghoul.RespawnS = "ghoul", "DOL.GS.NightSpawn", 0, 5
	bird := wolf(2, 31000, 30000, 5)
	bird.Name, bird.ClassType, bird.AggroLevel = "bird", "DOL.GS.DaySpawn", 0
	r := newRig(t, ghoul, bird)
	now := atGame(t1200)
	r.w.Clock = func() time.Time { return now }
	r.z.checkTimed()
	return r, &now
}

func TestNightSpawnsKeepToNight(t *testing.T) {
	r, now := timedRig(t)
	ghoul, bird := r.z.byID[1], r.z.byID[2]
	if !ghoul.hidden || bird.hidden {
		t.Fatalf("at noon the ghoul hidden=%v and the bird hidden=%v, want the ghoul gone and the bird about", ghoul.hidden, bird.hidden)
	}
	p := r.player("Alice", 5, 30500, 30000, false)
	*now = atGame(t1758)
	r.z.checkTimed()
	if !ghoul.hidden {
		t.Error("the ghoul was there before dusk")
	}
	*now = atGame(t1801)
	r.z.checkTimed()
	if ghoul.hidden || !bird.hidden {
		t.Fatalf("after dusk the ghoul hidden=%v and the bird hidden=%v, want the ghoul about and the bird gone", ghoul.hidden, bird.hidden)
	}
	r.ticks(2)
	seen := false
	for _, s := range snaps(p) {
		for _, e := range s.Ents {
			seen = seen || e.ID == 1
		}
	}
	if !seen {
		t.Error("a player nearby was not sent the ghoul at dusk")
	}
}

func TestNightSpawnsArriveOverTheFirstMinute(t *testing.T) {
	r, now := timedRig(t)
	ghoul := r.z.byID[1]
	ghoul.appear = 40
	*now = atGame(18*3600 + 24) // 24 game seconds in
	r.z.checkTimed()
	if !ghoul.hidden {
		t.Error("a creature due 40 game seconds in was there at 24")
	}
	*now = atGame(18*3600 + 48)
	r.z.checkTimed()
	if ghoul.hidden {
		t.Error("a creature due 40 game seconds in was not there at 48")
	}
}

func TestDawnMidFightLeavesAfterTheFight(t *testing.T) {
	r, now := timedRig(t)
	ghoul := r.z.byID[1]
	*now = atGame(t1801)
	r.z.checkTimed()
	p := r.player("Alice", 5, 30050, 30000, true)
	r.z.provoke(ghoul, p, 10)
	r.ticks(2)
	*now = atGame(t0601)
	r.z.checkTimed()
	if ghoul.hidden || !ghoul.leaving {
		t.Fatalf("at dawn mid-fight hidden=%v leaving=%v, want it to stay and be leaving", ghoul.hidden, ghoul.leaving)
	}
	ghoul.forget(p)
	ghoul.state = npcIdle
	r.z.checkTimed()
	if !ghoul.hidden {
		t.Error("it stayed after the fight was over")
	}
}

func TestTimedSpawnRespawnsOnlyAtNight(t *testing.T) {
	r, now := timedRig(t)
	ghoul := r.z.byID[1]
	*now = atGame(t1801)
	r.z.checkTimed()
	r.z.killNPC(ghoul)
	if !ghoul.dead() {
		t.Fatal("the ghoul did not die")
	}
	*now = atGame(t1200) // the day comes before it is due back
	ghoul.respawnAt = r.z.now + 1
	r.z.now += 10
	ghoul.update(r.z)
	if !ghoul.hidden {
		t.Fatal("it respawned by day")
	}
	*now = atGame(t1801)
	r.z.checkTimed()
	if ghoul.hidden || ghoul.dead() || ghoul.HP != ghoul.MaxHP {
		t.Errorf("at dusk it is hidden=%v dead=%v hp=%d/%d, want about and whole", ghoul.hidden, ghoul.dead(), ghoul.HP, ghoul.MaxHP)
	}
}

func TestLoadCountNamesTimedSpawns(t *testing.T) {
	r, _ := timedRig(t)
	var lines []string
	r.w.Log = func(f string, a ...any) { lines = append(lines, strings.TrimSpace(strings.ReplaceAll(f, "%d", "N"))) }
	r.z.startTimed()
	if len(lines) != 1 || !strings.Contains(lines[0], "night-only") {
		t.Errorf("log %v, want one line counting the night-only spawns", lines)
	}
}

func TestClientsAreSentTheSameClock(t *testing.T) {
	r, now := timedRig(t)
	*now = atGame(t1801)
	a := r.player("Alice", 5, 30500, 30000, false)
	b := r.player("Bobby", 5, 30600, 30000, false)
	gs := func(p *Player) []int {
		var out []int
		types, raws := sent(p)
		for i, ty := range types {
			if ty == protocol.TClock {
				var c protocol.Clock
				json.Unmarshal(raws[i], &c)
				out = append(out, c.GS)
			}
		}
		return out
	}
	ga, gb := gs(a), gs(b)
	if len(ga) != 1 || len(gb) != 1 || ga[0] != gb[0] || ga[0] != t1801 {
		t.Errorf("clocks %v and %v, want both %d on joining", ga, gb, t1801)
	}
	later := 0
	for i := 0; i < TickRate*clockEvery+1; i++ {
		r.ticks(1)
		later += len(gs(a))
		drain(b)
	}
	if later == 0 {
		t.Error("no clock sent after joining")
	}
}
