// Headless checks on the sun's shadow camera.
//
// Run: node web/lighting.test.mjs
//
// The shaders cannot run here, but the matrix that decides where every
// shadow lands can: the square must be centred on the player, reach the
// radius it claims and no further, keep everything near the ground inside
// its depth range, and put a caster between the sun and the ground nearer
// the light than the ground it shades.

import { lightMatrix, SUN_DIR } from './lighting.js';
import { check, finish } from './testkit.mjs';

const apply = (m, p) => [0, 1, 2].map((r) => m[r] * p[0] + m[4 + r] * p[1] + m[8 + r] * p[2] + m[12 + r]);
const radius = 2400, size = 2048;
const centre = [50000, 4700, 52000];
const m = lightMatrix(centre, { size, radius });

const c = apply(m, centre);
check('the player is at the middle of the map', Math.abs(c[0]) < 2 / size * 2 && Math.abs(c[1]) < 2 / size * 2,
  'ndc ' + c[0].toFixed(4) + ', ' + c[1].toFixed(4));

// Ground points across the square, in light space, must reach about the
// radius and fit inside the depth range.
let worst = 0, depthOk = true;
for (const [dx, dz] of [[radius, 0], [-radius, 0], [0, radius], [0, -radius]]) {
  const p = apply(m, [centre[0] + dx, centre[1], centre[2] + dz]);
  worst = Math.max(worst, Math.abs(p[0]), Math.abs(p[1]));
  depthOk = depthOk && p[2] > -1 && p[2] < 1;
}
check('ground a radius away lands near the edge, not off it', worst > 0.5 && worst < 1.05, 'furthest ' + worst.toFixed(3));
check('ground and a tall caster both fit the depth range', depthOk &&
  Math.abs(apply(m, [centre[0], centre[1] + 3000, centre[2]])[2]) < 1);

// A point up the sun's beam from the player shades the player: same place
// on the map, nearer the light.
const up = apply(m, centre.map((v, i) => v + SUN_DIR[i] * 500));
check('a caster toward the sun covers the player and is nearer the light',
  Math.hypot(up[0] - c[0], up[1] - c[1]) < 1e-3 && up[2] < c[2],
  'depth ' + up[2].toFixed(4) + ' vs ' + c[2].toFixed(4));

// As the player walks, a fixed point on the ground may shift across the map
// only by whole texels; anything in between makes shadow edges crawl.
let offGrid = 0;
const fixed = [centre[0] + 700, centre[1], centre[2] - 300];
const p0 = apply(m, fixed);
for (let k = 1; k <= 40; k++) {
  const mk = lightMatrix([centre[0] + k * 0.37, centre[1], centre[2] + k * 0.21], { size, radius });
  const pk = apply(mk, fixed);
  for (const i of [0, 1]) {
    const t = (pk[i] - p0[i]) / (2 / size);
    offGrid = Math.max(offGrid, Math.abs(t - Math.round(t)));
  }
}
check('the map moves in whole texels', offGrid < 1e-2, 'worst fraction of a texel ' + offGrid.toFixed(4));

finish();
