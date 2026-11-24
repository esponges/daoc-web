import assert from 'node:assert/strict';
import { delveEntries, rightClick, DELVE_HINT, REMOVE_HINT } from './delve.js';

const col = (lv, me) => (lv - me >= 4 ? 'red' : 'yellow');

// Gear: the name carries its con colour, the details follow in the plain colour.
{
  const e = delveEntries({ nm: 'bronze hammer', lv: 12, con: true, ln: ['Level 12', 'Damage: 1.8 DPS'] }, 5, col);
  assert.equal(e.length, 3);
  assert.deepEqual(e[0], { text: 'bronze hammer (level 12)', kind: 'info', color: 'red' });
  assert.equal(e[1].text, '  Level 12');
  assert.equal(e[1].color, undefined);
  console.log('ok: a delve names the item in its con colour, then lists its details');
}

// A spell is not con coloured; a level of 0 is not shown.
{
  const e = delveEntries({ nm: 'Frost Bolt', lv: 3, ln: ['Range: 1500'] }, 20, col);
  assert.equal(e[0].color, undefined);
  assert.equal(e[0].text, 'Frost Bolt (level 3)');
  const z = delveEntries({ nm: 'A scrap', ln: [] }, 20, col);
  assert.equal(z[0].text, 'A scrap');
  assert.equal(z.length, 1);
  console.log('ok: spells are not con coloured and a missing level is not written');
}

// Shift+Right-click clears a clearable slot; every other right-click delves.
assert.equal(rightClick({ shiftKey: true }, { clearable: true }), 'clear');
assert.equal(rightClick({ shiftKey: false }, { clearable: true }), 'delve');
assert.equal(rightClick({ shiftKey: true }, { clearable: false }), 'delve');
assert.equal(rightClick({ shiftKey: false }), 'delve');
assert.match(DELVE_HINT, /Right-click/);
assert.match(REMOVE_HINT, /Shift\+Right-click/);
console.log('ok: Shift+Right-click clears a clearable slot, a plain right-click delves');
