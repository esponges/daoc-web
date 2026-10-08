// Melee: targeting, auto-attack, health, NPCs that fight back, death.
//
// Deliberately the simplest version of DAoC's model. You select a target and
// switch attack on; while it is on, you swing whenever your swing timer is up,
// the target is in reach and roughly in front of you. Each swing rolls once to
// hit and once for damage, and a shield, if you carry one and face the blow,
// may block it first. No styles, no parry or evade, no levels -- each side
// has hit points, a damage range, a swing time and a hit chance, and the
// NPCs' come from the spawn file.
//
// None of this touches WebGL, so npc.test.mjs and combat.test.mjs can drive it
// headlessly.

import { turnToward, angleOff, TURN_RATE } from './npc.js';

// How far past both bodies' edges a swing reaches, in world units. A unit is
// about an inch, so a little under a metre.
export const REACH = 34;

// How far off your facing a target can be and still be hit: a little either
// side of straight ahead, as DAoC's own front arc is.
const FRONT_ARC = 1.2; // radians, each side

// When in a swing the blow lands, as a fraction of the clip. animnifs.csv has
// a "hit time" column, but every attack clip used here leaves it at zero.
const HIT_AT = 0.45;

// How long a corpse lies before it vanishes, and how long until the NPC
// respawns after that.
const CORPSE_TIME = 15;
const RESPAWN_TIME = 20;

// An NPC chasing further than this from home gives up and walks back.
const LEASH = 2500;

// Social aggro: attacking one of a group brings the others within this
// distance into the fight too.
export const ASSIST = 900;

// Out of combat this long, the player regenerates.
const REGEN_DELAY = 6;
const REGEN_RATE = 0.05; // of max hit points per second

// Tab-targeting looks this far.
const TARGET_RANGE = 3000;

export function distance(a, b) { return Math.hypot(a.x - b.x, a.y - b.y); }

// inReach is whether a can hit b from where they stand.
export function inReach(a, b) { return distance(a, b) <= a.body + b.body + REACH; }

// facing is whether b is inside a's front arc.
export function facing(a, b) {
  return angleOff(Math.atan2(b.x - a.x, -(b.y - a.y)), a.yaw) <= FRONT_ARC;
}

