// node web/group.test.mjs
import assert from 'node:assert/strict';
import { readFileSync, existsSync } from 'node:fs';
import { groupRows, chatLine, createStockGroup, createPlainGroup, createInvitePrompt } from './group.js';
import { UIData } from './uiwin.js';

// A stand-in document: enough surface for the builders.
const doc = {
  createElement(tag) {
    const classes = new Set();
    const e = {
      tag, style: {}, dataset: {}, children: [], listeners: {}, title: '',
      appendChild(c) { this.children.push(c); return c; },
      append(...cs) { cs.forEach((c) => this.children.push(c)); },
      replaceChildren(...cs) { this.children = cs; },
      addEventListener(t, f) { this.listeners[t] = f; },
      classList: {
        add: (c) => classes.add(c), remove: (c) => classes.delete(c), contains: (c) => classes.has(c),
        toggle: (c, on = !classes.has(c)) => (on ? classes.add(c) : classes.delete(c), on),
      },
    };
    let text = '';
    Object.defineProperty(e, 'textContent', { get: () => text, set: (v) => { text = v; if (v === '') e.children = []; } });
    Object.defineProperty(e, 'className', { get: () => [...classes].join(' '), set: (v) => { classes.clear(); String(v).split(/\s+/).filter(Boolean).forEach((c) => classes.add(c)); } });
    return e;
  },
};
const walk = (e, f) => { f(e); (e.children || []).forEach((c) => walk(c, f)); };
const texts = (e) => { const out = []; walk(e, (x) => { if (x.textContent && !x.children.length) out.push(x.textContent); }); return out; };

const members = [
  { name: 'Ragnar', level: 10, class: 'Warrior', hp: 100, end: 80, zone: 100, zoneName: 'Vale of Mularn', inRange: true },
  { name: 'Sigrid', level: 9, class: 'Warrior', hp: 40, end: 100, zone: 105, zoneName: 'Myrkwood Forest', inRange: false, you: true },
  { name: 'Bjorn', gone: true },
];

// Rows: fractions for the bars, the leader first, and a reason when dim.
const rows = groupRows(members);
assert.equal(rows.length, 3);
assert.equal(rows[0].leader, true);
assert.equal(rows[0].dim, false);
assert.equal(rows[1].hp, 0.4);
assert.equal(rows[1].where, 'in Myrkwood Forest');
assert.equal(rows[1].you, true);
assert.equal(rows[2].where, 'not in the game');
assert.equal(rows[2].hp, 0);
assert.match(rows[0].title, /Ragnar \(leader\)/);
assert.deepEqual(groupRows([]), []);
assert.equal(groupRows(Array.from({ length: 10 }, (_, i) => ({ name: 'P' + i }))).length, 8, 'never more than eight rows');
console.log('ok: group rows');

// Chat lines by channel.
assert.deepEqual(chatLine({ ch: 'tell', from: 'Ragnar', text: 'hi' }), { text: 'Ragnar sends, "hi"', kind: 'tell' });
assert.deepEqual(chatLine({ ch: 'tell', to: 'Sigrid', text: 'hi' }), { text: 'You send, "hi" to Sigrid.', kind: 'tell' });
assert.deepEqual(chatLine({ ch: 'group', from: 'Sigrid', text: 'pull now' }), { text: '[Group] Sigrid: "pull now"', kind: 'group' });
assert.deepEqual(chatLine({ ch: 'sys', text: 'You are not in a group.' }), { text: 'You are not in a group.', kind: 'sys' });
assert.deepEqual(chatLine({ ch: 'say', from: 'Ragnar', text: 'hello' }), { text: 'Ragnar: hello', kind: '' });
assert.equal(chatLine({ ch: 'all', from: 'Ragnar', text: 'hey' }).kind, 'all');
console.log('ok: chat lines by channel');

