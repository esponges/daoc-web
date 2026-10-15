// The connection to the game server, and the prediction that hides its
// latency.
//
// The server owns the world. This module sends it what the player holds down
// as runs of fixed 1/60 s movement steps, predicts the result locally with
// the very maths the server runs (move.js), and when a snapshot arrives it
// rewinds to the server's position for the last step it acknowledged and
// replays the steps still in flight. Small differences are eased out; large
// ones, or a correction the server flags, snap.
//
// Other players and creatures arrive in snapshots about twenty times a
// second and are drawn a little in the past, interpolated between the two
// samples around that time, so they move smoothly.
//
// No WebGL here: net.test.mjs drives it headlessly against a fake server.

import { stepBody, STEP_DT } from './move.js';

export const PROTOCOL = 1;

// Remote entities are drawn this far in the past, in seconds.
const INTERP_DELAY = 0.12;
// A prediction error beyond this many units snaps instead of easing.
const SNAP_ERROR = 160;
// How quickly a smoothed error dies away, per second.
const EASE_RATE = 12;
// Inputs go out at least this often, in seconds.
const SEND_EVERY = 0.05;
// A step run is kept moving this long after its last step, for animation.
const MOTION_HOLD = 0.12;

const NO_KEYS = { fwd: 0, strafe: 0, turn: 0, run: false };

export class NetClient {
  // ws is an open or opening WebSocket-like (send, close, onmessage, onopen,
  // onclose). ground is { groundAt, clamp }; mover is optional collision.
  constructor({ ws, ground, mover = null, now = () => performance.now() / 1000 }) {
    this.ws = ws;
    this.ground = ground;
    this.mover = mover;
    this.now = now;
    this.me = null;            // predicted body and the server's vitals
    this.ents = new Map();     // id -> remote entity
    this.events = [];          // server events since the UI last took them
    this.inv = { items: [], mo: 0 };
    this.bar = [];
    this.corpse = null;
    this.zone = 0;
    this.sent = 0;             // steps sent so far
    this.history = [];         // steps sent and not yet acknowledged: { seq, k }
    this.pendingRuns = [];     // runs not yet sent
    this.pendingSteps = 0;
    this.lastSend = 0;
    this.acc = 0;
    this.err = { x: 0, y: 0 }; // prediction error being eased out
    this.motion = { speed: 0, back: false, side: 0 };
    this.motionAt = -1;
    this.tick = 0;
    this.on = {};              // callbacks: welcome, zone, err, chat, inv, bar, corpse, close
    this.ready = new Promise((res, rej) => { this._ready = { res, rej }; });
    this.closed = false;
    ws.onmessage = (e) => this._message(typeof e.data === 'string' ? e.data : '');
    // A tab in the background gets no frames, so nothing else would speak;
    // a ping every so often keeps the server from taking it for gone.
    this._ping = setInterval(() => this.ping(), 15000);
    ws.onclose = () => { clearInterval(this._ping); this.closed = true; this._ready.rej(new Error('connection closed')); this.on.close?.(); };
    ws.onerror = () => {};
  }

  hello(charId) {
    const send = () => this._send({ t: 'hello', v: PROTOCOL, char: charId });
    if (this.ws.readyState === 1) send(); else this.ws.onopen = send;
    return this.ready;
  }

  _send(o) { if (this.ws.readyState === 1) this.ws.send(JSON.stringify(o)); }

  // --- requests ---
  target(id) { this._send({ t: 'target', id }); if (this.me) this.me.target = id; }
  attack(on) { this._send({ t: 'attack', on }); if (this.me) this.me.attacking = on; }
  style(slot) { this._send({ t: 'style', slot }); }
  release() { this._send({ t: 'release' }); }
  loot(corpse, index) { this._send({ t: 'loot', corpse, index }); }
  equip(from, to) { this._send({ t: 'equip', from, to }); }
  chat(ch, text) { this._send({ t: 'chat', ch, text }); }
  gm(line) { this._send({ t: 'gm', line }); }
  ping() { this._send({ t: 'ping' }); }

