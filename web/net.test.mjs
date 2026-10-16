// Headless checks of the movement maths and the client's network prediction.
//   node web/net.test.mjs
//
// The movement vectors are written by the Go tests
// (go test ./internal/server/world -update); this checks the browser's
// stepBody lands on the same numbers as the server's Step.

import { readFileSync } from 'node:fs';
import { stepBody, runKeys, planeGround, STEP_DT } from './move.js';
import { NetClient } from './net.js';

let failed = 0;
const ok = (cond, msg) => { if (!cond) { failed++; console.error('FAIL: ' + msg); } else console.log('ok: ' + msg); };

// --- shared vectors ---
const vectors = JSON.parse(readFileSync(new URL('../internal/server/world/testdata/move_vectors.json', import.meta.url)));
for (const v of vectors) {
  const b = { x: v.start[0], y: v.start[1], z: v.start[2], yaw: v.start[3] };
  for (const r of v.runs) {
    const { count, keys } = runKeys(r);
    for (let i = 0; i < count; i++) stepBody(planeGround, b, keys);
  }
  const got = [b.x, b.y, b.z, b.yaw];
  const worst = Math.max(...got.map((g, i) => Math.abs(g - v.end[i])));
  ok(worst < 1e-6, `vector "${v.name}" agrees with the server (worst difference ${worst.toExponential(1)})`);
}

// --- a fake server and a virtual clock ---
class FakeWS {
  constructor(sim) { this.sim = sim; this.readyState = 1; this.sentLog = []; }
  send(s) { this.sim.toServer(JSON.parse(s)); }
  close() {}
}

function makeSim({ latency = 0.1, ground = planeGround } = {}) {
  const sim = {
    t: 0, latency, queueToServer: [], queueToClient: [],
    srv: { x: 30000, y: 30000, z: 0, yaw: 0, applied: 0, backlog: [] },
    nextSnap: 0, tick: 0, fix: false,
    toServer(m) { this.queueToServer.push({ at: this.t + this.latency, m }); },
    toClient(m) { this.queueToClient.push({ at: this.t + this.latency, m }); },
  };
  sim.ws = new FakeWS(sim);
  sim.client = new NetClient({ ws: sim.ws, ground, now: () => sim.t });
  // The server, 20 Hz: apply three steps a tick from what it has been sent.
  sim.serverTick = () => {
    for (let i = 0; i < 3 && sim.srv.backlog.length; i++) {
      const k = sim.srv.backlog.shift();
      stepBody(ground, sim.srv, k);
      sim.srv.applied++;
    }
    sim.tick++;
    sim.toClient({
      t: 'snap', tick: sim.tick, ack: sim.srv.applied,
      you: { x: sim.srv.x, y: sim.srv.y, z: sim.srv.z, h: sim.srv.yaw, hp: 100, mhp: 100, end: 100, mend: 100, lv: 1, xp: 0, xpn: 1, mo: 0, tg: 0, at: false, fix: sim.fix, sp: 0 },
      ents: sim.ents || [],
    });
    sim.fix = false;
  };
  sim.advance = (dt) => {
    sim.t += dt;
    while (sim.queueToServer.length && sim.queueToServer[0].at <= sim.t) {
      const { m } = sim.queueToServer.shift();
      if (m.t === 'input') {
        for (const r of m.steps) { const { count, keys } = runKeys(r); for (let i = 0; i < count; i++) sim.srv.backlog.push(keys); }
      }
    }
    while (sim.nextSnap <= sim.t) { sim.nextSnap += 0.05; sim.serverTick(); }
    while (sim.queueToClient.length && sim.queueToClient[0].at <= sim.t) {
      sim.ws.onmessage({ data: JSON.stringify(sim.queueToClient.shift().m) });
    }
  };
  // The welcome.
  sim.toClient({ t: 'welcome', v: 1, you: 7, name: 'Test', char: 'troll-warrior', zone: 105, x: 30000, y: 30000, z: 0, yaw: 0, level: 1 });
  sim.advance(sim.latency + 0.001);
  return sim;
}

