// node web/torch.test.mjs — the torch: when it shows, which lights, and the saved choice.
import { torchStrength, torchLights, nearestLights, savedTorch, saveTorch, TORCH_HEIGHT } from './torch.js';

let failed = 0;
const ok = (cond, msg) => { if (!cond) { failed++; console.error('FAIL: ' + msg); } else console.log('ok: ' + msg); };

ok(torchStrength(0, false) === 0 && torchStrength(1, false) === 1 && torchStrength(0.5, false) === 0.5, 'outdoors a torch shows with the dark');
ok(torchStrength(0, true) === 1, 'inside a torch always shows');
ok(torchLights([{ x: 1, y: 2, h: 3 }], 0, [0, 0, 0]).length === 0, 'by day there is no torch light');

const lit = torchLights([{ x: 100, y: 200, h: 50 }], 0.5, [0, 0, 0]);
ok(lit.length === 1 && lit[0].pos[0] === 100 && lit[0].pos[1] === 50 + TORCH_HEIGHT && lit[0].pos[2] === 200, 'the light is above the bearer, scene space (height second)');
ok(lit[0].color[0] === 0.5, 'its colour scales with the dark');

const many = Array.from({ length: 12 }, (_, i) => ({ x: i * 100, y: 0, h: 0 }));
const near = torchLights(many, 1, [0, 0, 0], 8);
ok(near.length === 8 && near[0].pos[0] === 0 && near[7].pos[0] === 700, 'the eight nearest the eye are kept');

const wall = { pos: [50, 0, 0], color: [1, 1, 1], radius: 700 };
const merged = nearestLights([[wall], near], [0, 0, 0], 3);
ok(merged.length === 3 && merged[0] === wall && merged[1] === near[0], 'zone lights and torches share the nearest slots');

const mem = new Map();
const store = { getItem: (k) => mem.get(k) ?? null, setItem: (k, v) => mem.set(k, v) };
ok(savedTorch(store) === 'on', 'with nothing saved the torch is lit');
saveTorch(store, false);
ok(savedTorch(store) === 'off', 'putting it out is remembered');
saveTorch(store, true);
ok(savedTorch(store) === 'on', 'lighting it is remembered');
const broken = { getItem() { throw new Error('blocked'); }, setItem() { throw new Error('blocked'); } };
ok(savedTorch(broken) === 'on' && (saveTorch(broken, false), true), 'a blocked store is no saved choice and no error');

if (failed) { console.error(failed + ' failed'); process.exit(1); }
