// Skeleton posing for DAoC character models. No WebGL here on purpose: this
// is the part most likely to be subtly wrong, and keeping it free of the
// graphics context means it can be checked headlessly (see skeleton.test.mjs).
//
// A bone's local frame is whatever the artist left it as, so joints are not
// rotated in their own axes. A rotation is given in the body's frame and
// conjugated into the joint's:
//
//     R_local = transpose(W) * R_body * W
//
// where W is the joint's bind world rotation. Setting the joint's local
// transform to bindLocal * R_local makes its world rotation R_body * W, so
// "swing the thigh forward" means the same thing at every joint however it
// happens to be oriented. The model is Z-up with X lateral, so a limb swings
// about X and the torso leans about Z.

// A rigid transform is three rows of four: [r00 r01 r02 tx  r10 .. ty  r20 .. tz].
export function xform() {
  const m = new Float32Array(12);
  m[0] = m[5] = m[10] = 1;
  return m;
}

// fromBone builds a transform from the exported translation, 3x3 rotation and
// uniform scale, folding the scale into the rotation rows.
export function fromBone(t, r, s, out = xform()) {
  for (let row = 0; row < 3; row++) {
    out[row * 4 + 0] = r[row * 3 + 0] * s;
    out[row * 4 + 1] = r[row * 3 + 1] * s;
    out[row * 4 + 2] = r[row * 3 + 2] * s;
    out[row * 4 + 3] = t[row];
  }
  return out;
}

// compose writes a * b, i.e. apply b then a.
export function compose(a, b, out) {
  for (let r = 0; r < 3; r++) {
    const a0 = a[r * 4 + 0], a1 = a[r * 4 + 1], a2 = a[r * 4 + 2];
    out[r * 4 + 0] = a0 * b[0] + a1 * b[4] + a2 * b[8];
    out[r * 4 + 1] = a0 * b[1] + a1 * b[5] + a2 * b[9];
    out[r * 4 + 2] = a0 * b[2] + a1 * b[6] + a2 * b[10];
    out[r * 4 + 3] = a0 * b[3] + a1 * b[7] + a2 * b[11] + a[r * 4 + 3];
  }
  return out;
}

// apply maps a point through a transform.
export function apply(m, x, y, z, out = [0, 0, 0]) {
  out[0] = m[0] * x + m[1] * y + m[2] * z + m[3];
  out[1] = m[4] * x + m[5] * y + m[6] * z + m[7];
  out[2] = m[8] * x + m[9] * y + m[10] * z + m[11];
  return out;
}

export function rotX(angle, out) {
  const c = Math.cos(angle), s = Math.sin(angle);
  out.fill(0);
  out[0] = 1; out[5] = c; out[6] = -s; out[9] = s; out[10] = c;
  return out;
}

export function rotY(angle, out) {
  const c = Math.cos(angle), s = Math.sin(angle);
  out.fill(0);
  out[0] = c; out[2] = s; out[5] = 1; out[8] = -s; out[10] = c;
  return out;
}

export function rotZ(angle, out) {
  const c = Math.cos(angle), s = Math.sin(angle);
  out.fill(0);
  out[0] = c; out[1] = -s; out[4] = s; out[5] = c; out[10] = 1;
  return out;
}

// rotOnly copies just the rotation of a transform, dropping the translation.
// Conjugating a rotation needs this: composing with the full transform would
// drag the joint's own position into what must stay a pure rotation, which
// detaches the limb from the joint.
export function rotOnly(a, out) {
  for (let r = 0; r < 3; r++) {
    out[r * 4 + 0] = a[r * 4 + 0];
    out[r * 4 + 1] = a[r * 4 + 1];
    out[r * 4 + 2] = a[r * 4 + 2];
    out[r * 4 + 3] = 0;
  }
  return out;
}

// rotTransposeMul writes transpose(rot of a) * b, ignoring translations.
export function rotTransposeMul(a, b, out) {
  for (let r = 0; r < 3; r++) {
    for (let c = 0; c < 3; c++) {
      out[r * 4 + c] =
        a[0 * 4 + r] * b[0 * 4 + c] +
        a[1 * 4 + r] * b[1 * 4 + c] +
        a[2 * 4 + r] * b[2 * 4 + c];
    }
    out[r * 4 + 3] = 0;
  }
  return out;
}

