// The ground target: a point on the terrain that abilities such as Distraction
// are aimed at. The server owns it; the client only turns a click into a point
// and draws the marker the server confirms.
//
// Render space is x east, y up, z north: the world's (x, y, height) is
// (x, height, y) here, as the NPC picking in npc.js has it.

// pickGround walks a ray from origin along dir until it passes below the
// terrain, and returns the world point {x, y, z} there, or null when it never
// does within maxDist. groundAt(wx, wy) is the terrain height.
export function pickGround(origin, dir, groundAt, maxDist = 6000, step = 16) {
  const [ox, oy, oz] = origin;
  const len = Math.hypot(dir[0], dir[1], dir[2]) || 1;
  const d = dir.map((v) => v / len);
  let prev = 0;
  for (let t = step; t <= maxDist; t += step) {
    const x = ox + d[0] * t, y = oy + d[1] * t, z = oz + d[2] * t;
    if (y <= groundAt(x, z)) {
      // Narrow the last step by bisection so the point sits on the surface.
      let lo = prev, hi = t;
      for (let i = 0; i < 12; i++) {
        const mid = (lo + hi) / 2;
        if (oy + d[1] * mid <= groundAt(ox + d[0] * mid, oz + d[2] * mid)) hi = mid; else lo = mid;
      }
      const wx = ox + d[0] * hi, wy = oz + d[2] * hi;
      return { x: wx, y: wy, z: groundAt(wx, wy) };
    }
    prev = t;
  }
  return null;
}

// groundTargetOf is the point the server says is your ground target, or null.
export function groundTargetOf(you) {
  const g = you && you.gt;
  return Array.isArray(g) && g.length === 3 ? { x: g[0], y: g[1], z: g[2] } : null;
}
