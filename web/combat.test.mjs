// Headless checks on fighting: the real spawn file, the real characters and
// their converted combat clips, the player as a plate-armoured troll, and
// WebGL stubbed out. The player is the troll warrior, hammer and shield.
//
// Run: node web/combat.test.mjs
//
// It plays out the fights the viewer would: picking a target, closing in and
// killing it, a pack joining in, the player being killed, NPCs giving up and
// going home, corpses vanishing and respawning. Underneath that it checks the
// two animation features combat needed -- clips that play once and hold, and
// an attack laid over the upper body only.

import { here, gl, helpers, check, finish } from './testkit.mjs';
import { createNPCs, bodyRadius, pickNPC } from './npc.js';
import { createCharacter } from './character.js';
import { createCombat, inReach, distance, ASSIST } from './combat.js';
import { Animator } from './animator.js';
import { makePose } from './skeleton.js';

void here;
const cell = 256, extent = 65536;
const world = { cell, extent, groundAt: () => 0, isWet: () => false };

async function setup() {
  const npcs = await createNPCs(gl, 'spawns/zone100.json', helpers, world);
  const char = await createCharacter(gl, 'data/char/troll-warrior', helpers);
  const anim = new Animator(char);
  const player = {
    // The viewer's numbers for a hammer and shield.
    name: 'you', x: 0, y: 0, yaw: 0, hp: 220, maxHp: 220, damage: [15, 27], swing: 3.0,
    blockChance: 0.2,
    hitChance: 0.85, body: bodyRadius(char, char.scale), dead: false, anim,
  };
  const lines = [];
  let released = 0;
  // A fixed sequence of rolls, so a failure reproduces.
  let s = 7;
  const rng = () => ((s = (s * 16807) % 2147483647) / 2147483647);
  const combat = createCombat({
    player, npcs, rng,
    log: (t, k) => lines.push([k, t]),
    onPlayerDeath: () => { released++; player.x = 0; player.y = 0; },
  });
  return { npcs, char, anim, player, combat, lines, released: () => released };
}

// run steps the world, with the player walking up to its target when
// chase is set, as someone at the keyboard would.
function run(w, seconds, { chase = false, until = () => false } = {}) {
  const dt = 1 / 30;
  for (let i = 0; i < seconds * 30; i++) {
    let moving = false;
    const t = w.combat.target;
    if (chase && t && !t.dead && !w.player.dead && !inReach(w.player, t)) {
      const d = distance(w.player, t), s = Math.min(d, 191 * dt);
      w.player.x += ((t.x - w.player.x) / d) * s;
      w.player.y += ((t.y - w.player.y) / d) * s;
      w.player.yaw = Math.atan2(t.x - w.player.x, -(t.y - w.player.y));
      moving = true;
    }
    w.npcs.update(dt);
    w.combat.update(dt, { moving });
    w.anim.setBase(moving ? 'run' : 'idle');
    w.anim.update(dt);
    if (until()) return i * dt;
  }
  return seconds;
}

console.log('one-shot clips and layering');
{
  const w = await setup();
  const skel = w.char.skeleton, clips = w.char.clips;
  check('death is marked play-once and idle loops', clips.death.loop === false && clips.idle.loop === true);
  // A play-once clip past its end is its last frame, not its first again.
  const n = skel.bones.length, a = makePose(n), b = makePose(n);
  skel.samplePose(clips.death, clips.death.duration, a);
  skel.samplePose(clips.death, clips.death.duration + 1.7, b);
  let diff = 0;
  for (let i = 0; i < a.rot.length; i++) diff = Math.max(diff, Math.abs(a.rot[i] - b.rot[i]));
  check('a death holds its last frame', diff < 1e-6, `pose change past the end ${diff.toExponential(1)}`);

  // An attack laid over the upper body moves the arms and leaves the legs
  // exactly where the run put them.
  const thigh = skel.boneId('Bip01 L Thigh'), arm = skel.boneId('Bip01 R UpperArm');
  skel.poseCross(null, 0, clips.run, 0.3, 1);
  const legRun = [...skel.world[thigh]], armRun = [...skel.world[arm]];
  skel.poseLayered(null, 0, clips.run, 0.3, 1, clips.attack1, 0.4, 1, skel.upperBodyMask());
  const legDiff = Math.max(...legRun.map((v, i) => Math.abs(v - skel.world[thigh][i])));
  const armDiff = Math.max(...armRun.map((v, i) => Math.abs(v - skel.world[arm][i])));
  check('an upper-body attack leaves the running legs alone', legDiff < 1e-4, `thigh moved ${legDiff.toExponential(1)}`);
  check('and moves the arms', armDiff > 1, `arm moved ${armDiff.toFixed(1)}u`);
  const mask = skel.upperBodyMask();
  check('the upper body is the torso, arms and head, never the legs',
    mask[skel.boneId('Bip01 Head')] === 1 && mask[skel.boneId('Bip01 R Hand')] === 1 &&
    mask[skel.boneId('Bip01 Pelvis')] === 0 && mask[skel.boneId('Bip01 L Thigh')] === 0 &&
    mask[skel.boneId('Bip01 L Foot')] === 0);
}

