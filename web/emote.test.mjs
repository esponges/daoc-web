// Headless checks of emotes on a figure.
//   node web/emote.test.mjs
import { eventAnim, playEmote } from './remote.js';

let failed = 0;
const ok = (c, m) => { if (!c) { failed++; console.error('FAIL: ' + m); } else console.log('ok: ' + m); };

function figure(loaded = [], lazy = ['em_wave']) {
  const clips = Object.fromEntries(loaded.map((n) => [n, true]));
  const played = [];
  const loads = [];
  const ch = {
    clips,
    loadClip: (n) => { loads.push(n); return Promise.resolve(lazy.includes(n) ? (clips[n] = true) : null); },
  };
  return { ch, played, loads, has: (n) => !!clips[n], play: (n) => { played.push(n); return 1; }, act: null };
}

{
  const f = figure(['em_wave']);
  eventAnim(f, { k: 'emote', x: 'wave' }, false);
  ok(f.played.join() === 'em_wave', 'an emote event plays its clip on a figure that has it');
}
{
  const f = figure([]);
  eventAnim(f, { k: 'emote', x: 'wave' }, false);
  ok(f.loads.join() === 'em_wave' && f.played.length === 0, 'a clip not yet loaded is fetched first');
  await Promise.resolve(); await Promise.resolve();
  ok(f.played.join() === 'em_wave', 'and plays when it arrives');
}
{
  const f = figure([], []);
  eventAnim(f, { k: 'emote', x: 'dance' }, false);
  await Promise.resolve(); await Promise.resolve();
  ok(f.played.length === 0, 'a clip the conversion lacks plays nothing and does not fail');
}
{
  let threw = false;
  try { playEmote({ has: () => false, play() {} }, 'bow'); } catch { threw = true; }
  ok(!threw, 'a figure with no loader does not fail');
}
if (failed) process.exit(1);
console.log('all passed');
