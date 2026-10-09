// Grass: the region's sprite table scattered over the ground near the camera.
//
// zoneconv wrote the sprites (grass_mid.csv for the Vale of Mularn), their
// atlas, and a byte pair per heightmap cell: which group grows there and how
// thickly. Each cell's plants are generated from a generator seeded by the
// cell, so the same tuft stands in the same place every time it comes back
// into range, and only cells within RANGE of the camera hold any.
//
// The table's shapes are not documented; read against the atlas they are:
// 0 and 1 upright plants, drawn as two crossed cards; 2 the rocks and 4 the
// lily leaf, pictures seen from above, laid flat; 3 the fern, a frond drawn
// four times radiating outward. Length is a card's width -- grass clump 1 is
// 12 tall and 8 long, and its picture is 62 by 87 pixels -- and a flat
// sprite's footprint is its length by its width.

import { LIGHT_GLSL, LIGHT_UNIFORMS, bindLight } from './lighting.js';

// How far from the camera grass grows, and where it starts to shrink away.
const RANGE = 2200;
const FADE_FROM = 1500;
// Plants in a cell at full density. A cell is 256 units square, and the
// zone's densities mostly run 40-120 of 255, so this gives one plant every
// 20 units or so: a meadow rather than a scatter. How dense the game draws
// them is a client setting, not data; this is chosen by eye.
const PER_CELL = 550;

// Groups that grow in water. Everything else stops at the shore.
const WET_GROUPS = new Set([3, 4, 5, 8]);

const GRASS_VS = `#version 300 es
precision highp float;
layout(location=0) in vec3 aPos;    // card corner, unit size, y up
layout(location=1) in vec2 aUV;     // 0..1 across the sprite's rectangle
layout(location=2) in vec4 aInst;   // x, ground height, z, yaw
layout(location=3) in vec3 aSize;   // length, height, width
layout(location=4) in vec4 aRect;   // u0, v0, u1, v1
layout(location=5) in vec3 aTint;
uniform mat4 uViewProj;
uniform vec3 uCamPos;
uniform float uTime;
uniform vec2 uFade;
out vec2 vUV;
out vec3 vTint;
out vec3 vWorld;
out float vDist;
void main() {
  float d = distance(aInst.xz, uCamPos.xz);
  // Shrink into the ground towards the edge of the range, so the edge is
  // not a line where grass begins.
  float s = 1.0 - smoothstep(uFade.x, uFade.y, d);
  vec3 p = aPos * aSize.xyz * s;
  float c = cos(aInst.w), sn = sin(aInst.w);
  vec3 w = vec3(c * p.x - sn * p.z, p.y, sn * p.x + c * p.z);
  // Sway: the tops of the cards move, the roots stay put.
  float sway = aPos.y * (sin(uTime * 1.7 + aInst.x * 0.013 + aInst.z * 0.011) * 0.08);
  w.xz += vec2(sway, sway * 0.6) * aSize.y;
  vWorld = aInst.xyz + w;
  vUV = mix(aRect.xy, aRect.zw, aUV);
  vTint = aTint;
  vDist = length(vWorld - uCamPos);
  gl_Position = uViewProj * vec4(vWorld, 1.0);
}
`;

const GRASS_FS = `#version 300 es
precision highp float;
in vec2 vUV;
in vec3 vTint;
in vec3 vWorld;
in float vDist;
uniform sampler2D uAtlas;
uniform vec3 uFogColor;
uniform float uFogStart;
uniform float uFogEnd;
${LIGHT_GLSL}
out vec4 outColor;
void main() {
  vec4 t = texture(uAtlas, vUV);
  if (t.a < 0.35) discard;
  // Lit as the ground it stands on: from above, shadowed where the ground is.
  vec3 lit = t.rgb * vTint * lightAt(vec3(0.0, 1.0, 0.0), vWorld, false);
  float fog = clamp((vDist - uFogStart) / max(uFogEnd - uFogStart, 1.0), 0.0, 1.0);
  outColor = vec4(mix(lit, uFogColor, fog), t.a);
}
`;

