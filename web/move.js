// How a character moves, in one place.
//
// This is the same maths as internal/server/world/move.go, step for step:
// the server runs it on a player's inputs, and the browser runs it to predict
// the result before the server's answer arrives. web/net.test.mjs checks the
// two agree on vectors the Go tests write (internal/server/world/testdata).
//
// A step is 1/60 of a second. Yaw faces (sin yaw, -cos yaw); world X and Y
// are the ground plane, as everywhere else in the viewer.

export const WALK = 85;   // units a second
export const RUN = 191;
export const TURN = 2.8;  // radians a second
export const STEP_DT = 1 / 60;
export const STEPS_PER_SECOND = 60;

// The numbers of internal/server/rules/movement.go (rules-movement): the
// vectors fail if either side drifts.
export const SPRINT_FACTOR = 1.25;
export const SWIM_SPEED = 120;
export const JUMP_SPEED = 210;   // units a second, upward
export const GRAVITY = 500;      // units a second squared
export const STEP_DOWN = 12;     // the most the ground may drop in a step before the body falls
export const SWIM_DEPTH = 50;    // how deep the water must be to swim in it; also how far the feet hang
export const FLAG_RUN = 1;
export const FLAG_JUMP = 2;

// stepBody moves b one fixed step under the keys k. ground answers
// groundAt(x, y, z), clamp(x, y) and, if there are lakes, waterAt(x, y) -> the
// surface height or null; mover, if given, slides the move along whatever
// blocks it: slide(x0, y0, x1, y1, z) -> [x, y]. b carries the vertical state
// too: vz, peakZ, air and jumped. It returns what the step did, for choosing
// a clip.
export function stepBody(ground, b, k, mover = null) {
  const m = { speed: 0, back: false, side: 0, swim: false, air: false, landed: false, water: false, fallHeight: 0 };
  const slow = k.slow || 0; // the share of speed a slow or hold takes away; 1 is held
  const held = slow >= 1;
  if (!held) {
    const afloat = !b.air && swimming(ground, b.x, b.y);
    b.yaw += k.turn * TURN * STEP_DT;
    const fx = Math.sin(b.yaw), fy = -Math.cos(b.yaw);
    const rx = Math.cos(b.yaw), ry = Math.sin(b.yaw);
    let mx = fx * k.fwd + rx * k.strafe;
    let my = fy * k.fwd + ry * k.strafe;
    const l = Math.hypot(mx, my);
    if (l > 0) {
      mx /= l; my /= l;
      m.back = k.fwd < 0;
      if (k.fwd === 0) m.side = k.strafe;
      let speed = WALK;
      if (afloat) speed = SWIM_SPEED;
      else if (k.run && !m.back && m.side === 0) speed = k.sprint ? RUN * SPRINT_FACTOR : RUN;
      const d = speed * (1 - slow) * STEP_DT;
      let [nx, ny] = ground.clamp(b.x + mx * d, b.y + my * d);
      if (mover) [nx, ny] = mover.slide(b.x, b.y, nx, ny, b.z);
      m.speed = Math.hypot(nx - b.x, ny - b.y) / STEP_DT;
      b.x = nx; b.y = ny;
    }
  }
  vertical(ground, b, !!k.jump && !held, m);
  m.air = !!b.air;
  return m;
}

function waterAt(ground, x, y) {
  const s = ground.waterAt ? ground.waterAt(x, y) : null;
  return s === null || s === undefined ? null : s;
}

function swimming(ground, x, y) {
  const surface = waterAt(ground, x, y);
  return surface !== null && surface - ground.groundAt(x, y, surface) > SWIM_DEPTH;
}

// vertical is the part of a step that moves a body up and down, as
// vertical() in internal/server/world/move.go.
function vertical(ground, b, jump, m) {
  const g = ground.groundAt(b.x, b.y, b.z);
  const surface = waterAt(ground, b.x, b.y);
  const deep = surface !== null && surface - g > SWIM_DEPTH;
  if (!b.air) {
    if (deep) { b.z = surface - SWIM_DEPTH; m.swim = true; return; }
    if (jump) { b.air = true; b.jumped = true; b.vz = JUMP_SPEED; b.peakZ = b.z; }
    else if (b.z - g > STEP_DOWN) { b.air = true; b.jumped = false; b.vz = 0; b.peakZ = b.z; }
    else { b.z = g; return; }
  }
  b.vz -= GRAVITY * STEP_DT;
  b.z += b.vz * STEP_DT;
  b.peakZ = Math.max(b.peakZ, b.z);
  if (deep && b.z <= surface - SWIM_DEPTH) { b.z = surface - SWIM_DEPTH; m.water = true; }
  else if (!deep && b.z <= g) { b.z = g; }
  else return;
  m.landed = true; m.fallHeight = b.peakZ - b.z;
  b.air = false; b.jumped = false; b.vz = 0;
}

// runKeys decodes one run of an input message, [count, fwd, strafe, turn,
// flags], clamping as the server does.
export function runKeys(r) {
  const c = (v) => Math.max(-1, Math.min(1, v | 0));
  return {
    count: Math.max(0, Math.min(STEPS_PER_SECOND * 2, r[0] | 0)),
    keys: { fwd: c(r[1]), strafe: c(r[2]), turn: c(r[3]), run: (r[4] & FLAG_RUN) !== 0, jump: (r[4] & FLAG_JUMP) !== 0 },
  };
}

// The synthetic slope the shared vectors are written against.
export const planeGround = {
  groundAt: (x, y) => 4000 + 0.1 * x + 0.05 * y,
  clamp: (x, y) => [Math.max(0, Math.min(65280, x)), Math.max(0, Math.min(65280, y))],
};

// The two other synthetic grounds of the shared vectors (move_test.go): a
// cliff, and a lake with a shore.
export const cliffGround = {
  groundAt: (x) => (x < 30100 ? 4000 : 1000),
  clamp: planeGround.clamp,
};
export const lakeGround = {
  groundAt: (x, y) => (y >= 30200 ? 3500 : 4000),
  waterAt: (x, y) => (y >= 30200 ? 4000 : null),
  clamp: planeGround.clamp,
};
