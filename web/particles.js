// Particles: the fires' flames and smoke, the forges' sparks, the portals'
// streams, run from the emitter settings propconv took from each model.
//
// Nothing is simulated frame to frame. An emitter that makes R particles a
// second, each living up to L seconds, never has more than N = R x L alive,
// so it gets N slots, and slot k is reborn every N / R seconds, k / R after
// slot 0. Its age is then a function of the clock alone, and so is
// everything else: where in the birth box it starts, which way in the cone
// it flies and how fast all come from a hash of the slot and which rebirth
// this is. The vertex shader works out each particle from the time and its
// slot number, and the CPU only issues one draw per emitter per placement.
//
// The model's settings, as NetImmerse used them: vertical direction is the
// angle from the emitter's +Z, so 0 sprays up and pi down; horizontal turns
// about Z; each has a half-width. Size is the particle's radius. GrowFade
// takes it from nothing to full over its first seconds and back over its
// last; the colour curve, over the particle's life, multiplies the texture.
// Glows add to what is behind them; smoke is blended over it. Neither writes
// depth, so particles hide behind walls but not behind each other.

const VS = `#version 300 es
precision highp float;
precision highp int;
layout(location=0) in vec2 aCorner;   // -1..1 square
uniform mat4 uViewProj;
uniform vec3 uCamPos, uCamRight, uCamUp;
uniform float uTime;
// The placement: position (scene x, height, z), yaw, scale, and a seed so
// copies of one model do not burn in step.
uniform vec4 uPlace;
uniform float uScale;
uniform uint uSeed;
// The emitter, in model space (x, y across, z up).
uniform vec3 uOrigin;
uniform mat3 uAxes;
uniform vec3 uBox;
uniform float uRate;
uniform vec2 uLife, uSpeed, uVert, uHoriz;
uniform float uSize, uGrow, uFade, uSpin;
uniform vec4 uColor[6];  // r, g, b, a at uColorT
uniform float uColorT[6];
uniform int uColorN;
out vec2 vUV;
out vec4 vColor;
out float vDist;

uint hash(uint x) {
  x ^= x >> 16; x *= 0x7feb352du;
  x ^= x >> 15; x *= 0x846ca68bu;
  x ^= x >> 16;
  return x;
}
float rnd(inout uint s) { s = hash(s); return float(s) / 4294967296.0; }
float signed(inout uint s) { return rnd(s) * 2.0 - 1.0; }

vec4 colorAt(float t) {
  if (uColorN == 0) return vec4(1.0);
  vec4 c = uColor[0];
  for (int i = 1; i < 6; i++) {
    if (i >= uColorN) break;
    if (t <= uColorT[i]) {
      float a = uColorT[i - 1], b = uColorT[i];
      return mix(uColor[i - 1], uColor[i], clamp((t - a) / max(b - a, 1e-4), 0.0, 1.0));
    }
    c = uColor[i];
  }
  return c;
}

void main() {
  float k = float(gl_InstanceID);
  float maxLife = uLife.x + uLife.y;
  float slots = ceil(uRate * maxLife);
  float period = slots / uRate;
  float t = uTime - k / uRate;
  float cycle = floor(t / period);
  float age = t - cycle * period;

  uint s = hash(uint(gl_InstanceID) * 0x9e3779b9u ^ uint(int(cycle)) * 0x85ebca6bu ^ uSeed);
  float life = uLife.x + signed(s) * uLife.y;
  vec3 start = vec3(signed(s), signed(s), signed(s)) * uBox;
  float pol = uVert.x + signed(s) * uVert.y;
  float az = uHoriz.x + signed(s) * uHoriz.y;
  float speed = uSpeed.x + signed(s) * uSpeed.y;
  float turn = rnd(s) * 6.2831853;
  vDist = 0.0;
  if (age >= life || life <= 0.0) {
    // This slot is between lives: put it behind the camera.
    gl_Position = vec4(0.0, 0.0, -2.0, 1.0);
    vUV = vec2(0.0); vColor = vec4(0.0);
    return;
  }

  vec3 dir = vec3(sin(pol) * cos(az), sin(pol) * sin(az), cos(pol));
  vec3 local = start + dir * speed * age;
  vec3 m = uOrigin + uAxes * local;  // model space, z up

  // Placed as the props are: yaw about the vertical, scale, then position;
  // model z becomes height.
  float c = cos(uPlace.w), sn = sin(uPlace.w);
  vec3 world = vec3(
    uScale * (c * m.x - sn * m.y) + uPlace.x,
    uScale * m.z + uPlace.y,
    uScale * (sn * m.x + c * m.y) + uPlace.z);

  float size = uSize * uScale;
  if (uGrow > 0.0) size *= clamp(age / uGrow, 0.0, 1.0);
  if (uFade > 0.0) size *= clamp((life - age) / uFade, 0.0, 1.0);
  float a = turn + uSpin * age;
  vec2 q = mat2(cos(a), sin(a), -sin(a), cos(a)) * aCorner;
  world += (uCamRight * q.x + uCamUp * q.y) * size;

  vUV = aCorner * 0.5 + 0.5;
  vUV.y = 1.0 - vUV.y;
  vColor = colorAt(age / life);
  vDist = length(world - uCamPos);
  gl_Position = uViewProj * vec4(world, 1.0);
}
`;

const FS = `#version 300 es
precision highp float;
in vec2 vUV;
in vec4 vColor;
in float vDist;
uniform sampler2D uTex;
uniform float uAdditive;
uniform float uFogStart, uFogEnd;
uniform vec3 uFogColor;
out vec4 outColor;
void main() {
  vec4 t = texture(uTex, vUV) * vColor;
  float fog = clamp((vDist - uFogStart) / max(uFogEnd - uFogStart, 1.0), 0.0, 1.0);
  if (uAdditive > 0.5) {
    // Light: fades out in fog rather than into it.
    outColor = vec4(t.rgb * (1.0 - fog), t.a);
  } else {
    outColor = vec4(mix(t.rgb, uFogColor, fog), t.a);
  }
}
`;

