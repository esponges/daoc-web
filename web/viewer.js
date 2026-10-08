// Minimal WebGL2 viewer for a converted DAoC zone.
//
// Deliberately dependency-free: the point of this spike is to prove the
// extract -> convert -> render pipeline works, so it should not be able to
// fail because a CDN is unreachable.
//
// Coordinate mapping: DAoC is Z-up with the zone spanning 0..65536 on X and Y.
// Here that becomes a Y-up scene: world X -> GL x, world Y -> GL z, height -> GL y.

import { createCharacter, modelMatrix } from './character.js';
import { createProps } from './props.js';
import { createNPCs, bodyRadius, pickNPC } from './npc.js';
import { createRing } from './ring.js';
import { createCombat } from './combat.js';
import { Animator } from './animator.js';

const ZONE = 'data/zone100';

// Which converted character to load. ?char=norseman switches back without a
// rebuild; each one carries its own animations, so the clips follow the model.
const CHARACTER = 'data/char/' +
  (new URLSearchParams(location.search).get('char') || 'troll-warrior')
    .replace(/[^a-zA-Z0-9_-]/g, '');

// DAoC's own movement rates, in world units per second. A zone is 65536 units
// across and the Norseman model is 71.8 tall, which puts a unit at roughly an
// inch and makes these numbers read as sensible human speeds.
const WALK = 85;
const RUN = 191;
const SPRINT = 420;

// Ground distance covered by one complete two-step gait cycle. Taken from the
// model's leg length so the feet keep pace with the ground instead of skating.
const STRIDE = 105;

// How long a change of gait takes to cross-fade. Long enough that the switch
// is not a snap, short enough that the character still feels responsive to the
// key that caused it; a stride is about a second, so this is a fifth of one.
const FADE = 0.18;

// Clips that are stride cycles, and so can hand their phase to one another.
const LOCOMOTION = new Set(['walk', 'run']);

// Vertical field of view, shared by the projection and click-picking.
const FOV = Math.PI / 3;

// The player's hit points. A grey wolf takes a third of them in a fair fight.
const PLAYER_HP = 220;

const $ = (id) => document.getElementById(id);

function fail(msg) {
  const el = $('err');
  el.textContent = 'viewer error\n\n' + msg;
  el.style.display = 'grid';
  console.error(msg);
}

// ---------------------------------------------------------------- mat4 helpers

function perspective(fovy, aspect, near, far) {
  const f = 1 / Math.tan(fovy / 2), nf = 1 / (near - far);
  return new Float32Array([
    f / aspect, 0, 0, 0,
    0, f, 0, 0,
    0, 0, (far + near) * nf, -1,
    0, 0, 2 * far * near * nf, 0,
  ]);
}

function lookAt(eye, center, up) {
  let zx = eye[0] - center[0], zy = eye[1] - center[1], zz = eye[2] - center[2];
  let zl = Math.hypot(zx, zy, zz) || 1;
  zx /= zl; zy /= zl; zz /= zl;
  let xx = up[1] * zz - up[2] * zy, xy = up[2] * zx - up[0] * zz, xz = up[0] * zy - up[1] * zx;
  let xl = Math.hypot(xx, xy, xz) || 1;
  xx /= xl; xy /= xl; xz /= xl;
  const yx = zy * xz - zz * xy, yy = zz * xx - zx * xz, yz = zx * xy - zy * xx;
  return new Float32Array([
    xx, yx, zx, 0,
    xy, yy, zy, 0,
    xz, yz, zz, 0,
    -(xx * eye[0] + xy * eye[1] + xz * eye[2]),
    -(yx * eye[0] + yy * eye[1] + yz * eye[2]),
    -(zx * eye[0] + zy * eye[1] + zz * eye[2]),
    1,
  ]);
}

function mul(a, b) {
  const o = new Float32Array(16);
  for (let c = 0; c < 4; c++) {
    for (let r = 0; r < 4; r++) {
      o[c * 4 + r] = a[r] * b[c * 4] + a[4 + r] * b[c * 4 + 1] + a[8 + r] * b[c * 4 + 2] + a[12 + r] * b[c * 4 + 3];
    }
  }
  return o;
}

// --------------------------------------------------------------- gl boilerplate

function compile(gl, type, src, label) {
  const s = gl.createShader(type);
  gl.shaderSource(s, src);
  gl.compileShader(s);
  if (!gl.getShaderParameter(s, gl.COMPILE_STATUS)) {
    throw new Error(label + ' failed to compile:\n' + gl.getShaderInfoLog(s));
  }
  return s;
}

function program(gl, vsSrc, fsSrc, label) {
  const p = gl.createProgram();
  gl.attachShader(p, compile(gl, gl.VERTEX_SHADER, vsSrc, label + ' vertex shader'));
  gl.attachShader(p, compile(gl, gl.FRAGMENT_SHADER, fsSrc, label + ' fragment shader'));
  gl.linkProgram(p);
  if (!gl.getProgramParameter(p, gl.LINK_STATUS)) {
    throw new Error(label + ' failed to link:\n' + gl.getProgramInfoLog(p));
  }
  return p;
}

