// The interior world: a city or dungeon, which has no heightmap. It is all
// scenery -- interiorconv writes its pieces as props -- so this supplies
// what the heightmap supplies outdoors: where the ground is, where to start,
// and the look of a place with no sky.
//
// The ground is found by casting down onto the interior's own floors. The
// triangles that face up are gathered once, placed where their instances
// put them, and binned on a grid over the ground plane, so a lookup tests
// only the few under the character. From a given height it takes the
// highest floor no more than a step above it: the upper level of a
// two-storey hall is never taken for the floor of the lower, nor a ceiling
// for a floor.

import { instancePoint } from './props.js';
import { buildInteriorWalls, ledgeMover } from './collision.js';

// A light from lights.csv: a warm torch reaching TORCH_RADIUS units.
export const TORCH = [1.15, 0.78, 0.45];
export const TORCH_RADIUS = 700;

// How far above the feet a floor may be and still be stepped up onto.
export const STEP = 60;
const BIN = 256;
// A triangle counts as floor when its normal is within about 70° of up.
const FLOOR_Z = 0.34;

// buildFloor gathers the upward-facing triangles of a converted interior.
// manifest is props.json and raw props.bin; positions come back in zone
// coordinates (x, y on the ground, z up).
export function buildFloor(manifest, raw) {
  const stride = manifest.stride / 4;
  const vf = new Float32Array(raw, 0, manifest.vertexBytes / 4);
  const ix = new Uint32Array(raw, manifest.vertexBytes, manifest.indexBytes / 4);
  const perModel = manifest.models.map(() => []);
  for (const inst of manifest.instances) perModel[inst.m]?.push(inst);

  const tris = []; // x0 y0 z0 x1 y1 z1 x2 y2 z2, zone coordinates
  const v = (i) => [vf[i * stride], vf[i * stride + 1], vf[i * stride + 2]];
  manifest.models.forEach((model, mi) => {
    const insts = perModel[mi];
    if (!insts.length) return;
    for (const g of model.groups) {
      if (g.billboard || g.additive) continue;
      for (let t = g.first; t + 2 < g.first + g.count; t += 3) {
        const a = v(ix[t]), b = v(ix[t + 1]), c = v(ix[t + 2]);
        for (const inst of insts) {
          // instancePoint answers in world space, height second.
          const pa = instancePoint(inst, a), pb = instancePoint(inst, b), pc = instancePoint(inst, c);
          const A = [pa[0], pa[2], pa[1]], B = [pb[0], pb[2], pb[1]], C = [pc[0], pc[2], pc[1]];
          const ux = B[0] - A[0], uy = B[1] - A[1], uz = B[2] - A[2];
          const wx = C[0] - A[0], wy = C[1] - A[1], wz = C[2] - A[2];
          const nx = uy * wz - uz * wy, ny = uz * wx - ux * wz, nz = ux * wy - uy * wx;
          const n = Math.hypot(nx, ny, nz);
          if (n === 0) continue;
          // Facing either way: some floors are wound for the far side, and
          // a ceiling is ruled out by height, not by winding.
          if (Math.abs(nz) / n < FLOOR_Z) continue;
          tris.push(A[0], A[1], A[2], B[0], B[1], B[2], C[0], C[1], C[2]);
        }
      }
    }
  });
  return makeFloor(new Float32Array(tris));
}