  // --- incoming ---
  _message(text) {
    let m;
    try { m = JSON.parse(text); } catch { return; }
    switch (m.t) {
      case 'welcome': this._welcome(m); break;
      case 'snap': this._snap(m); break;
      case 'err':
        if (m.code === 'version' || m.code === 'join' || m.code === 'hello' || m.code === 'account') {
          this._ready.rej(new Error(m.text));
        }
        this.on.err?.(m);
        break;
      case 'zone': this.on.zone?.(m); break;
      case 'chatmsg': this.on.chat?.(m); break;
      case 'inv': this.inv = m; this.on.inv?.(m); break;
      case 'bar': this.bar = m.slots || []; this.on.bar?.(this.bar); break;
      case 'corpse': this.corpse = m; this.on.corpse?.(m); break;
      default: break; // an unknown message is ignored
    }
  }

  _welcome(m) {
    this.zone = m.zone;
    this.id = m.you;
    this.me = {
      id: m.you, name: m.name, char: m.char, level: m.level, admin: !!m.admin,
      x: m.x, y: m.y, z: m.z, yaw: m.yaw,
      hp: 1, maxHp: 1, end: 100, maxEnd: 100, xp: 0, xpNext: 1, money: 0,
      target: 0, attacking: false, dead: false, fighting: false, avail: [],
    };
    this._ready.res(m);
    this.on.welcome?.(m);
  }

  _snap(m) {
    const me = this.me;
    if (!me) return;
    this.tick = m.tick;
    const s = m.you;
    // Reconcile: the server's body at the last acknowledged step, then the
    // steps it has not seen yet, replayed on top.
    this.history = this.history.filter((h) => h.seq > m.ack);
    const b = { x: s.x, y: s.y, z: s.z, yaw: s.h };
    for (const h of this.history) stepBody(this.ground, b, h.k, this.mover);
    const ex = me.x - b.x, ey = me.y - b.y;
    if (s.fix || Math.hypot(ex, ey) > SNAP_ERROR) {
      this.err.x = this.err.y = 0;
    } else {
      this.err.x += ex; this.err.y += ey; // shown, then eased away
    }
    me.x = b.x; me.y = b.y; me.z = b.z; me.yaw = b.yaw;
    me.hp = s.hp; me.maxHp = s.mhp; me.end = s.end; me.maxEnd = s.mend;
    me.level = s.lv; me.xp = s.xp; me.xpNext = s.xpn; me.money = s.mo;
    me.target = s.tg; me.attacking = s.at; me.dead = !!s.dead; me.fighting = !!s.fg;
    me.avail = s.av || [];
    me.serverSpeed = s.sp;

    const t = this.now();
    const seen = new Set();
    for (const e of m.ents || []) {
      seen.add(e.id);
      let r = this.ents.get(e.id);
      if (!r) {
        r = { id: e.id, samples: [], x: e.x, y: e.y, z: e.z, yaw: e.h };
        this.ents.set(e.id, r);
      }
      r.k = e.k; r.name = e.nm; r.model = e.m || 0; r.guild = e.gd || ''; r.char = e.ch || ''; r.level = e.lv || 0;
      r.speed = e.sp || 0; r.hp = e.hp; r.state = e.st || ''; r.scale = e.sc || 0;
      r.aggro = !!e.ag; r.loot = !!e.lt; r.friendly = !!e.fr; r.back = !!e.bk; r.side = e.sd || 0;
      r.samples.push({ t, x: e.x, y: e.y, z: e.z, yaw: e.h });
      if (r.samples.length > 8) r.samples.shift();
    }
    for (const id of this.ents.keys()) if (!seen.has(id)) this.ents.delete(id);
    if (m.ev && m.ev.length) this.events.push(...m.ev);
  }