// Unit meshes, one per kind of shape. Corners are x across, y up, z across;
// uv runs 0..1 over the sprite with v = 0 at the top of the picture.
function quad(a, b, c, d) {
  // a b top, c d bottom; each [x, y, z]
  return [
    [...c, 0, 1], [...d, 1, 1], [...b, 1, 0],
    [...c, 0, 1], [...b, 1, 0], [...a, 0, 0],
  ];
}
const CROSS = [
  ...quad([-0.5, 1, 0], [0.5, 1, 0], [-0.5, 0, 0], [0.5, 0, 0]),
  ...quad([0, 1, -0.5], [0, 1, 0.5], [0, 0, -0.5], [0, 0, 0.5]),
];
// A flat picture lying on the ground, a little above it.
const FLAT = quad([-0.5, 0.15, -0.5], [0.5, 0.15, -0.5], [-0.5, 0.15, 0.5], [0.5, 0.15, 0.5]);
// Four fronds leaning out from the root. The fern is scaled by its height
// alone: a frond is as long as the plant is tall, and as wide as its
// picture is -- the fern rectangle is 79 by 179 pixels, 0.44 -- across.
const FERN = [0, 1, 2, 3].flatMap((k) => {
  const a = (k * Math.PI) / 2, ca = Math.cos(a), sa = Math.sin(a);
  const lean = 0.75, up = 0.66, half = 0.22; // tip out and up, per unit length
  const side = (x, r, y) => [x * -sa + r * ca, y, x * ca + r * sa];
  return quad(side(-half, lean, up), side(half, lean, up), side(-half, 0, 0), side(half, 0, 0));
});

function mulberry32(seed) {
  let a = seed >>> 0;
  return () => {
    a = (a + 0x6D2B79F5) >>> 0;
    let t = a;
    t = Math.imul(t ^ (t >>> 15), t | 1);
    t ^= t + Math.imul(t ^ (t >>> 7), t | 61);
    return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
  };
}

