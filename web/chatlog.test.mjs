// node web/chatlog.test.mjs
import assert from 'node:assert/strict';
import { CHAT_MAX, CHAT_SIZE, pushRing, atBottom, chatSize } from './chatlog.js';

{
  const l = [];
  for (let i = 0; i < CHAT_MAX + 30; i++) pushRing(l, i);
  assert.equal(l.length, 500, 'the ring keeps 500 lines');
  assert.equal(l[0], 30, 'the oldest fall off');
  assert.equal(l[l.length - 1], CHAT_MAX + 29);
  console.log('ok: the chat keeps a ring of 500 lines');
}
{
  assert.equal(atBottom(70, 100, 170), true, 'at the end');
  assert.equal(atBottom(68, 100, 170), true, 'within slack of the end');
  assert.equal(atBottom(20, 100, 170), false, 'scrolled up: a new line leaves the view alone');
  assert.equal(atBottom(0, 100, 100), true, 'everything fits');
  console.log('ok: stick-to-bottom only when at the bottom');
}
{
  assert.deepEqual(CHAT_SIZE, { w: 450, h: 230 });
  assert.deepEqual(chatSize(100, 50, 1024, 768), { w: 260, h: 110 }, 'not smaller than the minimum');
  assert.deepEqual(chatSize(5000, 5000, 1024, 768), { w: 1004, h: 748 }, 'not larger than the screen');
  assert.deepEqual(chatSize(500.4, 300.6, 1024, 768), { w: 500, h: 301 });
  console.log('ok: resizing stays within bounds');
}
