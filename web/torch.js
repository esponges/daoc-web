// The torch a player carries: T (the stock keyboard's Torch action) or
// /torch lights and puts it out, and the server tells every client whose torch
// is lit (an entity's `tc`, your own on the self part of the snapshot). After
// dusk each lit torch is a warm light around its bearer, so a player who puts
// theirs out is harder to see. By day it gives nothing. Inside a dungeon or a
// city there is no day, and a lit torch always shows.

export const TORCH_COLOR = [1.0, 0.7, 0.4];
export const TORCH_REACH = 520; // world units the light falls off to nothing over
export const TORCH_HEIGHT = 70; // carried at about shoulder height
const KEY = 'daoc.torch';

// torchStrength is how much a lit torch shows: 0 by day outdoors, rising with
// the dark to 1, and 1 inside.
export function torchStrength(nightness, indoor) {
  if (indoor) return 1;
  return Math.max(0, Math.min(1, nightness || 0));
}

// torchLights gives the point lights for the lit torches: bearers is
// [{ x, y, h }] in world space (h the height of their feet), eye the camera
// in scene space ([x, height, z]); the n nearest the eye.
export function torchLights(bearers, strength, eye, n = 8) {
  if (strength <= 0) return [];
  const color = TORCH_COLOR.map((c) => c * strength);
  return bearers
    .map((b) => ({ pos: [b.x, b.h + TORCH_HEIGHT, b.y], color, radius: TORCH_REACH }))
    .map((l) => ({ l, d: Math.hypot(l.pos[0] - eye[0], l.pos[1] - eye[1], l.pos[2] - eye[2]) }))
    .sort((a, b) => a.d - b.d).slice(0, n).map(({ l }) => l);
}

// nearestLights merges light lists, keeping the n nearest the eye.
export function nearestLights(lists, eye, n = 8) {
  return lists.flat()
    .map((l) => ({ l, d: Math.hypot(l.pos[0] - eye[0], l.pos[1] - eye[1], l.pos[2] - eye[2]) }))
    .sort((a, b) => a.d - b.d).slice(0, n).map(({ l }) => l);
}

// The stock client keeps the torch choice between sessions (torchmode); so
// does this one, in the browser. A store that throws counts as nothing saved.
export function savedTorch(store) {
  try { return store && store.getItem(KEY) === 'off' ? 'off' : 'on'; } catch { return 'on'; }
}
export function saveTorch(store, lit) {
  try { if (store) store.setItem(KEY, lit ? 'on' : 'off'); } catch { /* not kept */ }
}
