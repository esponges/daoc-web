// node web/command.test.mjs
import assert from 'node:assert/strict';
import { readFileSync, existsSync } from 'node:fs';
import { createCommandWindow, COMMAND_ACTIONS, COMMAND_WHY } from './command.js';
import { UIData } from './uiwin.js';

const doc = {
  createElement(tag) {
    const classes = new Set();
    const e = {
      tag, style: {}, dataset: {}, title: '', children: [], listeners: {},
      appendChild(c) { this.children.push(c); return c; },
      append(...cs) { cs.forEach((c) => this.children.push(c)); },
      replaceChildren(...cs) { this.children = cs; },
      addEventListener(t, f) { this.listeners[t] = f; },
      classList: { add: (c) => classes.add(c), remove: (c) => classes.delete(c), contains: (c) => classes.has(c) },
    };
    let text = '';
    Object.defineProperty(e, 'textContent', { get: () => text, set: (v) => { text = v; } });
    Object.defineProperty(e, 'className', { get: () => [...classes].join(' '), set: (v) => { classes.clear(); String(v).split(/\s+/).filter(Boolean).forEach((c) => classes.add(c)); } });
    return e;
  },
};

// The buttons that do something, and the ones greyed with a reason.
assert.deepEqual(Object.values(COMMAND_ACTIONS).sort(), ['attack', 'face', 'follow', 'pray', 'release', 'sit', 'sprint', 'stick']);
for (const ev of ['Bonuses']) assert.ok(COMMAND_WHY[ev], ev + ' has a reason');
assert.equal(COMMAND_ACTIONS.Stick, 'stick');
assert.ok(!('Stick' in COMMAND_WHY), 'Stick is built, so it is not greyed');
console.log('ok: which command buttons work and which say why not');

// Without the converted interface: two plain buttons.
{
  const asked = [];
  const w = createCommandWindow(doc, null, { onAction: (a) => asked.push(a) });
  const buttons = w.el.children;
  assert.deepEqual(buttons.map((b) => b.textContent), ['Sit', 'Sprint', 'Follow', 'Stick', 'Face']);
  buttons[0].listeners.click();
  buttons[1].listeners.click();
  buttons[2].listeners.click();
  buttons[3].listeners.click();
  buttons[4].listeners.click();
  assert.deepEqual(asked, ['sit', 'sprint', 'follow', 'stick', 'face']);
  w.toggle(false);
  assert.equal(w.el.style.display, 'none');
  console.log('ok: the plain command buttons sit, sprint, follow, stick and face');
}

// On the stock frame, when the converter has been run.
const real = 'web/data/ui/atlantis/layout.json';
if (existsSync(real)) {
  const data = new UIData(JSON.parse(readFileSync(real, 'utf8')), 'data/ui/atlantis/');
  const asked = [];
  const w = createCommandWindow(doc, data, { onAction: (a) => asked.push(a) });
  assert.equal(w.el.style.display, 'none');
  w.toggle(true);
  assert.equal(w.el.style.display, 'block');
  assert.equal(w.isOpen, true);
  w.toggle(false);
  assert.equal(w.isOpen, false);
  console.log('ok: the stock command window opens and closes');
} else console.log('skip stock command window (run go run ./cmd/uiconv)');
