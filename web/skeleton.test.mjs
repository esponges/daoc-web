// Headless checks on the exported character and the posing maths.
//
// Run: node web/skeleton.test.mjs
//
// The key check is end-to-end. charconv computed the model's bounding box by
// transforming each vertex by its shape's world placement, straight out of the
// NIF. This skins every vertex the way the shader does, from the exported
// skeleton, inverse bind matrices, bone indices and weights, and compares the
// resulting box. The two routes share no code, so agreement means the whole
// chain decodes correctly: the bone hierarchy, the transform composition, the
// vertex layout and the weights.
//
// It cannot check the shader itself, only that the arithmetic it implements is
// fed the right numbers.

import { readFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';
import { Skeleton, skinVertex, apply, Clip } from './skeleton.js';


const here = dirname(fileURLToPath(import.meta.url));
const base = join(here, 'data', 'char', 'norseman');

let failures = 0;
function check(name, ok, detail) {
  console.log(`  ${ok ? 'ok  ' : 'FAIL'}  ${name}${detail ? '   ' + detail : ''}`);
  if (!ok) failures++;
}

const man = JSON.parse(await readFile(join(base, 'char.json'), 'utf8'));
const raw = await readFile(join(base, 'mesh.bin'));
const buf = raw.buffer.slice(raw.byteOffset, raw.byteOffset + raw.byteLength);

console.log(`${man.source} -> ${man.name}`);
console.log(`  ${man.vertexCount} vertices, ${man.indexCount / 3} triangles, ` +
  `${man.bones.length} bones, ${man.shapes.length} shapes`);

// --- buffer shape ---
const vertBytes = man.vertexCount * man.stride;
check('mesh.bin size matches the manifest',
  buf.byteLength === vertBytes + man.indexCount * 4,
  `${buf.byteLength} bytes`);

const f32 = new Float32Array(buf, 0, vertBytes / 4);
const u8 = new Uint8Array(buf, 0, vertBytes);
const idx = new Uint32Array(buf, vertBytes, man.indexCount);
const perVert = man.stride / 4;

// --- skeleton ---
const skel = new Skeleton(man.bones);
check('every bone is reachable from a root', skel.order.length === man.bones.length,
  `${skel.order.length}/${man.bones.length}`);
check('all gait joints are present', skel.missingJoints.length === 0,
  skel.missingJoints.join(', ') || 'Bip01 L Thigh … Bip01 Spine');

// Indices must address real vertices.
let maxIdx = 0;
for (const v of idx) if (v > maxIdx) maxIdx = v;
check('triangle indices stay inside the vertex buffer', maxIdx < man.vertexCount,
  `max index ${maxIdx}`);

// Weights must sum to one, or the figure shrinks toward the origin.
let worstWeight = 0;
for (let v = 0; v < man.vertexCount; v++) {
  const o = v * perVert + 9; // 3 pos + 3 normal + 2 uv + 1 packed bone word
  const s = f32[o] + f32[o + 1] + f32[o + 2] + f32[o + 3];
  worstWeight = Math.max(worstWeight, Math.abs(s - 1));
}
check('vertex weights sum to 1', worstWeight < 1e-5, `worst error ${worstWeight.toExponential(2)}`);

// --- the end-to-end check ---
// Skin every vertex in the bind pose and compare the box with the converter's.
skel.poseBind();
const lo = [Infinity, Infinity, Infinity], hi = [-Infinity, -Infinity, -Infinity];
const pos = [0, 0, 0], bone = [0, 0, 0, 0], wts = [0, 0, 0, 0], out = [0, 0, 0];
let nonFinite = 0;

for (const shape of man.shapes) {
  // Walk this shape's own triangles so each vertex is skinned with the shape
  // whose bone list its indices refer to.
  const seen = new Set();
  for (let i = shape.first; i < shape.first + shape.count; i++) seen.add(idx[i]);
  for (const v of seen) {
    const o = v * perVert;
    pos[0] = f32[o]; pos[1] = f32[o + 1]; pos[2] = f32[o + 2];
    const bo = v * man.stride + 32;
    for (let k = 0; k < 4; k++) {
      bone[k] = u8[bo + k];
      wts[k] = f32[o + 9 + k];
    }
    skinVertex(skel, shape, pos, bone, wts, out);
    for (let k = 0; k < 3; k++) {
      if (!Number.isFinite(out[k])) { nonFinite++; continue; }
      if (out[k] < lo[k]) lo[k] = out[k];
      if (out[k] > hi[k]) hi[k] = out[k];
    }
  }
}

check('no vertex skins to a non-finite position', nonFinite === 0, `${nonFinite} bad`);

const tol = 0.25; // float32 down a ten-joint hierarchy
let worstBox = 0;
for (let k = 0; k < 3; k++) {
  worstBox = Math.max(worstBox, Math.abs(lo[k] - man.min[k]), Math.abs(hi[k] - man.max[k]));
}
const fmt = (a) => '[' + a.map((v) => v.toFixed(2)).join(', ') + ']';
check('CPU-skinned bind pose reproduces the converter bounding box',
  worstBox < tol, `worst delta ${worstBox.toFixed(4)}`);
console.log(`        converter  min ${fmt(man.min)}  max ${fmt(man.max)}`);
console.log(`        skinned    min ${fmt(lo)}  max ${fmt(hi)}`);

// --- the gait ---
// The legs must actually alternate, and the figure must not drift or explode.
const footY = (phase, joint) => {
  skel.poseWalk(phase, 1);
  const m = skel.world[skel.joints[joint]];
  return apply(m, 0, 0, 0)[1]; // model Y is the fore/aft axis
};
const headZ = (phase) => {
  skel.poseWalk(phase, 1);
  const h = skel.boneId('Bip01 Head');
  return apply(skel.world[h], 0, 0, 0)[2];
};

const lAt0 = footY(0.9, 'lFoot'), rAt0 = footY(0.9, 'rFoot');
const lAtPi = footY(0.9 + Math.PI, 'lFoot'), rAtPi = footY(0.9 + Math.PI, 'rFoot');
check('feet swing in antiphase',
  Math.sign(lAt0 - rAt0) === -Math.sign(lAtPi - rAtPi) && Math.abs(lAt0 - rAt0) > 1,
  `separation ${(lAt0 - rAt0).toFixed(1)} -> ${(lAtPi - rAtPi).toFixed(1)} units`);

// A rotation at a joint must pivot its descendants without moving the joint
// itself. This is the invariant that breaks if the conjugation is composed
// with full transforms instead of rotations: the limb detaches and flies off.
// Checking world positions would be wrong, because a joint legitimately moves
// when an ancestor rotates -- the shoulders ride on the spine. The exact
// invariant is local: whatever the gait does, a joint's offset from its parent
// is fixed by the skeleton and must survive posing untouched.
let worstPivot = 0, worstPivotName = '';
for (let phase = 0; phase < Math.PI * 2; phase += 0.37) {
  skel.poseWalk(phase, 1);
  for (let j = 0; j < skel.bones.length; j++) {
    for (let k = 3; k <= 11; k += 4) {
      const d = Math.abs(skel.animLocal[j][k] - skel.bindLocal[j][k]);
      if (d > worstPivot) { worstPivot = d; worstPivotName = skel.bones[j].name; }
    }
  }
}
check('posing rotates joints in place, never translating them',
  worstPivot < 1e-5, `worst offset change ${worstPivot.toExponential(2)} units` +
  (worstPivotName ? ` at ${worstPivotName}` : ''));

let swing = 0;
for (let i = 0; i < 24; i++) {
  const p = (i / 24) * Math.PI * 2;
  swing = Math.max(swing, Math.abs(footY(p, 'lFoot') - footY(p + Math.PI, 'lFoot')));
}
check('stride length is plausible for a 72-unit figure',
  swing > 10 && swing < 80, `peak fore/aft foot travel ${swing.toFixed(1)} units`);

let headDrift = 0, headBase = headZ(0);
for (let i = 0; i < 24; i++) {
  headDrift = Math.max(headDrift, Math.abs(headZ((i / 24) * Math.PI * 2) - headBase));
}
check('the head stays put while the legs move', headDrift < 4,
  `max head height change ${headDrift.toFixed(2)} units`);

// Posing must be a pure function of its arguments: no accumulation per call.
skel.poseWalk(1.3, 1);
const once = Array.from(skel.world[skel.joints.lCalf]);
for (let i = 0; i < 5; i++) skel.poseWalk(1.3, 1);
const again = Array.from(skel.world[skel.joints.lCalf]);
check('repeated posing is idempotent',
  once.every((v, i) => Math.abs(v - again[i]) < 1e-6));

// And the bind pose must be recoverable exactly.
skel.poseWalk(2.0, 1);
skel.poseBind();
const bindAgain = Array.from(skel.world[skel.joints.lCalf]);
skel.poseWalk(0, 0);
check('an idle pose equals the bind pose',
  bindAgain.every((v, i) => Math.abs(v - skel.world[skel.joints.lCalf][i]) < 1e-6));


// --- recorded animation ---------------------------------------------------
//
// The clips come from DAoC's own .kfa files, resolved to bones through
// animnode.dat. These checks are the same ones the procedural gait has to
// pass: if the quaternion component order were wrong, or the bone mapping
// off, the figure would not stand up straight or walk.

let clips = null;
try {
  const idx = JSON.parse(await readFile(join(base, 'anim', 'index.json'), 'utf8'));
  clips = {};
  for (const a of idx.animations) {
    clips[a.name] = new Clip(JSON.parse(await readFile(join(base, 'anim', a.file), 'utf8')), skel);
  }
} catch (e) {
  console.log('  --    recorded animation not converted; run animconv  (' + e.message + ')');
}

if (clips) {
  const names = Object.keys(clips);
  check('every converted clip binds to the skeleton',
    names.length > 0 && names.every((n) => clips[n].tracks.length > 0 && clips[n].missing.length === 0),
    names.map((n) => n + ':' + clips[n].tracks.length + 'tr').join(' '));

  // Sample the whole of every clip and demand the figure stays a figure.
  const boneHeight = (i) => skel.world[i][11];
  const head = skel.boneId('Bip01 Head');
  const lFoot = skel.boneId('Bip01 L Foot'), rFoot = skel.boneId('Bip01 R Foot');
  let worstHead = 0, badFrames = 0, minFoot = Infinity;
  for (const n of names) {
    const c = clips[n];
    for (let k = 0; k <= 48; k++) {
      skel.poseClip(c, (k / 48) * c.duration);
      for (const m of skel.world) {
        for (const v of m) if (!Number.isFinite(v)) badFrames++;
      }
      const h = boneHeight(head);
      // The model is 71.8 units tall; a standing head sits near the top.
      if (h < 55 || h > 80) worstHead = Math.max(worstHead, Math.abs(h - 67));
      minFoot = Math.min(minFoot, boneHeight(lFoot), boneHeight(rFoot));
    }
  }
  check('clips never produce a non-finite transform', badFrames === 0, badFrames + ' bad values');
  check('the head stays at a standing height through every clip',
    worstHead === 0, worstHead ? 'off by ' + worstHead.toFixed(1) : 'always 55..80u');
  check('feet never pass far below the ground plane', minFoot > -12,
    'lowest foot ' + minFoot.toFixed(1) + 'u');

  // A walk cycle has to actually swing the legs, in antiphase.
  const walk = clips.walk;
  if (walk) {
    let sep = [];
    for (let k = 0; k < 32; k++) {
      skel.poseClip(walk, (k / 32) * walk.duration);
      // Fore/aft is the model's Y axis.
      sep.push(skel.world[lFoot][7] - skel.world[rFoot][7]);
    }
    const lo = Math.min(...sep), hi = Math.max(...sep);
    check('the recorded walk swings the feet in antiphase', lo < -8 && hi > 8,
      'separation ' + lo.toFixed(1) + ' .. ' + hi.toFixed(1) + ' units');

    // And it has to loop: the pose at t=0 and t=duration must agree, or the
    // character snaps every cycle.
    skel.poseClip(walk, 0);
    const first = skel.world.map((m) => Array.from(m));
    skel.poseClip(walk, walk.duration);
    let loopErr = 0;
    for (let i = 0; i < first.length; i++) {
      for (let k = 0; k < 12; k++) loopErr = Math.max(loopErr, Math.abs(first[i][k] - skel.world[i][k]));
    }
    check('the walk cycle loops seamlessly', loopErr < 2.0, 'worst joint delta ' + loopErr.toFixed(3));
  }

  // Idle should be near-still compared with walking.
  const idle = clips.idle;
  if (idle && walk) {
    const travel = (c) => {
      let lo = Infinity, hi = -Infinity;
      for (let k = 0; k < 32; k++) {
        skel.poseClip(c, (k / 32) * c.duration);
        const d = skel.world[lFoot][7];
        lo = Math.min(lo, d); hi = Math.max(hi, d);
      }
      return hi - lo;
    };
    const ti = travel(idle), tw = travel(walk);
    check('idle moves the feet far less than walking', ti < tw * 0.5,
      'idle ' + ti.toFixed(1) + 'u vs walk ' + tw.toFixed(1) + 'u');
  }
}

console.log(failures ? `\n${failures} check(s) failed` : '\nall checks passed');
process.exit(failures ? 1 : 0);