console.log('playback rate');
{
  const w = await setup();
  const ch = w.char;
  check('the troll idle plays at the rate the game gives it', Math.abs(ch.clipRate('idle') - 2 / 15) < 1e-6,
    `${ch.clipRate('idle').toFixed(3)}x`);
  const a = new Animator(ch);
  a.t = 0;
  a.update(1);
  check('so a second of idle advances its clock by 2/15 s', Math.abs(a.t - 2 / 15) < 1e-6, a.t.toFixed(4) + 's');
  a.setBase('walk'); a.fade = 1; a.prev = null;
  const t0 = a.t;
  a.update(1, () => 1.5);
  check('a stride cycle follows the caller\'s ground-speed rate instead', Math.abs(a.t - t0 - 1.5) < 1e-6);
  check('an attack lasts its clip, at rate 1', Math.abs(a.play('attack1') - ch.clipDuration('attack1')) < 1e-6);
}

console.log('clicking on an NPC');
{
  const w = await setup();
  const ground = () => 0;
  const wolf = w.npcs.npcs.find((n) => n.char === 'large-grey-wolf');
  // A camera 400 units off and 60 up, looking at the wolf's middle: low
  // enough that something stood between the two is genuinely in the way.
  const eye = [wolf.x + 400, 60, wolf.y];
  const mid = [wolf.x, wolf.type.ch.height * wolf.scale * 0.5, wolf.y];
  const dir = mid.map((v, i) => v - eye[i]);
  check('a ray through an NPC picks it', pickNPC(w.npcs.npcs, eye, dir, ground) === wolf);
  const beside = [dir[0], dir[1], dir[2] + 300];
  const other = pickNPC(w.npcs.npcs, eye, beside, ground);
  check('a ray beside it does not', other !== wolf, other ? 'hit ' + other.name : 'nothing');
  const sky = [dir[0], dir[1] + 600, dir[2]];
  check('a ray over its head does not', pickNPC(w.npcs.npcs, eye, sky, ground) !== wolf);
  // Another NPC stood in front takes the click.
  const badger = w.npcs.npcs.find((n) => n.char === 'badger');
  const [bx, by] = [badger.x, badger.y];
  badger.x = wolf.x + 150; badger.y = wolf.y;
  check('the nearer of two in line is picked', pickNPC(w.npcs.npcs, eye, dir, ground) === badger);
  badger.dead = true;
  check('a corpse is not', pickNPC(w.npcs.npcs, eye, dir, ground) === wolf);
  badger.dead = false; badger.x = bx; badger.y = by;
  check('select makes it the target', w.combat.select(wolf) === wolf && w.combat.target === wolf);
}

console.log('hammer and shield');
{
  const w = await setup();
  const ch = w.char, skel = ch.skeleton;
  check('the warrior holds a hammer and carries a shield', ch.armed && ch.shield,
    ch.equip.map((e) => e.slot + ': ' + e.item.manifest.source + ' on ' + e.bone).join(', '));
  check('armed, it swings the one-handed attacks',
    ch.clips.attack1.source.toLowerCase().startsWith('a_h_1s'), ch.clips.attack1.source);
  check('and has a block', !!ch.clips.block && ch.clips.block.loop === false, ch.clips.block?.source);
  // Through a swing the hammer goes wherever the hand's socket goes.
  const right = ch.equip.find((e) => e.slot === 'right');
  const ctx = { viewProj: new Float32Array(16), model: new Float32Array(16), camPos: [0, 0, 0],
    fogColor: [0, 0, 0], fogStart: 1, fogEnd: 2, lightDir: [0, 1, 0] };
  let worst = 0, travel = 0, last = null;
  for (let k = 0; k <= 10; k++) {
    skel.poseClip(ch.clips.attack1, (k / 10) * ch.clips.attack1.duration);
    ch.draw(ctx);
    const held = right.item.skeleton.world[0], socket = skel.world[right.socket];
    worst = Math.max(worst, ...[...held].map((v, i) => Math.abs(v - socket[i])));
    if (last) travel = Math.max(travel, Math.hypot(held[3] - last[0], held[7] - last[1], held[11] - last[2]));
    last = [held[3], held[7], held[11]];
  }
  check('the hammer follows the hand through a swing', worst < 1e-6 && travel > 1,
    `off its socket by ${worst.toExponential(1)}, moves up to ${travel.toFixed(1)}u a step`);
}

