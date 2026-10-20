package world

import (
	"database/sql"
	"math"
	"strings"
	"sync"

	"daocweb/internal/server/store"
)

// KindStable is what a stable master does when interacted with: sells tickets
// like a merchant, and takes a ticket for one of its routes from the pack.
const KindStable = "stable"

const (
	// HorseSpeed is the speed of a ride, in units a second, where a route point
	// gives none or a faster one. Provisional.
	HorseSpeed = 1000.0
	// stableReach is how near a route's first point must be to the stable master
	// for the route to start there.
	stableReach = 800.0
)

// RoutePoint is one point of a horse route, in its zone's coordinates.
type RoutePoint struct {
	Zone    int
	X, Y, Z float64
	Speed   int // 0 for the default
	Wait    int
}

// Route is an imported horse route: the path a ticket of the same id names.
type Route struct {
	Path   string
	Start  int // the zone of the stable that sells it
	Points []RoutePoint
}

// RouteFor finds the route a ticket item id names, nil if it is not a ticket.
func (c *Catalog) RouteFor(item string) *Route { return c.Routes[strings.ToLower(item)] }

// loadRoutes reads the horse routes. A store imported before they were kept
// has none, and stable masters then only sell.
func (c *Catalog) loadRoutes(db *sql.DB, q func(string, func(*sql.Rows) error) error) error {
	c.Routes = map[string]*Route{}
	var has int
	db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name IN ('route','route_point')`).Scan(&has)
	if has != 2 {
		return nil
	}
	if err := q(`SELECT path,start_zone FROM route`, func(r *sql.Rows) error {
		rt := &Route{}
		if err := r.Scan(&rt.Path, &rt.Start); err != nil {
			return err
		}
		c.Routes[strings.ToLower(rt.Path)] = rt
		return nil
	}); err != nil {
		return err
	}
	if err := q(`SELECT path,zone,x,y,z,speed,wait FROM route_point ORDER BY path,step`, func(r *sql.Rows) error {
		var path string
		var p RoutePoint
		if err := r.Scan(&path, &p.Zone, &p.X, &p.Y, &p.Z, &p.Speed, &p.Wait); err != nil {
			return err
		}
		if rt := c.Routes[strings.ToLower(path)]; rt != nil {
			rt.Points = append(rt.Points, p)
		}
		return nil
	}); err != nil {
		return err
	}
	for k, rt := range c.Routes {
		if len(rt.Points) < 2 {
			delete(c.Routes, k)
		}
	}
	return nil
}

// ride is a rider's progress along a route: the next point to reach. It lives
// on the World, by character, so that it follows a rider into the next zone,
// where the player is made again.
type ride struct {
	route *Route
	next  int
}

type rides struct {
	mu sync.Mutex
	by map[int64]*ride
}

func (r *rides) set(char int64, rd *ride) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.by == nil {
		r.by = map[int64]*ride{}
	}
	r.by[char] = rd
}

func (r *rides) take(char int64) *ride {
	r.mu.Lock()
	defer r.mu.Unlock()
	rd := r.by[char]
	delete(r.by, char)
	return rd
}

// Riding says whether a player is on a horse.
func (p *Player) Riding() bool { return p.ride != nil }

// openStable answers a stable master's interaction: a ticket in the pack for a
// route that starts here is handed in and rides; any other route ticket is
// refused and kept; otherwise (or after a refusal) the master's tickets are
// for sale.
func (z *Zone) openStable(p *Player, n *NPC) {
	cat := z.W.Cat
	wrong := false
	for _, slot := range sortedSlots(p.inv) {
		it := p.inv[slot]
		rt := cat.RouteFor(it.ID)
		if rt == nil {
			continue
		}
		if rt.Start != z.ID || math.Hypot(rt.Points[0].X-n.Body.X, rt.Points[0].Y-n.Body.Y) > stableReach {
			wrong = true
			continue
		}
		if err := z.W.Store.Tx(func(tx *store.Tx) error { return tx.RemoveItem(p.Char.ID, slot) }); err != nil {
			p.errMsg("ride", "Could not take the ticket; try again.")
			return
		}
		delete(p.inv, slot)
		p.sendInv(cat, p.Char.Money)
		z.startRide(p, rt)
		return
	}
	if wrong {
		sys(p, "That ticket is not for this stable.")
	}
	z.openShop(p, n)
}

// startRide puts a player on a horse at the first point of a route.
func (z *Zone) startRide(p *Player, rt *Route) {
	z.closeShop(p)
	p.attacking, p.pending, p.target = false, -1, 0
	p.queue, p.queuedSteps = nil, 0
	p.ride = &ride{route: rt, next: 1}
	first := rt.Points[0]
	p.Body.X, p.Body.Y = first.X, first.Y
	p.Body.Z = z.G.GroundAt(first.X, first.Y, first.Z)
	p.correct = true
	sys(p, "You mount the horse and ride toward "+rideName(rt)+".")
}

// rideName is the end of a route as a player would say it: "Ticket to ..." is
// not at hand here, so the path's own words are used.
func rideName(rt *Route) string {
	parts := strings.Split(rt.Path, "_")
	return strings.Join(parts[1:], " ")
}

// rideStep moves a rider along their route for one tick, and ends the ride at
// the last point. Inputs do nothing meanwhile.
func (z *Zone) rideStep(p *Player) {
	rd := p.ride
	p.queue, p.queuedSteps, p.budget = nil, 0, 0
	for rd.next < len(rd.route.Points) {
		t := rd.route.Points[rd.next]
		if t.Zone != z.ID {
			// The next point is in another zone: go there and carry on from the one after.
			rd.next++
			z.W.rides.set(p.Char.ID, rd)
			z.W.moveToZone(z, p, t.Zone, t.X, t.Y, t.Z)
			p.ride = nil
			return
		}
		speed := HorseSpeed
		if t.Speed > 0 {
			speed = math.Min(float64(t.Speed), HorseSpeed)
		}
		dx, dy := t.X-p.Body.X, t.Y-p.Body.Y
		d := math.Hypot(dx, dy)
		step := speed * TickDT
		if d > step {
			p.Body.Yaw = angleTo(p.Body.X, p.Body.Y, t.X, t.Y)
			p.Body.X += dx / d * step
			p.Body.Y += dy / d * step
			p.Body.Z = z.G.GroundAt(p.Body.X, p.Body.Y, t.Z)
			p.motion = Motion{Speed: speed}
			p.correct = true
			return
		}
		p.Body.X, p.Body.Y, p.Body.Z = t.X, t.Y, z.G.GroundAt(t.X, t.Y, t.Z)
		rd.next++
	}
	p.ride = nil
	p.motion = Motion{}
	p.correct = true
	sys(p, "You dismount. You have arrived.")
	if err := z.W.Store.Save(p.Char); err != nil {
		z.W.Log("saving %s: %v", p.Char.Name, err)
	}
}

// resumeRide gives a player arriving in a new zone the ride they were on.
func (z *Zone) resumeRide(p *Player) {
	if rd := z.W.rides.take(p.Char.ID); rd != nil {
		p.ride = rd
	}
}
