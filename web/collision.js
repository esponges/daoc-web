// What a walker cannot walk through: trees, buildings, the walls of a
// dungeon, and the edges of a ledge.
//
// The same maths as internal/server/world/collision.go -- the server decides
// where a player stands, and this lets the browser predict it. Both are
// checked against vectors the Go tests write
// (internal/server/world/testdata/collision_vectors.json).
//
// A triangle blocks when it is steep (its normal is flat) and stands between
// knee and head height above the walker's feet; a trunk blocks as a cylinder.
// A walker is a circle, pushed out of whatever it would overlap, which slides
// it along a wall.

import { instancePoint } from './props.js';

export const BODY_RADIUS = 32;
export const STEP_HEIGHT = 60;
export const BODY_HEIGHT = 170;
export const MAX_DROP = 110;
const WALL_NZ = 0.5;
const BIN = 256;

// closestOnSegment: the point on ab nearest p, and the distance.
function closestOnSegment(px, py, ax, ay, bx, by) {
  const dx = bx - ax, dy = by - ay;
  const l2 = dx * dx + dy * dy;
  let t = 0;
  if (l2 > 0) t = Math.max(0, Math.min(1, ((px - ax) * dx + (py - ay) * dy) / l2));
  const cx = ax + t * dx, cy = ay + t * dy;
  return [cx, cy, Math.hypot(px - cx, py - cy)];
}

// A Collision holds wall triangles ([ax,ay,az, bx,by,bz, cx,cy,cz] each) and
// trunk cylinders ([x, y, r, zmin, zmax]), binned over the ground plane.
export class Collision {
  constructor(tris, cyls, bounds) {
    this.tris = tris.map((t) => ({ t, zmin: Math.min(t[2], t[5], t[8]), zmax: Math.max(t[2], t[5], t[8]) }));
    this.cyls = cyls;
    this.x0 = bounds[0]; this.y0 = bounds[1];
    this.w = Math.ceil((bounds[2] - bounds[0]) / BIN) + 1;
    this.h = Math.ceil((bounds[3] - bounds[1]) / BIN) + 1;
    this.bins = Array.from({ length: this.w * this.h }, () => []);
    const span = (x0, y0, x1, y1, fn) => {
      const i0 = Math.floor((x0 - this.x0) / BIN), i1 = Math.floor((x1 - this.x0) / BIN);
      const j0 = Math.floor((y0 - this.y0) / BIN), j1 = Math.floor((y1 - this.y0) / BIN);
      for (let j = Math.max(j0, 0); j <= Math.min(j1, this.h - 1); j++) {
        for (let i = Math.max(i0, 0); i <= Math.min(i1, this.w - 1); i++) fn(j * this.w + i);
      }
    };
    this.tris.forEach(({ t }, k) => {
      span(Math.min(t[0], t[3], t[6]) - BODY_RADIUS, Math.min(t[1], t[4], t[7]) - BODY_RADIUS,
        Math.max(t[0], t[3], t[6]) + BODY_RADIUS, Math.max(t[1], t[4], t[7]) + BODY_RADIUS, (b) => this.bins[b].push(k));
    });
    this.cyls.forEach((c, k) => {
      span(c[0] - c[2] - BODY_RADIUS, c[1] - c[2] - BODY_RADIUS, c[0] + c[2] + BODY_RADIUS, c[1] + c[2] + BODY_RADIUS,
        (b) => this.bins[b].push(this.tris.length + k));
    });
  }

  // slide returns where a walker moving from (x0,y0) to (x1,y1), feet at z,
  // ends up: the destination pushed out of what it overlaps.
  slide(x0, y0, x1, y1, z) {
    let px = x1, py = y1;
    const zlo = z + STEP_HEIGHT, zhi = z + BODY_HEIGHT;
    for (let iter = 0; iter < 3; iter++) {
      let moved = false;
      const i = Math.floor((px - this.x0) / BIN), j = Math.floor((py - this.y0) / BIN);
      if (i < 0 || j < 0 || i >= this.w || j >= this.h) break;
      for (const idx of this.bins[j * this.w + i]) {
        if (idx >= this.tris.length) {
          const k = this.cyls[idx - this.tris.length];
          if (k[4] <= zlo || k[3] >= zhi) continue;
          let dx = px - k[0], dy = py - k[1];
          let d = Math.hypot(dx, dy);
          const lim = k[2] + BODY_RADIUS;
          if (d < lim) {
            if (d < 1e-6) {
              dx = x0 - k[0]; dy = y0 - k[1]; d = Math.hypot(dx, dy);
              if (d < 1e-6) continue;
            }
            px = k[0] + dx / d * lim; py = k[1] + dy / d * lim;
            moved = true;
          }
          continue;
        }
        const { t, zmin, zmax } = this.tris[idx];
        if (zmax <= zlo || zmin >= zhi) continue;
        for (const [ax, ay, bx, by] of [[t[0], t[1], t[3], t[4]], [t[3], t[4], t[6], t[7]], [t[6], t[7], t[0], t[1]]]) {
          const [cx, cy, d] = closestOnSegment(px, py, ax, ay, bx, by);
          if (d < BODY_RADIUS) {
            let nx, ny;
            if (d < 1e-6) {
              nx = x0 - px; ny = y0 - py;
              const l = Math.hypot(nx, ny);
              if (l > 1e-9) { nx /= l; ny /= l; } else continue;
            } else { nx = (px - cx) / d; ny = (py - cy) / d; }
            px = cx + nx * BODY_RADIUS; py = cy + ny * BODY_RADIUS;
            moved = true;
          }
        }
      }
      if (!moved) break;
    }
    return [px, py];
  }
}