// The joints a gait drives. Everything else rides along through the hierarchy.
// How far the arms come down from the T-pose, and the resting elbow bend.
const ARM_DROP = 1.32;
const ARM_SWING = 0.42;
const ELBOW = 0.34;

const JOINT_NAMES = {
  lThigh: 'Bip01 L Thigh', rThigh: 'Bip01 R Thigh',
  lCalf: 'Bip01 L Calf', rCalf: 'Bip01 R Calf',
  lFoot: 'Bip01 L Foot', rFoot: 'Bip01 R Foot',
  lArm: 'Bip01 L UpperArm', rArm: 'Bip01 R UpperArm',
  lFore: 'Bip01 L Forearm', rFore: 'Bip01 R Forearm',
  spine: 'Bip01 Spine',
};

export class Skeleton {
  constructor(bones) {
    this.bones = bones;
    const n = bones.length;
    this.bindLocal = [];
    this.bindWorld = [];
    this.animLocal = [];
    this.world = [];
    // The scale is folded into the rotation rows, so a recorded clip that
    // rewrites the rotation would silently drop it. Keep it to put back.
    this.bindScale = new Float32Array(n);
    // The bind pose as rotation/translation/scale rather than matrices.
    // Blending has to interpolate rotations as quaternions -- averaging two
    // rotation matrices does not give a rotation -- and a bone with no track
    // in either clip still has to contribute its bind value, so the bind pose
    // is kept in the same form a sampled pose uses.
    this.bindQuat = new Float32Array(n * 4);
    this.bindTrans = new Float32Array(n * 3);
    for (let i = 0; i < n; i++) {
      this.bindLocal.push(fromBone(bones[i].t, bones[i].r, bones[i].s));
      this.bindWorld.push(xform());
      this.animLocal.push(xform());
      this.world.push(xform());
      this.bindScale[i] = bones[i].s || 1;
      rowsToQuat(this.bindLocal[i], this.bindScale[i], this.bindQuat, i * 4);
      this.bindTrans[i * 3 + 0] = this.bindLocal[i][3];
      this.bindTrans[i * 3 + 1] = this.bindLocal[i][7];
      this.bindTrans[i * 3 + 2] = this.bindLocal[i][11];
    }

    // Parents are not guaranteed to precede children, so resolve the order
    // depth-first from the roots rather than assuming.
    const children = Array.from({ length: n }, () => []);
    const roots = [];
    for (let i = 0; i < n; i++) {
      if (bones[i].parent >= 0) children[bones[i].parent].push(i);
      else roots.push(i);
    }
    this.order = [];
    const stack = roots.slice().reverse();
    while (stack.length) {
      const i = stack.pop();
      this.order.push(i);
      for (let k = children[i].length - 1; k >= 0; k--) stack.push(children[i][k]);
    }
    if (this.order.length !== n) {
      throw new Error(
        'skeleton: ' + (n - this.order.length) + ' of ' + n + ' bones are unreachable from a root');
    }

    for (const i of this.order) {
      const p = bones[i].parent;
      if (p < 0) this.bindWorld[i].set(this.bindLocal[i]);
      else compose(this.bindWorld[p], this.bindLocal[i], this.bindWorld[i]);
    }

    this.byName = new Map();
    for (let i = 0; i < n; i++) this.byName.set(bones[i].name.toLowerCase(), i);
    this.joints = {};
    this.missingJoints = [];
    for (const [key, name] of Object.entries(JOINT_NAMES)) {
      const id = this.boneId(name);
      this.joints[key] = id;
      if (id < 0) this.missingJoints.push(name);
    }

    this._a = xform();
    this._r1 = xform();
    this._r2 = xform();
    this._r3 = xform();
    this._d = xform();
    this._b = xform();
    this._c = xform();
    // Which way each arm points along the lateral axis decides the sign of the
    // drop: rotating both the same way would raise one and lower the other.
    const side = (j) => (j >= 0 && this.bindWorld[j][3] < 0 ? -1 : 1);
    this.armSide = { l: side(this.joints.lArm), r: side(this.joints.rArm) };

    this.poseBind();
  }

