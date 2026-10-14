// Headless checks on placing props: the instance turn the vertex shader
// applies, mirrored on the CPU by instancePoint.
//
// Run: node web/props.test.mjs
//
// An instance without an axis turns about up, and must land exactly where
// the plain heading always put it; a turn about another axis must keep
// lengths and move the axis's own direction nowhere.

import { instancePoint } from './props.js';
import { check, finish } from './testkit.mjs';

// The heading formula the shader used before axes: model X and Y on the
// ground, Z up, world Y up.
const heading = (inst, v) => {
  const c = Math.cos(inst.yaw), s = Math.sin(inst.yaw);
  return [
    inst.s * (c * v[0] - s * v[1]) + inst.p[0],
    inst.s * v[2] + inst.p[2],
    inst.s * (s * v[0] + c * v[1]) + inst.p[1],
  ];
};
const close = (a, b) => a.every((x, i) => Math.abs(x - b[i]) < 1e-6);

let same = true;
for (const yaw of [0, 0.7, -1.570796, Math.PI, 2.5]) {
  for (const v of [[1, 0, 0], [0, 1, 0], [3, -2, 5], [-120, 40, 7]]) {
    const inst = { p: [52736, 19200, 4960], yaw, s: 1.5 };
    same = same && close(instancePoint(inst, v), heading(inst, v));
    same = same && close(instancePoint({ ...inst, axis: [0, 0, 1] }, v), heading(inst, v));
  }
}
check('an instance without an axis turns exactly as the heading did', same);

const tipped = { p: [0, 0, 0], yaw: 1.1, s: 1, axis: [0.6, 0, 0.8] };
const v = [3, -2, 5];
const w = instancePoint(tipped, v);
const len = (a) => Math.hypot(...a);
check('a turn about a tipped axis keeps lengths', Math.abs(len(w) - len(v)) < 1e-9, len(w) + ' vs ' + len(v));
const onAxis = instancePoint(tipped, [6, 0, 8]); // world [x, z, y] of model (6, 0, 8)
check('a point on the axis stays put', close(onAxis, [6, 8, 0]), onAxis.join(', '));

finish();