function makeFloor(tris) {
  let x0 = Infinity, y0 = Infinity, x1 = -Infinity, y1 = -Infinity, z0 = Infinity, z1 = -Infinity;
  const n = tris.length / 9;
  for (let i = 0; i < n * 3; i++) {
    x0 = Math.min(x0, tris[i * 3]); x1 = Math.max(x1, tris[i * 3]);
    y0 = Math.min(y0, tris[i * 3 + 1]); y1 = Math.max(y1, tris[i * 3 + 1]);
    z0 = Math.min(z0, tris[i * 3 + 2]); z1 = Math.max(z1, tris[i * 3 + 2]);
  }
  const w = Math.max(1, Math.ceil((x1 - x0) / BIN) + 1), h = Math.max(1, Math.ceil((y1 - y0) / BIN) + 1);
  const bins = Array.from({ length: w * h }, () => []);
  for (let t = 0; t < n; t++) {
    const o = t * 9;
    const tx0 = Math.min(tris[o], tris[o + 3], tris[o + 6]), tx1 = Math.max(tris[o], tris[o + 3], tris[o + 6]);
    const ty0 = Math.min(tris[o + 1], tris[o + 4], tris[o + 7]), ty1 = Math.max(tris[o + 1], tris[o + 4], tris[o + 7]);
    for (let j = Math.floor((ty0 - y0) / BIN); j <= Math.floor((ty1 - y0) / BIN); j++) {
      for (let i = Math.floor((tx0 - x0) / BIN); i <= Math.floor((tx1 - x0) / BIN); i++) bins[j * w + i].push(t);
    }
  }
  // The highest floor under (x, y) at or below fromZ + STEP, or null.
  const below = (x, y, fromZ) => {
    const i = Math.floor((x - x0) / BIN), j = Math.floor((y - y0) / BIN);
    if (i < 0 || j < 0 || i >= w || j >= h) return null;
    let best = null;
    const top = fromZ + STEP;
    for (const t of bins[j * w + i]) {
      const o = t * 9;
      const ax = tris[o], ay = tris[o + 1], bx = tris[o + 3], by = tris[o + 4], cx = tris[o + 6], cy = tris[o + 7];
      const d = (by - cy) * (ax - cx) + (cx - bx) * (ay - cy);
      if (d === 0) continue;
      const l1 = ((by - cy) * (x - cx) + (cx - bx) * (y - cy)) / d;
      const l2 = ((cy - ay) * (x - cx) + (ax - cx) * (y - cy)) / d;
      const l3 = 1 - l1 - l2;
      if (l1 < -1e-4 || l2 < -1e-4 || l3 < -1e-4) continue;
      const z = l1 * tris[o + 2] + l2 * tris[o + 5] + l3 * tris[o + 8];
      if (z <= top && (best === null || z > best)) best = z;
    }
    return best;
  };
  return { below, triangles: n, bounds: { x0, y0, z0, x1, y1, z1 } };
}

// spawnAt finds where to stand at an entrance: on the floor under the
// box's centre, or failing that the nearest floor within a few hundred
// units around it, searched in rings. An exit's box stands on its floor, so
// the floor is sought from just above the box's foot; from its top, the
// ceiling over a low passage would be taken instead.
export function spawnAt(floor, box) {
  const cx = (box.min[0] + box.max[0]) / 2, cy = (box.min[1] + box.max[1]) / 2;
  const from = box.min[2] + 100;
  for (let r = 0; r <= 600; r += 40) {
    const steps = r === 0 ? 1 : Math.ceil((2 * Math.PI * r) / 40);
    for (let k = 0; k < steps; k++) {
      const a = (k / steps) * Math.PI * 2;
      const x = cx + Math.cos(a) * r, y = cy + Math.sin(a) * r;
      const z = floor.below(x, y, from);
      if (z !== null) return [x, y, z];
    }
  }
  return [cx, cy, box.min[2]];
}

// openWay finds which way the floor runs from p: of sixteen headings, the
// one along which there is walkable floor -- each 50-unit step no more than
// a step up or a short drop from the last -- for longest, up to 1200 units.
// It returns the heading (as the viewer counts yaw: 0 towards -y) and how
// far the floor went.
export function openWay(floor, p) {
  let best = 0, reach = -1;
  for (let k = 0; k < 16; k++) {
    const yaw = (k / 16) * Math.PI * 2;
    const dx = Math.sin(yaw), dy = -Math.cos(yaw);
    let z = p[2], d = 0;
    for (; d < 1200; d += 50) {
      const nz = floor.below(p[0] + dx * (d + 50), p[1] + dy * (d + 50), z);
      if (nz === null || z - nz > 120) break;
      z = nz;
    }
    if (d > reach) { reach = d; best = yaw; }
  }
  return { yaw: best, reach };
}

