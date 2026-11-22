// How spells look in the world: what the caster does, what flies to the target
// and what plays on it, from the game's own tables (docs/formats.md, "How a
// spell looks").
//
// The server says what happened (`cast`, `spell`, `effect`, `resist` events,
// each carrying the spell's client effect as `fx`). effectconv has written the
// effects the world's spells use into web/data/effects: an index tying each
// client effect to its caster animations and its effects, and a pack of their
// models and particle emitters. This file turns the two into pictures:
//
//   - a list of live effect instances, each hung on a bone of a figure for its
//     length of time and dropped when it ends (never more than MAX_LIVE);
//   - the state of a cast's animation, start then held then release;
//   - a projectile's flight from one figure to another.
//
// Without web/data/effects nothing here runs and spells look as they did.
// The pure parts are tested by spellfx.test.mjs; createSpellFx is the GL glue.

import { createProps } from './props.js';
import { createParticles } from './particles.js';

// Beyond this many live effects the oldest is dropped, so a crowd of casters
// cannot swamp the frame.
export const MAX_LIVE = 64;
// Effects further than this from the camera are not drawn.
export const DRAW_DISTANCE = 9000;
// A projectile's speed, units a second, and the shortest and longest it flies.
export const FLIGHT_SPEED = 1800;
export const FLIGHT_MIN = 0.12;
export const FLIGHT_MAX = 0.7;

const createFxBookHas = (index, id) => !!(index.effects && index.effects[id] && index.effects[id].model >= 0);

// createFxBook reads the converter's index.json: spell(fx) is a client effect's
// parts, effect(id) an effect that has a model to draw (else null).
export function createFxBook(index) {
  return {
    spell: (fx) => (index && index.spells && index.spells[fx]) || null,
    effect: (id) => {
      const e = index && index.effects && index.effects[id];
      return e && e.model >= 0 ? e : null;
    },
    // arrow is the effect of a bow shot, {travel, strike}, or null without it.
    arrow: () => (index && index.arrow && createFxBookHas(index, index.arrow.travel) ? index.arrow : null),
  };
}

// castPlan names the three clips of a spell's cast: its start, its held loop
// and its release, as cast<number> after the anims.csv column the table gives;
// null for a part the spell has none of.
export function castPlan(spell) {
  const a = (spell && spell.anims) || [0, 0, 0];
  const name = (n) => (n > 0 ? 'cast' + n : null);
  return { start: name(a[0]), mid: name(a[1]), end: name(a[2]) };
}

// bonePoint is where a bone of a posed figure is in the world, as
// [x, y, height], the order the placements use. skel is the figure's skeleton
// (world[i] is a 3x4 row-major transform, model space Z up), view its place:
// {x, y, z, yaw, scale}, z being the ground height under it. The figure's
// model maps as modelMatrix does: model X and Y spread over the ground and Z
// is up. A bone the figure lacks, or none named, gives the feet, or with
// chest true the middle of the body.
export function bonePoint(skel, name, view, { chest = 0 } = {}) {
  const id = skel && skel.boneId && name && name !== 'rootnode' ? skel.boneId(name) : -1;
  const m = id >= 0 && skel.world ? skel.world[id] : null;
  const px = m ? m[3] : 0, py = m ? m[7] : 0, pz = m ? m[11] : (chest || 0);
  const s = view.scale || 1;
  const c = Math.cos(view.yaw || 0) * s, sn = Math.sin(view.yaw || 0) * s;
  return [c * px - sn * py + view.x, sn * px + c * py + view.y, s * pz + (view.z || 0)];
}

