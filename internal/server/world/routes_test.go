package world

import (
	"math"
	"testing"
)

const stableID = 950

// stableRig is a zone with a stable master at (30000, 30000) who sells the
// ticket to B, and one to C that starts in another stable; the route to B runs
// east in three points, and the one to C starts far away.
func stableRig(t *testing.T, more ...Spawn) *rig {
	t.Helper()
	stable := Spawn{ID: stableID, Name: "Gularg", Guild: "Stable Master", Model: 56, X: 30000, Y: 30000, Z: 5000, Level: 10, Size: 50,
		Speed: 191, Realm: 2, ClassType: "DOL.GS.GameStableMaster", ItemsList: "ST"}
	r := newRig(t, append([]Spawn{stable}, more...)...)
	r.w.Cat.Stock = map[string][]string{"ST": {"HS_A_B", "HS_C_D"}}
	r.w.Cat.Items["HS_A_B"] = &Item{ID: "HS_A_B", Name: "Ticket to B", Price: 5000, Droppable: true}
	r.w.Cat.Items["HS_C_D"] = &Item{ID: "HS_C_D", Name: "Ticket to D", Price: 5000, Droppable: true}
	r.w.Cat.Routes = map[string]*Route{
		"hs_a_b": {Path: "HS_A_B", Start: 100, Points: []RoutePoint{{Zone: 100, X: 30100, Y: 30000}, {Zone: 100, X: 30600, Y: 30000}, {Zone: 100, X: 31200, Y: 30000, Speed: 350}}},
		"hs_c_d": {Path: "HS_C_D", Start: 100, Points: []RoutePoint{{Zone: 100, X: 50000, Y: 50000}, {Zone: 100, X: 50500, Y: 50000}}},
	}
	return r
}

func TestStableMasterSellsTickets(t *testing.T) {
	r := stableRig(t)
	p := r.player("Rider", 5, 30050, 30000, true)
	p.Char.Money = 100000
	r.z.Interact(p, stableID)
	shop, ok := lastShop(p)
	if !ok || len(shop.Items) != 2 || shop.Items[0].ID != "HS_A_B" || shop.Items[0].Price != 5000 {
		t.Fatalf("shop %+v (%v), want the two tickets for sale", shop, ok)
	}
	before := p.Char.Money
	r.z.Buy(p, "HS_A_B")
	got := false
	for _, it := range p.inv {
		got = got || it.ID == "HS_A_B"
	}
	if !got || p.Char.Money != before-5000 {
		t.Errorf("ticket in pack %v, money %d -> %d, want the ticket and 5000 taken", got, before, p.Char.Money)
	}
}

func TestTicketStartsRide(t *testing.T) {
	r := stableRig(t)
	p := r.player("Rider", 5, 30050, 30000, true)
	if _, err := r.w.addItem(p, "HS_A_B", 1); err != nil {
		t.Fatal(err)
	}
	r.z.Interact(p, stableID)
	if !p.Riding() {
		t.Fatal("the player is not riding after handing in the ticket")
	}
	for _, it := range p.inv {
		if it.ID == "HS_A_B" {
			t.Error("the ticket is still in the pack")
		}
	}
	if p.Body.X != 30100 || p.Body.Y != 30000 {
		t.Errorf("the ride starts at %v,%v, want the route's first point", p.Body.X, p.Body.Y)
	}
}

func TestWrongStableRefusesAndKeepsTheTicket(t *testing.T) {
	r := stableRig(t)
	p := r.player("Rider", 5, 30050, 30000, true)
	if _, err := r.w.addItem(p, "HS_C_D", 1); err != nil {
		t.Fatal(err)
	}
	r.z.Interact(p, stableID)
	if p.Riding() {
		t.Error("a ticket for another stable started a ride")
	}
	kept := false
	for _, it := range p.inv {
		kept = kept || it.ID == "HS_C_D"
	}
	if !kept {
		t.Error("the refused ticket was taken")
	}
}

