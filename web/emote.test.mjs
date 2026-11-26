// Headless checks of emotes on a figure.
//   node web/emote.test.mjs
import { eventAnim, playEmote, setStyleAnims, styleClip, combatStance } from './remote.js';

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
// A styled swing plays the style's own clip, named by styles.json.
setStyleAnims({ 150: 60 });
ok(styleClip(150) === 'style60' && styleClip(999) === null, 'a style number names its clip through the table');
{
  const f = figure(['attack1', 'style60']);
  eventAnim(f, { k: 'swing', s: 150 }, false);
  ok(f.played.join() === 'style60', 'a styled swing plays the style clip');
}
{
  const f = figure(['attack1'], ['style60']);
  eventAnim(f, { k: 'swing', s: 150 }, false);
  ok(f.loads.join() === 'style60' && f.played.length === 0, 'a style clip not yet loaded is fetched first');
  await Promise.resolve(); await Promise.resolve();
  ok(f.played.join() === 'style60', 'and plays when it arrives');
}
{
  const f = figure(['attack1'], []);
  eventAnim(f, { k: 'swing', s: 150 }, false, () => 0);
  await Promise.resolve(); await Promise.resolve();
  ok(f.played.join() === 'attack1', 'a body converted without the style clip swings plainly');
}
{
  const f = figure(['attack1']);
  eventAnim(f, { k: 'swing', s: 999 }, false, () => 0);
  ok(f.played.join() === 'attack1', 'a style the table lacks swings plainly');
}
setStyleAnims({});

// A player swings as the weapon held, not with the race body's fists.
{
  const f = figure(['attack1', 'att_2h1']);
  f.wield = '2h';
  eventAnim(f, { k: 'swing' }, false, () => 0);
  ok(f.played.join() === 'att_2h1', 'a two-hander swings its two-handed clip');
}
{
  const f = figure(['attack1'], ['att_1h1']);
  f.wield = '';
  eventAnim(f, { k: 'swing' }, false, () => 0);
  await Promise.resolve(); await Promise.resolve();
  ok(f.loads.join() === 'att_1h1' && f.played.join() === 'att_1h1', 'one-handed (wd absent) fetches and plays the one-handed swing');
}
{
  const f = figure(['attack1'], []);
  f.wield = 'pole';
  eventAnim(f, { k: 'swing' }, false, () => 0);
  await Promise.resolve(); await Promise.resolve();
  ok(f.played.join() === 'attack1', 'a body converted without the set swings its own attack');
}
{
  const f = figure(['attack1', 'att_1h1']);
  eventAnim(f, { k: 'swing' }, false, () => 0);
  ok(f.played.join() === 'attack1', 'a creature (no wield) keeps its own clips');
}
{
  const f = figure(['cidle', 'cidle_2h']);
  f.wield = '2h';
  ok(combatStance(f) === 'cidle_2h', 'the stance follows the weapon');
  const g = figure(['cidle'], ['cidle_staff']);
  g.wield = 'staff';
  ok(combatStance(g) === 'cidle' && g.loads.join() === 'cidle_staff', 'a stance not yet loaded is fetched while the plain one stands in');
}
if (failed) process.exit(1);
console.log('all passed');