// createFxLive keeps the live effect instances. An instance is
// {effect, model, scale, start, dur, anchor, tag, seed}: start and dur in
// seconds of the frame clock, anchor {id, bone} says what it hangs on, tag
// groups instances so that they can be stopped together.
export function createFxLive(cap = MAX_LIVE) {
  let list = [];
  let seq = 0;
  return {
    add(inst) {
      list.push({ seed: ++seq * 2654435761, ...inst });
      while (list.length > cap) list.shift();
      return list[list.length - 1];
    },
    // update drops what has ended or been stopped and returns what is left.
    update(now) {
      list = list.filter((i) => !i.stopped && now < i.start + i.dur);
      return list;
    },
    // stop ends every instance with a tag now.
    stop(tag) { for (const i of list) if (i.tag === tag) i.stopped = true; },
    get length() { return list.length; },
    get list() { return list; },
  };
}

// createCastSeq runs the animation of one figure's cast. It does not play
// anything: each call returns what to do, or null for nothing. An action is
// {play, hold} or {clear: true}.
//   begin(plan, now, clipOf)  the cast starts; clipOf(name) is the clip's
//                             length in seconds when the figure has it, else 0
//   tick(now, clipOf)         call each frame: moves on from the start to the hold
//   land()                    the spell took effect: the release
//   stop()                    interrupted or failed
// A figure without the spell's own clips holds the generic cast clip, and moves
// onto the spell's when a clip that was still being fetched arrives.
export function createCastSeq() {
  let st = null;
  const have = (name, clipOf) => !!name && clipOf(name) > 0;
  return {
    get active() { return !!st; },
    begin(plan, now, clipOf) {
      st = { plan, stage: 'generic', at: Infinity };
      if (have(plan.start, clipOf)) {
        st.stage = 'start';
        st.at = now + clipOf(plan.start);
        return { play: plan.start, hold: true };
      }
      if (have(plan.mid, clipOf)) {
        st.stage = 'mid';
        return { play: plan.mid, hold: true };
      }
      return { play: 'cast', hold: true };
    },
    tick(now, clipOf) {
      if (!st) return null;
      const { plan } = st;
      if (st.stage === 'start' && now >= st.at) {
        st.stage = 'mid';
        return have(plan.mid, clipOf) ? { play: plan.mid, hold: true } : null;
      }
      if (st.stage === 'generic' && have(plan.mid, clipOf)) {
        st.stage = 'mid';
        return { play: plan.mid, hold: true };
      }
      return null;
    },
    land(clipOf) {
      const plan = st && st.plan;
      st = null;
      if (plan && have(plan.end, clipOf)) return { play: plan.end, hold: false };
      return { clear: true };
    },
    stop() {
      const was = !!st;
      st = null;
      return was ? { clear: true } : null;
    },
  };
}

// flightTime is how long a projectile takes to cover a distance.
export function flightTime(dist) {
  return Math.min(FLIGHT_MAX, Math.max(FLIGHT_MIN, dist / FLIGHT_SPEED));
}

// flightAt is where a projectile is at a time: along the line between its
// ends, [x, y, height], or null before it starts or after it lands.
export function flightAt(f, now) {
  const t = (now - f.start) / f.dur;
  if (t < 0 || t >= 1) return null;
  return [0, 1, 2].map((k) => f.from[k] + (f.to[k] - f.from[k]) * t);
}

