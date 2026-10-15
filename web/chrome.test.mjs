// Headless checks of the window chrome's small pure parts.
//   node web/chrome.test.mjs

import { clockText } from './chrome.js';

let failed = 0;
const check = (name, ok, extra = '') => {
  if (!ok) {
    failed++;
    console.log(`FAIL ${name} ${extra}`);
  } else console.log(`ok   ${name}`);
};
const at = (h, m) => new Date(2026, 9, 15, h, m);

check('midnight is 12 AM', clockText(at(0, 5)) === '12:05 AM', clockText(at(0, 5)));
check('morning', clockText(at(9, 30)) === '9:30 AM');
check('noon is 12 PM', clockText(at(12, 0)) === '12:00 PM');
check('afternoon', clockText(at(15, 7)) === '3:07 PM');
check('last minute of the day', clockText(at(23, 59)) === '11:59 PM');

console.log(failed ? `${failed} FAILED` : 'all passed');
process.exit(failed ? 1 : 0);