  boneId(name) {
    const i = this.byName.get(String(name).toLowerCase());
    return i === undefined ? -1 : i;
  }

  // setJoint rotates one joint by an angle given in the body's frame.
  //
  // Both sides of the conjugation use the rotation of the bind world transform
  // only. Using the full transform would leave a translation in R_local and
  // the limb would fly off its joint.
  setJoint(i, angle, axis = 'x') {
    if (i < 0 || !angle) return;
    const r = axis === 'z' ? rotZ : axis === 'y' ? rotY : rotX;
    this.setJointRot(i, r(angle, this._a));
  }

  // setJointRot applies an arbitrary body-frame rotation to one joint.
  setJointRot(i, R) {
    if (i < 0) return;
    rotOnly(this.bindWorld[i], this._d);                  // Wr
    rotTransposeMul(this._d, R, this._b);                 // Wr' * R
    compose(this._b, this._d, this._c);                   // Wr' * R * Wr
    compose(this.bindLocal[i], this._c, this.animLocal[i]);
  }

  // resolve accumulates animLocal down the hierarchy into world.
  resolve() {
    for (const i of this.order) {
      const p = this.bones[i].parent;
      if (p < 0) this.world[i].set(this.animLocal[i]);
      else compose(this.world[p], this.animLocal[i], this.world[i]);
    }
  }

  poseBind() {
    for (let i = 0; i < this.animLocal.length; i++) {
      this.animLocal[i].set(this.bindLocal[i]);
    }
    this.resolve();
  }

  // poseWalk is a gait, not a recording: phase advances with ground covered so
  // the feet keep pace at any speed, and intensity scales the throw so the
  // same cycle serves a walk and a sprint.
  poseWalk(phase, intensity) {
    for (let i = 0; i < this.animLocal.length; i++) {
      this.animLocal[i].set(this.bindLocal[i]);
    }
    const J = this.joints;
    const s = Math.sin(phase), c = Math.cos(phase);

    // The model is authored in a T-pose, arms straight out along the lateral
    // axis. Swinging them fore and aft is a rotation about that same axis,
    // which does nothing to an arm lying along it, so the arms have to come
    // down to the sides first. This happens even when standing still: a
    // T-posed idle is not an idle.
    const drop = ARM_DROP;
    const swingL = ARM_SWING * intensity * s;
    const swingR = -ARM_SWING * intensity * s;
    rotY(drop * this.armSide.l, this._r1);
    rotX(swingL, this._r2);
    compose(this._r2, this._r1, this._r3);
    this.setJointRot(J.lArm, this._r3);
    rotY(drop * this.armSide.r, this._r1);
    rotX(swingR, this._r2);
    compose(this._r2, this._r1, this._r3);
    this.setJointRot(J.rArm, this._r3);
    // A relaxed elbow, deepening slightly as the arm swings through.
    this.setJoint(J.lFore, -ELBOW - 0.14 * intensity * c);
    this.setJoint(J.rFore, -ELBOW + 0.14 * intensity * c);

    if (intensity > 0.001) {
      const hip = 0.62 * intensity;
      const knee = 0.95 * intensity;
      this.setJoint(J.lThigh, -hip * s);
      this.setJoint(J.rThigh, hip * s);
      // A knee only folds one way, so the bend is clamped to one half-cycle.
      this.setJoint(J.lCalf, knee * Math.max(0, -Math.sin(phase - 0.6)));
      this.setJoint(J.rCalf, knee * Math.max(0, -Math.sin(phase + Math.PI - 0.6)));
      this.setJoint(J.lFoot, -0.35 * intensity * Math.max(0, -s));
      this.setJoint(J.rFoot, -0.35 * intensity * Math.max(0, s));
      this.setJoint(J.spine, 0.055 * intensity * c, 'z');
    }
    this.resolve();
  }

  // skinMatrix writes world[bone] * invBind, which is what the shader wants.
  skinMatrix(bone, invBind, out) {
    return compose(this.world[bone], invBind, out);
  }
}

