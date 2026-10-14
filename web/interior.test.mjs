// Headless checks on the interior world: finding the floor in a converted
// dungeon, where to start, and keeping the chase camera out of the walls.
//
// Run: node web/interior.test.mjs
//
// They read Spindelhalla (zone 125) as interiorconv writes it into
// web/data; without it they are skipped.

import { existsSync, readFileSync } from 'node:fs';
import { buildFloor, buildSolid, spawnAt, openWay, STEP } from './interior.js';
import { here, check, finish } from './testkit.mjs';

const dir = new URL('./data/zone125/', import.meta.url);
if (!existsSync(new URL('props/props.bin', dir))) {
  console.log('  skip  no converted zone 125 (run go run ./cmd/interiorconv -zone 125)');
  process.exit(0);
}
void here;
const man = JSON.parse(readFileSync(new URL('zone.json', dir)));
const pm = JSON.parse(readFileSync(new URL('props/props.json', dir)));
const buf = readFileSync(new URL('props/props.bin', dir));
const raw = buf.buffer.slice(buf.byteOffset, buf.byteOffset + buf.byteLength);
const floor = buildFloor(pm, raw);

const exit = man.entrances.find((e) => e.name === 'ExittoMyrk');
const start = spawnAt(floor, exit);
check('the start stands on a floor within the exit box\'s height',
  start[2] >= exit.min[2] - 200 && start[2] <= exit.max[2], start.map(Math.round).join(', '));

// Walk the open way from the start: every step finds floor, never more
// than a step up from the last, so a ceiling or an upper level is never
// taken for the ground.
const way = openWay(floor, start);
check('the floor runs on from the entrance', way.reach >= 300, way.reach + ' units');
let z = start[2], ok = true, worstUp = 0;
for (let d = 25; d <= way.reach; d += 25) {
  const x = start[0] + Math.sin(way.yaw) * d, y = start[1] - Math.cos(way.yaw) * d;
  const nz = floor.below(x, y, z);
  if (nz === null) { ok = false; break; }
  worstUp = Math.max(worstUp, nz - z);
  z = nz;
}
check('walking in, the feet stay on the floor', ok && worstUp <= STEP, 'largest step up ' + worstUp.toFixed(1));

// Straight up from the start, the ceiling is not the floor: from far above,
// the floor found is still within a storey of the start.
const fromHigh = floor.below(start[0], start[1], start[2] + 40);
check('from just above the feet the floor is the one stood on', Math.abs(fromHigh - start[2]) < 1, String(fromHigh));

// The chase camera: a segment from the character's head straight up through
// the cave's roof meets it.
const solid = buildSolid(pm, raw);
const head = [start[0], start[1], start[2] + 60];
const t = solid.first(head, [head[0], head[1], head[2] + 3000]);
check('a line up from the head meets the roof', t < 1, 'at ' + (t * 3000).toFixed(0) + ' units');

finish();