// Emitters further than this from the camera are not drawn.
const DRAW_DISTANCE = 9000;

export function createParticles(gl, props, { program, uniforms }) {
  const prog = program(gl, VS, FS, 'particles');
  const U = uniforms(gl, prog, ['uViewProj', 'uCamPos', 'uCamRight', 'uCamUp', 'uTime', 'uPlace',
    'uScale', 'uSeed', 'uOrigin', 'uAxes', 'uBox', 'uRate', 'uLife', 'uSpeed', 'uVert', 'uHoriz',
    'uSize', 'uGrow', 'uFade', 'uSpin', 'uColor', 'uColorT', 'uColorN', 'uTex', 'uAdditive',
    'uFogStart', 'uFogEnd', 'uFogColor']);

  const vao = gl.createVertexArray();
  gl.bindVertexArray(vao);
  const vbo = gl.createBuffer();
  gl.bindBuffer(gl.ARRAY_BUFFER, vbo);
  gl.bufferData(gl.ARRAY_BUFFER, new Float32Array([-1, -1, 1, -1, -1, 1, 1, 1]), gl.STATIC_DRAW);
  gl.enableVertexAttribArray(0);
  gl.vertexAttribPointer(0, 2, gl.FLOAT, false, 0, 0);
  gl.bindVertexArray(null);

  // Every placed emitter: the model's settings and the placement's.
  const man = props.manifest;
  const emitters = [];
  man.instances.forEach((inst, n) => {
    const model = man.models[inst.m];
    for (const p of model.particles || []) {
      const colors = (p.color || []).slice(0, 6);
      emitters.push({
        p,
        place: [inst.p[0], inst.p[2], inst.p[1], inst.yaw],
        scale: inst.s,
        seed: (n * 2654435761) >>> 0,
        slots: Math.ceil(p.rate * (p.life[0] + p.life[1])),
        // mat3 uniforms are column-major; the axes are row-major.
        axes: new Float32Array([p.axes[0], p.axes[3], p.axes[6], p.axes[1], p.axes[4], p.axes[7], p.axes[2], p.axes[5], p.axes[8]]),
        colorT: new Float32Array(6).map((_, i) => (colors[i] ? colors[i][0] : 1)),
        color: new Float32Array(24).map((_, i) => (colors[i >> 2] ? colors[i >> 2][1 + (i & 3)] : 0)),
        colorN: colors.length,
      });
    }
  });

  let drawn = 0;
  function draw(ctx) {
    drawn = 0;
    if (!emitters.length) return;
    gl.useProgram(prog);
    gl.uniformMatrix4fv(U.uViewProj, false, ctx.viewProj);
    gl.uniform3fv(U.uCamPos, ctx.camPos);
    gl.uniform3fv(U.uCamRight, ctx.camRight);
    gl.uniform3fv(U.uCamUp, ctx.camUp);
    gl.uniform1f(U.uTime, ctx.time);
    gl.uniform1f(U.uFogStart, ctx.fogStart);
    gl.uniform1f(U.uFogEnd, ctx.fogEnd);
    gl.uniform3fv(U.uFogColor, ctx.fogColor);
    gl.uniform1i(U.uTex, 0);
    gl.activeTexture(gl.TEXTURE0);
    gl.bindVertexArray(vao);
    gl.enable(gl.BLEND);
    gl.depthMask(false);
    gl.disable(gl.CULL_FACE);
    // Smoke first, glows over it.
    for (const additive of [false, true]) {
      gl.blendFunc(gl.SRC_ALPHA, additive ? gl.ONE : gl.ONE_MINUS_SRC_ALPHA);
      gl.uniform1f(U.uAdditive, additive ? 1 : 0);
      for (const e of emitters) {
        const p = e.p;
        if (p.additive !== additive) continue;
        if (Math.hypot(e.place[0] - ctx.camPos[0], e.place[2] - ctx.camPos[2]) > DRAW_DISTANCE) continue;
        gl.bindTexture(gl.TEXTURE_2D, props.texture(p.texture));
        gl.uniform4fv(U.uPlace, e.place);
        gl.uniform1f(U.uScale, e.scale);
        gl.uniform1ui(U.uSeed, e.seed);
        gl.uniform3fv(U.uOrigin, p.origin);
        gl.uniformMatrix3fv(U.uAxes, false, e.axes);
        gl.uniform3fv(U.uBox, p.box);
        gl.uniform1f(U.uRate, p.rate);
        gl.uniform2fv(U.uLife, p.life);
        gl.uniform2fv(U.uSpeed, p.speed);
        gl.uniform2fv(U.uVert, p.vertical);
        gl.uniform2fv(U.uHoriz, p.horizontal);
        gl.uniform1f(U.uSize, p.size);
        gl.uniform1f(U.uGrow, p.grow);
        gl.uniform1f(U.uFade, p.fade);
        gl.uniform1f(U.uSpin, p.spin);
        gl.uniform4fv(U.uColor, e.color);
        gl.uniform1fv(U.uColorT, e.colorT);
        gl.uniform1i(U.uColorN, e.colorN);
        gl.drawArraysInstanced(gl.TRIANGLE_STRIP, 0, 4, e.slots);
        drawn += e.slots;
      }
    }
    gl.enable(gl.CULL_FACE);
    gl.depthMask(true);
    gl.disable(gl.BLEND);
    gl.bindVertexArray(null);
  }

  return { draw, emitters, get drawn() { return drawn; } };
}