// createCombat ties the player to the NPCs. player needs the same fighter
// fields an NPC has -- x, y, yaw, hp, maxHp, body, damage, swing, hitChance,
// anim -- plus name. npcs is what createNPCs returned. log(text, kind) shows a
// line of combat text; onPlayerDeath is called once the death has played out.
export function createCombat({ player, npcs, log = () => {}, onPlayerDeath = () => {}, rng = Math.random }) {
  const st = {
    target: null,
    attacking: false,
    swingTimer: 0,
    pending: [], // blows in flight: { at, from, to }
    sinceCombat: Infinity,
    deadTimer: 0,
    lastTooFar: -Infinity,
    clock: 0,
  };

  const the = (n) => 'the ' + n.name;

  // --- targeting ---

  // targetNext picks the nearest living NPC in front of the camera, or the
  // next one along on repeated presses, the way Tab cycles in the game.
  function targetNext(viewYaw = player.yaw) {
    const live = npcs.npcs.filter((n) => !n.dead && !n.gone && distance(player, n) < TARGET_RANGE);
    if (!live.length) { st.target = null; return null; }
    // Rank by distance, heavily penalising anything behind the view.
    const score = (n) => {
      const off = angleOff(Math.atan2(n.x - player.x, -(n.y - player.y)), viewYaw);
      return distance(player, n) * (off > Math.PI / 2 ? 4 : 1);
    };
    live.sort((a, b) => score(a) - score(b));
    const i = live.indexOf(st.target);
    st.target = live[(i + 1) % live.length];
    return st.target;
  }

  function clearTarget() { st.target = null; st.attacking = false; }

  // toggleAttack switches auto-attack, picking a target first if there is
  // none, which is what pressing attack with nothing selected does in-game.
  function toggleAttack(viewYaw) {
    if (player.dead) return false;
    if (st.attacking) { st.attacking = false; log('You stop attacking.', 'info'); return false; }
    if (!st.target || st.target.dead) targetNext(viewYaw);
    if (!st.target) { log('You have no target.', 'info'); return false; }
    st.attacking = true;
    log('You prepare to attack ' + the(st.target) + '.', 'info');
    return true;
  }

  // --- blows ---

  // swing starts an attack animation and schedules the blow to land partway
  // through it. moving lays the attack over the upper body only.
  function swing(from, to, moving) {
    const which = 'attack' + (1 + Math.floor(rng() * 3));
    const d = from.anim.play(which, { upper: moving }) || from.anim.play('attack1', { upper: moving }) || 1;
    st.pending.push({ at: st.clock + Math.max(0.2, d * HIT_AT), from, to });
  }

  function land(from, to) {
    if (from.dead || to.dead) return;
    // A blow already thrown still lands if the target stepped just out of
    // reach, but not if it has gone far.
    if (distance(from, to) > from.body + to.body + REACH * 2.5) return;
    const you = from === player, them = to === player;
    // A shield stops a blow outright, but only one it can see coming.
    if (to.blockChance && facing(to, from) && rng() < to.blockChance) {
      to.anim.play('block', { upper: true });
      log(them ? 'You block ' + the(from) + '\'s attack!' : cap(the(to)) + ' blocks your attack!',
        them ? 'block' : 'miss');
      if (!them) provoke(to);
      return;
    }
    if (rng() > from.hitChance) {
      log(you ? 'You miss ' + the(to) + '!' : cap(the(from)) + ' misses you!', you ? 'miss' : 'miss-in');
      return;
    }
    const [lo, hi] = from.damage;
    const dmg = Math.round(lo + rng() * (hi - lo));
    to.hp = Math.max(0, to.hp - dmg);
    log(you ? 'You hit ' + the(to) + ' for ' + dmg + ' damage!'
      : cap(the(from)) + ' hits you for ' + dmg + ' damage!', you ? 'hit' : 'hit-in');
    if (to.hp <= 0) return die(to, from);
    // Flinch, unless it is busy swinging itself.
    if (!to.anim.acting) to.anim.play('flinch', { upper: them && player.moving });
    if (!them) provoke(to);
  }

  // provoke pulls an NPC, and any of its group close by, into a fight.
  function provoke(n) {
    if (n.dead || player.dead) return;
    for (const m of npcs.npcs) {
      if (m.dead || m.fight) continue;
      if (m === n || (m.group === n.group && distance(m, n) < ASSIST)) {
        m.fight = player;
        m.swingTimer = 0.4 + rng() * 0.6;
        m.returning = false;
      }
    }
  }

  function die(who, killer) {
    who.dead = true;
    who.hp = 0;
    who.anim.play('death', { hold: true });
    st.pending = st.pending.filter((p) => p.from !== who && p.to !== who);
    if (who === player) {
      log('You have been killed by ' + the(killer) + '.', 'death');
      st.attacking = false;
      st.deadTimer = 5;
      for (const n of npcs.npcs) if (n.fight === player) disengage(n);
    } else {
      log(cap(the(who)) + ' dies!', 'kill');
      who.fight = null;
      who.corpse = CORPSE_TIME;
      if (st.target === who) st.attacking = false;
    }
  }

  // disengage sends an NPC home, where it heals.
  function disengage(n) {
    n.fight = null;
    n.returning = true;
    npcs.wanderTo(n, n.hx, n.hy);
  }

  // --- per frame ---

  // update runs the fight. moving says whether the player is moving this
  // frame, which decides whether an attack is full-body or upper-body.
  function update(dt, { moving = false } = {}) {
    st.clock += dt;
    player.moving = moving;

    // Blows land on their own schedule.
    const due = st.pending.filter((p) => p.at <= st.clock);
    st.pending = st.pending.filter((p) => p.at > st.clock);
    for (const p of due) land(p.from, p.to);

    // --- the player ---
    if (player.dead) {
      st.deadTimer -= dt;
      if (st.deadTimer <= 0 && player.dead) {
        player.dead = false;
        player.hp = player.maxHp;
        player.anim.clear();
        st.target = null;
        onPlayerDeath();
        log('You have been released from death.', 'info');
      }
    } else {
      const t = st.target;
      if (t && (t.dead || t.gone)) {
        if (t.gone || st.attacking) st.attacking = false;
        if (t.gone) st.target = null;
      }
      st.swingTimer -= dt;
      if (st.attacking && t && !t.dead) {
        // Standing still, square up to the target as the game does.
        if (!moving) player.yaw = turnToward(player.yaw, Math.atan2(t.x - player.x, -(t.y - player.y)), 8 * dt);
        if (st.swingTimer <= 0) {
          if (!inReach(player, t)) {
            if (st.clock - st.lastTooFar > 2.5) {
              log(cap(the(t)) + ' is too far away to attack!', 'info');
              st.lastTooFar = st.clock;
            }
          } else if (facing(player, t)) {
            swing(player, t, moving);
            st.swingTimer = player.swing;
            provoke(t);
          }
        }
      }
      // Regenerate once nothing has been fighting for a while.
      const engaged = st.attacking || npcs.npcs.some((n) => n.fight === player);
      st.sinceCombat = engaged ? 0 : st.sinceCombat + dt;
      if (st.sinceCombat > REGEN_DELAY && player.hp < player.maxHp) {
        player.hp = Math.min(player.maxHp, player.hp + REGEN_RATE * player.maxHp * dt);
      }
    }

    // --- the NPCs ---
    for (const n of npcs.npcs) {
      if (n.dead) {
        if (!n.gone) {
          n.corpse -= dt;
          if (n.corpse <= 0) { n.gone = true; n.respawnIn = RESPAWN_TIME; }
        } else {
          n.respawnIn -= dt;
          if (n.respawnIn <= 0) npcs.respawn(n);
        }
        continue;
      }
      // An aggressive NPC notices a living player who comes too close.
      if (!n.fight && !n.returning && n.aggro > 0 && !player.dead && distance(n, player) < n.aggro) {
        log(cap(the(n)) + ' attacks you!', 'hit-in');
        provoke(n);
      }
      if (n.fight !== player) continue;
      if (player.dead || Math.hypot(n.x - n.hx, n.y - n.hy) > LEASH + n.radius) {
        disengage(n);
        continue;
      }
      n.swingTimer = (n.swingTimer ?? 0) - dt;
      if (!inReach(n, player)) {
        // Close in, aiming just short of the player rather than into them.
        const d = distance(n, player), stop = n.body + player.body + REACH * 0.6;
        const k = Math.max(0, (d - stop) / d);
        npcs.step(n, n.x + (player.x - n.x) * k, n.y + (player.y - n.y) * k, n.runSpeed, dt);
        n.anim.setBase(n.anim.has('run') ? 'run' : 'walk');
      } else {
        n.yaw = turnToward(n.yaw, Math.atan2(player.x - n.x, -(player.y - n.y)), TURN_RATE * dt);
        n.anim.setBase(n.anim.has('cidle') ? 'cidle' : 'idle');
        if (n.swingTimer <= 0 && facing(n, player)) {
          swing(n, player, false);
          n.swingTimer = n.swing;
        }
      }
    }
  }

  return {
    state: st,
    get target() { return st.target; },
    get attacking() { return st.attacking; },
    targetNext,
    clearTarget,
    toggleAttack,
    update,
  };
}

function cap(s) { return s.charAt(0).toUpperCase() + s.slice(1); }
