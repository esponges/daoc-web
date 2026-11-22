// node web/conc.test.mjs
import assert from 'node:assert/strict';
import { readFileSync, existsSync } from 'node:fs';
import { concRows, heldSpells, createConcWindow } from './conc.js';
import { UIData } from './uiwin.js';
import { createSpellBar } from './spells.js';

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
const walk = (e, f) => { f(e); (e.children || []).forEach((c) => walk(c, f)); };
const texts = (e) => { const out = []; walk(e, (x) => { if (x.textContent && !x.children.length) out.push(x.textContent); }); return out; };

const m = { max: 92, used: 12, list: [{ id: 1, sp: 8001, nm: 'Blessing of Might', tg: 'Ragnar', cost: 6 }, { id: 2, sp: 8002, nm: 'Blessing of Iron', tg: 'Sigrid', cost: 6 }] };

// Rows and the pool line.
const v = concRows(m);
assert.equal(v.title, 'Concentration 12 / 92');
assert.deepEqual(v.rows.map((r) => r.text), ['Blessing of Might on Ragnar', 'Blessing of Iron on Sigrid']);
assert.equal(v.empty, '');
assert.equal(concRows({ max: 92, used: 0, list: [] }).empty, 'You hold no buffs.');
assert.equal(concRows(null).title, 'Concentration 0 / 0');
assert.deepEqual([...heldSpells(m)], [8001, 8002]);
assert.equal(heldSpells(null).size, 0);
console.log('ok: concentration rows');

// The spell bar marks the buffs that are held.
const book = { lines: [{ spec: 'Mending', nm: 'Mending', sp: [{ id: 8001, nm: 'Blessing of Might', lv: 1, pw: 0, ct: 1, ok: true }, { id: 8003, nm: 'Minor Mending', lv: 1, pw: 1, ct: 2, ok: true }] }] };
const sb = createSpellBar(doc, {});
sb.set([8001, 8003, 0], book, 0, heldSpells(m));
const marked = sb.el.children.map((b) => b.classList.contains('on'));
assert.deepEqual(marked, [true, false, false], 'only the held buff is marked');
assert.match(sb.el.children[0].title, /you hold this buff/);
sb.set([8001, 8003, 0], book, 0);
assert.deepEqual(sb.el.children.map((b) => b.classList.contains('on')), [false, false, false], 'nothing is marked when nothing is held');
console.log('ok: held buffs marked on the spell bar');

// The plain panel: lists the buffs, and a cancel button asks for that buff.
const asked = [];
const plain = createConcWindow(doc, null, { onCancel: (id) => asked.push(id) });
plain.update(m);
assert.equal(plain.isOpen, false);
plain.toggle();
assert.equal(plain.isOpen, true);
assert.ok(plain.el.classList.contains('open'));
assert.match(texts(plain.el).join('|'), /Concentration 12 \/ 92/);
assert.match(texts(plain.el).join('|'), /Blessing of Iron on Sigrid/);
const buttons = [];
walk(plain.el, (x) => { if (x.tag === 'button') buttons.push(x); });
assert.equal(buttons.length, 2);
buttons[1].listeners.click();
assert.deepEqual(asked, [2]);
// The list follows the server: the freed concentration shows at once.
plain.update({ max: 92, used: 6, list: [m.list[0]] });
assert.match(texts(plain.el).join('|'), /Concentration 6 \/ 92/);
assert.doesNotMatch(texts(plain.el).join('|'), /Sigrid/);
plain.toggle();
assert.equal(plain.isOpen, false);
console.log('ok: plain concentration list');

// On the stock frame, when the converter has been run.
const real = 'web/data/ui/atlantis/layout.json';
if (existsSync(real)) {
  const data = new UIData(JSON.parse(readFileSync(real, 'utf8')), 'data/ui/atlantis/');
  const stock = createConcWindow(doc, data, { onCancel: (id) => asked.push(id) });
  stock.update(m);
  stock.toggle(true);
  assert.equal(stock.el.style.display, 'block');
  assert.equal(stock.isOpen, true);
  stock.toggle(false);
  assert.equal(stock.el.style.display, 'none');
  console.log('ok: stock concentration list');
} else console.log('skip stock concentration list (run go run ./cmd/uiconv)');

// The stock window's bar and list.
{
  const { concBar, concListRows } = await import('./conc.js');
  const m = { max: 10, used: 4, list: [{ id: 1, nm: 'Rune of Strength', tg: 'you', cost: 4, sp: 9 }] };
  assert.equal(concBar(m), 0.4);
  assert.equal(concBar(null), 0);
  const rows = concListRows(m);
  assert.equal(rows[0].header, true);
  assert.equal(rows[1].id, 1);
  assert.equal(rows[1].text, 'Rune of Strength on you  (4)');
  assert.equal(concListRows({ max: 10, used: 0, list: [] })[1].text, 'You hold no buffs.');
  console.log('ok: the stock concentration window shows its bar and buffs');
}
