// Other players and creatures, as the server describes them.
//
// In network mode nothing here decides anything: the server says where each
// entity is, how fast it is moving, whether it is fighting or dead, and
// what happened this tick. This module turns that into pictures -- loading
// the right converted character for each (a creature by its model number, a
// player by its character), choosing the clip that matches how it moves, and
// playing swings, flinches and deaths as the server's events arrive.
//
// It keeps the fields npc.js's pickNPC and the target ring expect (x, y,
// scale, body, type, dead, gone), so clicking and targeting work unchanged.

import { createCharacter, modelMatrix } from './character.js';
import { Animator } from './animator.js';
import { bodyRadius } from './npc.js';

// Beyond this distance from the camera an entity is neither posed nor drawn.
const DRAW_DISTANCE = 14000;
const FADE = 0.3;
const LOCOMOTION = new Set(['walk', 'run', 'back', 'strafeL', 'strafeR']);

// The clip an event plays on the entity it happened to or came from.
export function eventAnim(anim, ev, moving, rng = Math.random) {
  if (!anim) return;
  switch (ev.k) {
    case 'swing': {
      const which = 'attack' + (1 + Math.floor(rng() * 3));
      anim.play(which, { upper: moving }) || anim.play('attack1', { upper: moving });
      break;
    }
    case 'hit':
      if (!anim.acting) anim.play('flinch', { upper: moving });
      break;
    case 'block': case 'parry':
      anim.play('block', { upper: true });
      break;
    case 'evade':
      break;
    default: break;
  }
}

export function createRemote(gl, helpers, { cell = 256 } = {}) {
  const types = new Map(); // key -> { state, ch, walk, run, steps }
  const views = new Map(); // entity id -> view
  let drawn = 0;

  // The directory a character is converted into: creatures by model number.
  const keyOf = (e) => (e.k === 'p' ? e.char : 'm' + e.model);

  function loadType(key) {
    let t = types.get(key);
    if (t) return t;
    t = { state: 'loading', ch: null };
    types.set(key, t);
    const safe = key.replace(/[^a-zA-Z0-9_-]/g, '');
    createCharacter(gl, 'data/char/' + safe, helpers).then((ch) => {
      const walk = ch.groundSpeed('walk') || 40;
      let run = ch.groundSpeed('run');
      if (!(run > 1.5 * walk && run < 8 * walk)) run = 3 * walk;
      const steps = {};
      for (const n of ['back', 'strafeL', 'strafeR']) if (ch.clips[n]) steps[n] = ch.groundSpeed(n);
      Object.assign(t, { state: 'ok', ch, walk, run, steps });
    }).catch((e) => {
      t.state = 'fail';
      console.warn('character ' + key + ' not loaded:', e.message);
    });
    return t;
  }

  // sync brings the views in line with the client's entities.
  function sync(net) {
    for (const e of net.ents.values()) {
      let v = views.get(e.id);
      if (!v) {
        v = { e, id: e.id, anim: null, type: null, x: e.x, y: e.y, yaw: e.yaw, scale: 1, body: 40,
          dead: false, gone: false, deathPlayed: false, moving: false };
        views.set(e.id, v);
      }
      v.e = e;
      v.x = e.x; v.y = e.y; v.yaw = e.yaw;
      v.name = e.name; v.level = e.level; v.hp = e.hp; v.maxHp = 1;
      v.dead = e.state === 'dead';
      v.fight = e.state === 'fight';
      v.isPlayer = e.k === 'p';
      v.friendly = e.friendly;
      v.guild = e.guild || '';
      if (!v.type || v.type.state !== 'ok') {
        const t = loadType(keyOf(e));
        if (t.state === 'ok' && !v.failed) {
          v.type = { ch: t.ch, ...t };
          v.anim = new Animator(t.ch, { fade: FADE, carry: LOCOMOTION });
          v.scale = t.ch.scale * (e.scale || 1);
          v.body = bodyRadius(t.ch, v.scale);
        }
      }
    }
    for (const id of views.keys()) if (!net.ents.has(id)) views.delete(id);
  }

  // events plays the server's events on the views they concern. selfId is
  // the local player, whose own animation the viewer handles.
  function events(list, selfId) {
    for (const ev of list) {
      if (ev.from && ev.from !== selfId) {
        const v = views.get(ev.from);
        if (v && v.anim) eventAnim(v.anim, ev, v.e.speed > 5);
      }
      if (ev.to && ev.to !== selfId && (ev.k === 'hit')) {
        const v = views.get(ev.to);
        if (v && v.anim) eventAnim(v.anim, ev, v.e.speed > 5);
      }
    }
  }

  function update(dt) {
    for (const v of views.values()) {
      const a = v.anim;
      if (!a) continue;
      const e = v.e, t = v.type;
      if (v.dead) {
        if (!v.deathPlayed) { a.play('death', { hold: true }); v.deathPlayed = true; }
      } else if (v.deathPlayed) {
        a.clear(); v.deathPlayed = false;
      }
      const sp = e.speed || 0;
      v.moving = sp > 5;
      let base = 'idle';
      if (v.moving && !v.dead) {
        if (e.back && a.has('back')) base = 'back';
        else if (e.side && a.has(e.side > 0 ? 'strafeR' : 'strafeL')) base = e.side > 0 ? 'strafeR' : 'strafeL';
        else base = sp > t.walk * v.scale * 1.7 ? 'run' : 'walk';
        if (!a.has(base)) base = 'walk';
      } else if (v.fight && a.has('cidle')) {
        base = 'cidle';
      }
      a.setBase(base);
      a.update(dt, (name) => {
        if (name === 'walk') return (e.back && !a.has('back') ? -1 : 1) * Math.max(sp, 1) / (t.walk * v.scale);
        if (name === 'run') return Math.max(sp, 1) / (t.run * v.scale);
        if (t.steps[name] > 5) return Math.max(sp, 1) / (t.steps[name] * v.scale);
        return 1;
      });
    }
  }

  function draw(ctx) {
    drawn = 0;
    const [cx, , cz] = ctx.camPos;
    for (const v of views.values()) {
      if (!v.anim || v.gone || Math.hypot(v.x - cx, v.y - cz) > DRAW_DISTANCE) continue;
      // One creature that cannot be drawn (a skin needing more bones than
      // the shader has, say) must not take the frame down with it: drop
      // that kind and carry on.
      try {
        v.anim.pose();
        v.type.ch.draw({ ...ctx, model: modelMatrix(v.x, v.e.z, v.y, v.yaw, v.scale) });
        drawn++;
      } catch (e) {
        const key = keyOf(v.e);
        const t = types.get(key);
        if (t && t.state === 'ok') { t.state = 'fail'; console.warn('character ' + key + ' cannot be drawn:', e.message); }
        v.anim = null; v.type = null;
      }
    }
  }

  return {
    views, sync, events, update, draw,
    // The creatures a click or Tab can select.
    get pickables() { return [...views.values()].filter((v) => !v.isPlayer && v.anim); },
    get drawn() { return drawn; },
    get types() { return types; },
  };
}
