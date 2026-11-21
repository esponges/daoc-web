// node web/spellfx.test.mjs
import assert from 'node:assert/strict';
import {
  createFxBook, castPlan, bonePoint, createFxLive, createCastSeq, flightTime, flightAt,
  MAX_LIVE, FLIGHT_MIN, FLIGHT_MAX,
} from './spellfx.js';

// The index says how each client effect looks; an effect with no model is not drawn.
{
  const book = createFxBook({
    spells: { 2501: { name: 'Minor Dusk Dart', anims: [41, 42, 43], caster: [160, 161], target: 256 } },
    effects: { 160: { model: 3, node: 'Bip01 L Held', ms: 3000 }, 161: { model: -1, ms: 100 } },
  });
  assert.equal(book.spell(2501).name, 'Minor Dusk Dart');
  assert.equal(book.spell(9), null);
  assert.equal(book.effect(160).node, 'Bip01 L Held');
  assert.equal(book.effect(161), null, 'an effect with nothing to draw is skipped');
  assert.equal(book.effect(5), null);
  assert.equal(createFxBook(null).spell(1), null, 'no index, no spells');
  console.log('ok: the book finds a spell and its drawable effects');
}

// The three cast clips are named after the columns of the anim table.
{
  assert.deepEqual(castPlan({ anims: [41, 42, 43] }), { start: 'cast41', mid: 'cast42', end: 'cast43' });
  assert.deepEqual(castPlan({ anims: [16, 0, 0] }), { start: 'cast16', mid: null, end: null });
  assert.deepEqual(castPlan(null), { start: null, mid: null, end: null });
  console.log('ok: a cast plan names its three clips');
}

// A bone's world place follows the model matrix: X and Y spread over the ground, Z is up.
{
  const skel = { boneId: (n) => ({ hand: 1 }[n] ?? -1), world: [null, new Float32Array([1, 0, 0, 10, 0, 1, 0, 20, 0, 0, 1, 30])] };
  const at = (view) => bonePoint(skel, 'hand', view).map((v) => Math.round(v * 1000) / 1000);
  assert.deepEqual(at({ x: 100, y: 200, z: 50, yaw: 0, scale: 1 }), [110, 220, 80], 'no turn');
  // A quarter turn: model X goes to world +Y, model Y to world -X.
  assert.deepEqual(at({ x: 0, y: 0, z: 0, yaw: Math.PI / 2, scale: 1 }), [-20, 10, 30]);
  assert.deepEqual(at({ x: 0, y: 0, z: 0, yaw: 0, scale: 2 }), [20, 40, 60], 'a larger figure reaches further');
  assert.deepEqual(bonePoint(skel, 'rootnode', { x: 5, y: 6, z: 7, yaw: 1, scale: 1 }), [5, 6, 7], 'rootnode is the feet');
  assert.deepEqual(bonePoint(skel, 'no such bone', { x: 5, y: 6, z: 7, yaw: 0, scale: 1 }, { chest: 40 }), [5, 6, 47], 'a missing bone falls back to the body');
  assert.deepEqual(bonePoint(null, 'hand', { x: 1, y: 2, z: 3, yaw: 0, scale: 1 }), [1, 2, 3], 'no skeleton, the feet');
  console.log('ok: a bone is placed in the world like the model');
}

// Live instances end on time, are stopped by tag, and never exceed the cap.
{
  const live = createFxLive(3);
  live.add({ effect: 1, start: 0, dur: 2, tag: 'cast:7' });
  live.add({ effect: 2, start: 0, dur: 5, tag: 'land' });
  assert.equal(live.update(1).length, 2);
  assert.equal(live.update(2.5).length, 1, 'the first has ended');
  live.stop('land');
  assert.equal(live.update(3).length, 0, 'a stopped one goes at the next update');
  for (let i = 0; i < 5; i++) live.add({ effect: i, start: 0, dur: 9 });
  assert.equal(live.length, 3, 'the cap drops the oldest');
  assert.deepEqual(live.list.map((i) => i.effect), [2, 3, 4], 'the newest are kept');
  const seeds = new Set(live.list.map((i) => i.seed));
  assert.equal(seeds.size, 3, 'each instance has its own seed, so copies do not burn in step');
  assert.equal(MAX_LIVE, 64);
  console.log('ok: live effects end, stop by tag and stay under the cap');
}