// loadInterior builds the interior world from its zone.json (man) and its
// converted props.
export async function loadInterior(man, zoneDir, { fetchJSON, fetchBuffer }) {
  const manifest = await fetchJSON(zoneDir + '/props/props.json');
  const raw = await fetchBuffer(zoneDir + '/props/props.bin');
  const floor = buildFloor(manifest, raw);
  const solid = buildSolid(manifest, raw);
  // What stops a player: the walls, and the edges of the floor.
  const mover = ledgeMover(buildInteriorWalls(manifest, raw, man.bounds), floor);
  const b = man.bounds;
  const entrance = (man.entrances || [])[0];
  const start = entrance ? spawnAt(floor, entrance) : [(b.min[0] + b.max[0]) / 2, (b.min[1] + b.max[1]) / 2, b.max[2]];
  // Face the way the floor runs from the entrance, and stand a few steps
  // along it, so the chase camera behind has the passage to hang in rather
  // than the rock the exit is cut into.
  const way = openWay(floor, start);
  const facing = way.yaw;
  const ahead = Math.min(300, way.reach / 2);
  start[0] += Math.sin(facing) * ahead;
  start[1] -= Math.cos(facing) * ahead;
  start[2] = floor.below(start[0], start[1], start[2] + STEP) ?? start[2];
  const mx = (b.min[0] + b.max[0]) / 2, my = (b.min[1] + b.max[1]) / 2;
  const cell = 256;
  const lights = (man.lights || []).map((l) => [l.pos[0], l.pos[2], l.pos[1]]);
  const overview = {
    pos: [mx, b.max[2] + 3000, b.max[1] + 4000],
    target: [mx, (b.min[2] + b.max[2]) / 2, my],
  };
  const groundAt = (x, y, fromZ = b.max[2]) => {
    const z = floor.below(x, y, fromZ);
    return z === null ? fromZ : z;
  };
  return {
    kind: man.kind, cell, grid: 0, extent: Math.max(b.max[0], b.max[1]),
    overview,
    spawn: [start[0], start[1]], spawnZ: start[2], spawnYaw: facing,
    hud: [
      ['i-grid', man.kind + ', no heightmap'],
      ['i-height', Math.round(b.min[2]) + ' .. ' + Math.round(b.max[2]) + 'u'],
      ['i-tris', floor.triangles.toLocaleString() + ' floor triangles'],
      ['i-atlas', '—'],
      ['i-align', 'origin ' + man.transform.origin.map((v) => Math.round(v)).join(', ') +
        ' · ' + man.transform.onSurface + '/' + man.transform.points + ' exits and lights on the geometry'],
      ['i-water', (man.entrances || []).length + ' exits, ' + (man.lights || []).length + ' lights'],
    ],
    heightAtCell: (cx, cy) => groundAt(cx * cell, cy * cell),
    groundAt,
    isWet: () => false,
    clamp: (x, y) => [Math.max(b.min[0], Math.min(b.max[0], x)), Math.max(b.min[1], Math.min(b.max[1], y))],
    // No sky: a dark fog closes the view at a cave's own distances.
    fog: { color: [0.02, 0.02, 0.025], start: 1500, end: 9000 },
    // Pulls the chase camera in to just short of whatever solid lies between
    // the character and it; eye and target in world space (height second).
    clipCamera(target, eye) {
      const a = [target[0], target[2], target[1]], b = [eye[0], eye[2], eye[1]];
      const t = solid.first(a, b);
      if (t >= 1) return eye;
      const len = Math.hypot(b[0] - a[0], b[1] - a[1], b[2] - a[2]) || 1;
      const k = Math.max(0, t - 12 / len);
      return [target[0] + (eye[0] - target[0]) * k, target[1] + (eye[1] - target[1]) * k, target[2] + (eye[2] - target[2]) * k];
    },
    // The lights nearest the eye, world space (height second), for
    // lighting.js. Each is a warm torch: lights.csv says only a type number,
    // and nothing in the install says what the types are.
    lightsNear(eye, n = 8) {
      return lights
        .map((p) => ({ p, d: Math.hypot(p[0] - eye[0], p[1] - eye[1], p[2] - eye[2]) }))
        .sort((a, b) => a.d - b.d).slice(0, n)
        .map(({ p }) => ({ pos: p, color: TORCH, radius: TORCH_RADIUS }));
    },
    drawGround: () => {},
    drawWater: () => {},
    floor,
    mover,
  };
}

