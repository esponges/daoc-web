// Headless checks of how a creature's equipment becomes held items.
//   node web/gear.test.mjs
//
// Uses made-up lookup rows; the real lookup is generated from an install
// (charconv -gear) and is never in the repository.

import { kindFitsSlot, socketFor, resolveHeld, createGear, tiersOf, pickTier, defaultRealm, resolveArmour } from './gear.js';

let failed = 0;
const ok = (cond, msg) => { if (!cond) { failed++; console.error('FAIL: ' + msg); } else console.log('ok: ' + msg); };

const table = {
  items: {
    311: { item: 'h_2h_celticmace01', kind: 'twohand' },
    473: { item: 'b_shield_kitebuckler', kind: 'shield' },
    79: { item: 'b_bo_shortbow01', kind: 'ranged' },
    16: { item: 'b_cr_mace01', kind: 'melee' },
  },
};

// --- kind against slot ---
ok(kindFitsSlot('twohand', 10), 'a two-hander fits the right hand (the data puts them there)');
ok(!kindFitsSlot('shield', 10), 'a shield does not fit the right hand');
ok(kindFitsSlot('shield', 11), 'a shield fits the left hand');
ok(!kindFitsSlot('ranged', 11), 'a bow does not fit a hand slot');
ok(kindFitsSlot('ranged', 13), 'a bow fits the ranged slot');
ok(!kindFitsSlot('melee', 25), 'a torso slot holds nothing');

// --- sockets ---
ok(socketFor(11, 'shield') === 'Bip01 L Shield', 'a left-hand shield goes on the shield socket');
ok(socketFor(11, 'melee') === 'Bip01 L Held', 'an off-hand weapon goes in the left hand');
ok(socketFor(10, 'melee') === 'Bip01 R Held' && socketFor(12, 'twohand') === 'Bip01 R Held', 'right and two-hand go in the right hand');

// --- resolving ---
{
  const eq = [[25, 662, 17, 0], [10, 311, 0, 0], [11, 473, 0, 0], [11, 79, 0, 0], [12, 9999, 0, 0], [10, 473, 0, 0]];
  const r = resolveHeld(eq, table);
  ok(r.pieces.length === 2, 'only the two fitting held pieces are drawn');
  ok(r.pieces[0].item === 'h_2h_celticmace01' && r.pieces[1].socket === 'Bip01 L Shield', 'they resolve to the right items and sockets');
  ok(r.skipped === 3, 'a bow in the left hand, an unknown number and a shield in the right hand are counted as skipped (armour is not)');
  ok(resolveHeld(eq, null).pieces.length === 0, 'no lookup draws nothing');
  ok(resolveHeld(undefined, table).pieces.length === 0, 'no equipment draws nothing');
}

// --- sharing and fallback ---
{
  let loads = 0;
  const world = [];
  const mk = () => ({ skeleton: { world: [new Float32Array(12)] }, draw: (ctx) => world.push(ctx) });
  const gear = createGear({
    loadItem: async () => { loads++; return mk(); },
    fetchJson: async () => table,
  });
  await new Promise((r) => setTimeout(r, 0));
  ok(gear.ready, 'the lookup loads');
  const skel = { boneId: (n) => (n === 'Bip01 R Held' ? 3 : -1), world: [0, 1, 2, [9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9, 9]] };
  const eq = [[10, 311, 0, 0]];
  const a = {}, b = {};
  gear.draw(a, eq, skel, { n: 1 });
  gear.draw(b, eq, skel, { n: 2 }); // first call only starts the load
  await new Promise((r) => setTimeout(r, 0));
  gear.draw(a, eq, skel, { n: 3 });
  gear.draw(b, eq, skel, { n: 4 });
  ok(loads === 1, 'two creatures with the same mace load it once');
  ok(world.length === 2, 'both draw it once it has loaded');
  const missing = createGear({ loadItem: async () => mk(), fetchJson: async () => { throw new Error('404'); } });
  await new Promise((r) => setTimeout(r, 0));
  missing.draw({}, eq, skel, {});
  ok(!missing.ready && missing.skipped === 0, 'with no lookup file nothing is drawn and the game carries on');
}