// The cast animation: start, hold, release; the generic clip for a figure without the spell's.
{
  const lens = { cast41: 0.5, cast42: 1, cast43: 0.4, cast: 0.8 };
  const clipOf = (n) => lens[n] || 0;
  const plan = castPlan({ anims: [41, 42, 43] });
  const seq = createCastSeq();
  assert.deepEqual(seq.begin(plan, 10, clipOf), { play: 'cast41', hold: true });
  assert.equal(seq.tick(10.3, clipOf), null, 'the start is still playing');
  assert.deepEqual(seq.tick(10.5, clipOf), { play: 'cast42', hold: true }, 'then the hold');
  assert.equal(seq.tick(11, clipOf), null, 'and it stays held');
  assert.deepEqual(seq.land(clipOf), { play: 'cast43', hold: false }, 'the release when the spell lands');
  assert.equal(seq.active, false);

  // Interrupted.
  seq.begin(plan, 0, clipOf);
  assert.deepEqual(seq.stop(), { clear: true });
  assert.equal(seq.stop(), null, 'nothing to stop twice');

  // A figure that does not have the clips yet holds the generic cast, and moves onto the spell's when they arrive.
  const late = { cast: 0.8 };
  const clipLate = (n) => late[n] || 0;
  assert.deepEqual(seq.begin(plan, 0, clipLate), { play: 'cast', hold: true });
  assert.equal(seq.tick(0.2, clipLate), null);
  late.cast42 = 1;
  assert.deepEqual(seq.tick(0.4, clipLate), { play: 'cast42', hold: true }, 'the fetched clip takes over');
  late.cast43 = 0.4;
  assert.deepEqual(seq.land(clipLate), { play: 'cast43', hold: false });

  // Landing on a figure without a release clip just ends the pose.
  seq.begin(plan, 0, clipLate);
  assert.deepEqual(seq.land(() => 0), { clear: true });
  console.log('ok: a cast starts, holds and releases, with the generic clip as the fallback');
}

// A projectile flies in a time that fits its distance, along the line.
{
  assert.equal(flightTime(10), FLIGHT_MIN, 'a point-blank shot is not instant');
  assert.equal(flightTime(1e6), FLIGHT_MAX, 'a far one is not slow');
  assert.ok(flightTime(900) > flightTime(300));
  const f = { start: 10, dur: 1, from: [0, 0, 0], to: [100, 200, 40] };
  assert.equal(flightAt(f, 9.9), null, 'not yet');
  assert.deepEqual(flightAt(f, 10.5), [50, 100, 20], 'halfway');
  assert.equal(flightAt(f, 11), null, 'landed');
  console.log('ok: projectiles fly along the line in a fitting time');
}