// --- prediction under latency ---
{
  const sim = makeSim({ latency: 0.1 });
  const c = sim.client;
  ok(c.me && c.me.name === 'Test', 'welcome creates the player');
  let maxJump = 0, last = { ...c.shown };
  const frame = 1 / 60;
  // Run north 1.5 s, turn right while running 1 s, strafe 0.5 s, stop 1 s.
  const script = [[90, { fwd: 1, strafe: 0, turn: 0, run: true }], [60, { fwd: 1, strafe: 0, turn: 1, run: true }],
    [30, { fwd: 0, strafe: 1, turn: 0, run: true }], [60, { fwd: 0, strafe: 0, turn: 0, run: false }]];
  for (const [n, keys] of script) {
    for (let i = 0; i < n; i++) {
      sim.advance(frame);
      c.update(frame, keys);
      const s = c.shown;
      maxJump = Math.max(maxJump, Math.hypot(s.x - last.x, s.y - last.y));
      last = { ...s };
    }
  }
  for (let i = 0; i < 60; i++) { sim.advance(frame); c.update(frame, { fwd: 0, strafe: 0, turn: 0, run: false }); }
  const d = Math.hypot(c.me.x - sim.srv.x, c.me.y - sim.srv.y);
  ok(d < 0.5, `after the dust settles prediction and server differ by ${d.toFixed(3)} units`);
  ok(sim.srv.applied === c.sent, `server applied every step sent (${sim.srv.applied}/${c.sent})`);
  ok(maxJump < 191 / 60 * 1.6, `the shown position never jumps (largest frame step ${maxJump.toFixed(2)} units)`);
  ok(c.history.length === 0, 'acknowledged steps are forgotten');
}

// --- a correction snaps ---
{
  const sim = makeSim({ latency: 0.05 });
  const c = sim.client;
  for (let i = 0; i < 30; i++) { sim.advance(1 / 60); c.update(1 / 60, { fwd: 1, strafe: 0, turn: 0, run: true }); }
  for (let i = 0; i < 30; i++) { sim.advance(1 / 60); c.update(1 / 60, { fwd: 0, strafe: 0, turn: 0, run: false }); }
  // The server moves the player somewhere else, as a teleport or rejection would.
  sim.srv.x += 5000; sim.srv.y += 100;
  sim.fix = true;
  for (let i = 0; i < 20; i++) { sim.advance(1 / 60); c.update(1 / 60); }
  ok(Math.hypot(c.me.x - sim.srv.x, c.me.y - sim.srv.y) < 0.5, 'a server correction moves the player to where the server says');
  ok(c.err.x === 0 && c.err.y === 0, 'a snap leaves no error to ease');
}

// --- remote entities are interpolated ---
{
  const sim = makeSim({ latency: 0.02 });
  const c = sim.client;
  let wx = 31000;
  const xs = [];
  for (let i = 0; i < 180; i++) {
    // A wolf walking east at 85 u/s.
    wx = 31000 + sim.t * 85;
    sim.ents = [{ id: 5, k: 'n', nm: 'wolf', m: 56, x: wx, y: 30000, z: 0, h: Math.PI / 2, sp: 85, hp: 1, lv: 3 }];
    sim.advance(1 / 60);
    c.update(1 / 60);
    if (sim.t > 1) xs.push(c.ents.get(5).x);
  }
  const e = c.ents.get(5);
  ok(e && e.name === 'wolf' && e.model === 56, 'a creature in a snapshot appears with its name and model');
  let worst = 0;
  for (let i = 1; i < xs.length; i++) worst = Math.max(worst, Math.abs(xs[i] - xs[i - 1] - 85 / 60));
  ok(worst < 0.6, `the wolf slides smoothly (worst frame-to-frame deviation ${worst.toFixed(3)})`);
  ok(Math.abs(e.x - (wx - 85 * 0.12)) < 4, `it is drawn about 120 ms in the past (${(wx - e.x).toFixed(1)} units behind)`);
  sim.ents = [];
  for (let i = 0; i < 10; i++) { sim.advance(1 / 60); c.update(1 / 60); }
  ok(!c.ents.has(5), 'an entity that leaves the snapshot is dropped');
}