// ------------------------------------------------------------------- shaders

const TERRAIN_VS = `#version 300 es
precision highp float;
layout(location=0) in vec3 aPos;
layout(location=1) in vec3 aNormal;
layout(location=2) in vec2 aUV;
uniform mat4 uViewProj;
uniform vec3 uCamPos;
out vec3 vNormal;
out vec2 vUV;
out float vDist;
void main() {
  vNormal = aNormal;
  vUV = aUV;
  vDist = length(aPos - uCamPos);
  gl_Position = uViewProj * vec4(aPos, 1.0);
}
`;

const TERRAIN_FS = `#version 300 es
precision highp float;
in vec3 vNormal;
in vec2 vUV;
in float vDist;
uniform sampler2D uAtlas;
uniform vec3 uFogColor;
uniform float uFogStart;
uniform float uFogEnd;
uniform vec3 uLightDir;
uniform float uWireframe;
out vec4 outColor;
void main() {
  vec3 base = texture(uAtlas, vUV).rgb;
  if (uWireframe > 0.5) base = vec3(0.55, 0.62, 0.70);
  vec3 n = normalize(vNormal);
  // Hemisphere ambient plus one directional term: enough to read the relief.
  float diff = max(dot(n, normalize(uLightDir)), 0.0);
  float ambient = 0.45 + 0.20 * n.y;
  vec3 lit = base * (ambient + 0.75 * diff);
  float fog = clamp((vDist - uFogStart) / max(uFogEnd - uFogStart, 1.0), 0.0, 1.0);
  outColor = vec4(mix(lit, uFogColor, fog), 1.0);
}
`;

const WATER_VS = `#version 300 es
precision highp float;
layout(location=0) in vec3 aPos;
uniform mat4 uViewProj;
uniform vec3 uCamPos;
out float vDist;
out vec2 vXZ;
void main() {
  vDist = length(aPos - uCamPos);
  vXZ = aPos.xz;
  gl_Position = uViewProj * vec4(aPos, 1.0);
}
`;

const WATER_FS = `#version 300 es
precision highp float;
in float vDist;
in vec2 vXZ;
uniform vec3 uFogColor;
uniform float uFogStart;
uniform float uFogEnd;
uniform float uTime;
out vec4 outColor;
void main() {
  // Two low-frequency waves at an angle to each other, kept subtle: a single
  // separable sin*sin product reads as a checkerboard rather than as water.
  float a = sin(dot(vXZ, vec2(0.00042, 0.00031)) * 6.2831 + uTime * 0.9);
  float b = sin(dot(vXZ, vec2(-0.00026, 0.00049)) * 6.2831 + uTime * 0.6);
  float r = 0.5 + 0.25 * (a + b) * 0.5;
  vec3 water = mix(vec3(0.07, 0.22, 0.32), vec3(0.13, 0.35, 0.46), r);
  float fog = clamp((vDist - uFogStart) / max(uFogEnd - uFogStart, 1.0), 0.0, 1.0);
  outColor = vec4(mix(water, uFogColor, fog), 0.78);
}
`;

// ---------------------------------------------------------------------- main