// --- armour ---
{
  const shapes = [
    { slot: 'body', tier: 1, base: true, texture: 'f_cth0_body1m.png' }, { slot: 'body', tier: 4 }, { slot: 'body', tier: 5 },
    { slot: 'lbody', tier: 1, base: true }, { slot: 'lbody', tier: 3 },
    { slot: 'legs', tier: 1, base: true }, { slot: 'legs', tier: 3 }, { slot: 'legs', tier: 4 },
    { slot: 'boots', tier: 1, base: true }, { slot: 'boots', tier: 5 },
    { name: 'HeadA1' },
  ];
  const sets = (l) => ({ body: l + '_body1m', bodyF: l + '_body1f', lbody: l + '_lbody1', arms: l + '_arm1', gloves: l + '_glov1', legs: l + '_leg1', boots: l + '_boot1' });
  const armorTable = { armor: {
    pieces: { '27:47': { level: 4, realm: 1 }, '25:428': { level: 3, realm: 3 }, '25:245': { level: 0 }, '23:50': { level: 4, realm: 1 } },
    textures: { 'b:4': sets('b_plt1'), 'h:3': sets('h_scl1'), 'n:0': sets('n_cth1'), 'b:4x': sets('unused') },
    tints: { 17: [125, 126, 133], 40: [255, 255, 255] },
  } };
  ok(JSON.stringify(tiersOf(shapes)) === JSON.stringify({ body: [1, 4, 5], lbody: [1, 3], legs: [1, 3, 4], boots: [1, 5] }), 'tiers are listed per slot, ignoring shapes without one');
  ok(pickTier([1, 4, 5], 5) === 5 && pickTier([1, 4, 5], 3) === 1 && pickTier([2, 4], 1) === 2, 'a tier the model lacks falls back to the nearest lighter one, or its lightest');
  ok(defaultRealm('HFirbolgM.NIF') === 'h' && defaultRealm('NVikingM.NIF') === 'n' && defaultRealm('BritonM.NIF') === 'b' && defaultRealm('') === 'n', 'a creature defaults to the realm its model file names');

  let look = resolveArmour([[27, 47, 0, 0], [23, 50, 17, 0]], armorTable, shapes);
  ok(look.tiers.legs === 4 && look.tiers.boots === 5, 'plate legs and boots pick the heaviest tiers');
  ok(look.tex.legs === 'data/armor/b_plt1_leg1.png' && look.tex.boots === 'data/armor/b_plt1_boot1.png', 'they use the Albion plate texture set');
  ok(look.tint.boots && Math.abs(look.tint.boots[0] - 125 / 255) < 1e-9 && !look.tint.legs, 'a colour tints its slot, and none leaves it alone');
  ok(!look.tiers.body && !look.tiers.arms, 'slots with no worn piece, or that the model lacks, are left to the base model');

  look = resolveArmour([[25, 428, 0, 0]], armorTable, shapes);
  ok(look.tiers.body === 4 && look.tiers.lbody === 3, 'a torso piece dresses both the short and the long body');
  ok(look.tex.body === 'data/armor/h_scl1_body1m.png', 'the realm comes from the database items');
  look = resolveArmour([[25, 428, 0, 0]], armorTable, shapes, { female: true });
  ok(look.tex.body === 'data/armor/h_scl1_body1f.png' && look.tex.lbody === 'data/armor/h_scl1_lbody1.png', 'a female model gets the female body texture');
  look = resolveArmour([[25, 245, 40, 0]], armorTable, shapes, { realm: 'n' });
  ok(look.tex.body === 'data/armor/n_cth1_body1m.png' && !look.tint.body, 'a number naming no realm uses the creature realm, and white is no tint');
  look = resolveArmour([[25, 9999, 0, 0], [27, 47, 0, 0]], armorTable, shapes);
  ok(look.skipped === 1 && look.tiers.legs === 4, 'an unknown number is skipped and counted without losing the rest');
  ok(Object.keys(resolveArmour([[27, 47, 0, 0]], null, shapes).tiers).length === 0, 'no lookup dresses nothing');
  look = resolveArmour([[27, 47, 999, 0]], armorTable, shapes);
  ok(look.tiers.legs === 4 && !look.tint.legs, 'a colour the tints table lacks leaves the texture as it is');
}

if (failed) { console.error(failed + ' failed'); process.exit(1); }
console.log('all passed');
