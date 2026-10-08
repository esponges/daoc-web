// Non-player characters: creatures and villagers that wander a zone.
//
// Where an NPC stands is not in the client. The server places every creature
// and tells the client its model number, so nothing in the install says
// "three wolves live here". spawns/zone100.json is therefore this project's
// own placement, written by hand, and it names converted characters rather
// than anything the live game sent.
//
// Each spawn group puts a few of one character around a home point and lets
// them wander it: idle a while, pick a spot nearby, walk there, repeat. They
// share one loaded model per character -- the mesh, textures and clips --
// and each keeps only its own position and animation clock, posing the shared
// skeleton just before it is drawn.

import { createCharacter, modelMatrix } from './character.js';

// Cross-fade between idle and walk; a little longer than the player's, since
// nothing is waiting on an NPC to respond.
const FADE = 0.3;

// Beyond this distance from the camera an NPC is neither posed nor drawn.
// Its wandering still runs, so it is somewhere sensible when you arrive.
const DRAW_DISTANCE = 14000;

// How fast an NPC turns, in radians per second.
const TURN_RATE = 3.5;

// Small deterministic generator, so the same file spawns the same scene.
function mulberry32(seed) {
  let a = seed >>> 0;
  return () => {
    a = (a + 0x6D2B79F5) >>> 0;
    let t = a;
    t = Math.imul(t ^ (t >>> 15), t | 1);
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61);
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

function turnToward(cur, target, maxStep) {
  let d = (target - cur) % (Math.PI * 2);
  if (d > Math.PI) d -= Math.PI * 2;
  if (d < -Math.PI) d += Math.PI * 2;
  return cur + Math.max(-maxStep, Math.min(maxStep, d));
}

// createNPCs loads the spawn file and every character it names. world gives
// the terrain: groundAt(x, y), isWet(x, y), the cell size and zone extent.
export async function createNPCs(gl, url, helpers, world) {
  const res = await fetch(url);
  if (!res.ok) throw new Error('GET ' + url + ' -> ' + res.status);
  const spec = await res.json();
  const { cell, extent, groundAt, isWet } = world;

  // One load per character, however many spawn from it. A character that has
  // not been converted is skipped with a warning rather than failing the lot.
  const types = new Map();
  await Promise.all([...new Set(spec.groups.map((g) => g.char))].map(async (name) => {
    const safe = name.replace(/[^a-zA-Z0-9_-]/g, '');
    try {
      const ch = await createCharacter(gl, 'data/char/' + safe, helpers);
      if (!ch.clips.idle || !ch.clips.walk) {
        throw new Error('needs idle and walk clips; run animconv');
      }
      // The walk's own pace, so the feet stay planted at 1x.
      const walk = ch.groundSpeed('walk');
      types.set(name, { ch, walk: walk > 0 ? walk : 40 });
    } catch (e) {
      console.warn('npc ' + name + ' not loaded:', e.message);
    }
  }));

  const rng = mulberry32(spec.seed ?? 1);
  // A dry spot within radius of a centre, or the centre itself if the
  // neighbourhood is all lake.
  const pickSpot = (cx, cy, radius) => {
    for (let tries = 0; tries < 24; tries++) {
      const a = rng() * Math.PI * 2, r = Math.sqrt(rng()) * radius;
      const x = Math.max(0, Math.min(extent, cx + Math.cos(a) * r));
      const y = Math.max(0, Math.min(extent, cy + Math.sin(a) * r));
      if (!isWet(x, y)) return [x, y];
    }
    return [cx, cy];
  };

  const npcs = [];
  for (const g of spec.groups) {
    const type = types.get(g.char);
    if (!type) continue;
    const hx = g.x * cell, hy = g.y * cell, radius = (g.radius ?? 4) * cell;
    for (let i = 0; i < (g.count ?? 1); i++) {
      const [x, y] = pickSpot(hx, hy, radius);
      // Scale from the table, times any the spawn adds: monsters.csv draws
      // the small grey wolf at half the large one's size from one model.
      const scale = type.ch.scale * (g.scale ?? 1);
      npcs.push({
        name: g.char, type, x, y, hx, hy, radius, scale,
        yaw: rng() * Math.PI * 2,
        speed: type.walk * scale,
        idleMin: g.idle?.[0] ?? 3, idleMax: g.idle?.[1] ?? 10,
        state: 'idle', timer: rng() * 6, tx: x, ty: y,
        // Staggered clocks, or a pack would breathe in unison.
        anim: { cur: 'idle', prev: null, t: rng() * 10, prevT: 0, fade: 1 },
      });
    }
  }

  function update(dt) {
    for (const n of npcs) {
      if (n.state === 'idle') {
        n.timer -= dt;
        if (n.timer <= 0) {
          [n.tx, n.ty] = pickSpot(n.hx, n.hy, n.radius);
          n.state = 'walk';
          // Give up on a target it cannot reach in reasonable time.
          n.timer = 4 * Math.hypot(n.tx - n.x, n.ty - n.y) / n.speed + 2;
        }
      } else {
        const dx = n.tx - n.x, dy = n.ty - n.y, d = Math.hypot(dx, dy);
        n.timer -= dt;
        if (d < Math.max(12, n.speed * dt) || n.timer <= 0) {
          n.state = 'idle';
          n.timer = n.idleMin + rng() * (n.idleMax - n.idleMin);
        } else {
          // Same heading convention as the player: yaw y faces (sin y, -cos y).
          const want = Math.atan2(dx, -dy);
          n.yaw = turnToward(n.yaw, want, TURN_RATE * dt);
          // Walk only once roughly facing the target, so it turns on the
          // spot instead of sweeping round in a wide arc.
          const off = Math.abs(((want - n.yaw) % (Math.PI * 2) + Math.PI * 3) % (Math.PI * 2) - Math.PI);
          if (off < 0.6) {
            const step = Math.min(d, n.speed * dt);
            const nx = n.x + Math.sin(n.yaw) * step, ny = n.y - Math.cos(n.yaw) * step;
            if (!isWet(nx, ny)) { n.x = nx; n.y = ny; } else n.timer = 0;
          }
        }
      }

      const a = n.anim;
      const want = n.state === 'walk' ? 'walk' : 'idle';
      if (want !== a.cur) {
        a.prev = a.cur; a.prevT = a.t;
        a.cur = want; a.t = 0; a.fade = 0;
      }
      a.t += dt;
      if (a.prev) a.prevT += dt;
      if (a.fade < 1) {
        a.fade = Math.min(1, a.fade + dt / FADE);
        if (a.fade >= 1) a.prev = null;
      }
    }
  }

  let drawn = 0;
  function draw(ctx) {
    drawn = 0;
    const [cx, , cz] = ctx.camPos;
    for (const n of npcs) {
      if (Math.hypot(n.x - cx, n.y - cz) > DRAW_DISTANCE) continue;
      const { ch } = n.type, a = n.anim;
      ch.poseBlend(a.prev, a.prevT, a.cur, a.t, a.fade);
      ch.draw({ ...ctx, model: modelMatrix(n.x, groundAt(n.x, n.y), n.y, n.yaw, n.scale) });
      drawn++;
    }
  }

  return {
    npcs,
    types,
    update,
    draw,
    get drawn() { return drawn; },
  };
}
