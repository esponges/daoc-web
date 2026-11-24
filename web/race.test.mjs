import assert from 'node:assert/strict';
import { baseFor, candidates, loadRaces, skinFor, loadAppearance } from './race.js';

const ok = (c, m) => { assert.ok(c, m); console.log('ok: ' + m); };

const table = { races: { 6: { m: 'm137', f: 'm145' }, 5: { m: 'm153' } } };

ok(baseFor(table, 6, false) === 'm137' && baseFor(table, 6, true) === 'm145', 'race and gender pick their own base');
ok(baseFor(table, 5, true) === null, 'a gender the race lacks has no base');
ok(baseFor(table, 99, false) === null, 'an unknown race has no base');
ok(baseFor(null, 6, false) === null && baseFor({}, 6, false) === null, 'no table gives no base');

ok(JSON.stringify(candidates(table, { race: 6, female: true, char: 'troll-warrior' })) === '["m145","troll-warrior"]',
  'the base comes first and the outfit is the fallback');
ok(JSON.stringify(candidates(table, { race: 5, female: true, char: 'norseman' })) === '["norseman"]',
  'with no base only the outfit is tried');
ok(JSON.stringify(candidates(null, { race: 6, female: false, char: '' })) === '[]', 'nothing to try when the server named nothing');

const missing = loadRaces(() => Promise.reject(new Error('404')));
await missing.done;
ok(missing.state === 'none' && missing.table === null, 'a missing races.json leaves players on the outfit');
const present = loadRaces(() => Promise.resolve(table));
await present.done;
ok(present.state === 'ok' && present.table === table, 'a present races.json is kept');

const look = { races: { 6: { m: { faces: ['t_mhed1', 't_mhed2'], hair: ['t_hair1', 't_hair2'] } } } };
ok(JSON.stringify(skinFor(look, 6, false, 2, 1)) === '{"Head":"data/appearance/t_mhed2.png","Hair":"data/appearance/t_hair1.png"}', 'a chosen face and hair give their textures');
ok(skinFor(look, 6, false, 0, 0) === null, 'the default look keeps the model own look');
ok(JSON.stringify(skinFor(look, 6, false, 0, 2)) === '{"Hair":"data/appearance/t_hair2.png"}', 'hair alone is fine');
ok(skinFor(look, 6, true, 1, 1) === null && skinFor(look, 6, false, 9, 9) === null && skinFor(null, 6, false, 1, 1) === null, 'a missing gender, number or table gives nothing');
const none = loadAppearance(() => Promise.reject(new Error('404')));
await none.done;
ok(none.table === null, 'a missing appearance.json is not an error');

console.log('all passed');
