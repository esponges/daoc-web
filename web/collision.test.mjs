// The browser's collision against the server's.
//   node web/collision.test.mjs
//
// Vectors come from the Go tests (go test ./internal/server/world -update);
// the real-zone check loads Mularn's collision files through the loader the
// viewer uses and walks into a tree and a building.

import { readFileSync, existsSync } from 'node:fs';
import { Collision, isWall, loadOutdoorCollision, BODY_RADIUS } from './collision.js';

let failed = 0;
const ok = (cond, msg) => { if (!cond) { failed++; console.error('FAIL: ' + msg); } else console.log('ok: ' + msg); };

const scenes = JSON.parse(readFileSync(new URL('../internal/server/world/testdata/collision_vectors.json', import.meta.url)));
for (const s of scenes) {
  const walls = (s.tris || []).filter((t) => isWall(t));
  const c = new Collision(walls, s.cyls || [], [0, 0, 10000, 10000]);
  let [x, y] = [s.start[0], s.start[1]];
  for (let i = 0; i < s.steps; i++) [x, y] = c.slide(x, y, x + s.step[0], y + s.step[1], s.start[2]);
  const d = Math.max(Math.abs(x - s.end[0]), Math.abs(y - s.end[1]));
  ok(d < 1e-3, `"${s.name}" ends where the server's does (off by ${d.toExponential(1)})`);
}

// --- real zone ---
const root = new URL('./data/', import.meta.url);
if (existsSync(new URL('zone100/props/collision.json', root))) {
  const fetchJSON = async (p) => JSON.parse(readFileSync(new URL(p.replace(/^data\//, ''), root)));
  const fetchBuffer = async (p) => { const b = readFileSync(new URL(p.replace(/^data\//, ''), root)); return b.buffer.slice(b.byteOffset, b.byteOffset + b.byteLength); };
  const c = await loadOutdoorCollision('data/zone100', { fetchJSON, fetchBuffer });
  ok(c && c.cyls.length >= 700, `Mularn's ${c?.cyls.length} tree trunks load`);
  ok(c && c.tris.length > 100, `and ${c?.tris.length} wall triangles`);
  const tr = c.cyls[0];
  let [x, y] = [tr[0] - 400, tr[1] + 3];
  let inside = false;
  for (let i = 0; i < 400; i++) {
    [x, y] = c.slide(x, y, x + 3.2, y, tr[3] + 50);
    if (Math.hypot(x - tr[0], y - tr[1]) < tr[2] + BODY_RADIUS - 1) inside = true;
  }
  ok(!inside, 'walking at a tree never enters it');
} else {
  console.log('skip: zone 100 has no collision files (go run ./cmd/propconv -zone 100)');
}

console.log(failed ? `\n${failed} FAILED` : '\nall passed');
process.exit(failed ? 1 : 0);