// --- a player's look ---
{
  const sim = makeSim({ latency: 0.02 });
  const c = sim.client;
  const player = (extra) => ({ id: 9, k: 'p', nm: 'Grimm', ch: 'troll-warrior', rc: 6, fm: true, x: 30100, y: 30000, z: 0, h: 0, hp: 1, lv: 4, ...extra });
  const settle = () => { for (let i = 0; i < 10; i++) { sim.advance(1 / 60); c.update(1 / 60); } };
  sim.ents = [player({ eq: [[10, 311, 0, 0]] })];
  settle();
  let e = c.ents.get(9);
  ok(e.char === 'troll-warrior' && e.level === 4, 'a player keeps its outfit and level');
  ok(e.race === 6 && e.female === true, 'a player carries its race and gender');
  ok(JSON.stringify(e.eq) === '[[10,311,0,0]]', 'a player\'s gear arrives with it');
  sim.ents = [player({})]; // later snapshots do not repeat it
  settle();
  ok(JSON.stringify(c.ents.get(9).eq) === '[[10,311,0,0]]', 'gear is kept while the player stays in view');
  sim.ents = [player({ gs: true, eq: [[10, 444, 0, 0]] })];
  settle();
  ok(JSON.stringify(c.ents.get(9).eq) === '[[10,444,0,0]]', 'changed gear replaces the old');
  sim.ents = [player({ gs: true })]; // took everything off: present and empty
  settle();
  ok(Array.isArray(c.ents.get(9).eq) && c.ents.get(9).eq.length === 0, 'gear can change to nothing');
  sim.ents = [];
  settle();
  sim.toClient({ t: 'inv', items: [], mo: 0, eq: [[11, 473, 0, 0]] });
  settle();
  ok(JSON.stringify(c.eq) === '[[11,473,0,0]]', 'your own gear arrives with the inventory');
  sim.toClient({ t: 'inv', items: [], mo: 0 });
  settle();
  ok(Array.isArray(c.eq) && c.eq.length === 0, 'an inventory from an older server leaves you with no gear');
}

// --- shops ---
{
  const sim = makeSim();
  const c = sim.client;
  const sent = [];
  const send = sim.ws.send.bind(sim.ws);
  sim.ws.send = (s) => { sent.push(JSON.parse(s)); send(s); };
  c.interact(900); c.buy('hammer'); c.sell(3);
  ok(JSON.stringify(sent) === JSON.stringify([{ t: 'interact', id: 900 }, { t: 'buy', id: 'hammer' }, { t: 'sell', slot: 3 }]), 'interact, buy and sell are sent as the protocol names them');
  let opened = null, closed = 0;
  c.on.shop = (m) => { opened = m; };
  c.on.shopclose = () => { closed++; };
  sim.toClient({ t: 'shop', npc: 900, nm: 'Smith', items: [{ slot: 0, id: 'hammer', nm: 'bronze hammer', pr: 1250 }] });
  sim.advance(sim.latency + 0.001);
  ok(opened && opened.npc === 900 && c.shop && c.shop.items[0].pr === 1250, 'a shop message opens the shop and is kept');
  sim.toClient({ t: 'shopclose' });
  sim.advance(sim.latency + 0.001);
  ok(closed === 1 && c.shop === null, 'shopclose closes it');
  sim.toClient({ t: 'shop', npc: 900, nm: 'Smith', items: [] });
  sim.advance(sim.latency + 0.001);
  sim.toClient({ t: 'zone', zone: 101 });
  sim.advance(sim.latency + 0.001);
  ok(c.shop === null, 'changing zone leaves no shop open');
  sim.ents = [{ id: 900, k: 'n', nm: 'Smith', ik: 'merchant', fr: true, x: 30000, y: 30000, z: 0, h: 0, hp: 1 }, { id: 901, k: 'n', nm: 'Guard', fr: true, x: 30100, y: 30000, z: 0, h: 0, hp: 1 }];
  for (let i = 0; i < 10; i++) { sim.advance(1 / 60); c.update(1 / 60); }
  ok(c.ents.get(900)?.kind === 'merchant' && c.ents.get(901)?.kind === '', 'an entity carries what interacting with it does');
}

// --- robustness ---
{
  const sim = makeSim();
  const c = sim.client;
  sim.ws.onmessage({ data: '{"t":"from-the-future","x":1}' });
  sim.ws.onmessage({ data: 'not json' });
  ok(c.me !== null, 'unknown and malformed messages are ignored');
  let errored = null;
  c.on.err = (m) => { errored = m; };
  sim.ws.onmessage({ data: '{"t":"err","code":"style","text":"not enough endurance"}' });
  ok(errored && errored.code === 'style', 'server errors reach the UI');
}

console.log(failed ? `\n${failed} FAILED` : '\nall passed');
process.exit(failed ? 1 : 0);
