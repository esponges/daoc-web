// Headless checks of the task window's pure parts.
//   node web/task.test.mjs

import assert from 'node:assert/strict';
import { taskView, timeLeft, journalGroups } from './task.js';

let failed = 0;
const ok = (cond, name, extra = '') => {
  if (!cond) { failed++; console.log(`FAIL ${name} ${extra}`); } else console.log(`ok   ${name}`);
};

ok(taskView(null, 0, 0) === null, 'no message, no view');
const none = taskView({ on: false, n: 2, max: 8 }, 0, 0);
ok(!none.hasTask && none.lines[0] === 'You have no task.' && none.foot === '2 of 8 tasks done at this level', 'no task says so and counts');
const kill = taskView({ on: true, k: 'kill', nm: 'Kill Task', tx: 'Find a wolf and kill it, then return to Sir Gudrun for your reward.', gv: 'Sir Gudrun', left: 7200, n: 1, max: 8 }, 100, 100);
ok(kill.hasTask && kill.title === 'Kill Task' && kill.lines.length === 2 && kill.lines[1] === 'Given by Sir Gudrun', 'a kill task shows what to do and who gave it');
ok(kill.foot === '2h 00m left, 1 of 8 tasks done at this level', 'the time left and the count', kill.foot);
const later = taskView({ on: true, nm: 'Kill Task', tx: 'x', left: 7200, n: 0, max: 1 }, 160, 100);
ok(later.foot.startsWith('1h 59m'), 'the countdown runs from when the message arrived', later.foot);
ok(taskView({ on: true, nm: 'T', tx: 'x', left: 5, n: 0, max: 1 }, 500, 0).foot.startsWith('0m 00s'), 'time never goes below zero');
ok(timeLeft(3725) === '1h 02m' && timeLeft(125) === '2m 05s' && timeLeft(-4) === '0m 00s', 'times read as hours and minutes or minutes and seconds');
ok(taskView({ on: true, nm: 'Kill Task', tx: 'Return to X', dn: true, left: 10, n: 0, max: 1 }, 0, 0).done === true, 'a finished kill is marked');

if (failed) { console.log(`\n${failed} failed`); process.exit(1); }
// The journal tree: the task is one entry with its details under it; no task says so.
{
  const v = taskView({ on: true, nm: 'Kill 3 goblins', tx: 'Kill 3 little water goblins.', gv: 'Guard Flintrock', left: 3600, n: 1, max: 12 }, 0, 0);
  const g = journalGroups(v);
  assert.equal(g.length, 1);
  assert.equal(g[0].title, 'Kill 3 goblins');
  assert.equal(g[0].open, true);
  assert.deepEqual(g[0].rows.map((r) => r.text).slice(0, 2), ['Kill 3 little water goblins.', 'Given by Guard Flintrock']);
  assert.match(g[0].rows[2].text, /left, 1 of 12 tasks done/);
  assert.equal(journalGroups(v, false)[0].open, false, 'a folded entry stays folded');
  assert.equal(journalGroups({ ...v, done: true })[0].title, 'Kill 3 goblins  (done)');
  const none = journalGroups(taskView({ on: false }, 0, 0));
  assert.equal(none[0].title, 'Tasks');
  assert.deepEqual(none[0].rows.map((r) => r.text), ['You have no task.', 'Ask a town guard or a trainer.']);
  assert.deepEqual(journalGroups(null)[0].rows, []);
  console.log('ok: the journal lists the task as an entry with its details');
}

console.log('\nall passed');