// wallTriangle: true when a triangle is steep enough to block.
export function isWall(t) {
  const ux = t[3] - t[0], uy = t[4] - t[1], uz = t[5] - t[2];
  const wx = t[6] - t[0], wy = t[7] - t[1], wz = t[8] - t[2];
  const nx = uy * wz - uz * wy, ny = uz * wx - ux * wz, nz = ux * wy - uy * wx;
  const n = Math.hypot(nx, ny, nz);
  return n > 0 && Math.abs(nz) / n < WALL_NZ;
}

// loadOutdoorCollision reads a zone's collision.json/bin and places each hull
// with props.json's instances; a model with no hull but a trunk radius blocks
// as a cylinder. It resolves to null for a zone with no collision files.
export async function loadOutdoorCollision(zoneDir, { fetchJSON, fetchBuffer }, extent = 65536) {
  let cj, cb, pj;
  try {
    [cj, cb, pj] = await Promise.all([
      fetchJSON(zoneDir + '/props/collision.json'),
      fetchBuffer(zoneDir + '/props/collision.bin'),
      fetchJSON(zoneDir + '/props/props.json'),
    ]);
  } catch { return null; }
  if (cj.models.length !== pj.models.length) return null;
  const verts = new Float32Array(cb, 0, cj.vertexBytes / 4);
  const idx = new Uint32Array(cb, cj.vertexBytes);
  const tris = [], cyls = [];
  for (const inst of pj.instances) {
    const m = cj.models[inst.m];
    if (!m) continue;
    if (m.ni > 0) {
      const v = (i) => [verts[(m.v + i) * 3], verts[(m.v + i) * 3 + 1], verts[(m.v + i) * 3 + 2]];
      for (let t = 0; t + 2 < m.ni; t += 3) {
        const tri = [];
        for (let k = 0; k < 3; k++) {
          // instancePoint answers (x, height, y); the server's wants (x, y, z).
          const p = instancePoint(inst, v(idx[m.i + t + k]));
          tri.push(p[0], p[2], p[1]);
        }
        if (isWall(tri)) tris.push(tri);
      }
    } else if (m.r > 0) {
      const sc = inst.s || 1;
      cyls.push([inst.p[0], inst.p[1], m.r * sc, inst.p[2] - 50, inst.p[2] + 400]);
    }
  }
  return new Collision(tris, cyls, [0, 0, extent, extent]);
}

// interiorCollision gathers the walls of a city or dungeon from its visible
// geometry, and wraps a slide that also refuses ledges and voids, as the
// server's Interior does. floor is interior.js's floor (below(x, y, fromZ)).
export function buildInteriorWalls(manifest, raw, bounds) {
  const stride = manifest.stride / 4;
  const vf = new Float32Array(raw, 0, manifest.vertexBytes / 4);
  const ix = new Uint32Array(raw, manifest.vertexBytes, manifest.indexBytes / 4);
  const perModel = manifest.models.map(() => []);
  for (const inst of manifest.instances) perModel[inst.m]?.push(inst);
  const tris = [];
  const v = (i) => [vf[i * stride], vf[i * stride + 1], vf[i * stride + 2]];
  manifest.models.forEach((model, mi) => {
    if (!perModel[mi].length) return;
    for (const g of model.groups) {
      if (g.billboard || g.additive) continue;
      for (let t = g.first; t + 2 < g.first + g.count; t += 3) {
        const a = v(ix[t]), b = v(ix[t + 1]), c = v(ix[t + 2]);
        for (const inst of perModel[mi]) {
          const tri = [];
          for (const q of [a, b, c]) { const p = instancePoint(inst, q); tri.push(p[0], p[2], p[1]); }
          if (isWall(tri)) tris.push(tri);
        }
      }
    }
  });
  return new Collision(tris, [], [bounds.min[0] - BIN, bounds.min[1] - BIN, bounds.max[0] + BIN, bounds.max[1] + BIN]);
}

// ledgeMover turns walls plus a floor into the mover a walker uses indoors.
export function ledgeMover(walls, floor) {
  const walkable = (x, y, z) => { const f = floor.below(x, y, z); return f !== null && z - f <= MAX_DROP; };
  return {
    slide(x0, y0, x1, y1, z) {
      const [nx, ny] = walls.slide(x0, y0, x1, y1, z);
      if (!walkable(x0, y0, z)) return [nx, ny];
      if (walkable(nx, ny, z)) return [nx, ny];
      if (walkable(nx, y0, z)) return [nx, y0];
      if (walkable(x0, ny, z)) return [x0, ny];
      return [x0, y0];
    },
  };
}