// skinVertex computes a bind-or-posed vertex position on the CPU. Used by the
// tests to check the shader's arithmetic without a GL context.
export function skinVertex(skel, shape, pos, boneIdx, weights, out = [0, 0, 0]) {
  out[0] = out[1] = out[2] = 0;
  const m = xform();
  for (let k = 0; k < 4; k++) {
    const w = weights[k];
    if (w <= 0) continue;
    const local = boneIdx[k];
    skel.skinMatrix(shape.bones[local], shape.invBind[local], m);
    out[0] += w * (m[0] * pos[0] + m[1] * pos[1] + m[2] * pos[2] + m[3]);
    out[1] += w * (m[4] * pos[0] + m[5] * pos[1] + m[6] * pos[2] + m[7]);
    out[2] += w * (m[8] * pos[0] + m[9] * pos[1] + m[10] * pos[2] + m[11]);
  }
  return out;
}

// --- recorded animation ---------------------------------------------------
//
// A clip from DAoC's own .kfa files. Unlike poseWalk, which rotates joints by
// deltas conjugated into each joint's frame, a recorded track supplies the
// joint's local transform outright: the rotation replaces the bind rotation
// rather than composing with it. Bones with no track keep their bind pose, and
// a channel with no keys keeps that component of it -- a track may rotate a
// bone without moving it.

// quatToRows writes a w,x,y,z quaternion into the rotation part of a
// transform, scaled, leaving the translation alone.
export function quatToRows(q, scale, out) {
  const w = q[0], x = q[1], y = q[2], z = q[3];
  const xx = x * x, yy = y * y, zz = z * z;
  const xy = x * y, xz = x * z, yz = y * z;
  const wx = w * x, wy = w * y, wz = w * z;
  out[0] = (1 - 2 * (yy + zz)) * scale;
  out[1] = (2 * (xy - wz)) * scale;
  out[2] = (2 * (xz + wy)) * scale;
  out[4] = (2 * (xy + wz)) * scale;
  out[5] = (1 - 2 * (xx + zz)) * scale;
  out[6] = (2 * (yz - wx)) * scale;
  out[8] = (2 * (xz - wy)) * scale;
  out[9] = (2 * (yz + wx)) * scale;
  out[10] = (1 - 2 * (xx + yy)) * scale;
  return out;
}

// span finds the key interval containing t and the blend factor within it.
// Keys are [time, ...] rows in ascending time.
function span(keys, t) {
  const n = keys.length;
  if (n === 0) return null;
  if (n === 1 || t <= keys[0][0]) return { a: 0, b: 0, f: 0 };
  if (t >= keys[n - 1][0]) return { a: n - 1, b: n - 1, f: 0 };
  let lo = 0, hi = n - 1;
  while (hi - lo > 1) {
    const mid = (lo + hi) >> 1;
    if (keys[mid][0] <= t) lo = mid; else hi = mid;
  }
  const t0 = keys[lo][0], t1 = keys[hi][0];
  const d = t1 - t0;
  return { a: lo, b: hi, f: d > 1e-9 ? (t - t0) / d : 0 };
}

// sampleQuat slerps between the two keys bracketing t. The shorter arc is
// always taken: a quaternion and its negation are the same rotation, so a
// negative dot means the raw pair would spin the long way round.
function sampleQuat(keys, t, out) {
  const s = span(keys, t);
  if (!s) return null;
  const A = keys[s.a], B = keys[s.b];
  let w0 = A[1], x0 = A[2], y0 = A[3], z0 = A[4];
  let w1 = B[1], x1 = B[2], y1 = B[3], z1 = B[4];
  let dot = w0 * w1 + x0 * x1 + y0 * y1 + z0 * z1;
  if (dot < 0) { w1 = -w1; x1 = -x1; y1 = -y1; z1 = -z1; dot = -dot; }
  let s0 = 1 - s.f, s1 = s.f;
  if (dot < 0.9995) {
    const theta = Math.acos(Math.min(1, dot));
    const sin = Math.sin(theta);
    if (sin > 1e-6) {
      s0 = Math.sin((1 - s.f) * theta) / sin;
      s1 = Math.sin(s.f * theta) / sin;
    }
  }
  out[0] = s0 * w0 + s1 * w1;
  out[1] = s0 * x0 + s1 * x1;
  out[2] = s0 * y0 + s1 * y1;
  out[3] = s0 * z0 + s1 * z1;
  const len = Math.hypot(out[0], out[1], out[2], out[3]) || 1;
  out[0] /= len; out[1] /= len; out[2] /= len; out[3] /= len;
  return out;
}