// The director: events in, animations and effects out.
{
  const { createDirector } = await import('./spellfx.js');
  const calls = [];
  const mkAnim = () => ({ act: null, play(n, o) { calls.push(['play', n, o.hold, o.upper]); this.act = n; }, clear() { calls.push(['clear']); this.act = null; } });
  const lens = { cast: 0.8, cast41: 0.5, cast42: 1, cast43: 0.4 };
  const clips = {};
  const mkCh = (have) => ({
    clips: Object.fromEntries(have.map((n) => [n, true])),
    clipDuration: (n) => lens[n] || 0, clipRate: () => 1,
    loadClip(n) { calls.push(['load', n]); return Promise.resolve(null); },
  });
  const caster = { anim: mkAnim(), ch: mkCh(['cast', 'cast41', 'cast42', 'cast43']), moving: false };
  const figs = { 7: caster };
  const played = [], flown = [], stopped = [];
  const book = createFxBook({
    spells: { 2501: { name: 'Minor Dusk Dart', anims: [41, 42, 43], caster: [160, 161], projectile: 77, target: 256, resist: 5 }, 9: { name: 'Bare', anims: [0, 0, 0] } },
    effects: { 160: { model: 1, node: 'Bip01 L Held', ms: 3000 }, 161: { model: 2, node: 'Bip01 R Held', ms: 3000 }, 77: { model: 3, node: 'rootnode', ms: 1000 },
      256: { model: 4, node: 'rootnode', ms: 2000 }, 5: { model: 5, node: 'rootnode', ms: 500 } },
  });
  const live = createFxLive();
  const fx = { book, live, play: (id, anchor, start, dur, tag) => { played.push({ id, anchor, start, dur, tag }); return live.add({ effect: id, anchor, start, dur: dur || 1, tag }); },
    fly: (id, from, to, start, dur) => flown.push({ id, from, to, start, dur }) };
  const where = { 7: [0, 0, 0], 8: [600, 0, 0] };
  const dir = createDirector({ getFx: () => fx, figure: (id) => figs[id] || null, locate: (a) => where[a.id] || null });

  // A cast: the start clip, the caster's effects on their bones for the cast and a moment.
  dir.event({ k: 'cast', from: 7, to: 8, x: 'Minor Dusk Dart', fx: 2501, dur: 2.6 }, 10);
  assert.deepEqual(calls.shift(), ['play', 'cast41', true, false]);
  assert.deepEqual(played.map((p) => [p.id, p.anchor.bone, p.tag]), [[160, 'Bip01 L Held', 'cast:7'], [161, 'Bip01 R Held', 'cast:7']]);
  assert.ok(Math.abs(played[0].dur - 2.9) < 1e-9, 'for the cast and a tail');
  dir.tick(10.2);
  assert.equal(calls.length, 0, 'nothing yet');
  dir.tick(10.6);
  assert.deepEqual(calls.shift(), ['play', 'cast42', true, false], 'the hold follows the start');

  // It lands: the release, the caster effects end, the bolt flies and the hit waits for it.
  dir.event({ k: 'spell', from: 7, to: 8, x: 'Minor Dusk Dart', n: 6, fx: 2501 }, 12.6);
  assert.deepEqual(calls.shift(), ['play', 'cast43', false, false]);
  assert.equal(live.update(12.7).filter((i) => i.tag === 'cast:7').length, 0, 'the caster effects end with the cast');
  assert.equal(flown.length, 1);
  assert.deepEqual(flown[0].from, where[7]);
  assert.deepEqual(flown[0].to, where[8]);
  const hit = played.find((p) => p.tag === 'hit');
  assert.equal(hit.id, 256);
  assert.ok(Math.abs(hit.start - (12.6 + flown[0].dur)) < 1e-9, 'the hit effect plays when the bolt arrives');
  assert.equal(hit.anchor.id, 8);

  // A tick of a lasting spell, soon after: the hit again, but no second bolt.
  played.length = 0;
  dir.event({ k: 'spell', from: 7, to: 8, x: 'Minor Dusk Dart', n: 3, fx: 2501 }, 14);
  assert.equal(flown.length, 1, 'the bolt does not fly for each tick');
  assert.equal(played.length, 1);

  // A resist shows the bolt but not the hit.
  played.length = 0;
  dir.event({ k: 'cast', from: 7, to: 8, fx: 2501, dur: 1 }, 30);
  calls.length = 0;
  dir.event({ k: 'resist', from: 7, to: 8, x: 'Minor Dusk Dart', fx: 2501 }, 31);
  assert.equal(flown.length, 2);
  assert.deepEqual(played.filter((p) => p.tag === 'hit').map((p) => p.id), [5], 'the resist effect, not the hit effect');

  // Interrupted: the pose and the caster effects stop.
  dir.event({ k: 'cast', from: 7, to: 8, fx: 2501, dur: 2 }, 40);
  calls.length = 0;
  dir.event({ k: 'interrupt', from: 7 }, 41);
  assert.deepEqual(calls, [['clear']]);

  // A spell with no client effect still animates, with the generic clip, and shows nothing else.
  played.length = 0; calls.length = 0;
  dir.event({ k: 'cast', from: 7, to: 8, dur: 1 }, 50);
  assert.deepEqual(calls.shift(), ['play', 'cast', true, false]);
  dir.event({ k: 'spell', from: 7, to: 8, n: 1 }, 51);
  assert.deepEqual(calls.shift(), ['clear']);
  assert.equal(played.length, 0);

  // A caster whose clips are still being fetched holds the generic clip and asks for them.
  figs[7] = { anim: mkAnim(), ch: mkCh(['cast']), moving: true };
  calls.length = 0;
  dir.event({ k: 'cast', from: 7, to: 8, fx: 2501, dur: 2 }, 60);
  assert.deepEqual(calls.filter((c) => c[0] === 'load').map((c) => c[1]), ['cast41', 'cast42', 'cast43']);
  assert.deepEqual(calls.find((c) => c[0] === 'play'), ['play', 'cast', true, true], 'a moving caster casts with the upper body');

  // No effects loaded at all: the cast still animates and nothing throws.
  const bare = createDirector({ getFx: () => null, figure: (id) => figs[id] || null, locate: () => null });
  calls.length = 0;
  bare.event({ k: 'cast', from: 7, fx: 2501, dur: 1 }, 70);
  bare.event({ k: 'spell', from: 7, to: 8, fx: 2501 }, 71);
  assert.deepEqual(calls.map((c) => c[0]), ['load'].slice(0, 0).concat(['play', 'clear']), 'generic cast, then it ends');
  console.log('ok: the director plays casts, releases, bolts, hits, resists and interrupts');
}

