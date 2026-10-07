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
    for (let i = 0; i < n; i++) {
      this.bindLocal.push(fromBone(bones[i].t, bones[i].r, bones[i].s));
      this.bindWorld.push(xform());
      this.animLocal.push(xform());
      this.world.push(xform());
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