function sampleVec(keys, t, out) {
  const s = span(keys, t);
  if (!s) return null;
  const A = keys[s.a], B = keys[s.b];
  for (let k = 0; k < 3; k++) out[k] = A[k + 1] + (B[k + 1] - A[k + 1]) * s.f;
  return out;
}

function sampleFloat(keys, t) {
  const s = span(keys, t);
  if (!s) return null;
  return keys[s.a][1] + (keys[s.b][1] - keys[s.a][1]) * s.f;
}

// Clip binds a converted animation to a particular skeleton, resolving each
// track's bone name to an index once rather than per frame.
export class Clip {
  constructor(json, skeleton) {
    this.name = json.name;
    this.source = json.source;
    this.duration = json.duration || 0;
    // A clip that does not loop -- an attack, a flinch, a death -- plays
    // once and holds its last frame rather than starting over.
    this.loop = json.loop !== false;
    // How fast the game plays it relative to its keys (animnifs.csv fps
    // over base fps): 1 for most, far less for the idles.
    this.rate = json.rate > 0 ? json.rate : 1;
    this.tracks = [];
    this.missing = [];
    for (const t of json.tracks || []) {
      const i = skeleton.boneId(t.bone);
      if (i < 0) { this.missing.push(t.bone); continue; }
      this.tracks.push({
        bone: i, name: t.bone,
        rot: t.rot || [], trans: t.trans || [], scale: t.scale || [],
      });
    }
  }
}

Skeleton.prototype.poseClip = function (clip, time) {
  const n = this.animLocal.length;
  for (let i = 0; i < n; i++) this.animLocal[i].set(this.bindLocal[i]);
  if (!clip || !clip.tracks.length) { this.resolve(); return; }

  // Looping clips wrap, and a negative time is as valid as a large one;
  // the rest clamp to their ends.
  const d = clip.duration;
  let t = time;
  if (d > 0) t = clip.loop ? ((time % d) + d) % d : Math.max(0, Math.min(d, time));

  const q = this._q || (this._q = new Float32Array(4));
  const v = this._v || (this._v = new Float32Array(3));
  for (const tr of clip.tracks) {
    const m = this.animLocal[tr.bone];
    if (tr.rot.length && sampleQuat(tr.rot, t, q)) {
      // Scale lives folded into the rotation rows, so rewriting the rotation
      // would drop it. Take the clip's scale if it has one, else the bind
      // scale this bone already carried.
      let s = tr.scale.length ? sampleFloat(tr.scale, t) : null;
      if (s === null) s = this.bindScale[tr.bone];
      quatToRows(q, s, m);
    }
    if (tr.trans.length && sampleVec(tr.trans, t, v)) {
      m[3] = v[0]; m[7] = v[1]; m[11] = v[2];
    }
  }
  this.resolve();
};

// --- blending -------------------------------------------------------------
//
// Switching clips outright snaps the figure: a foot forward in one cycle is a
// foot back in the next. Cross-fading means interpolating two poses, and that
// has to happen on rotations rather than on the matrices poseClip writes --
// the average of two rotation matrices is not a rotation, and blending them
// shears the limb instead of turning it.
//
// So a pose here is the same decomposition the file itself uses: a quaternion,
// a translation and a scale per bone, defaulting to the bind value for any
// bone a clip does not drive. Poses blend cleanly; matrices are built once at
// the end.