// An arrow flies from the archer to the target, turned along its path.
{
  const { createDirector } = await import('./spellfx.js');
  const flown = [];
  const book = createFxBook({ spells: {}, effects: { 593: { model: 7, node: 'Bip01 L Held', ms: 2000 } }, arrow: { travel: 593, strike: 594 } });
  const fx = { book, live: createFxLive(), play() {}, fly: (id, from, to, start, dur) => flown.push({ id, from, to, start, dur }) };
  const dir = createDirector({ getFx: () => fx, figure: () => null, locate: (a) => ({ 1: [0, 0, 100], 2: [900, 0, 100] }[a.id]) });
  dir.event({ k: 'arrow', from: 1, to: 2, dur: 0.6 }, 5);
  assert.equal(flown.length, 1);
  assert.deepEqual([flown[0].id, flown[0].start, flown[0].dur], [593, 5, 0.6]);
  assert.deepEqual(flown[0].to, [900, 0, 100]);
  dir.event({ k: 'arrow', from: 1, to: 2, dur: 0 }, 6);
  assert.equal(flown[1].dur, FLIGHT_MIN, 'a point-blank shot is still seen');
  // Without the arrow in the data, or without effects, a shot shows nothing and nothing throws.
  createDirector({ getFx: () => ({ ...fx, book: createFxBook({ spells: {}, effects: {} }) }), figure: () => null, locate: () => null }).event({ k: 'arrow', from: 1, to: 2, dur: 1 }, 7);
  createDirector({ getFx: () => null, figure: () => null, locate: () => null }).event({ k: 'arrow', from: 1, to: 2, dur: 1 }, 7);
  assert.equal(flown.length, 2);
  console.log('ok: an arrow flies to its target, and a missing arrow model shows nothing');
}
