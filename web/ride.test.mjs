// node web/ride.test.mjs — a rider's client: keys do nothing, the server's
// position is taken, others riding are marked.
import { NetClient } from './net.js';
import { planeGround } from './move.js';

let failed = 0;
const ok = (cond, msg) => { if (!cond) { failed++; console.error('FAIL: ' + msg); } else console.log('ok: ' + msg); };

let t = 0;
const sent = [];
const c = new NetClient({ ws: { readyState: 1, send: (s) => sent.push(s), close() {} }, ground: planeGround, now: () => t });
c._welcome({ zone: 100, you: 1, name: 'Rider', char: 'troll-warrior', x: 30000, y: 30000, z: 0, yaw: 0, level: 5 });
const you = (x, rd) => ({ x, y: 30000, z: 0, h: 1.5, hp: 100, mhp: 100, end: 100, mend: 100, lv: 5, xp: 0, xpn: 1, mo: 0, tg: 0, at: false, fix: true, sp: rd ? 1000 : 0, rd });
c._snap({ tick: 1, ack: 0, you: you(30100, true), ents: [{ id: 5, k: 'p', nm: 'Friend', x: 30200, y: 30000, z: 0, h: 1.5, hp: 1, sp: 1000, rd: true }, { id: 6, k: 'p', nm: 'Walker', x: 30300, y: 30000, z: 0, h: 1.5, hp: 1, sp: 100 }] });
t = 0.5;
const m = c.update(0.25, { fwd: 1, strafe: 0, turn: 0, run: true });
ok(c.me.x === 30100 && c.me.ride, 'a rider stays where the server put them, whatever keys are held');
ok(c.sent === 0 && !sent.some((s) => s.includes('"input"')), 'a rider sends no movement');
ok(m.speed === 191, 'a rider is drawn running');
ok(c.ents.get(5).ride === true && !c.ents.get(6).ride, 'other riders are marked, walkers are not');
c._snap({ tick: 2, ack: 0, you: you(31200, false), ents: [] });
t = 0.8;
c.update(0.25, { fwd: 1, strafe: 0, turn: 0, run: true });
ok(!c.me.ride && c.sent > 0, 'after the ride the keys move them again');
process.exit(failed ? 1 : 0);
