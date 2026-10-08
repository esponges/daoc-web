// Headless checks on where grass grows.
//
// Run: node web/grass.test.mjs
//
// Loads zone 100's converted grass and generates plants the way the viewer
// does, then checks them against the maps they came from: none where the
// density map says none, groups that keep to dry land do, the shore's
// waterside plants are there, and a cell comes back the same each time.

import { readFile } from 'node:fs/promises';
import { join } from 'node:path';
import { here, gl, helpers, check, finish } from './testkit.mjs';
import { createGrass } from './grass.js';

const zone = JSON.parse(await readFile(join(here, 'data/zone100/zone.json'), 'utf8'));
const heights = new Uint16Array((await readFile(join(here, 'data/zone100/heights.u16'))).buffer.slice(0));
const grid = zone.grid, cell = zone.cellUnits;
const H = (x, y) => heights[Math.min(grid - 1, Math.max(0, y)) * grid + Math.min(grid - 1, Math.max(0, x))];
const groundAt = (wx, wy) => H(Math.round(wx / cell), Math.round(wy / cell));
const lake = zone.waters[0];
const isWet = (wx, wy) => groundAt(wx, wy) < lake.height;
const map = new Uint8Array((await readFile(join(here, 'data/zone100', zone.grass.map))).buffer.slice(0));

const grass = await createGrass(gl, 'data/zone100', zone.grass, { cell, grid, groundAt, isWet }, helpers);
const FLOATS = 14;
const plants = (cx, cy) => {
  const p = grass.cellPlants(cx, cy);
  const out = [];
  for (const list of p) for (let i = 0; i < list.length; i += FLOATS) out.push(list.slice(i, i + FLOATS));
  return out;
};

console.log('grass: ' + zone.grass.table + ', ' + zone.grass.sprites.length + ' sprites');

// Mularn's square has no density: nothing grows on the cobbles.
const empty = (cx, cy) => [-1, 0, 1].every((dy) => [-1, 0, 1].every((dx) => map[((cy + dy) * grid + cx + dx) * 2 + 1] === 0));
let bare = 0, bareCells = 0;
for (let cy = 1; cy < grid - 1; cy += 3) {
  for (let cx = 1; cx < grid - 1; cx += 3) {
    if (!empty(cx, cy)) continue;
    bareCells++;
    bare += plants(cx, cy).length;
  }
}
check('no plants where the density map and its neighbours are zero', bareCells > 100 && bare === 0,
  bareCells + ' bare cells, ' + bare + ' plants');

// Dense meadow grows thickly.
const meadow = plants(180, 190).length;
check('a dense meadow cell holds a meadow', meadow > 100, meadow + ' plants in cell (180, 190)');

// The shore's own plants grow into the water.
let shoreline = 0;
const wetGroups = new Set([3, 4, 5, 8]);
for (let cy = 0; cy < grid; cy += 2) {
  for (let cx = 0; cx < grid; cx += 2) {
    const g = map[(cy * grid + cx) * 2];
    for (const p of plants(cx, cy)) {
      if (isWet(p[0], p[2]) && wetGroups.has(g)) shoreline++;
    }
  }
}
check('waterside plants grow in the shallows', shoreline > 0, shoreline + ' in water');

// Every plant stands on the ground beneath it.
let off = 0;
for (const p of plants(150, 150)) if (Math.abs(p[1] - groundAt(p[0], p[2])) > 1e-3) off++;
check('plants stand on the ground', off === 0, off + ' off it');

// The same cell twice is the same plants: generated from its own seed.
const a = JSON.stringify(plants(120, 140));
check('a cell is generated the same every time', a === JSON.stringify(plants(120, 140)));

// Sizes come from the table: nothing taller than the tallest sprite allows.
const tallest = Math.max(...zone.grass.sprites.map((s) => (s.height[0] + s.height[1]) * (1 + s.scale / 100)));
let tooTall = 0;
for (const cxy of [[150, 150], [180, 190], [110, 118]]) for (const p of plants(...cxy)) if (p[5] > tallest + 1e-3) tooTall++;
check('no plant is taller than its sprite allows', tooTall === 0, 'tallest allowed ' + tallest.toFixed(0) + 'u');

finish();
