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

// stepBody moves b one fixed step under the keys k. ground answers
// groundAt(x, y, z) and clamp(x, y); mover, if given, slides the move along
// whatever blocks it: slide(x0, y0, x1, y1, z) -> [x, y]. It returns what the
// step did, for choosing a clip.
export function stepBody(ground, b, k, mover = null) {
  b.yaw += k.turn * TURN * STEP_DT;
  const fx = Math.sin(b.yaw), fy = -Math.cos(b.yaw);
  const rx = Math.cos(b.yaw), ry = Math.sin(b.yaw);
  let mx = fx * k.fwd + rx * k.strafe;
  let my = fy * k.fwd + ry * k.strafe;
  const m = { speed: 0, back: false, side: 0 };
  const l = Math.hypot(mx, my);
  if (l > 0) {
    mx /= l; my /= l;
    m.back = k.fwd < 0;
    if (k.fwd === 0) m.side = k.strafe;
    const speed = k.run && !m.back && m.side === 0 ? RUN : WALK;
    const d = speed * STEP_DT;
    let [nx, ny] = ground.clamp(b.x + mx * d, b.y + my * d);
    if (mover) [nx, ny] = mover.slide(b.x, b.y, nx, ny, b.z);
    m.speed = Math.hypot(nx - b.x, ny - b.y) / STEP_DT;
    b.x = nx; b.y = ny;
  }
  b.z = ground.groundAt(b.x, b.y, b.z);
  return m;
}

// runKeys decodes one run of an input message, [count, fwd, strafe, turn,
// flags], clamping as the server does.
export function runKeys(r) {
  const c = (v) => Math.max(-1, Math.min(1, v | 0));
  return {
    count: Math.max(0, Math.min(STEPS_PER_SECOND * 2, r[0] | 0)),
    keys: { fwd: c(r[1]), strafe: c(r[2]), turn: c(r[3]), run: (r[4] & 1) !== 0 },
  };
}

// The synthetic slope the shared vectors are written against.
export const planeGround = {
  groundAt: (x, y) => 4000 + 0.1 * x + 0.05 * y,
  clamp: (x, y) => [Math.max(0, Math.min(65280, x)), Math.max(0, Math.min(65280, y))],
};
