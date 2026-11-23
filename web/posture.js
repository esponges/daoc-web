// Which clip a figure's body is in, apart from walking and fighting: sitting
// down, sitting, standing up, in the air, landing and swimming.
//
// The server says the posture ("sit", "air", "swim" or ""); this turns a
// stream of those into the base clip to play, with the one-shot clips on the
// way in and out (sit down once, then the seated loop; stand up once).
// A figure without a clip for a step goes straight to the next one, so an
// older conversion still sits, as an idle.
//
// One PostureClips per figure. step(dt, { posture, moving }) returns the base
// clip to play, or null when locomotion should choose.

// A landing is a short thing: the clip is long, but the next step cuts it.
export const LAND_MAX = 0.45;

// The clips a figure should have loaded to be drawn in any posture.
export const POSTURE_CLIPS = ['jump', 'land', 'sitdown', 'sit', 'standup', 'swim'];

export class PostureClips {
  // has(name) says a clip is loaded; seconds(name) is how long it plays.
  constructor({ has, seconds }) {
    this.has = has;
    this.seconds = seconds;
    this.mode = 'none';
    this.t = 0;
  }

  _enter(mode) { this.mode = mode; this.t = 0; }

  step(dt, { posture = '', moving = false } = {}) {
    this.t += dt;
    const has = this.has;
    switch (this.mode) {
      case 'sitdown':
        if (posture !== 'sit') return this._standUp();
        if (this.t >= this.seconds('sitdown')) this._enter('seated');
        return this.mode === 'sitdown' ? 'sitdown' : this._seated();
      case 'seated':
        if (posture !== 'sit') return this._standUp();
        return this._seated();
      case 'standup':
        if (posture === 'sit') return this._sitDown();
        if (this.t >= this.seconds('standup')) { this._enter('none'); return this._from(posture, moving); }
        return 'standup';
      case 'air':
        if (posture === 'air') return has('jump') ? 'jump' : null;
        if (posture === '' && has('land') && !moving) { this._enter('landing'); return 'land'; }
        this._enter('none');
        return this._from(posture, moving);
      case 'landing':
        if (posture !== '' || moving || this.t >= Math.min(LAND_MAX, this.seconds('land'))) {
          this._enter('none');
          return this._from(posture, moving);
        }
        return 'land';
      case 'swim':
        if (posture === 'swim') return has('swim') ? 'swim' : null;
        this._enter('none');
        return this._from(posture, moving);
      default:
        return this._from(posture, moving);
    }
  }

  // from none: what the posture asks for.
  _from(posture, moving) {
    switch (posture) {
      case 'sit': return this._sitDown();
      case 'air': this._enter('air'); return this.has('jump') ? 'jump' : null;
      case 'swim': this._enter('swim'); return this.has('swim') ? 'swim' : null;
      default: return null;
    }
  }

  _sitDown() {
    if (this.has('sitdown')) { this._enter('sitdown'); return 'sitdown'; }
    this._enter('seated');
    return this._seated();
  }

  _seated() { return this.has('sit') ? 'sit' : null; }

  _standUp() {
    if (this.has('standup')) { this._enter('standup'); return 'standup'; }
    this._enter('none');
    return null;
  }
}
