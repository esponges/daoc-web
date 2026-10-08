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
import { Skeleton, skinVertex, apply, Clip, rowsToQuat, quatToRows, makePose, xform } from './skeleton.js';
const here = dirname(fileURLToPath(import.meta.url));
// Which converted character to check; pass a name to test another outfit.
const who = process.argv[2] || 'norseman';
const base = join(here, 'data', 'char', who);

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
  // "Stays a figure" is a matter of structure, not height. A height band is
  // the Norseman's alone: a wolf's head is lower, and a badger's run is a
  // bound that lifts its whole body seventeen units mid-leap. What no sound
  // clip does is fold the skeleton, so measure the head against the pelvis,
  // relative to the character's own idle, and keep the head above the feet.
  const pelvis = skel.boneId('Bip01 Pelvis');
  const reach = () => Math.hypot(skel.world[head][3] - skel.world[pelvis][3],
    skel.world[head][7] - skel.world[pelvis][7], skel.world[head][11] - skel.world[pelvis][11]);
  skel.poseClip(clips.idle || clips[names[0]], 0);
  const reachRef = reach();
  let worstReach = 0, badFrames = 0, minFoot = Infinity, inverted = 0;
  for (const n of names) {
    const c = clips[n];
    for (let k = 0; k <= 48; k++) {
      skel.poseClip(c, (k / 48) * c.duration);
      for (const m of skel.world) {
        for (const v of m) if (!Number.isFinite(v)) badFrames++;
      }
      worstReach = Math.max(worstReach, Math.abs(reach() / reachRef - 1));
      const footTop = Math.max(boneHeight(lFoot), boneHeight(rFoot));
      // A death is supposed to end lying down; only the loops must stand.
      if (c.loop && boneHeight(head) <= footTop) inverted++;
      minFoot = Math.min(minFoot, boneHeight(lFoot), boneHeight(rFoot));
    }
  }
  check('clips never produce a non-finite transform', badFrames === 0, badFrames + ' bad values');
  check('head-to-pelvis distance stays within a quarter of idle\'s',
    worstReach < 0.25, 'worst ' + (100 * worstReach).toFixed(1) + '% of ' + reachRef.toFixed(1) + 'u');
  check('the head never drops to the feet in a looping clip', inverted === 0, inverted + ' frames');
  if (clips.death) {
    skel.poseClip(clips.idle || clips[names[0]], 0);
    const standing = boneHeight(head);
    skel.poseClip(clips.death, clips.death.duration);
    const lying = boneHeight(head);
    check('a death ends with the head down near the ground', lying < 0.5 * standing,
      'head ' + standing.toFixed(1) + 'u standing, ' + lying.toFixed(1) + 'u at the end of death');
  }
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


// --- blending -------------------------------------------------------------
//
// Blending decomposes each pose into quaternion/translation/scale, so the
// decomposition has to be lossless before any blend can be trusted.
//
// It is not quite lossless, and the reason is in the data rather than the
// code: the bind matrices DAoC ships are not exactly orthonormal. The worst
// bone's rows are 0.9986 long instead of 1, and two of them are 1.5e-4 off
// perpendicular. A quaternion can only represent a true rotation, so the round
// trip returns a cleaned-up matrix that differs from the original by exactly
// that much. The check below allows for the input's own error and no more.

{
  let worstTrip = 0, worstSkew = 0, tripBone = '', skewBone = '';
  const m = xform();
  const q = new Float32Array(4);
  for (let i = 0; i < skel.bones.length; i++) {
    const b = skel.bindLocal[i], s = skel.bindScale[i];
    // How far this matrix is from being a rotation times a uniform scale.
    const row = (r) => [b[r * 4], b[r * 4 + 1], b[r * 4 + 2]];
    const len = (r) => Math.hypot(...row(r));
    const dot = (a, c) => row(a).reduce((t, v, k) => t + v * row(c)[k], 0) / (len(a) * len(c) || 1);
    const skew = Math.max(
      Math.abs(len(0) - s), Math.abs(len(1) - s), Math.abs(len(2) - s),
      Math.abs(dot(0, 1)), Math.abs(dot(0, 2)), Math.abs(dot(1, 2)));
    if (skew > worstSkew) { worstSkew = skew; skewBone = skel.bones[i].name; }

    rowsToQuat(b, s, q, 0);
    quatToRows(q, s, m);
    for (const k of [0, 1, 2, 4, 5, 6, 8, 9, 10]) {
      const d = Math.abs(m[k] - b[k]);
      if (d > worstTrip) { worstTrip = d; tripBone = skel.bones[i].name; }
    }
  }
  check('the shipped bind matrices are nearly, not exactly, orthonormal',
    worstSkew > 0, 'worst skew ' + worstSkew.toExponential(2) + ' on ' + skewBone);
  check('the quaternion round trip loses no more than that skew',
    worstTrip <= worstSkew * 1.5 + 1e-6,
    'round trip ' + worstTrip.toExponential(2) + ' on ' + tripBone +
    ' vs skew ' + worstSkew.toExponential(2));
}