func TestRideFollowsThePathAndDismounts(t *testing.T) {
	r := stableRig(t)
	p := r.player("Rider", 5, 30050, 30000, true)
	r.w.addItem(p, "HS_A_B", 1)
	r.z.Interact(p, stableID)
	r.ticks(int(0.6 / TickDT)) // past 500 units at the default speed, into the second leg
	if !p.Riding() || p.Body.X < 30400 || p.Body.X > 30700 || p.Body.Yaw < math.Pi/2-0.01 || p.Body.Yaw > math.Pi/2+0.01 {
		t.Fatalf("mid ride at x=%v yaw=%v riding=%v, want about 30500 heading east", p.Body.X, p.Body.Yaw, p.Riding())
	}
	r.ticks(int(4 / TickDT))
	if p.Riding() || p.Body.X != 31200 {
		t.Errorf("riding=%v x=%v after the route, want dismounted at the last point 31200", p.Riding(), p.Body.X)
	}
	// Free again: a step of input moves them.
	before := p.Body.X
	p.queue, p.queuedSteps = []stepRun{{n: 10, k: Keys{Fwd: 1, Run: true}}}, 10
	r.ticks(2)
	if p.Body.X == before && p.Body.Y == 30000 {
		t.Error("the dismounted player cannot move")
	}
}

func TestRiderIsNotAttackedAndCannotFight(t *testing.T) {
	r := stableRig(t, Spawn{ID: 951, Name: "wolf", Model: 56, X: 30300, Y: 30100, Z: 5000, Level: 5, Size: 50, Speed: 191, AggroLevel: 1, AggroRange: 700})
	n := r.z.byID[951]
	p := r.player("Rider", 5, 30050, 30000, true)
	r.w.addItem(p, "HS_A_B", 1)
	r.z.Interact(p, stableID)
	hp := p.hp
	r.ticks(int(1.5 / TickDT))
	if n.target != nil || p.hp != hp {
		t.Errorf("the wolf fought the rider: target %v, hp %v -> %v", n.target, hp, p.hp)
	}
	r.z.SetTarget(p, 951)
	r.z.SetAttack(p, true)
	if p.attacking {
		t.Error("a rider started to attack")
	}
	p.queue, p.queuedSteps = []stepRun{{n: 30, k: Keys{Fwd: -1, Run: true}}}, 30
	x := p.Body.X
	r.ticks(1)
	if p.Body.X < x {
		t.Error("movement keys moved a rider")
	}
}

func TestRideCrossesAZoneBorder(t *testing.T) {
	r := stableRig(t)
	other := NewZone(r.w, 101, "Next", flat{}, nil, nil)
	r.w.zones[101] = other
	r.w.Cat.Routes["hs_a_b"].Points = []RoutePoint{{Zone: 100, X: 30100, Y: 30000}, {Zone: 100, X: 30500, Y: 30000}, {Zone: 101, X: 400, Y: 30000}, {Zone: 101, X: 20000, Y: 30000}}
	p := r.player("Rider", 5, 30050, 30000, true)
	r.w.addItem(p, "HS_A_B", 1)
	r.z.Interact(p, stableID)
	r.ticks(int(2 / TickDT))
	if !p.zoning.Load() || !p.closed.Load() {
		t.Fatalf("the rider did not leave for the next zone (zoning %v closed %v)", p.zoning.Load(), p.closed.Load())
	}
	// The client reconnects to zone 101; the ride is still theirs.
	var q *Player
	var err error
	done := make(chan struct{})
	go func() { q, err = r.w.Join(p.Acct, p.Char.ID); close(done) }()
	for joined := false; !joined; {
		other.Tick()
		select {
		case <-done:
			joined = true
		default:
		}
	}
	if err != nil {
		t.Fatal(err)
	}
	if q.Zone != other || !q.Riding() || q.Body.X < 400 {
		t.Fatalf("after joining: zone %v riding %v x=%v, want zone 101, still riding, at the border point", q.Zone != nil && q.Zone.ID == 101, q.Riding(), q.Body.X)
	}
	for i := 0; i < int(25/TickDT); i++ {
		other.Tick()
	}
	if q.Riding() || q.Body.X != 20000 {
		t.Errorf("riding=%v x=%v, want dismounted at the last point 20000", q.Riding(), q.Body.X)
	}
}