async function main() {
  const canvas = $('gl');
  const gl = canvas.getContext('webgl2', { antialias: true, depth: true });
  if (!gl) throw new Error('WebGL2 is not available in this browser.');

  // --- load converted data ---
  const man = await fetchJSON(ZONE + '/zone.json');
  const [heightBuf, atlasImg] = await Promise.all([
    fetchBuffer(ZONE + '/' + man.heights),
    loadImage(ZONE + '/' + man.atlas),
  ]);

  const grid = man.grid;
  const heights = new Uint16Array(heightBuf);
  if (heights.length !== grid * grid) {
    throw new Error('heights.u16 holds ' + heights.length + ' samples, manifest says ' + grid * grid);
  }
  const cell = man.cellUnits;
  const H = (x, y) => heights[Math.min(grid - 1, Math.max(0, y)) * grid + Math.min(grid - 1, Math.max(0, x))];

  // The height of the ground as drawn. Each cell of the mesh below is two
  // triangles split along the (x+1, y)-(x, y+1) diagonal, so this
  // interpolates across the same two triangles. Bilinear sampling would agree
  // at the samples but not between them -- by up to a quarter of the cell's
  // twist, tens of units on a steep hillside -- which left the target ring
  // buried under the slope it was meant to lie on, and feet a little in or
  // above the visible ground.
  const groundAt = (wx, wy) => {
    const fx = wx / cell, fy = wy / cell;
    const x0 = Math.floor(fx), y0 = Math.floor(fy);
    const tx = fx - x0, ty = fy - y0;
    const h00 = H(x0, y0), h10 = H(x0 + 1, y0), h01 = H(x0, y0 + 1), h11 = H(x0 + 1, y0 + 1);
    if (tx + ty <= 1) return h00 + (h10 - h00) * tx + (h01 - h00) * ty;
    return h11 + (h01 - h11) * (1 - tx) + (h10 - h11) * (1 - ty);
  };

  // --- terrain mesh ---
  const vertCount = grid * grid;
  const positions = new Float32Array(vertCount * 3);
  const normals = new Float32Array(vertCount * 3);
  const uvs = new Float32Array(vertCount * 2);
  for (let y = 0; y < grid; y++) {
    for (let x = 0; x < grid; x++) {
      const i = y * grid + x;
      positions[i * 3 + 0] = x * cell;
      positions[i * 3 + 1] = H(x, y);
      positions[i * 3 + 2] = y * cell;
      // Central differences; the 2*cell span cancels into the normalise.
      const nx = (H(x - 1, y) - H(x + 1, y));
      const nz = (H(x, y - 1) - H(x, y + 1));
      const ny = 2 * cell;
      const len = Math.hypot(nx, ny, nz) || 1;
      normals[i * 3 + 0] = nx / len;
      normals[i * 3 + 1] = ny / len;
      normals[i * 3 + 2] = nz / len;
      uvs[i * 2 + 0] = x / (grid - 1);
      uvs[i * 2 + 1] = y / (grid - 1);
    }
  }
  const quads = (grid - 1) * (grid - 1);
  const indices = new Uint32Array(quads * 6);
  let k = 0;
  for (let y = 0; y < grid - 1; y++) {
    for (let x = 0; x < grid - 1; x++) {
      const a = y * grid + x, b = a + 1, c = a + grid, d = c + 1;
      indices[k++] = a; indices[k++] = c; indices[k++] = b;
      indices[k++] = b; indices[k++] = c; indices[k++] = d;
    }
  }

  // Separate line-index buffer so the debug toggle draws a real wireframe
  // rather than reinterpreting triangle indices as line pairs.
  const lineIdx = new Uint32Array(quads * 4);
  let m = 0;
  for (let y = 0; y < grid - 1; y++) {
    for (let x = 0; x < grid - 1; x++) {
      const a = y * grid + x;
      lineIdx[m++] = a; lineIdx[m++] = a + 1;      // along x
      lineIdx[m++] = a; lineIdx[m++] = a + grid;   // along y
    }
  }

  const terrainProg = program(gl, TERRAIN_VS, TERRAIN_FS, 'terrain');
  const terrainVAO = gl.createVertexArray();
  gl.bindVertexArray(terrainVAO);
  bindAttrib(gl, 0, positions, 3);
  bindAttrib(gl, 1, normals, 3);
  bindAttrib(gl, 2, uvs, 2);
  const ibo = gl.createBuffer();
  gl.bindBuffer(gl.ELEMENT_ARRAY_BUFFER, ibo);
  gl.bufferData(gl.ELEMENT_ARRAY_BUFFER, indices, gl.STATIC_DRAW);
  const lbo = gl.createBuffer();
  gl.bindBuffer(gl.ELEMENT_ARRAY_BUFFER, lbo);
  gl.bufferData(gl.ELEMENT_ARRAY_BUFFER, lineIdx, gl.STATIC_DRAW);
  gl.bindVertexArray(null);

  // --- atlas texture ---
  const tex = gl.createTexture();
  gl.bindTexture(gl.TEXTURE_2D, tex);
  // No Y flip: the converter stitched the atlas in heightmap row order, so
  // atlas row 0 corresponds to heightmap row 0, which is v = 0.
  gl.texImage2D(gl.TEXTURE_2D, 0, gl.RGBA, gl.RGBA, gl.UNSIGNED_BYTE, atlasImg);
  gl.generateMipmap(gl.TEXTURE_2D);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);
  const aniso = gl.getExtension('EXT_texture_filter_anisotropic');
  if (aniso) {
    const maxA = gl.getParameter(aniso.MAX_TEXTURE_MAX_ANISOTROPY_EXT);
    gl.texParameterf(gl.TEXTURE_2D, aniso.TEXTURE_MAX_ANISOTROPY_EXT, Math.min(8, maxA));
  }

  // --- water surfaces ---
  // SECTOR.DAT gives each body as two index-aligned shore chains. The surface
  // is the strip between them, so it covers only the actual lake rather than
  // flooding every part of the zone that happens to sit below the waterline.
  const waterProg = program(gl, WATER_VS, WATER_FS, 'water');
  const extent = man.zoneUnits;
  const waters = [];
  for (const w of man.waters || []) {
    const L = w.left || [], R = w.right || [];
    const n = Math.min(L.length, R.length);
    if (n < 2) continue;
    const verts = [];
    const push = (p) => verts.push(p[0] * cell, w.height, p[1] * cell);
    for (let i = 0; i < n - 1; i++) {
      push(L[i]); push(R[i]); push(L[i + 1]);
      push(L[i + 1]); push(R[i]); push(R[i + 1]);
    }
    const vao = gl.createVertexArray();
    gl.bindVertexArray(vao);
    bindAttrib(gl, 0, new Float32Array(verts), 3);
    gl.bindVertexArray(null);
    waters.push({ ...w, vao, count: verts.length / 3 });
  }

  // Whether a point is under one of the lakes: inside its shoreline's box and
  // below its surface. NPCs use it to stay out of the water.
  const wetBoxes = waters.map((w) => {
    const pts = [...(w.left || []), ...(w.right || [])];
    return {
      h: w.height,
      x0: Math.min(...pts.map((p) => p[0])) * cell, x1: Math.max(...pts.map((p) => p[0])) * cell,
      y0: Math.min(...pts.map((p) => p[1])) * cell, y1: Math.max(...pts.map((p) => p[1])) * cell,
    };
  });
  const isWet = (wx, wy) => wetBoxes.some((b) =>
    wx >= b.x0 && wx <= b.x1 && wy >= b.y0 && wy <= b.y1 && groundAt(wx, wy) < b.h);

  // --- character ---
  const char = await createCharacter(gl, CHARACTER, { program, uniforms, loadImage });
  if (char.missingJoints.length) {
    console.warn('skeleton is missing expected joints:', char.missingJoints.join(', '));
  }

  // --- scenery ---
  // Props are optional: a zone that has not been through propconv should
  // still render its terrain rather than failing to start.
  let props = null;
  try {
    props = await createProps(gl, ZONE + '/props');
  } catch (e) {
    console.warn('scenery not loaded:', e.message);
  }

  // --- NPCs ---
  // Also optional: they need their characters converted, and the zone is
  // worth seeing without them.
  let npcs = null;
  try {
    // ?npcs=0 leaves them out, for looking at the zone on its own.
    if (new URLSearchParams(location.search).get('npcs') !== '0') {
      npcs = await createNPCs(gl, 'spawns/zone100.json', { program, uniforms, loadImage },
        { cell, extent, groundAt, isWet });
    }
  } catch (e) {
    console.warn('npcs not loaded:', e.message);
  }

  // Spawn on the rising ground south-east of the lake, which is open enough
  // to see the gait and close enough to walk to the shore.
  const SPAWN = [112 * cell, 118 * cell];
  const anim = new Animator(char, { carry: LOCOMOTION, fade: FADE });
  anim.t = 0;
  const player = {
    x: SPAWN[0], y: SPAWN[1], // world X and Y, i.e. heightmap cells
    yaw: Math.PI, phase: 0, gait: 0, clip: null, blend: 1,
    // A fighter, as combat.js sees one. What it holds sets the numbers:
    // a warhammer hits harder and swings slower than a fist, and a shield
    // blocks some of what comes from the front. They are this project's
    // numbers, not the game's.
    name: 'you', hp: PLAYER_HP, maxHp: PLAYER_HP,
    damage: char.armed ? [15, 27] : [9, 17], swing: char.armed ? 3.0 : 2.2,
    blockChance: char.shield ? 0.2 : 0,
    hitChance: 0.85, body: bodyRadius(char, char.scale), dead: false, anim,
  };

  // --- combat ---
  const logEl = $('log');
  const log = (text, kind = 'info') => {
    const line = document.createElement('div');
    line.className = kind;
    line.textContent = text;
    logEl.appendChild(line);
    while (logEl.childElementCount > 7) logEl.firstChild.remove();
  };
  const combat = npcs ? createCombat({
    player, npcs, log,
    // Back on your feet where you started, as at a bindstone.
    onPlayerDeath: () => { [player.x, player.y] = SPAWN; },
  }) : null;
  const ring = combat ? createRing(gl, { program, uniforms }) : null;

  // --- cameras ---
  // Two modes: an orbit camera chasing the character, and the original
  // free-fly for looking at the zone as a whole.
  const orbit = { yaw: Math.PI, pitch: 0.30, dist: 320 };
  const mid = extent / 2;
  const cam = {
    pos: [mid - 26000, man.maxHeight + 9000, mid + 34000],
    yaw: 0, pitch: 0,
    speed: 4000, // world units per second
  };
  {
    const t = [mid, man.minHeight + 1500, mid];
    const dx = t[0] - cam.pos[0], dy = t[1] - cam.pos[1], dz = t[2] - cam.pos[2];
    cam.yaw = Math.atan2(dx, -dz);
    cam.pitch = Math.atan2(dy, Math.hypot(dx, dz));
  }
  let followMode = true;
  let lastEye = [...cam.pos];
  // Where the camera was last frame, for turning a click into a ray.
  const view = { eye: null, target: null };

  // Shortest-way-round turn, so the character never spins the long way to
  // face a new heading.
  const turnToward = (cur, target, maxStep) => {
    let d = (target - cur) % (Math.PI * 2);
    if (d > Math.PI) d -= Math.PI * 2;
    if (d < -Math.PI) d += Math.PI * 2;
    return cur + Math.max(-maxStep, Math.min(maxStep, d));
  };

  // Debug handle: lets you jump the camera from the console, e.g.
  //   daoc.goto(120, 90, 800)   // heightmap cell x, y, metres above ground
  globalThis.daoc = {
    cam, orbit, player, char, props, npcs, combat, anim, manifest: man, heights,
    heightAt: (cx, cy) => H(Math.round(cx), Math.round(cy)),
    groundAt,
    // Put the character on a given heightmap cell, e.g. daoc.warp(60, 70).
    warp(cx, cy) {
      player.x = cx * cell;
      player.y = cy * cell;
      return { x: player.x, y: player.y, ground: groundAt(player.x, player.y) };
    },
    goto(cx, cy, above = 1200) {
      cam.pos = [cx * cell, H(Math.round(cx), Math.round(cy)) + above, cy * cell];
      return cam.pos;
    },
    look(cx, cy) {
      const dx = cx * cell - cam.pos[0];
      const dz = cy * cell - cam.pos[2];
      const dy = H(Math.round(cx), Math.round(cy)) - cam.pos[1];
      cam.yaw = Math.atan2(dx, -dz);
      cam.pitch = Math.atan2(dy, Math.hypot(dx, dz));
    },
  };

  const keys = new Set();
  let wireframe = false;
  addEventListener('keydown', (e) => {
    keys.add(e.code);
    if (e.code === 'KeyF') wireframe = !wireframe;
    // Combat. Tab would otherwise move focus out of the page.
    if (combat && e.code === 'Tab') {
      e.preventDefault();
      const t = combat.targetNext(orbit.yaw);
      if (t) log('You target ' + 'the ' + t.name + '.', 'info');
    }
    if (combat && e.code === 'Digit1' && !e.repeat) combat.toggleAttack(orbit.yaw);
    if (combat && e.code === 'Escape') combat.clearTarget();
    if (e.code === 'KeyC') {
      followMode = !followMode;
      // Hand the free camera the view it was just looking from, so toggling
      // does not teleport.
      if (!followMode) {
        cam.pos = [...lastEye];
        cam.yaw = orbit.yaw;
        cam.pitch = -orbit.pitch;
      }
    }
  });
  addEventListener('keyup', (e) => keys.delete(e.code));
  let dragging = false;
  // A click -- pressed and released without dragging -- selects whatever NPC
  // is under the cursor; a drag orbits the camera as before.
  let downAt = null;
  canvas.addEventListener('mousedown', (e) => {
    dragging = true;
    downAt = [e.clientX, e.clientY];
    e.preventDefault();
  });
  addEventListener('mouseup', (e) => {
    dragging = false;
    if (!downAt || !combat || !view.eye) return;
    const moved = Math.hypot(e.clientX - downAt[0], e.clientY - downAt[1]);
    downAt = null;
    if (moved > 4) return;
    // A ray from the eye through the clicked pixel, built from the camera's
    // own basis and field of view.
    const r = canvas.getBoundingClientRect();
    const nx = ((e.clientX - r.left) / r.width) * 2 - 1;
    const ny = 1 - ((e.clientY - r.top) / r.height) * 2;
    const [ex, ey, ez] = view.eye, [tx, ty, tz] = view.target;
    let f = [tx - ex, ty - ey, tz - ez];
    const fl = Math.hypot(...f); f = f.map((v) => v / fl);
    let s = [-f[2], 0, f[0]]; // forward x up(0,1,0)
    const sl = Math.hypot(...s) || 1; s = s.map((v) => v / sl);
    const u = [s[1] * f[2] - s[2] * f[1], s[2] * f[0] - s[0] * f[2], s[0] * f[1] - s[1] * f[0]];
    const k = Math.tan(FOV / 2), aspect = r.width / r.height;
    const dir = [0, 1, 2].map((i) => f[i] + s[i] * nx * k * aspect + u[i] * ny * k);
    const hit = pickNPC(npcs.npcs, view.eye, dir, groundAt);
    if (hit && hit !== combat.target) {
      combat.select(hit);
      log('You target the ' + hit.name + '.', 'info');
    }
  });
  addEventListener('mousemove', (e) => {
    if (!dragging) return;
    if (followMode) {
      orbit.yaw -= e.movementX * 0.005;
      orbit.pitch = Math.max(-0.35, Math.min(1.25, orbit.pitch + e.movementY * 0.004));
    } else {
      cam.yaw -= e.movementX * 0.0028;
      cam.pitch = Math.max(-1.5, Math.min(1.5, cam.pitch - e.movementY * 0.0028));
    }
  });
  canvas.addEventListener('wheel', (e) => {
    if (followMode) {
      orbit.dist = Math.max(90, Math.min(4000, orbit.dist * (e.deltaY > 0 ? 1.12 : 0.89)));
    } else {
      cam.speed = Math.max(200, Math.min(60000, cam.speed * (e.deltaY > 0 ? 0.85 : 1.18)));
    }
    e.preventDefault();
  }, { passive: false });

  // --- HUD ---
  const fog = man.fog;
  const fogColor = [fog.r / 255, fog.g / 255, fog.b / 255];
  $('title').textContent = man.name + '  (zone ' + man.zone + ')';
  $('i-grid').textContent = grid + ' x ' + grid + ' @ ' + cell + 'u';
  $('i-height').textContent = man.minHeight + ' .. ' + man.maxHeight + 'u';
  $('i-tris').textContent = (quads * 2).toLocaleString();
  $('i-atlas').textContent = man.atlasPx + ' x ' + man.atlasPx;
  $('i-align').textContent = man.orientation.replace(/\s+/g, ' ');
  $('i-water').textContent = waters.length
    ? waters.map((w) => w.name + ' @ ' + w.height).join(', ')
    : 'none';
  $('i-props').textContent = props
    ? props.instanceCount + ' placed, ' + props.models.length + ' models, ' +
      props.triangleCount.toLocaleString() + ' tris, ' + props.drawCalls + ' draws'
    : 'none';
  $('i-char').textContent = char.manifest.source + '  ' +
    char.triangles + ' tris, ' + char.boneCount + ' bones, ' +
    char.height.toFixed(1) + 'u, drawn at ' + char.scale + 'x';

  // --- uniform locations ---
  const tU = uniforms(gl, terrainProg, ['uViewProj', 'uCamPos', 'uAtlas', 'uFogColor', 'uFogStart', 'uFogEnd', 'uLightDir', 'uWireframe']);
  const wU = uniforms(gl, waterProg, ['uViewProj', 'uCamPos', 'uFogColor', 'uFogStart', 'uFogEnd', 'uTime']);

  gl.enable(gl.DEPTH_TEST);
  gl.enable(gl.CULL_FACE);
  gl.cullFace(gl.BACK);

  let last = performance.now(), fpsAccum = 0, fpsFrames = 0;
  function frame(now) {
    const dt = Math.min(0.1, (now - last) / 1000);
    last = now;

    // --- move the character ---
    // WASD is relative to where the camera points and the character turns to
    // face wherever it is heading, which is how an MMO handles it.
    const camF = [Math.sin(orbit.yaw), -Math.cos(orbit.yaw)];
    const camR = [Math.cos(orbit.yaw), Math.sin(orbit.yaw)];
    let mx = 0, my = 0;
    if (keys.has('KeyW')) { mx += camF[0]; my += camF[1]; }
    if (keys.has('KeyS')) { mx -= camF[0]; my -= camF[1]; }
    if (keys.has('KeyD')) { mx += camR[0]; my += camR[1]; }
    if (keys.has('KeyA')) { mx -= camR[0]; my -= camR[1]; }
    const moveLen = Math.hypot(mx, my);
    if (moveLen > 0 && followMode && !player.dead) {
      mx /= moveLen; my /= moveLen;
      const sprinting = keys.has('ShiftLeft') || keys.has('ShiftRight');
      const walking = keys.has('AltLeft') || keys.has('AltRight');
      const speed = sprinting ? SPRINT : walking ? WALK : RUN;
      const d = speed * dt;
      player.x = Math.max(0, Math.min(extent, player.x + mx * d));
      player.y = Math.max(0, Math.min(extent, player.y + my * d));
      player.yaw = turnToward(player.yaw, Math.atan2(mx, -my), 11 * dt);
      // Phase advances with ground covered, not with time, so the feet keep
      // pace at any speed.
      player.phase += (d / STRIDE) * Math.PI * 2;
      player.gait = speed;
    } else {
      player.gait = 0;
    }
    const intensity = Math.min(1, player.gait / RUN);

    // --- NPCs and the fight ---
    if (npcs) npcs.update(dt);
    if (combat) combat.update(dt, { moving: player.gait > 0 });

    // --- animation ---
    // Which clip the character should be in. Sprint has no cycle of its own,
    // so it reuses the run and lets the rate below carry the extra speed.
    // Standing in a fight is the combat stance rather than the idle.
    const fighting = combat && (combat.attacking || npcs.npcs.some((n) => n.fight === player));
    const still = fighting && anim.has('cidle') ? 'cidle' : 'idle';
    anim.setBase(player.gait === 0 ? still : player.gait <= WALK ? 'walk' : 'run');
    // A recorded clip is authored for one speed. Advancing its own clock in
    // proportion to how fast the character is actually moving keeps the feet
    // planted instead of skating -- the same rule as the procedural phase,
    // applied to a cycle someone else timed.
    anim.update(dt, (name) => {
      // Divided by the race's scale: a troll drawn 1.3x larger covers 1.3x
      // the ground per stride, so its cycle has to run that much slower.
      if (name === 'walk') return Math.max(player.gait, WALK) / (WALK * char.scale);
      if (name === 'run') return Math.max(player.gait, RUN) / (RUN * char.scale);
      return 1;
    });

    // Prefer DAoC's own animation; fall back to the procedural gait if
    // animconv has not been run.
    const blended = anim.pose();
    player.clip = blended ? anim.cur : null;
    player.blend = blended && anim.prev ? anim.fade : 1;
    if (!blended) {
      player.phase += player.gait > 0 ? (player.gait * dt / STRIDE) * Math.PI * 2 : 0;
      char.pose(player.phase, intensity);
    }

    const playerGround = groundAt(player.x, player.y);
    // A small vertical bob, twice per stride, sells the weight transfer. The
    // recorded clips already carry their own, so this is only for the
    // procedural fallback.
    const bob = player.clip
      ? 0
      : 1.6 * intensity * (1 - Math.cos(player.phase * 2)) * 0.5;
    const charModel = modelMatrix(player.x, playerGround + bob, player.y, player.yaw, char.scale);

    // --- place the camera ---
    let eye, target;
    const cp = Math.cos(cam.pitch), sp = Math.sin(cam.pitch);
    const fwd = [Math.sin(cam.yaw) * cp, sp, -Math.cos(cam.yaw) * cp];
    if (followMode) {
      const op = Math.cos(orbit.pitch), os = Math.sin(orbit.pitch);
      const look = [Math.sin(orbit.yaw) * op, -os, -Math.cos(orbit.yaw) * op];
      target = [player.x, playerGround + char.height * char.scale * 0.62, player.y];
      eye = [
        target[0] - look[0] * orbit.dist,
        target[1] - look[1] * orbit.dist,
        target[2] - look[2] * orbit.dist,
      ];
      // Never let the chase camera sink into the hillside behind.
      const floor = groundAt(eye[0], eye[2]) + 40;
      if (eye[1] < floor) eye[1] = floor;
      lastEye = eye;
    } else {
      const right = [Math.cos(cam.yaw), 0, Math.sin(cam.yaw)];
      const v = cam.speed * (keys.has('ShiftLeft') || keys.has('ShiftRight') ? 4 : 1) * dt;
      const step = (ax, s) => { cam.pos[0] += ax[0] * s; cam.pos[1] += ax[1] * s; cam.pos[2] += ax[2] * s; };
      if (keys.has('KeyW')) step(fwd, v);
      if (keys.has('KeyS')) step(fwd, -v);
      if (keys.has('KeyD')) step(right, v);
      if (keys.has('KeyA')) step(right, -v);
      if (keys.has('KeyE')) cam.pos[1] += v;
      if (keys.has('KeyQ')) cam.pos[1] -= v;
      eye = cam.pos;
      target = [cam.pos[0] + fwd[0], cam.pos[1] + fwd[1], cam.pos[2] + fwd[2]];
    }

    const w = canvas.clientWidth, h = canvas.clientHeight;
    const dpr = Math.min(devicePixelRatio || 1, 2);
    if (canvas.width !== Math.round(w * dpr) || canvas.height !== Math.round(h * dpr)) {
      canvas.width = Math.round(w * dpr);
      canvas.height = Math.round(h * dpr);
    }
    gl.viewport(0, 0, canvas.width, canvas.height);
    gl.clearColor(fogColor[0], fogColor[1], fogColor[2], 1);
    gl.clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT);

    const far = 160000;
    const near = followMode ? 15 : 20;
    const vp = mul(perspective(FOV, w / h, near, far), lookAt(eye, target, [0, 1, 0]));
    view.eye = eye; view.target = target;
    const fogStart = far * 0.25, fogEnd = far * 0.92;

    gl.useProgram(terrainProg);
    gl.uniformMatrix4fv(tU.uViewProj, false, vp);
    gl.uniform3fv(tU.uCamPos, eye);
    gl.uniform3fv(tU.uFogColor, fogColor);
    gl.uniform1f(tU.uFogStart, fogStart);
    gl.uniform1f(tU.uFogEnd, fogEnd);
    gl.uniform3fv(tU.uLightDir, [0.45, 0.78, 0.35]);
    gl.uniform1f(tU.uWireframe, wireframe ? 1 : 0);
    gl.uniform1i(tU.uAtlas, 0);
    gl.activeTexture(gl.TEXTURE0);
    gl.bindTexture(gl.TEXTURE_2D, tex);
    gl.bindVertexArray(terrainVAO);
    if (wireframe) {
      gl.bindBuffer(gl.ELEMENT_ARRAY_BUFFER, lbo);
      gl.drawElements(gl.LINES, lineIdx.length, gl.UNSIGNED_INT, 0);
    } else {
      gl.bindBuffer(gl.ELEMENT_ARRAY_BUFFER, ibo);
      gl.drawElements(gl.TRIANGLES, indices.length, gl.UNSIGNED_INT, 0);
    }

    if (props) {
      props.draw({
        viewProj: vp, camPos: eye,
        fogColor, fogStart, fogEnd, lightDir: [0.45, 0.78, 0.35],
      });
    }

    // The character goes in before the water so a submerged figure is tinted
    // by the translucent surface rather than drawn over it.
    char.draw({
      viewProj: vp, model: charModel, camPos: eye,
      fogColor, fogStart, fogEnd, lightDir: [0.45, 0.78, 0.35],
    });
    if (npcs) {
      npcs.draw({
        viewProj: vp, camPos: eye,
        fogColor, fogStart, fogEnd, lightDir: [0.45, 0.78, 0.35],
      });
    }
    // The target's ring: gold when selected, red while it is a fight.
    const tgt = combat && combat.target;
    if (tgt && !tgt.gone) {
      const hostile = combat.attacking || tgt.fight === player;
      const color = tgt.dead ? [0.55, 0.55, 0.55] : hostile ? [0.95, 0.25, 0.2] : [1.0, 0.82, 0.25];
      ring.draw(vp, tgt.x, tgt.y, tgt.body * 1.15 + 10, groundAt, color, now / 1000);
    }

    if (waters.length) {
      gl.useProgram(waterProg);
      gl.uniformMatrix4fv(wU.uViewProj, false, vp);
      gl.uniform3fv(wU.uCamPos, eye);
      gl.uniform3fv(wU.uFogColor, fogColor);
      gl.uniform1f(wU.uFogStart, fogStart);
      gl.uniform1f(wU.uFogEnd, fogEnd);
      gl.uniform1f(wU.uTime, now / 1000);
      gl.enable(gl.BLEND);
      gl.blendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA);
      gl.disable(gl.CULL_FACE);
      for (const wb of waters) {
        gl.bindVertexArray(wb.vao);
        gl.drawArrays(gl.TRIANGLES, 0, wb.count);
      }
      gl.enable(gl.CULL_FACE);
      gl.disable(gl.BLEND);
    }
    gl.bindVertexArray(null);

    fpsAccum += dt; fpsFrames++;
    if (fpsAccum >= 0.5) {
      $('i-fps').textContent = Math.round(fpsFrames / fpsAccum) + '  (speed ' + Math.round(cam.speed) + 'u/s)';
      fpsAccum = 0; fpsFrames = 0;
    }
    $('i-pos').textContent = followMode
      ? 'cell ' + (player.x / cell).toFixed(1) + ', ' + (player.y / cell).toFixed(1) +
        '  ground ' + Math.round(playerGround) + 'u'
      : cam.pos.map((n) => Math.round(n)).join(', ');
    $('i-gait').textContent = (player.gait
      ? Math.round(player.gait) + ' u/s ' +
        (player.gait >= SPRINT ? '(sprint)' : player.gait <= WALK ? '(walk)' : '(run)')
      : 'idle') +
      (player.clip
        ? '  · ' + (anim.prev
            ? anim.prev + ' → ' + anim.cur + ' ' + Math.round(anim.fade * 100) + '%'
            : anim.cur + '.kfa') + (anim.act ? ' + ' + anim.act + (anim.upper ? ' (upper)' : '') : '')
        : '  · procedural');
    $('i-mode').textContent = followMode ? 'third person' : 'free fly';
    if (combat) {
      const bar = (el, hp, max) => {
        const f = Math.max(0, hp / max);
        el.firstElementChild.style.width = (100 * f).toFixed(1) + '%';
        el.classList.toggle('low', f < 0.5 && f >= 0.25);
        el.classList.toggle('crit', f < 0.25);
      };
      $('you-name').textContent = char.manifest.title || 'You';
      $('you-hp').textContent = player.dead ? 'dead' : Math.ceil(player.hp) + ' / ' + player.maxHp;
      bar($('you-bar'), player.hp, player.maxHp);
      const t = combat.target, box = $('target-box');
      box.classList.toggle('on', !!t);
      if (t) {
        $('tgt-name').textContent = t.name;
        $('tgt-hp').textContent = t.dead ? 'dead'
          : Math.ceil(t.hp) + ' / ' + t.maxHp + '  ·  ' + Math.round(Math.hypot(t.x - player.x, t.y - player.y)) + 'u';
        bar($('tgt-bar'), t.hp, t.maxHp);
        box.classList.toggle('engaged', combat.attacking);
      }
    }
    if (npcs) {
      $('i-npc').textContent = npcs.npcs.length + ' placed, ' + npcs.drawn + ' in view, ' +
        npcs.types.size + ' kinds';
    }

    requestAnimationFrame(frame);
  }
  requestAnimationFrame(frame);
}

function bindAttrib(gl, loc, data, size) {
  const b = gl.createBuffer();
  gl.bindBuffer(gl.ARRAY_BUFFER, b);
  gl.bufferData(gl.ARRAY_BUFFER, data, gl.STATIC_DRAW);
  gl.enableVertexAttribArray(loc);
  gl.vertexAttribPointer(loc, size, gl.FLOAT, false, 0, 0);
  return b;
}

function uniforms(gl, prog, names) {
  const o = {};
  for (const n of names) o[n] = gl.getUniformLocation(prog, n);
  return o;
}

async function fetchJSON(url) {
  const r = await fetch(url);
  if (!r.ok) throw new Error('GET ' + url + ' -> ' + r.status + ' ' + r.statusText);
  return r.json();
}

async function fetchBuffer(url) {
  const r = await fetch(url);
  if (!r.ok) throw new Error('GET ' + url + ' -> ' + r.status + ' ' + r.statusText);
  return r.arrayBuffer();
}

function loadImage(url) {
  return new Promise((res, rej) => {
    const im = new Image();
    im.onload = () => res(im);
    im.onerror = () => rej(new Error('failed to load image ' + url));
    im.src = url;
  });
}

main().catch((e) => fail(e && e.stack ? e.stack : String(e)));