if (clips && clips.walk && clips.idle) {
  const { walk, idle } = clips;
  const snapshot = () => skel.world.map((x) => Array.from(x));
  const maxDiff = (a, b) => {
    let d = 0;
    for (let i = 0; i < a.length; i++) {
      for (let k = 0; k < 12; k++) d = Math.max(d, Math.abs(a[i][k] - b[i][k]));
    }
    return d;
  };

  // A cross-fade at its endpoints must be exactly the clips it fades between,
  // or every transition starts and ends with a jump.
  skel.poseClip(walk, 0.4);
  const direct = snapshot();
  skel.poseCross(walk, 0.4, idle, 1.0, 0);
  check('a cross-fade at f=0 is exactly the outgoing clip',
    maxDiff(direct, snapshot()) < 1e-3, maxDiff(direct, snapshot()).toExponential(2));

  skel.poseClip(idle, 1.0);
  const toIdle = snapshot();
  skel.poseCross(walk, 0.4, idle, 1.0, 1);
  check('a cross-fade at f=1 is exactly the incoming clip',
    maxDiff(toIdle, snapshot()) < 1e-3, maxDiff(toIdle, snapshot()).toExponential(2));

  const head = skel.boneId('Bip01 Head');
  const boneLen = (name, parentName) => {
    const i = skel.boneId(name), p = skel.boneId(parentName);
    return Math.hypot(skel.world[i][3] - skel.world[p][3],
                      skel.world[i][7] - skel.world[p][7],
                      skel.world[i][11] - skel.world[p][11]);
  };
  // Baseline from the same pipeline, so the comparison is not against the
  // un-normalised bind matrices the check above just characterised.
  skel.poseCross(null, 0, null, 0, 0);
  // The foot's own parent: the calf on a biped, the HorseLink on a
  // quadruped, whose hock sits between the two and bends.
  const footParent = man.bones[man.bones[skel.boneId('Bip01 L Foot')].parent ?? -1]?.name || 'Bip01 L Calf';
  const bindShin = boneLen('Bip01 L Foot', footParent);

  skel.poseClip(idle, 0);
  const standHead = skel.world[head][11];
  let bad = 0, hiHead = -Infinity, loHead = Infinity, worstShin = 0;
  for (let s = 0; s <= 10; s++) {
    for (let k = 0; k < 8; k++) {
      skel.poseCross(walk, (k / 8) * walk.duration, idle, (k / 8) * idle.duration, s / 10);
      for (const m of skel.world) for (const v of m) if (!Number.isFinite(v)) bad++;
      hiHead = Math.max(hiHead, skel.world[head][11]);
      loHead = Math.min(loHead, skel.world[head][11]);
      worstShin = Math.max(worstShin, Math.abs(boneLen('Bip01 L Foot', footParent) - bindShin));
    }
  }
  check('blended poses are finite', bad === 0, bad + ' bad values');
  check('the figure stays upright through a blend',
    loHead > 0.75 * standHead && hiHead < 1.25 * standHead, 'head ' + loHead.toFixed(1) + '..' + hiHead.toFixed(1) + 'u');
  check('bones keep their length mid-blend',
    worstShin < 0.02, 'shin drifts ' + worstShin.toExponential(2) + 'u of ' + bindShin.toFixed(1) + 'u');

  // The reason the decomposition exists. Lerping two rotation matrices
  // componentwise does not give a rotation: it shortens as it goes, so a limb
  // visibly contracts halfway through a transition. Measure both ways on the
  // same pair of poses and show the gap.
  {
    const lCalf = skel.boneId('Bip01 L Calf');
    skel.poseClip(walk, 0.4);
    const A = Array.from(skel.animLocal[lCalf]);
    skel.poseClip(idle, 1.0);
    const B = Array.from(skel.animLocal[lCalf]);
    const rowLen = (m, r) => Math.hypot(m[r * 4], m[r * 4 + 1], m[r * 4 + 2]);
    let worstNaive = 0;
    for (let s = 1; s < 10; s++) {
      const f = s / 10;
      const L = A.map((v, i) => v + (B[i] - v) * f);
      worstNaive = Math.max(worstNaive, Math.abs(rowLen(L, 0) - 1), Math.abs(rowLen(L, 1) - 1));
    }
    skel.poseCross(walk, 0.4, idle, 1.0, 0.5);
    const ours = Math.max(Math.abs(rowLen(skel.animLocal[lCalf], 0) - 1),
                          Math.abs(rowLen(skel.animLocal[lCalf], 1) - 1));
    check('slerping beats lerping the matrices outright',
      ours < worstNaive * 0.1,
      'matrix lerp shrinks a bone by up to ' + (worstNaive * 100).toFixed(1) +
      '%, slerp by ' + (ours * 100).toFixed(3) + '%');
  }

  // Halfway between two clips should sit between them, not outside.
  skel.poseCross(walk, 0.4, null, 0, 0);
  const a = skel.world[skel.boneId('Bip01 L Thigh')][7];
  skel.poseCross(idle, 1.0, null, 0, 0);
  const b = skel.world[skel.boneId('Bip01 L Thigh')][7];
  skel.poseCross(walk, 0.4, idle, 1.0, 0.5);
  const mid = skel.world[skel.boneId('Bip01 L Thigh')][7];
  check('a half blend lands between the two poses',
    mid >= Math.min(a, b) - 0.5 && mid <= Math.max(a, b) + 0.5,
    a.toFixed(2) + ' .. ' + mid.toFixed(2) + ' .. ' + b.toFixed(2));
}

console.log(failures ? `\n${failures} check(s) failed` : '\nall checks passed');
process.exit(failures ? 1 : 0);
