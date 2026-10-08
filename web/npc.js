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
// and each keeps only its own position and animation state, posing the
// shared skeleton just before it is drawn.
//
// Fighting is combat.js's business. While an NPC is in a fight, or dead, the
// wandering here leaves it alone.

import { createCharacter, modelMatrix } from './character.js';
import { Animator } from './animator.js';

// Cross-fade between base clips; a little longer than the player's, since
// nothing is waiting on an NPC to respond.
const FADE = 0.3;

// Beyond this distance from the camera an NPC is neither posed nor drawn.
// Its wandering still runs, so it is somewhere sensible when you arrive.
const DRAW_DISTANCE = 14000;

// How fast an NPC turns, in radians per second.
export const TURN_RATE = 3.5;

// Small deterministic generator, so the same file spawns the same scene.
export function mulberry32(seed) {
  let a = seed >>> 0;
  return () => {
    a = (a + 0x6D2B79F5) >>> 0;
    let t = a;
    t = Math.imul(t ^ (t >>> 15), t | 1);
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61);
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

export function turnToward(cur, target, maxStep) {
  let d = (target - cur) % (Math.PI * 2);
  if (d > Math.PI) d -= Math.PI * 2;
  if (d < -Math.PI) d += Math.PI * 2;
  return cur + Math.max(-maxStep, Math.min(maxStep, d));
}

// angleOff is how far heading a is from heading b, 0..PI.
export function angleOff(a, b) {
  return Math.abs((((a - b) % (Math.PI * 2)) + Math.PI * 3) % (Math.PI * 2) - Math.PI);
}

// bodyRadius is how far a figure's body reaches from its centre along the way
// it faces, which is what melee range is measured between. Its width is no
// use: a player race's bind pose has the arms straight out.
export function bodyRadius(ch, scale) {
  const m = ch.manifest;
  return 0.5 * (m.max[1] - m.min[1]) * scale;
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
      // The walk's own pace, so the feet stay planted at 1x. The run is
      // read the same way but less reliably -- a running foot is barely
      // down -- so it is only trusted inside a sane band of the walk's.
      const walk = ch.groundSpeed('walk') || 40;
      let run = ch.groundSpeed('run');
      if (!(run > 1.5 * walk && run < 8 * walk)) run = 3 * walk;
      types.set(name, { ch, walk, run, title: ch.manifest.title || name });
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
  spec.groups.forEach((g, gi) => {
    const type = types.get(g.char);
    if (!type) return;
    const hx = g.x * cell, hy = g.y * cell, radius = (g.radius ?? 4) * cell;
    for (let i = 0; i < (g.count ?? 1); i++) {
      const [x, y] = pickSpot(hx, hy, radius);
      // Scale from the table, times any the spawn adds.
      const scale = type.ch.scale * (g.scale ?? 1);
      const hp = g.hp ?? 60;
      npcs.push({
        name: type.title, char: g.char, group: gi, type, x, y, hx, hy, radius, scale,
        yaw: rng() * Math.PI * 2,
        speed: type.walk * scale,
        runSpeed: type.run * scale,
        idleMin: g.idle?.[0] ?? 3, idleMax: g.idle?.[1] ?? 10,
        state: 'idle', timer: rng() * 6, tx: x, ty: y,
        anim: new Animator(type.ch, { fade: FADE }),
        // Fighting. aggro is how close the player may come before an
        // aggressive NPC attacks unprovoked; 0 means it only fights back.
        hp, maxHp: hp,
        damage: g.damage ?? [3, 7],
        swing: g.swing ?? 2.5,
        hitChance: g.hit ?? 0.7,
        aggro: g.aggro ?? 0,
        body: bodyRadius(type.ch, scale),
        fight: null, dead: false, gone: false,
      });
    }
  });

  // respawn puts an NPC back at a fresh spot near home, whole and calm.
  function respawn(n) {
    [n.x, n.y] = pickSpot(n.hx, n.hy, n.radius);
    n.hp = n.maxHp;
    n.dead = n.gone = false;
    n.fight = null;
    n.state = 'idle';
    n.timer = n.idleMin + rng() * (n.idleMax - n.idleMin);
    n.anim.clear();
    n.anim.setBase('idle');
  }

  // wanderTo sends an NPC walking to a spot, on its own or because combat.js
  // has sent it home.
  function wanderTo(n, x, y) {
    n.tx = x; n.ty = y;
    n.state = 'walk';
    // Give up on a target it cannot reach in reasonable time.
    n.timer = 4 * Math.hypot(n.tx - n.x, n.ty - n.y) / n.speed + 2;
  }

  // step moves an NPC toward (tx, ty) at speed, turning first. Returns the
  // distance left, or -1 if water was in the way.
  function step(n, tx, ty, speed, dt) {
    const dx = tx - n.x, dy = ty - n.y, d = Math.hypot(dx, dy);
    if (d < 1e-3) return 0;
    // Same heading convention as the player: yaw y faces (sin y, -cos y).
    const want = Math.atan2(dx, -dy);
    n.yaw = turnToward(n.yaw, want, TURN_RATE * dt);
    // Move only once roughly facing the target, so it turns on the spot
    // instead of sweeping round in a wide arc.
    if (angleOff(want, n.yaw) >= 0.6) return d;
    const s = Math.min(d, speed * dt);
    const nx = n.x + Math.sin(n.yaw) * s, ny = n.y - Math.cos(n.yaw) * s;
    if (isWet(nx, ny)) return -1;
    n.x = nx; n.y = ny;
    return d - s;
  }

  function update(dt) {
    for (const n of npcs) {
      // In a fight or dead, combat.js drives position and clips; the clocks
      // still run here.
      if (!n.fight && !n.dead) {
        if (n.state === 'idle') {
          n.timer -= dt;
          if (n.timer <= 0) wanderTo(n, ...pickSpot(n.hx, n.hy, n.radius));
        } else {
          n.timer -= dt;
          // Arrival is judged before moving, so the frame that ends a walk
          // is not also a frame of walking.
          const left = Math.hypot(n.tx - n.x, n.ty - n.y);
          if (left >= Math.max(12, n.speed * dt) && n.timer > 0) {
            if (step(n, n.tx, n.ty, n.speed, dt) < 0) n.timer = 0; // water ahead
          } else {
            n.state = 'idle';
            n.timer = n.idleMin + rng() * (n.idleMax - n.idleMin);
            // Coming home from a fight heals.
            if (n.returning) { n.returning = false; n.hp = n.maxHp; }
          }
        }
        n.anim.setBase(n.state === 'walk' ? 'walk' : 'idle');
      }
      n.anim.update(dt);
    }
  }

  let drawn = 0;
  function draw(ctx) {
    drawn = 0;
    const [cx, , cz] = ctx.camPos;
    for (const n of npcs) {
      if (n.gone || Math.hypot(n.x - cx, n.y - cz) > DRAW_DISTANCE) continue;
      n.anim.pose();
      n.type.ch.draw({ ...ctx, model: modelMatrix(n.x, groundAt(n.x, n.y), n.y, n.yaw, n.scale) });
      drawn++;
    }
  }

  return {
    npcs,
    types,
    update,
    draw,
    respawn,
    wanderTo,
    step,
    rng,
    get drawn() { return drawn; },
  };
}