// rowsToQuat extracts a quaternion from the rotation part of a transform whose
// rows carry a uniform scale. Shepperd's method: pick the largest diagonal
// term so the square root never divides by something near zero.
export function rowsToQuat(m, scale, out, o = 0) {
  const s = scale || 1;
  const m00 = m[0] / s, m01 = m[1] / s, m02 = m[2] / s;
  const m10 = m[4] / s, m11 = m[5] / s, m12 = m[6] / s;
  const m20 = m[8] / s, m21 = m[9] / s, m22 = m[10] / s;
  const tr = m00 + m11 + m22;
  let w, x, y, z;
  if (tr > 0) {
    const k = Math.sqrt(tr + 1) * 2;
    w = 0.25 * k; x = (m21 - m12) / k; y = (m02 - m20) / k; z = (m10 - m01) / k;
  } else if (m00 > m11 && m00 > m22) {
    const k = Math.sqrt(1 + m00 - m11 - m22) * 2;
    w = (m21 - m12) / k; x = 0.25 * k; y = (m01 + m10) / k; z = (m02 + m20) / k;
  } else if (m11 > m22) {
    const k = Math.sqrt(1 + m11 - m00 - m22) * 2;
    w = (m02 - m20) / k; x = (m01 + m10) / k; y = 0.25 * k; z = (m12 + m21) / k;
  } else {
    const k = Math.sqrt(1 + m22 - m00 - m11) * 2;
    w = (m10 - m01) / k; x = (m02 + m20) / k; y = (m12 + m21) / k; z = 0.25 * k;
  }
  const len = Math.hypot(w, x, y, z) || 1;
  out[o] = w / len; out[o + 1] = x / len; out[o + 2] = y / len; out[o + 3] = z / len;
  return out;
}

// makePose allocates a pose buffer for a skeleton.
export function makePose(n) {
  return {
    rot: new Float32Array(n * 4),
    trans: new Float32Array(n * 3),
    scale: new Float32Array(n),
  };
}

// slerpQuat blends two quaternions along the shorter arc, reading and writing
// into flat arrays at the given offsets.
export function slerpQuat(a, ao, b, bo, f, out, oo) {
  let w0 = a[ao], x0 = a[ao + 1], y0 = a[ao + 2], z0 = a[ao + 3];
  let w1 = b[bo], x1 = b[bo + 1], y1 = b[bo + 2], z1 = b[bo + 3];
  let dot = w0 * w1 + x0 * x1 + y0 * y1 + z0 * z1;
  if (dot < 0) { w1 = -w1; x1 = -x1; y1 = -y1; z1 = -z1; dot = -dot; }
  let s0 = 1 - f, s1 = f;
  if (dot < 0.9995) {
    const theta = Math.acos(Math.min(1, dot));
    const sin = Math.sin(theta);
    if (sin > 1e-6) {
      s0 = Math.sin((1 - f) * theta) / sin;
      s1 = Math.sin(f * theta) / sin;
    }
  }
  let w = s0 * w0 + s1 * w1, x = s0 * x0 + s1 * x1;
  let y = s0 * y0 + s1 * y1, z = s0 * z0 + s1 * z1;
  const len = Math.hypot(w, x, y, z) || 1;
  out[oo] = w / len; out[oo + 1] = x / len; out[oo + 2] = y / len; out[oo + 3] = z / len;
  return out;
}

// samplePose fills a pose from a clip at time t, starting from the bind pose
// so bones the clip does not drive still have a value to blend.
Skeleton.prototype.samplePose = function (clip, time, pose) {
  pose.rot.set(this.bindQuat);
  pose.trans.set(this.bindTrans);
  pose.scale.set(this.bindScale);
  if (!clip || !clip.tracks.length) return pose;

  const d = clip.duration;
  let t = time;
  if (d > 0) t = clip.loop ? ((time % d) + d) % d : Math.max(0, Math.min(d, time));

  const q = this._q || (this._q = new Float32Array(4));
  const v = this._v || (this._v = new Float32Array(3));
  for (const tr of clip.tracks) {
    const i = tr.bone;
    if (tr.rot.length && sampleQuat(tr.rot, t, q)) {
      pose.rot[i * 4] = q[0]; pose.rot[i * 4 + 1] = q[1];
      pose.rot[i * 4 + 2] = q[2]; pose.rot[i * 4 + 3] = q[3];
    }
    if (tr.trans.length && sampleVec(tr.trans, t, v)) {
      pose.trans[i * 3] = v[0]; pose.trans[i * 3 + 1] = v[1]; pose.trans[i * 3 + 2] = v[2];
    }
    if (tr.scale.length) {
      const s = sampleFloat(tr.scale, t);
      if (s !== null) pose.scale[i] = s;
    }
  }
  return pose;
};