// createSpellFx loads the converted effects and returns the player of them, or
// null when they are not there. helpers are the viewer's program and uniforms
// builders, which the particles share.
//
//   play(effectId, anchor, now, dur, tag)  hang an effect on a figure's bone
//   fly(effectId, from, to, now, dur)      send an effect from one point to another
//   draw(ctx, locate)                      draw what is live; locate(anchor) gives
//                                          [x, y, height] of an anchor, or null
export async function createSpellFx(gl, helpers, { base = 'data/effects', fetchJson } = {}) {
  const get = fetchJson || ((u) => fetch(u).then((r) => (r.ok ? r.json() : null)));
  let index = null;
  try { index = await get(base + '/index.json'); } catch { index = null; }
  if (!index || !index.spells) return null;
  let props;
  try { props = await createProps(gl, base + '/pack'); } catch (e) {
    console.warn('spell effects not loaded:', e.message);
    return null;
  }
  const particles = createParticles(gl, props, helpers);
  const book = createFxBook(index);
  const live = createFxLive();
  const flights = [];
  let flown = 0;

  const place = (inst, pos) => {
    const e = book.effect(inst.effect);
    return { m: e.model, p: pos, yaw: inst.yaw || 0, s: (e.scale || 1) * (inst.scale || 1), t0: inst.start, seed: inst.seed };
  };

  return {
    book,
    live,
    // play hangs an effect on an anchor {id, bone} for dur seconds (the table's
    // own length when dur is left out), starting at now. Returns the instance,
    // or null when the effect has no model.
    play(effectId, anchor, now, dur, tag) {
      const e = book.effect(effectId);
      if (!e) return null;
      return live.add({ effect: effectId, anchor, start: now, dur: dur || e.ms / 1000, tag });
    },
    // stats is what is playing now, for checks in the browser.
    get stats() { return { flights: flights.length, flown }; },
    // fly sends an effect from one point to another over dur seconds.
    fly(effectId, from, to, now, dur) {
      if (!book.effect(effectId)) return null;
      // A model flies along the line it is sent: turned to face the way it goes.
      const f = { effect: effectId, from, to, start: now, dur, seed: flights.length + 1, yaw: Math.atan2(to[0] - from[0], -(to[1] - from[1])) };
      flights.push(f);
      flown++;
      return f;
    },
    draw(ctx, locate) {
      if (ctx.shadowPass) return;
      const now = ctx.time;
      const placed = [];
      for (const inst of live.update(now)) {
        // One set to start later (a hit effect waiting for its bolt) is not shown yet.
        if (now < inst.start) continue;
        const pos = locate(inst.anchor) || inst.last;
        if (!pos) continue;
        inst.last = pos;
        if (Math.hypot(pos[0] - ctx.camPos[0], pos[1] - ctx.camPos[2]) > DRAW_DISTANCE) continue;
        placed.push(place(inst, pos));
      }
      for (let i = flights.length - 1; i >= 0; i--) {
        const f = flights[i];
        if (now >= f.start + f.dur) { flights.splice(i, 1); continue; }
        const pos = flightAt(f, now);
        if (pos) placed.push(place({ ...f, scale: 1 }, pos));
      }
      if (!placed.length) { props.setInstances([]); particles.setDynamic([]); return; }
      props.setInstances(placed);
      particles.setDynamic(placed);
      props.draw(ctx);
      particles.draw(ctx);
    },
  };
}