console.log('killing a small grey wolf');
{
  const w = await setup();
  const pack = w.npcs.npcs.filter((n) => n.char === 'small-grey-wolf');
  // Stand a little way off the pack and take the nearest.
  w.player.x = pack[0].x + 600; w.player.y = pack[0].y;
  const t = w.combat.targetNext(-Math.PI / 2);
  check('Tab picks a living NPC', t && !t.dead, t ? t.name : 'nothing');
  const t2 = w.combat.targetNext(-Math.PI / 2);
  check('Tab again picks a different one', t2 && t2 !== t, t2 ? t2.name : 'nothing');
  w.combat.targetNext(-Math.PI / 2);
  while (w.combat.target !== t) w.combat.targetNext(-Math.PI / 2);

  w.combat.toggleAttack();
  check('attack switches on', w.combat.attacking);
  const took = run(w, 120, { chase: true, until: () => t.dead });
  check('the wolf dies', t.dead, `after ${took.toFixed(1)}s`);
  const hits = w.lines.filter(([k]) => k === 'hit').length;
  const misses = w.lines.filter(([k]) => k === 'miss').length;
  check('the player hits and misses', hits > 0 && misses >= 0, `${hits} hits, ${misses} misses`);
  check('swings come no faster than the swing timer',
    (hits + misses) <= Math.ceil(took / w.player.swing) + 1, `${hits + misses} swings in ${took.toFixed(1)}s`);
  const joined = pack.filter((n) => n !== t && n.fight === w.player).length;
  check('its packmates join the fight', joined > 0, `${joined} of ${pack.length - 1}`);
  check('the wolves fight back', w.lines.some(([k]) => k === 'hit-in' || k === 'miss-in'),
    `player at ${Math.round(w.player.hp)}/${w.player.maxHp}`);
  check('the corpse plays its death and holds it', t.anim.act === 'death' && t.anim.hold);
  check('attack stops when the target dies', !w.combat.attacking);

  // Corpse lies 15 s, then 20 s to respawn.
  const home = [t.hx, t.hy];
  run(w, 16);
  check('the corpse vanishes', t.gone);
  run(w, 21);
  check('and respawns whole near home', !t.dead && !t.gone && t.hp === t.maxHp &&
    Math.hypot(t.x - home[0], t.y - home[1]) <= t.radius + 1);
}

console.log('walking into a pack of large grey wolves');
{
  const w = await setup();
  const pack = w.npcs.npcs.filter((n) => n.char === 'large-grey-wolf');
  // Walk to within their aggro range, face them and do nothing else.
  w.player.x = pack[0].x + 300; w.player.y = pack[0].y;
  // Facing them: a shield only blocks what comes from in front.
  w.player.yaw = Math.atan2(pack[0].x - w.player.x, -(pack[0].y - w.player.y));
  run(w, 1);
  check('an aggressive wolf attacks unprovoked', pack[0].fight === w.player);
  // Packmates within assist range of it join; any further off do not.
  const near = pack.filter((n) => n === pack[0] || distance(n, pack[0]) < ASSIST);
  const fought = pack.filter((n) => n.fight === w.player);
  check('and brings the packmates in assist range, only them',
    near.every((n) => n.fight === w.player) && fought.length === near.length,
    `${fought.length} of ${pack.length} join, ${near.length} within ${ASSIST}u`);
  const t = run(w, 300, { until: () => w.player.dead });
  check('standing still, the player is killed', w.player.dead, `after ${t.toFixed(1)}s`);
  check('the death is announced', w.lines.some(([k]) => k === 'death'));
  const blocks = w.lines.filter(([k]) => k === 'block').length;
  const taken = w.lines.filter(([k]) => k === 'hit-in' || k === 'miss-in').length + blocks;
  check('the shield blocks some of what comes at it', blocks > 0 && blocks < taken,
    `${blocks} of ${taken} attacks blocked`);
  check('the wolves give up and head home', fought.every((n) => !n.fight && (n.returning || n.dead)));
  run(w, 6);
  check('the player is released whole', !w.player.dead && w.player.hp === w.player.maxHp && w.released() === 1);
  run(w, 90);
  check('the wolves are home and healed', pack.every((n) => n.dead || (!n.returning && n.hp === n.maxHp)));
}

console.log('running away');
{
  const w = await setup();
  const n = w.npcs.npcs.find((m) => m.char === 'badger');
  w.player.x = n.x + 200; w.player.y = n.y;
  while (w.combat.target !== n) w.combat.targetNext();
  w.combat.toggleAttack();
  run(w, 4, { chase: true });
  check('a passive badger fights back once hit', n.fight === w.player);
  // Teleport well past the leash; the badger chases until it hits it.
  w.combat.toggleAttack();
  w.player.x = n.hx + 9000;
  run(w, 60, { until: () => !n.fight });
  check('the badger gives up at its leash', !n.fight && n.returning,
    `${Math.round(Math.hypot(n.x - n.hx, n.y - n.hy))}u from home`);
}

finish();
