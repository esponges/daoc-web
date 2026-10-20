// Day and night on the client. The server sends the game time of day as
// seconds since midnight (a `clock` message), and a game day lasts 60 real
// minutes, so between messages the clock runs at 24 times real time. Night is
// 18:00 up to 06:00; the scene darkens over the game hour after dusk and
// lightens over the hour before dawn.

export const DAY = 86400;
export const RATE = 24;
const DUSK = 18 * 3600;
const DAWN = 6 * 3600;
const HOUR = 3600;

// gameSeconds is the game time of day now, from the last clock message
// ({ gs, at }, `at` in the seconds of the same clock as nowSec).
export function gameSeconds(clock, nowSec) {
  if (!clock) return null;
  return (((clock.gs + (nowSec - clock.at) * RATE) % DAY) + DAY) % DAY;
}

// isNight says whether a game time of day is night.
export const isNight = (gs) => gs >= DUSK || gs < DAWN;

// nightness is how dark the scene is, 0 by day to 1 at night: it rises over the
// hour after 18:00 and falls over the hour before 06:00.
export function nightness(gs) {
  if (gs == null) return 0;
  if (gs >= DUSK) return Math.min(1, (gs - DUSK) / HOUR);
  if (gs < DAWN) return Math.min(1, (DAWN - gs) / HOUR);
  return 0;
}

// clockFace is the game time as the clock window's Date-like: whole hours and
// minutes.
export function clockFace(gs) {
  return { getHours: () => Math.floor(gs / 3600) % 24, getMinutes: () => Math.floor((gs % 3600) / 60) };
}

// The night's colours: fog and horizon go to a deep blue, the sun's direct
// light nearly out, and the ambient light dim and blue.
const NIGHT_FOG = [0.03, 0.045, 0.1];
const NIGHT_AMBIENT = [0.09, 0.11, 0.2];
const NIGHT_SUN = [0.05, 0.06, 0.1];

const lerp3 = (a, b, t) => a.map((v, i) => v + (b[i] - v) * t);

// nightLook blends the day's fog colour, ambient and sun light toward night.
export function nightLook(day, n) {
  return { fog: lerp3(day.fog, NIGHT_FOG, n), ambient: lerp3(day.ambient, NIGHT_AMBIENT, n), sun: lerp3(day.sun, NIGHT_SUN, n) };
}

// The sky when no night sky art is converted: the day sky, darkened by a flat
// blue laid over it.
const FLAT_VS = `#version 300 es
void main() {
  vec2 p = vec2(float((gl_VertexID << 1) & 2), float(gl_VertexID & 2)) * 2.0 - 1.0;
  gl_Position = vec4(p, 0.0, 1.0);
}`;
const FLAT_FS = `#version 300 es
precision mediump float;
uniform vec4 uColor;
out vec4 outColor;
void main() { outColor = uColor; }`;

// createDarken builds a full-screen blue veil; draw(alpha) lays it over what is
// drawn so far (the sky, before the world goes in over it).
export function createDarken(gl) {
  const sh = (type, src) => {
    const s = gl.createShader(type);
    gl.shaderSource(s, src);
    gl.compileShader(s);
    if (!gl.getShaderParameter(s, gl.COMPILE_STATUS)) throw new Error(gl.getShaderInfoLog(s));
    return s;
  };
  const prog = gl.createProgram();
  gl.attachShader(prog, sh(gl.VERTEX_SHADER, FLAT_VS));
  gl.attachShader(prog, sh(gl.FRAGMENT_SHADER, FLAT_FS));
  gl.linkProgram(prog);
  const uColor = gl.getUniformLocation(prog, 'uColor');
  const vao = gl.createVertexArray();
  return {
    draw(alpha) {
      if (alpha <= 0) return;
      gl.useProgram(prog);
      gl.uniform4f(uColor, NIGHT_FOG[0], NIGHT_FOG[1], NIGHT_FOG[2], alpha);
      gl.enable(gl.BLEND);
      gl.blendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA);
      gl.disable(gl.DEPTH_TEST);
      gl.bindVertexArray(vao);
      gl.drawArrays(gl.TRIANGLES, 0, 3);
      gl.bindVertexArray(null);
      gl.enable(gl.DEPTH_TEST);
      gl.disable(gl.BLEND);
    },
  };
}
