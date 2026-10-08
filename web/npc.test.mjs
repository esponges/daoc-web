// Headless checks on the NPCs: load the real spawn file and the converted
// characters it names, then run ten simulated minutes of wandering.
//
// Run: node web/npc.test.mjs
//
// WebGL is stubbed out -- every call is a no-op -- so this exercises
// everything except the pixels: loading, the walk speed read from each
// clip, wandering, turning, water avoidance, cross-fading and posing.

import { readFile } from 'node:fs/promises';
import { join } from 'node:path';
import { here, gl, helpers, check, finish } from './testkit.mjs';
import { createNPCs } from './npc.js';

// Flat ground, and a lake laid across the large wolves' home so there is
// something to avoid: a disc 900 units across, centred on cell (128, 132).
const cell = 256, extent = 65536;
const lake = { x: 128 * cell, y: 132 * cell, r: 900 };
const isWet = (x, y) => Math.hypot(x - lake.x, y - lake.y) < lake.r;
const world = { cell, extent, groundAt: () => 0, isWet };

const spec = JSON.parse(await readFile(join(here, 'spawns', 'zone100.json'), 'utf8'));
const npcs = await createNPCs(gl, 'spawns/zone100.json', helpers, world);
const kinds = new Set(spec.groups.map((g) => g.char));
const want = spec.groups.reduce((s, g) => s + (g.count ?? 1), 0);

console.log(`${npcs.npcs.length} NPCs of ${npcs.types.size} kinds`);
check('every character in the spawn file loads', npcs.types.size === kinds.size,
  [...npcs.types.keys()].join(', '));
check('every spawn is placed', npcs.npcs.length === want, `${npcs.npcs.length}/${want}`);

for (const [name, t] of npcs.types) {
  check(`${name} walks at its clip's own pace`, t.walk > 10 && t.walk < 150,
    `${t.walk.toFixed(1)} u/s at 1x, drawn at ${t.ch.scale}x`);
}

const spots = new Set(npcs.npcs.map((n) => Math.round(n.x) + ',' + Math.round(n.y)));
check('no two NPCs start on the same spot', spots.size === npcs.npcs.length,
  `${spots.size} distinct of ${npcs.npcs.length}`);

// Ten minutes at 30 fps.
const dt = 1 / 30, steps = 10 * 60 * 30;
const travelled = npcs.npcs.map(() => 0);
let wet = 0, strayed = 0, overspeed = 0, worstStray = 0, walkingWrongClip = 0, drawErr = null;
for (let s = 0; s < steps; s++) {
  const before = npcs.npcs.map((n) => [n.x, n.y]);
  npcs.update(dt);
  npcs.npcs.forEach((n, i) => {
    const step = Math.hypot(n.x - before[i][0], n.y - before[i][1]);
    travelled[i] += step;
    if (step > n.speed * dt + 1e-6) overspeed++;
    if (isWet(n.x, n.y)) wet++;
    const away = Math.hypot(n.x - n.hx, n.y - n.hy) - n.radius;
    if (away > 1) { strayed++; worstStray = Math.max(worstStray, away); }
    if (step > 0 && n.anim.cur !== 'walk') walkingWrongClip++;
  });
  // Pose and draw now and then, from a camera in the middle of everything.
  if (s % 300 === 0) {
    try {
      npcs.draw({ viewProj: new Float32Array(16), camPos: [150 * cell, 0, 160 * cell],
        fogColor: [0, 0, 0], fogStart: 1, fogEnd: 2, lightDir: [0, 1, 0] });
    } catch (e) {
      drawErr = e;
    }
  }
}

const moved = travelled.filter((d) => d > 50).length;
check('every NPC wanders', moved === npcs.npcs.length, `${moved}/${npcs.npcs.length} moved`);
check('none ever moves faster than its walk', overspeed === 0, `${overspeed} steps`);
check('none strays beyond its spawn radius', strayed === 0,
  strayed ? `${strayed} steps, worst ${worstStray.toFixed(0)}u out` : 'all inside');
check('none walks into the water', wet === 0, `${wet} steps wet`);
check('anything moving is playing its walk', walkingWrongClip === 0, `${walkingWrongClip} steps`);
check('posing and drawing every NPC succeeds', !drawErr, drawErr ? drawErr.message : `${npcs.drawn} drawn`);
const finite = npcs.npcs.every((n) => Number.isFinite(n.x + n.y + n.yaw + n.anim.t));
check('state stays finite', finite);

finish();
