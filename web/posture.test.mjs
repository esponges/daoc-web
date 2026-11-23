// Headless checks of the posture clips.
//   node web/posture.test.mjs
import { PostureClips, LAND_MAX } from './posture.js';

let failed = 0;
const ok = (c, m) => { if (!c) { failed++; console.error('FAIL: ' + m); } else console.log('ok: ' + m); };

const dur = { sitdown: 0.5, sit: 2, standup: 0.4, jump: 0.4, land: 3, swim: 1.8 };
const make = (clips = Object.keys(dur)) => new PostureClips({ has: (n) => clips.includes(n), seconds: (n) => dur[n] || 0 });
const run = (pc, seconds, s, dt = 0.05) => { let out; for (let t = 0; t < seconds - 1e-9; t += dt) out = pc.step(dt, s); return out; };

{
  const pc = make();
  ok(pc.step(0.05, { posture: '' }) === null, 'standing: locomotion chooses');
  ok(pc.step(0.05, { posture: 'sit' }) === 'sitdown', 'sitting begins with the sit-down clip');
  ok(run(pc, 0.3, { posture: 'sit' }) === 'sitdown', 'the sit-down clip plays through');
  ok(run(pc, 0.4, { posture: 'sit' }) === 'sit', 'then the seated loop');
  ok(run(pc, 5, { posture: 'sit' }) === 'sit', 'and it stays seated');
  ok(pc.step(0.05, { posture: '' }) === 'standup', 'standing up plays the stand-up clip');
  ok(run(pc, 0.5, { posture: '' }) === null, 'then locomotion again');
}
{
  const pc = make();
  pc.step(0.05, { posture: 'sit' });
  ok(pc.step(0.05, { posture: '' }) === 'standup', 'standing in the middle of sitting down stands up');
  ok(pc.step(0.05, { posture: 'sit' }) === 'sitdown', 'sitting again in the middle of standing up sits down');
}
{
  const pc = make(['sit']);
  ok(pc.step(0.05, { posture: 'sit' }) === 'sit', 'a figure with no sit-down clip goes straight to the seated loop');
  ok(pc.step(0.05, { posture: '' }) === null, 'and with no stand-up clip straight to locomotion');
  const bare = make([]);
  ok(bare.step(0.05, { posture: 'sit' }) === null, 'a figure with no sit clips is drawn standing, with no error');
}
{
  const pc = make();
  ok(pc.step(0.05, { posture: 'air' }) === 'jump', 'a jump plays the jump clip');
  ok(run(pc, 1, { posture: 'air' }) === 'jump', 'which holds in the air');
  ok(pc.step(0.05, { posture: '', moving: false }) === 'land', 'landing standing still plays the land clip');
  ok(run(pc, LAND_MAX + 0.2, { posture: '' }) === null, 'cut short by LAND_MAX, not run to its three seconds');
}
{
  const pc = make();
  pc.step(0.05, { posture: 'air' });
  ok(pc.step(0.05, { posture: '', moving: true }) === null, 'landing on the run goes straight to the run');
  pc.step(0.05, { posture: 'air' });
  pc.step(0.05, { posture: '' });
  ok(pc.step(0.05, { posture: '', moving: true }) === null, 'moving cuts the landing short');
}
{
  const pc = make();
  ok(pc.step(0.05, { posture: 'swim' }) === 'swim', 'in water the swim clip plays');
  ok(run(pc, 3, { posture: 'swim' }) === 'swim', 'and holds, still or moving');
  ok(pc.step(0.05, { posture: '' }) === null, 'out of the water it is locomotion');
  const dry = make(['jump']);
  ok(dry.step(0.05, { posture: 'swim' }) === null, 'with no swim clip, locomotion stands in');
}
{
  const pc = make();
  pc.step(0.05, { posture: 'air' });
  ok(pc.step(0.05, { posture: 'swim' }) === 'swim', 'falling into water swims');
}
if (failed) process.exit(1);
console.log('all passed');
