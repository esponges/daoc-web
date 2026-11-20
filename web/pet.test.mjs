// node web/pet.test.mjs
import assert from 'node:assert/strict';
import { petView, ownerLabel, petOrderText, createPetWindow } from './pet.js';

const doc = {
  createElement(tag) {
    const classes = new Set();
    const e = {
      tag, style: {}, dataset: {}, children: [], listeners: {},
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
const walk = (e, f) => { f(e); (e.children || []).forEach((c) => walk(c, f)); };
const find = (e, pred) => { const out = []; walk(e, (x) => { if (pred(x)) out.push(x); }); return out; };

const m = { t: 'pet', id: 1500000001, nm: 'spirit fighter', lv: 8, hp: 0.5, st: 'defensive', md: 'follow', mn: [{ id: 1500000002, nm: 'bonemage', lv: 7, hp: 1 }] };

// What is drawn.
const v = petView(m);
assert.equal(v.title, 'spirit fighter  (level 8)');
assert.equal(v.hp, 0.5);
assert.equal(v.hpText, '50%');
assert.equal(v.stance, 'defensive');
assert.equal(v.minions[0].text, 'bonemage (7)  100%');
assert.equal(petView({ t: 'pet', cl: true, hp: 0 }), null, 'a closed pet draws nothing');
assert.equal(petView(null), null);
assert.equal(petView({ id: 3, nm: 'x', hp: 4 }).hp, 1, 'health is kept between 0 and 1');
assert.equal(ownerLabel({ ow: 'Alice' }), "Alice's pet");
assert.equal(ownerLabel({}), '');
assert.equal(petOrderText('stay'), '/pet stay');
console.log('ok: pet view');

// The plain panel opens with a pet, gives orders, and closes when it is gone.
const sent = [];
const win = createPetWindow(doc, null, { onOrder: (c) => sent.push(c) });
assert.equal(win.isOpen, false);
win.update(m);
assert.equal(win.isOpen, true);
assert.ok(win.el.classList.contains('open'));
const buttons = find(win.el, (x) => x.tag === 'button');
assert.deepEqual(buttons.map((b) => b.textContent), ['Attack', 'Follow', 'Stay', 'Release', 'Passive', 'Defensive', 'Aggressive']);
assert.ok(buttons[1].classList.contains('on'), 'the current order is marked');
assert.ok(buttons[5].classList.contains('on'), 'the current stance is marked');
buttons[0].listeners.click();
buttons[6].listeners.click();
assert.deepEqual(sent, ['attack', 'aggressive']);
win.update({ t: 'pet', cl: true, hp: 0 });
assert.equal(win.isOpen, false);
assert.equal(win.el.style.display, 'none');
console.log('ok: pet window');