// createDirector turns the server's spell events into what is seen. It holds no
// pictures itself: it asks for a figure to animate and a place to put an
// effect, and tells the effects player what to play.
//   getFx()        the effects player from createSpellFx, or null when the
//                  converted effects are absent (casting then only animates)
//   figure(id)     {anim, ch, moving} of the entity, or null
//   locate(anchor) [x, y, height] of {id, bone}, or null
// event(ev, now) takes each event as it arrives; tick(now) is called each frame.
export function createDirector({ getFx, figure, locate }) {
  const seqs = new Map();   // caster id -> {seq, since}
  const recent = new Map(); // `${from}:${fx}` -> when it last landed
  const clipOf = (ch) => (n) => (ch && ch.clips && ch.clips[n] ? ch.clipDuration(n) / (ch.clipRate(n) || 1) : 0);
  const apply = (fig, act) => {
    if (!act || !fig) return;
    if (act.clear) { if (/^cast/.test(fig.anim.act || '')) fig.anim.clear(); return; }
    fig.anim.play(act.play, { hold: act.hold, upper: !!fig.moving });
  };
  const nodeOf = (fx, id) => { const e = fx && fx.book.effect(id); return e ? e.node : 'rootnode'; };
  const bodyAt = (id) => ({ id, bone: 'rootnode', chest: true });

  function start(ev, now) {
    const fx = getFx();
    const spell = fx && ev.fx ? fx.book.spell(ev.fx) : null;
    const plan = castPlan(spell);
    const fig = figure(ev.from);
    if (fig) {
      // The spell's own clips are fetched the first time anything casts it.
      for (const n of [plan.start, plan.mid, plan.end]) {
        if (n && !(fig.ch.clips && fig.ch.clips[n]) && fig.ch.loadClip) fig.ch.loadClip(n);
      }
      const seq = createCastSeq();
      seqs.set(ev.from, { seq, since: now });
      apply(fig, seq.begin(plan, now, clipOf(fig.ch)));
    }
    if (fx && spell) {
      const len = ev.dur > 0 ? ev.dur + 0.3 : 0;
      for (const id of spell.caster || []) {
        const e = fx.book.effect(id);
        if (e) fx.play(id, { id: ev.from, bone: e.node }, now, len ? Math.min(len, e.ms / 1000) : 0, 'cast:' + ev.from);
      }
    }
  }

  function end(id, now, landed) {
    const fx = getFx();
    const s = seqs.get(id);
    seqs.delete(id);
    if (fx) fx.live.stop('cast:' + id);
    if (!s) return false;
    apply(figure(id), landed ? s.seq.land(clipOf((figure(id) || {}).ch)) : s.seq.stop());
    return true;
  }

  function land(ev, now) {
    const fx = getFx();
    const wasCasting = end(ev.from, now, true);
    const spell = fx && ev.fx ? fx.book.spell(ev.fx) : null;
    if (!spell) return;
    // A tick of a lasting spell lands again and again with no cast of its own:
    // it gets its hit effect, but the bolt does not fly each time.
    const key = ev.from + ':' + ev.fx;
    const tick = !wasCasting && now - (recent.get(key) ?? -Infinity) < 5;
    recent.set(key, now);
    let delay = 0;
    if (spell.projectile && !tick) {
      const e = fx.book.effect(spell.projectile);
      const from = locate({ id: ev.from, bone: nodeOf(fx, (spell.caster || [])[0]) === 'rootnode' ? 'Bip01 R Hand' : nodeOf(fx, spell.caster[0]) });
      const to = locate(bodyAt(ev.to));
      if (e && from && to) {
        delay = flightTime(Math.hypot(to[0] - from[0], to[1] - from[1], to[2] - from[2]));
        fx.fly(spell.projectile, from, to, now, delay);
      }
    }
    const hit = ev.k === 'resist' ? spell.resist : spell.target;
    const e = hit && fx.book.effect(hit);
    if (e) fx.play(hit, { id: ev.to, bone: e.node }, now + delay, 0, 'hit');
  }

  // An arrow flies from the archer's bow hand to the target's body for the
  // time the server gave, hit or miss.
  function shoot(ev, now) {
    const fx = getFx();
    const a = fx && fx.book.arrow();
    if (!a) return;
    const from = locate({ id: ev.from, bone: 'Bip01 L Held' });
    const to = locate(bodyAt(ev.to));
    if (from && to) fx.fly(a.travel, from, to, now, Math.max(FLIGHT_MIN, ev.dur || 0));
  }

  return {
    event(ev, now) {
      switch (ev.k) {
        case 'cast': start(ev, now); break;
        case 'arrow': shoot(ev, now); break;
        case 'spell': case 'effect': case 'resist': land(ev, now); break;
        case 'interrupt': case 'castfail': end(ev.from, now, false); break;
        default: break;
      }
    },
    tick(now) {
      for (const [id, s] of seqs) {
        const fig = figure(id);
        // A cast whose end never came (out of range, gone) is not held forever.
        if (!fig || now - s.since > 20) { seqs.delete(id); continue; }
        apply(fig, s.seq.tick(now, clipOf(fig.ch)));
      }
    },
  };
}
