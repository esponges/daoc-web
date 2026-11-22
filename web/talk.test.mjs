// node web/talk.test.mjs
import assert from 'node:assert/strict';
import { healerTalk } from './talk.js';

{
  const v = healerTalk({ nm: 'Brother Hrafn', lost: 3, total: 120, price: 40 }, (c) => c + 'c');
  assert.equal(v.title, 'Brother Hrafn');
  assert.equal(v.text, 'Brother Hrafn can restore your 3 lost constitution for 120c (40c a point).');
  assert.equal(v.accept, 'ACCEPT');
  console.log('ok: a healer\'s offer is the Interact window\'s title and text');
}
