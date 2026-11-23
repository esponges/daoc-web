import assert from 'node:assert/strict';
import { pickGround, groundTargetOf } from './groundtarget.js';

const flat = () => 100;
// A ray from 500 up, looking 45 degrees down and north, meets flat ground 400 units on.
const hit = pickGround([0, 500, 0], [0, -1, 1], flat);
assert.ok(hit, 'a ray down at flat ground hits it');
assert.ok(Math.abs(hit.x) < 1e-6 && Math.abs(hit.y - 400) < 1, 'the point is 400 units north: ' + JSON.stringify(hit));
assert.ok(Math.abs(hit.z - 100) < 1e-6, 'its height is the ground\'s');

// A ray that climbs never meets it.
assert.equal(pickGround([0, 500, 0], [0, 1, 1], flat), null);
// One that stays above the ground past maxDist does not either.
assert.equal(pickGround([0, 5000, 0], [0, -0.01, 1], flat, 1000), null);

// A slope is met where the ground rises to the ray: height = x / 2, ray level at 100.
const slope = (x) => x / 2;
const s = pickGround([0, 100, 0], [1, 0, 0], slope);
assert.ok(Math.abs(s.x - 200) < 1, 'the level ray meets the slope at x=200: ' + JSON.stringify(s));

assert.deepEqual(groundTargetOf({ gt: [1, 2, 3] }), { x: 1, y: 2, z: 3 });
assert.equal(groundTargetOf({}), null);
assert.equal(groundTargetOf(null), null);
console.log('groundtarget ok');