// createGrass needs the zone's grass spec from zone.json, the cell size and
// grid, and the world's groundAt and isWet. The grass map may be finer than
// the heightmap (twice as fine in the Shrouded Isles and the frontiers):
// spec.cellsPerSide says so, and its cells are smaller to match.
export async function createGrass(gl, base, spec, world, { program, uniforms, loadImage }) {
  const { groundAt, isWet } = world;
  const grid = spec.cellsPerSide || world.grid;
  const cell = (world.cell * world.grid) / grid;
  // Plants per cell, kept to the same number per unit of ground.
  const perCell = PER_CELL * (cell / world.cell) ** 2;
  const [img, mapBuf] = await Promise.all([
    loadImage(base + '/' + spec.atlas),
    fetch(base + '/' + spec.map).then((r) => {
      if (!r.ok) throw new Error(spec.map + ': HTTP ' + r.status);
      return r.arrayBuffer();
    }),
  ]);
  const map = new Uint8Array(mapBuf);
  if (map.length !== grid * grid * 2) throw new Error('grass map is ' + map.length + ' bytes');

  const prog = program(gl, GRASS_VS, GRASS_FS, 'grass');
  const U = uniforms(gl, prog, ['uViewProj', 'uCamPos', 'uTime', 'uFade', 'uAtlas',
    'uFogColor', 'uFogStart', 'uFogEnd', ...LIGHT_UNIFORMS]);

  const tex = gl.createTexture();
  gl.bindTexture(gl.TEXTURE_2D, tex);
  gl.texImage2D(gl.TEXTURE_2D, 0, gl.RGBA, gl.RGBA, gl.UNSIGNED_BYTE, img);
  gl.generateMipmap(gl.TEXTURE_2D);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);
  gl.bindTexture(gl.TEXTURE_2D, null);

  // Sprites by group, with each group's total chance for picking.
  const px = spec.atlasPx;
  const groups = new Map();
  for (const s of spec.sprites) {
    if (!(s.chance > 0)) continue;
    const g = groups.get(s.group) || { total: 0, sprites: [] };
    g.total += s.chance;
    g.sprites.push({
      ...s,
      // Half a texel in from the rectangle's edge, so mipmaps do not bleed
      // the neighbouring picture in.
      uv: [(s.rect[0] + 0.5) / px, (s.rect[2] + 0.5) / px, (s.rect[1] - 0.5) / px, (s.rect[3] - 0.5) / px],
      tint: s.tint ? s.tint.map((v) => v / 255) : [1, 1, 1],
      kind: s.shape === 2 || s.shape === 4 ? 1 : s.shape === 3 ? 2 : 0,
    });
    groups.set(s.group, g);
  }

  const KINDS = [CROSS, FLAT, FERN];
  const FLOATS = 14; // per instance: inst 4, size 3, rect 4, tint 3
  const kinds = KINDS.map((mesh) => {
    const vao = gl.createVertexArray();
    gl.bindVertexArray(vao);
    const vbo = gl.createBuffer();
    gl.bindBuffer(gl.ARRAY_BUFFER, vbo);
    gl.bufferData(gl.ARRAY_BUFFER, new Float32Array(mesh.flat()), gl.STATIC_DRAW);
    gl.enableVertexAttribArray(0);
    gl.vertexAttribPointer(0, 3, gl.FLOAT, false, 20, 0);
    gl.enableVertexAttribArray(1);
    gl.vertexAttribPointer(1, 2, gl.FLOAT, false, 20, 12);
    const ibo = gl.createBuffer();
    gl.bindBuffer(gl.ARRAY_BUFFER, ibo);
    const attr = (loc, n, off) => {
      gl.enableVertexAttribArray(loc);
      gl.vertexAttribPointer(loc, n, gl.FLOAT, false, FLOATS * 4, off * 4);
      gl.vertexAttribDivisor(loc, 1);
    };
    attr(2, 4, 0); attr(3, 3, 4); attr(4, 4, 7); attr(5, 3, 11);
    gl.bindVertexArray(null);
    return { vao, ibo, verts: mesh.length, count: 0 };
  });

  // One cell's plants, per kind, made once and kept.
  const cache = new Map();
  function cellPlants(cx, cy) {
    const key = cy * grid + cx;
    let out = cache.get(key);
    if (out) return out;
    out = [[], [], []];
    // A cell's group and density are sampled per plant rather than per
    // cell: density bilinearly between cell centres, and the group from a
    // point jittered up to half a cell away, so neighbouring groups
    // interleave along their border instead of meeting on a straight line.
    const at = (x, y) => (Math.min(grid - 1, Math.max(0, y)) * grid + Math.min(grid - 1, Math.max(0, x))) * 2;
    let most = 0;
    for (let dy = -1; dy <= 1; dy++) for (let dx = -1; dx <= 1; dx++) most = Math.max(most, map[at(cx + dx, cy + dy) + 1]);
    const n = Math.round((most / 255) * perCell);
    const rng = mulberry32(key * 2654435761);
    for (let i = 0; i < n; i++) {
      const fx = cx + rng(), fy = cy + rng();
      const x = fx * cell, y = fy * cell;
      const bx = fx - 0.5, by = fy - 0.5, x0 = Math.floor(bx), y0 = Math.floor(by), tx = bx - x0, ty = by - y0;
      const dens = (map[at(x0, y0) + 1] * (1 - tx) + map[at(x0 + 1, y0) + 1] * tx) * (1 - ty)
        + (map[at(x0, y0 + 1) + 1] * (1 - tx) + map[at(x0 + 1, y0 + 1) + 1] * tx) * ty;
      if (rng() * most >= dens) continue;
      const group = map[at(Math.floor(fx + rng() - 0.5), Math.floor(fy + rng() - 0.5))];
      const g = groups.get(group);
      if (!g) continue;
      let pick = rng() * g.total, s = g.sprites[0];
      for (const t of g.sprites) { pick -= t.chance; if (pick <= 0) { s = t; break; } }
      const yaw = rng() * Math.PI * 2;
      const spread = (b) => b[0] + (rng() * 2 - 1) * b[1];
      const k = 1 + ((rng() * 2 - 1) * s.scale) / 100;
      const len = Math.max(1, spread(s.length)) * k;
      const hgt = Math.max(1, spread(s.height)) * k;
      const wid = Math.max(1, spread(s.width)) * k;
      if (!WET_GROUPS.has(group) && isWet(x, y)) continue;
      const size = s.kind === 1 ? [len, Math.max(hgt, 2), wid] : s.kind === 2 ? [hgt, hgt, hgt] : [len, hgt, len];
      out[s.kind].push(x, groundAt(x, y), y, yaw, ...size,
        ...s.uv, ...s.tint);
    }
    cache.set(key, out);
    return out;
  }

  let centre = null, drawnPlants = 0;
  // refresh rebuilds the instance buffers when the camera has moved into
  // another cell.
  function refresh(camX, camZ) {
    const cx = Math.floor(camX / cell), cy = Math.floor(camZ / cell);
    if (centre && centre[0] === cx && centre[1] === cy) return;
    centre = [cx, cy];
    const r = Math.ceil(RANGE / cell) + 1;
    const lists = [[], [], []];
    for (let y = cy - r; y <= cy + r; y++) {
      for (let x = cx - r; x <= cx + r; x++) {
        if (x < 0 || y < 0 || x >= grid || y >= grid) continue;
        if (Math.hypot((x + 0.5) * cell - camX, (y + 0.5) * cell - camZ) > RANGE + cell) continue;
        const p = cellPlants(x, y);
        for (let k = 0; k < 3; k++) lists[k].push(p[k]);
      }
    }
    drawnPlants = 0;
    kinds.forEach((kd, k) => {
      const len = lists[k].reduce((n, a) => n + a.length, 0);
      const data = new Float32Array(len);
      let o = 0;
      for (const a of lists[k]) { data.set(a, o); o += a.length; }
      gl.bindBuffer(gl.ARRAY_BUFFER, kd.ibo);
      gl.bufferData(gl.ARRAY_BUFFER, data, gl.DYNAMIC_DRAW);
      kd.count = len / FLOATS;
      drawnPlants += kd.count;
    });
    // Forget cells far behind, so a long walk does not hoard them.
    if (cache.size > 4000) cache.clear();
  }

  function draw(ctx) {
    refresh(ctx.camPos[0], ctx.camPos[2]);
    gl.useProgram(prog);
    gl.uniformMatrix4fv(U.uViewProj, false, ctx.viewProj);
    gl.uniform3fv(U.uCamPos, ctx.camPos);
    gl.uniform1f(U.uTime, ctx.time);
    gl.uniform2f(U.uFade, FADE_FROM, RANGE);
    gl.uniform3fv(U.uFogColor, ctx.fogColor);
    gl.uniform1f(U.uFogStart, ctx.fogStart);
    gl.uniform1f(U.uFogEnd, ctx.fogEnd);
    bindLight(gl, U, ctx.light);
    gl.uniform1i(U.uAtlas, 0);
    gl.activeTexture(gl.TEXTURE0);
    gl.bindTexture(gl.TEXTURE_2D, tex);
    gl.disable(gl.CULL_FACE);
    // With multisampling, the cut-out's alpha becomes coverage, which
    // softens the edges of thousands of thin blades.
    gl.enable(gl.SAMPLE_ALPHA_TO_COVERAGE);
    for (const kd of kinds) {
      if (!kd.count) continue;
      gl.bindVertexArray(kd.vao);
      gl.drawArraysInstanced(gl.TRIANGLES, 0, kd.verts, kd.count);
    }
    gl.disable(gl.SAMPLE_ALPHA_TO_COVERAGE);
    gl.bindVertexArray(null);
    gl.enable(gl.CULL_FACE);
  }

  return {
    draw,
    get plants() { return drawnPlants; },
    // For tests: one cell's plants, by kind.
    cellPlants,
  };
}