// buildSolid keeps each model's solid triangles in its own space, with each
// instance's bounding sphere, for testing a short segment against what is
// near it: the chase camera must not end up inside a wall or above a
// ceiling. Testing in model space needs no world-space copy of the
// geometry, which for the largest dungeons would run to hundreds of MB.
export function buildSolid(manifest, raw) {
  const stride = manifest.stride / 4;
  const vf = new Float32Array(raw, 0, manifest.vertexBytes / 4);
  const ix = new Uint32Array(raw, manifest.vertexBytes, manifest.indexBytes / 4);
  const models = manifest.models.map((m) => {
    const t = [];
    for (const g of m.groups) {
      // Glows, cards and see-through things are not walls.
      if (g.billboard || g.additive || g.alpha) continue;
      for (let i = g.first; i + 2 < g.first + g.count; i++) {
        if ((i - g.first) % 3) continue;
        for (const k of [ix[i], ix[i + 1], ix[i + 2]]) {
          t.push(vf[k * stride], vf[k * stride + 1], vf[k * stride + 2]);
        }
      }
    }
    const r = Math.hypot(...[0, 1, 2].map((a) => Math.max(Math.abs(m.min[a]), Math.abs(m.max[a]))));
    return { tris: new Float32Array(t), radius: r };
  });
  const insts = manifest.instances.map((inst) => ({
    inst, model: models[inst.m],
    // Zone coordinates of the instance's origin; its sphere is about it.
    c: [inst.p[0], inst.p[1], inst.p[2]],
    r: models[inst.m].radius * inst.s,
  }));
  // toModel takes a zone-space point into an instance's model space: the
  // instance turn undone (yaw about the axis, negated) and the scale.
  const toModel = (it, q) => {
    const inst = it.inst, a = inst.axis || [0, 0, 1];
    const v = [(q[0] - inst.p[0]) / inst.s, (q[1] - inst.p[1]) / inst.s, (q[2] - inst.p[2]) / inst.s];
    const c = Math.cos(-inst.yaw), s = Math.sin(-inst.yaw);
    const d = (a[0] * v[0] + a[1] * v[1] + a[2] * v[2]) * (1 - c);
    return [
      v[0] * c + (a[1] * v[2] - a[2] * v[1]) * s + a[0] * d,
      v[1] * c + (a[2] * v[0] - a[0] * v[2]) * s + a[1] * d,
      v[2] * c + (a[0] * v[1] - a[1] * v[0]) * s + a[2] * d,
    ];
  };
  // first is the fraction along a to b, zone coordinates, at which the
  // segment first meets solid geometry, or 1.
  const first = (a, b) => {
    const mid = [(a[0] + b[0]) / 2, (a[1] + b[1]) / 2, (a[2] + b[2]) / 2];
    const half = Math.hypot(b[0] - a[0], b[1] - a[1], b[2] - a[2]) / 2;
    let best = 1;
    for (const it of insts) {
      const dc = Math.hypot(it.c[0] - mid[0], it.c[1] - mid[1], it.c[2] - mid[2]);
      if (dc > it.r + half) continue;
      const o = toModel(it, a), e = toModel(it, b);
      const dir = [e[0] - o[0], e[1] - o[1], e[2] - o[2]];
      const t = it.model.tris;
      for (let k = 0; k < t.length; k += 9) {
        const h = hit(o, dir, t, k);
        if (h !== null && h < best) best = h;
      }
    }
    return best;
  };
  return { first };
}

// hit is where a ray o + t*dir, t in [0, 1], meets the triangle at t[k..k+9]
// (Moller-Trumbore, either face), or null.
function hit(o, dir, t, k) {
  const e1 = [t[k + 3] - t[k], t[k + 4] - t[k + 1], t[k + 5] - t[k + 2]];
  const e2 = [t[k + 6] - t[k], t[k + 7] - t[k + 1], t[k + 8] - t[k + 2]];
  const p = [dir[1] * e2[2] - dir[2] * e2[1], dir[2] * e2[0] - dir[0] * e2[2], dir[0] * e2[1] - dir[1] * e2[0]];
  const det = e1[0] * p[0] + e1[1] * p[1] + e1[2] * p[2];
  if (Math.abs(det) < 1e-9) return null;
  const inv = 1 / det;
  const s = [o[0] - t[k], o[1] - t[k + 1], o[2] - t[k + 2]];
  const u = (s[0] * p[0] + s[1] * p[1] + s[2] * p[2]) * inv;
  if (u < 0 || u > 1) return null;
  const q = [s[1] * e1[2] - s[2] * e1[1], s[2] * e1[0] - s[0] * e1[2], s[0] * e1[1] - s[1] * e1[0]];
  const w = (dir[0] * q[0] + dir[1] * q[1] + dir[2] * q[2]) * inv;
  if (w < 0 || u + w > 1) return null;
  const tt = (e2[0] * q[0] + e2[1] * q[1] + e2[2] * q[2]) * inv;
  return tt >= 0 && tt <= 1 ? tt : null;
}