// blendPose writes (1-f)*a + f*b, slerping rotations and lerping the rest.
Skeleton.prototype.blendPose = function (a, b, f, out) {
  const n = this.animLocal.length;
  if (f <= 0) { out.rot.set(a.rot); out.trans.set(a.trans); out.scale.set(a.scale); return out; }
  if (f >= 1) { out.rot.set(b.rot); out.trans.set(b.trans); out.scale.set(b.scale); return out; }
  for (let i = 0; i < n; i++) {
    slerpQuat(a.rot, i * 4, b.rot, i * 4, f, out.rot, i * 4);
    for (let k = 0; k < 3; k++) {
      const j = i * 3 + k;
      out.trans[j] = a.trans[j] + (b.trans[j] - a.trans[j]) * f;
    }
    out.scale[i] = a.scale[i] + (b.scale[i] - a.scale[i]) * f;
  }
  return out;
};

// applyPose turns a pose into local matrices and resolves the hierarchy.
Skeleton.prototype.applyPose = function (pose) {
  const n = this.animLocal.length;
  const q = this._q4 || (this._q4 = new Float32Array(4));
  for (let i = 0; i < n; i++) {
    const m = this.animLocal[i];
    q[0] = pose.rot[i * 4]; q[1] = pose.rot[i * 4 + 1];
    q[2] = pose.rot[i * 4 + 2]; q[3] = pose.rot[i * 4 + 3];
    quatToRows(q, pose.scale[i], m);
    m[3] = pose.trans[i * 3];
    m[7] = pose.trans[i * 3 + 1];
    m[11] = pose.trans[i * 3 + 2];
  }
  this.resolve();
};

// poseCross is the whole cross-fade in one call: sample both clips, blend,
// apply. f is how far across, 0 meaning entirely clip a.
Skeleton.prototype.poseCross = function (a, ta, b, tb, f) {
  const n = this.animLocal.length;
  this._poseA = this._poseA || makePose(n);
  this._poseB = this._poseB || makePose(n);
  this._poseOut = this._poseOut || makePose(n);
  if (!b || f <= 0) {
    this.applyPose(this.samplePose(a, ta, this._poseA));
    return;
  }
  if (!a || f >= 1) {
    this.applyPose(this.samplePose(b, tb, this._poseB));
    return;
  }
  this.samplePose(a, ta, this._poseA);
  this.samplePose(b, tb, this._poseB);
  this.applyPose(this.blendPose(this._poseA, this._poseB, f, this._poseOut));
};

// groundSpeed reads how fast a locomotion clip travels, in model units per
// second, from the clip alone.
//
// A planted foot does not move against the ground, so relative to the body
// it slides backward at exactly the speed the body moves forward. Take the
// frames where a foot sits at its lowest -- within half a unit of it, which
// is the flat stretch of the stance -- and the median of its horizontal
// speed there is the clip's own ground speed. Move a figure at that speed
// with the clip at 1x and its feet stay put.
//
// The game's tables carry "stride" numbers for this, but they do not agree
// with the clips: the wolf's walk is listed at 52 and its feet say 40.
// Returns 0 when no foot ever plants, which is the honest answer for a clip
// that is not a gait.
Skeleton.prototype.groundSpeed = function (clip) {
  const N = 360, dt = clip.duration / N;
  const speeds = [];
  for (const side of ['L', 'R']) {
    const b = this.boneId('Bip01 ' + side + ' Foot');
    if (b < 0 || !(clip.duration > 0)) continue;
    const xs = [], ys = [], zs = [];
    for (let i = 0; i <= N; i++) {
      this.poseClip(clip, Math.min(i * dt, clip.duration - 1e-4));
      const m = this.world[b];
      xs.push(m[3]); ys.push(m[7]); zs.push(m[11]);
    }
    const low = Math.min(...zs) + 0.5;
    const v = [];
    for (let i = 1; i <= N; i++) {
      if (zs[i] < low && zs[i - 1] < low) {
        v.push(Math.hypot(xs[i] - xs[i - 1], ys[i] - ys[i - 1]) / dt);
      }
    }
    // A foot that touches down for a frame or two is not a stance.
    if (v.length >= N / 20) {
      v.sort((p, q) => p - q);
      speeds.push(v[v.length >> 1]);
    }
  }
  return speeds.length ? speeds.reduce((s, x) => s + x, 0) / speeds.length : 0;
};