  // takeEvents hands over the events since the last call.
  takeEvents() { const e = this.events; this.events = []; return e; }

  // --- per frame ---

  // update advances prediction by dt seconds under the held keys, sends any
  // inputs due, and moves the remote entities to where they are drawn. It
  // returns what the player's movement looked like, for animation.
  update(dt, keys = NO_KEYS) {
    const me = this.me;
    if (!me) return this.motion;
    const t = this.now();
    this.acc = Math.min(this.acc + dt, 0.25);
    const moving = keys.fwd || keys.strafe || keys.turn;
    while (this.acc >= STEP_DT) {
      this.acc -= STEP_DT;
      if (!moving || me.dead) continue;
      const k = { fwd: keys.fwd, strafe: keys.strafe, turn: keys.turn, run: !!keys.run };
      const b = { x: me.x, y: me.y, z: me.z, yaw: me.yaw };
      const m = stepBody(this.ground, b, k, this.mover);
      me.x = b.x; me.y = b.y; me.z = b.z; me.yaw = b.yaw;
      this.sent++;
      this.history.push({ seq: this.sent, k });
      this._queue(k);
      if (m.speed > 0 || k.turn) { this.motion = m; this.motionAt = t; }
    }
    if (this.pendingSteps > 0 && (t - this.lastSend >= SEND_EVERY || this.pendingSteps >= 6)) this._flush(t);
    if (t - this.motionAt > MOTION_HOLD) this.motion = { speed: 0, back: false, side: 0 };

    // Ease the smoothed error away.
    const f = Math.exp(-EASE_RATE * dt);
    this.err.x *= f; this.err.y *= f;
    if (Math.abs(this.err.x) < 0.05) this.err.x = 0;
    if (Math.abs(this.err.y) < 0.05) this.err.y = 0;

    // Remote entities, drawn in the past.
    const rt = t - INTERP_DELAY;
    for (const r of this.ents.values()) {
      const s = r.samples;
      if (!s.length) continue;
      let a = s[0], b = s[s.length - 1];
      if (rt >= b.t) { a = b; }
      else if (rt > a.t) {
        for (let i = s.length - 1; i > 0; i--) if (s[i - 1].t <= rt) { a = s[i - 1]; b = s[i]; break; }
      } else { b = a; }
      const u = a === b || b.t === a.t ? 0 : Math.min(1, Math.max(0, (rt - a.t) / (b.t - a.t)));
      // A jump too big to be walking (a respawn) is a cut, not a slide.
      if (Math.hypot(b.x - a.x, b.y - a.y) > 1500) { r.x = b.x; r.y = b.y; r.z = b.z; r.yaw = b.yaw; continue; }
      r.x = a.x + (b.x - a.x) * u; r.y = a.y + (b.y - a.y) * u; r.z = a.z + (b.z - a.z) * u;
      let dy = b.yaw - a.yaw;
      dy = ((dy + Math.PI) % (2 * Math.PI) + 2 * Math.PI) % (2 * Math.PI) - Math.PI;
      r.yaw = a.yaw + dy * u;
    }
    return this.motion;
  }

  // The player as drawn: the predicted body plus the error still easing out.
  get shown() {
    const me = this.me;
    return { x: me.x + this.err.x, y: me.y + this.err.y, z: me.z, yaw: me.yaw };
  }

  _queue(k) {
    const flags = k.run ? 1 : 0;
    const last = this.pendingRuns[this.pendingRuns.length - 1];
    if (last && last[1] === k.fwd && last[2] === k.strafe && last[3] === k.turn && last[4] === flags) last[0]++;
    else this.pendingRuns.push([1, k.fwd, k.strafe, k.turn, flags]);
    this.pendingSteps++;
  }

  _flush(t) {
    this._send({ t: 'input', seq: this.sent, steps: this.pendingRuns, px: this.me.x, py: this.me.y });
    this.pendingRuns = [];
    this.pendingSteps = 0;
    this.lastSend = t;
  }
}
