import assert from 'node:assert/strict';
import { baseFor, candidates, loadRaces } from './race.js';

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

console.log('all passed');