// --- layering ---------------------------------------------------------------
//
// A cross-fade swaps the whole body from one clip to another, which is right
// for idle to walk but wrong for swinging while running: the legs would stop
// to throw the punch. A layer plays one clip over another with a per-bone
// weight, so an attack can own the spine, arms and head while the run keeps
// the pelvis and legs.

// upperBodyMask is 1 for the torso, arms and head, 0 for the root, pelvis
// and legs, which stay with the base.
//
// Where the torso starts depends on the body. A humanoid Biped hangs its
// thighs off Bip01 Spine, not the pelvis, so taking Spine would take the
// legs with it; the upper body there is Spine1 and up. A quadruped's thighs
// hang off the pelvis. Either way, no thigh's subtree is ever included.
Skeleton.prototype.upperBodyMask = function () {
  if (this._upper) return this._upper;
  const n = this.bones.length, mask = new Float32Array(n);
  let top = this.boneId('Bip01 Spine1');
  if (top < 0) top = this.boneId('Bip01 Spine');
  const legs = new Set([this.boneId('Bip01 L Thigh'), this.boneId('Bip01 R Thigh')]);
  for (const i of this.order) {
    const p = this.bones[i].parent;
    if (legs.has(i)) continue;
    if (i === top || (p >= 0 && mask[p] === 1)) mask[i] = 1;
  }
  return (this._upper = mask);
};

// blendPoseMasked is blendPose with a weight per bone: f * mask[i], or f
// for every bone when mask is null.
Skeleton.prototype.blendPoseMasked = function (a, b, f, mask, out) {
  if (!mask) return this.blendPose(a, b, f, out);
  const n = this.animLocal.length;
  for (let i = 0; i < n; i++) {
    const w = f * mask[i];
    slerpQuat(a.rot, i * 4, b.rot, i * 4, w, out.rot, i * 4);
    for (let k = 0; k < 3; k++) {
      const j = i * 3 + k;
      out.trans[j] = a.trans[j] + (b.trans[j] - a.trans[j]) * w;
    }
    out.scale[i] = a.scale[i] + (b.scale[i] - a.scale[i]) * w;
  }
  return out;
};

// poseLayered is poseCross with an action on top: the a-to-b cross-fade
// gives the base, then clip act at time tAct is laid over it at weight w,
// on the bones mask selects (all of them when mask is null).
Skeleton.prototype.poseLayered = function (a, ta, b, tb, f, act, tAct, w, mask) {
  if (!act || w <= 0) { this.poseCross(a, ta, b, tb, f); return; }
  const n = this.animLocal.length;
  this._poseA = this._poseA || makePose(n);
  this._poseB = this._poseB || makePose(n);
  this._poseOut = this._poseOut || makePose(n);
  this._poseAct = this._poseAct || makePose(n);
  this._poseBase = this._poseBase || makePose(n);
  let base;
  if (!b || f <= 0) base = this.samplePose(a, ta, this._poseA);
  else if (!a || f >= 1) base = this.samplePose(b, tb, this._poseB);
  else {
    this.samplePose(a, ta, this._poseA);
    this.samplePose(b, tb, this._poseB);
    base = this.blendPose(this._poseA, this._poseB, f, this._poseBase);
  }
  this.samplePose(act, tAct, this._poseAct);
  this.applyPose(this.blendPoseMasked(base, this._poseAct, Math.min(1, w), mask, this._poseOut));
};