// The plain panel lists the members and asks through the handlers.
let invited = 0, left = 0;
const plain = createPlainGroup(doc, { onInvite: () => invited++, onDisband: () => left++ });
plain.update(members);
const shown = texts(plain.el).join('|');
assert.match(shown, /★ Ragnar 10 Warrior/);
assert.match(shown, /Sigrid 9 Warrior \(in Myrkwood Forest\)/);
plain.update([]);
assert.match(texts(plain.el).join('|'), /You are not in a group/);
const buttons = [];
walk(plain.el, (x) => { if (x.tag === 'button') buttons.push(x); });
buttons[0].listeners.click(); buttons[1].listeners.click();
assert.equal(invited, 1);
assert.equal(left, 1);
plain.toggle(true);
assert.equal(plain.isOpen, true);
console.log('ok: plain group panel');

// The plain invitation answers once and closes.
let accepted = 0;
const inv = createInvitePrompt(doc, null, { onAccept: () => accepted++ });
inv.show('Ragnar', 60);
assert.equal(inv.isOpen, true);
assert.match(texts(inv.el).join('|'), /Ragnar has invited you to join a group/);
const ib = [];
walk(inv.el, (x) => { if (x.tag === 'button') ib.push(x); });
ib[0].listeners.click();
assert.equal(accepted, 1);
assert.equal(inv.isOpen, false);
console.log('ok: plain invitation');

// The stock windows, when the converter has been run.
const real = 'web/data/ui/atlantis/layout.json';
if (existsSync(real)) {
  const data = new UIData(JSON.parse(readFileSync(real, 'utf8')), 'data/ui/atlantis/');
  let inv2 = 0, dis2 = 0;
  const g = createStockGroup(doc, data, { onInvite: () => inv2++, onDisband: () => dis2++ });
  assert.equal(g.isOpen, false, 'closed until asked for');
  g.update(members);
  const [big, mini] = g.el.children;
  assert.equal(big.dataset.window, 'StatsGroup');
  assert.equal(mini.dataset.window, 'MiniGroup');
  // INVITE and DISBAND are labelled stock buttons that ask.
  const byControl = {};
  walk(big, (x) => { if (x.dataset && x.dataset.control) byControl[x.dataset.control] = x; });
  for (const id of ['1000', '1001']) {
    const b = byControl[id];
    b.listeners.pointerenter(); b.listeners.pointerdown(); b.listeners.pointerup();
  }
  assert.equal(inv2, 1);
  assert.equal(dis2, 1);
  // MINI swaps to the small window; clicking it swaps back.
  const m = byControl['1002'];
  m.listeners.pointerenter(); m.listeners.pointerdown(); m.listeners.pointerup();
  assert.equal(big.style.display, 'none');
  assert.equal(mini.style.display, 'block');
  mini.listeners.click();
  assert.equal(big.style.display, 'block');
  // The first member's red bar is full, the second's at 40%.
  const bars = [];
  walk(big, (x) => { if (x.style && x.style.overflow === 'hidden') bars.push(x); });
  assert.equal(bars[0].style.width, '72px');
  assert.equal(bars[3].style.width, (72 * 0.4).toFixed(1).replace(/\.0$/, '') + 'px');
  // Emptying the group clears the rows.
  g.update([]);
  assert.equal(bars[0].style.width, '0px');
  console.log('ok: stock group window');
  let yes = 0, no = 0;
  const si = createInvitePrompt(doc, data, { onAccept: () => yes++, onDecline: () => no++ });
  si.show('Ragnar', 60);
  assert.equal(si.isOpen, true);
  const decline = si.el.children[3];
  decline.listeners.pointerenter(); decline.listeners.pointerdown(); decline.listeners.pointerup();
  assert.equal(no, 1);
  assert.equal(yes, 0);
  assert.equal(si.isOpen, false);
  console.log('ok: stock invitation');
} else console.log('skip stock group window (run go run ./cmd/uiconv)');
