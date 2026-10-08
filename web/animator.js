// Animation state for one figure: a looping base and a one-shot action.
//
// The base is what the figure is doing -- idle, walk, run, standing ready to
// fight -- and changes between those cross-fade. An action is something it
// does once on top: an attack, a flinch, a death. It fades in, plays through
// and fades back out to whatever the base has become meanwhile, or, for a
// death, holds its last frame for good.
//
// An action can be laid over the upper body only, so a figure can swing while
// its legs keep running.

// How long a change of base takes to cross-fade.
export const BASE_FADE = 0.18;
// How fast an action blends in and out. Quick: a punch that eases in for a
// quarter of a second has already missed.
const ACT_IN = 0.08;
const ACT_OUT = 0.15;

export class Animator {
  // ch is a character from createCharacter. carry names the base clips that
  // are stride cycles, which hand their phase to one another on a change.
  constructor(ch, { base = 'idle', carry = new Set(['walk', 'run']), fade = BASE_FADE } = {}) {
    this.ch = ch;
    this.carry = carry;
    this.fadeTime = fade;
    this.cur = base; this.t = Math.random() * 10;
    this.prev = null; this.prevT = 0; this.fade = 1;
    this.act = null; this.actT = 0; this.actW = 0;
    this.hold = false; this.upper = false;
  }

  has(name) { return !!this.ch.clips[name]; }

  // setBase changes what the figure is doing underneath any action.
  setBase(name) {
    if (name === this.cur || !this.has(name)) return;
    this.prev = this.cur;
    this.prevT = this.t;
    const dOld = this.ch.clipDuration(this.cur), dNew = this.ch.clipDuration(name);
    // Carry the normalised phase between stride cycles, so the swing foot
    // stays the swing foot; anything else starts from its beginning.
    this.t = this.carry.has(name) && this.carry.has(this.cur) && dOld > 0
      ? ((this.t % dOld) / dOld) * dNew : 0;
    this.cur = name;
    this.fade = 0;
  }

  // play starts a one-shot action. hold keeps its last frame (a death);
  // upper confines it to the spine and above. Returns the clip's duration,
  // or 0 if the character has no such clip. The duration is in seconds of
  // real time, after the clip's playback rate.
  play(name, { hold = false, upper = false } = {}) {
    const d = this.ch.clipDuration(name);
    if (!d) return 0;
    // Restarting an action already under way keeps its weight, so a new
    // swing does not pop back to the base for a frame first.
    if (!this.act) this.actW = 0;
    this.act = name; this.actT = 0;
    this.hold = hold; this.upper = upper;
    return d / this.ch.clipRate(name);
  }

  get acting() { return !!this.act; }
  get actionDone() { return !this.act || this.actT >= this.ch.clipDuration(this.act); }

  // clear drops any action at once, for a figure brought back to life.
  clear() { this.act = null; this.actW = 0; this.hold = false; }

  // speed is how fast a clip's clock runs. A stride cycle's is the caller's
  // to set from ground speed, which already keeps the feet planted; anything
  // else runs at the rate the game plays it, which for an idle is a fraction
  // of the speed it was keyed at.
  speed(name, rate) {
    return this.carry.has(name) ? rate(name) : rate(name) * this.ch.clipRate(name);
  }

  // update advances the clocks. rate(name) scales the base clips' speed,
  // so a walk can play faster when the figure is moving faster.
  update(dt, rate = () => 1) {
    this.t += dt * this.speed(this.cur, rate);
    if (this.prev) this.prevT += dt * this.speed(this.prev, rate);
    if (this.fade < 1) {
      this.fade = Math.min(1, this.fade + dt / this.fadeTime);
      if (this.fade >= 1) this.prev = null;
    }
    if (this.act) {
      // The action's clock is in clip time; the fades are in real time.
      const r = this.ch.clipRate(this.act);
      this.actT += dt * r;
      const d = this.ch.clipDuration(this.act);
      const left = (d - this.actT) / r;
      if (this.hold) {
        this.actW = Math.min(1, this.actW + dt / ACT_IN);
      } else if (left <= 0) {
        this.act = null; this.actW = 0;
      } else if (left < ACT_OUT) {
        this.actW = Math.max(0, Math.min(this.actW, left / ACT_OUT));
      } else {
        this.actW = Math.min(1, this.actW + dt / ACT_IN);
      }
    }
  }

  // pose writes this state into the character's shared skeleton. Returns
  // false if the character has no clips, so the caller can fall back.
  pose() {
    return this.ch.poseLayered(this.prev, this.prevT, this.cur, this.t, this.fade,
      this.act, this.actT, this.actW, this.upper);
  }
}
