// Scenery rendering: the trees, fences and buildings fixtures.csv places.
//
// Props never move, so propconv has already baked each model's scene graph
// into one flat vertex buffer. All that is left at runtime is placing copies,
// which is what instancing is for: one draw call per model-and-texture, with
// the position, heading and scale of every copy supplied as per-instance
// attributes. Zone 100 draws its ~970 props in about 120 calls that way.
//
// The vertex layout matches the character's convention: the model is Z-up,
// and the instance transform maps model Z to world height exactly as
// modelMatrix does in character.js.

const PROP_VS = `#version 300 es
precision highp float;
layout(location = 0) in vec3 aPos;
layout(location = 1) in vec3 aNormal;
layout(location = 2) in vec2 aUV;
layout(location = 3) in vec4 aInst;   // world x, height, world z, yaw
layout(location = 4) in float aScale;

uniform mat4 uViewProj;
uniform vec3 uCamPos;

out vec2 vUV;
out vec3 vNormal;
out float vDist;

void main() {
  float c = cos(aInst.w), s = sin(aInst.w);
  // Same basis as character.js modelMatrix: model X and Y span the ground
  // plane and model Z is up.
  vec3 world = vec3(
    aScale * (c * aPos.x - s * aPos.y) + aInst.x,
    aScale * aPos.z + aInst.y,
    aScale * (s * aPos.x + c * aPos.y) + aInst.z
  );
  vNormal = normalize(vec3(
    c * aNormal.x - s * aNormal.y,
    aNormal.z,
    s * aNormal.x + c * aNormal.y
  ));
  vUV = aUV;
  vDist = length(world - uCamPos);
  gl_Position = uViewProj * vec4(world, 1.0);
}`;

const PROP_FS = `#version 300 es
precision highp float;
in vec2 vUV;
in vec3 vNormal;
in float vDist;

uniform sampler2D uTex;
uniform vec3 uLightDir;
uniform vec3 uFogColor;
uniform float uFogStart;
uniform float uFogEnd;
uniform float uAlphaTest;

out vec4 outColor;

void main() {
  vec4 texel = texture(uTex, vUV);
  // Foliage is a cut-out: the canopy is a handful of quads whose texture is
  // mostly transparent. Discarding keeps the depth buffer honest without
  // needing the props sorted back to front.
  if (uAlphaTest > 0.5 && texel.a < 0.5) discard;
  vec3 n = normalize(vNormal);
  // Two-sided: foliage quads and thin boards are seen from both faces, and
  // the exporter does not wind them consistently.
  float lambert = abs(dot(n, normalize(uLightDir)));
  vec3 lit = texel.rgb * (0.45 + 0.55 * lambert);
  float fog = clamp((vDist - uFogStart) / max(uFogEnd - uFogStart, 1.0), 0.0, 1.0);
  outColor = vec4(mix(lit, uFogColor, fog), 1.0);
}`;

function compile(gl, type, src, label) {
  const sh = gl.createShader(type);
  gl.shaderSource(sh, src);
  gl.compileShader(sh);
  if (!gl.getShaderParameter(sh, gl.COMPILE_STATUS)) {
    throw new Error(label + ': ' + gl.getShaderInfoLog(sh));
  }
  return sh;
}

function program(gl, vsSrc, fsSrc, label) {
  const p = gl.createProgram();
  gl.attachShader(p, compile(gl, gl.VERTEX_SHADER, vsSrc, label + ' vs'));
  gl.attachShader(p, compile(gl, gl.FRAGMENT_SHADER, fsSrc, label + ' fs'));
  gl.linkProgram(p);
  if (!gl.getProgramParameter(p, gl.LINK_STATUS)) {
    throw new Error(label + ' link: ' + gl.getProgramInfoLog(p));
  }
  return p;
}

function loadImage(src) {
  return new Promise((resolve, reject) => {
    const img = new Image();
    img.onload = () => resolve(img);
    img.onerror = () => reject(new Error('texture failed to load: ' + src));
    img.src = src;
  });
}

export async function createProps(gl, base) {
  const manifest = await fetch(base + '/props.json').then((r) => {
    if (!r.ok) throw new Error('props.json: HTTP ' + r.status);
    return r.json();
  });
  const raw = await fetch(base + '/props.bin').then((r) => {
    if (!r.ok) throw new Error('props.bin: HTTP ' + r.status);
    return r.arrayBuffer();
  });

  const expected = manifest.vertexBytes + manifest.indexBytes;
  if (raw.byteLength !== expected) {
    throw new Error('props.bin is ' + raw.byteLength + ' bytes, manifest says ' + expected);
  }

  const prog = program(gl, PROP_VS, PROP_FS, 'props');
  const U = {};
  for (const n of ['uViewProj', 'uCamPos', 'uTex', 'uLightDir', 'uFogColor', 'uFogStart', 'uFogEnd', 'uAlphaTest']) {
    U[n] = gl.getUniformLocation(prog, n);
  }

  const vbo = gl.createBuffer();
  gl.bindBuffer(gl.ARRAY_BUFFER, vbo);
  gl.bufferData(gl.ARRAY_BUFFER, new Uint8Array(raw, 0, manifest.vertexBytes), gl.STATIC_DRAW);

  const ibo = gl.createBuffer();
  gl.bindBuffer(gl.ELEMENT_ARRAY_BUFFER, ibo);
  gl.bufferData(gl.ELEMENT_ARRAY_BUFFER,
    new Uint8Array(raw, manifest.vertexBytes, manifest.indexBytes), gl.STATIC_DRAW);

  // Group the instance list by model so each model gets one instance buffer.
  const perModel = manifest.models.map(() => []);
  for (const inst of manifest.instances) {
    if (inst.m >= 0 && inst.m < perModel.length) perModel[inst.m].push(inst);
  }

  const stride = manifest.stride;
  const models = manifest.models.map((m, i) => {
    const list = perModel[i];
    const data = new Float32Array(list.length * 5);
    for (let k = 0; k < list.length; k++) {
      const it = list[k];
      // The fixture's Z is absolute world height, so it goes straight in.
      data[k * 5 + 0] = it.p[0];
      data[k * 5 + 1] = it.p[2];
      data[k * 5 + 2] = it.p[1];
      data[k * 5 + 3] = it.yaw;
      data[k * 5 + 4] = it.s;
    }
    const ibuf = gl.createBuffer();
    gl.bindBuffer(gl.ARRAY_BUFFER, ibuf);
    gl.bufferData(gl.ARRAY_BUFFER, data, gl.STATIC_DRAW);

    const vao = gl.createVertexArray();
    gl.bindVertexArray(vao);
    gl.bindBuffer(gl.ARRAY_BUFFER, vbo);
    gl.enableVertexAttribArray(0);
    gl.vertexAttribPointer(0, 3, gl.FLOAT, false, stride, 0);
    gl.enableVertexAttribArray(1);
    gl.vertexAttribPointer(1, 3, gl.FLOAT, false, stride, 12);
    gl.enableVertexAttribArray(2);
    gl.vertexAttribPointer(2, 2, gl.FLOAT, false, stride, 24);
    gl.bindBuffer(gl.ARRAY_BUFFER, ibuf);
    gl.enableVertexAttribArray(3);
    gl.vertexAttribPointer(3, 4, gl.FLOAT, false, 20, 0);
    gl.vertexAttribDivisor(3, 1);
    gl.enableVertexAttribArray(4);
    gl.vertexAttribPointer(4, 1, gl.FLOAT, false, 20, 16);
    gl.vertexAttribDivisor(4, 1);
    gl.bindBuffer(gl.ELEMENT_ARRAY_BUFFER, ibo);
    gl.bindVertexArray(null);

    return { name: m.name, vao, count: list.length, groups: m.groups, max: m.max };
  });

  // One texture object per distinct image. A missing file is not fatal: the
  // prop draws white rather than taking the whole zone down with it.
  const names = new Set();
  for (const m of manifest.models) for (const g of m.groups) if (g.texture) names.add(g.texture);

  const white = gl.createTexture();
  gl.bindTexture(gl.TEXTURE_2D, white);
  gl.texImage2D(gl.TEXTURE_2D, 0, gl.RGBA, 1, 1, 0, gl.RGBA, gl.UNSIGNED_BYTE,
    new Uint8Array([255, 255, 255, 255]));

  const textures = new Map();
  let missing = 0;
  await Promise.all([...names].map(async (n) => {
    let img;
    try {
      img = await loadImage(base + '/tex/' + n + '.png');
    } catch {
      missing++;
      textures.set(n, white);
      return;
    }
    const t = gl.createTexture();
    gl.bindTexture(gl.TEXTURE_2D, t);
    gl.pixelStorei(gl.UNPACK_FLIP_Y_WEBGL, false);
    gl.texImage2D(gl.TEXTURE_2D, 0, gl.RGBA, gl.RGBA, gl.UNSIGNED_BYTE, img);
    gl.generateMipmap(gl.TEXTURE_2D);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT);
    textures.set(n, t);
  }));

  function draw(ctx) {
    gl.useProgram(prog);
    gl.uniformMatrix4fv(U.uViewProj, false, ctx.viewProj);
    gl.uniform3fv(U.uCamPos, ctx.camPos);
    gl.uniform3fv(U.uLightDir, ctx.lightDir);
    gl.uniform3fv(U.uFogColor, ctx.fogColor);
    gl.uniform1f(U.uFogStart, ctx.fogStart);
    gl.uniform1f(U.uFogEnd, ctx.fogEnd);
    gl.uniform1i(U.uTex, 0);
    gl.activeTexture(gl.TEXTURE0);
    // Cut-out foliage and single-sided boards are both two-sided in practice.
    gl.disable(gl.CULL_FACE);

    let calls = 0;
    for (const m of models) {
      if (!m.count) continue;
      gl.bindVertexArray(m.vao);
      for (const g of m.groups) {
        gl.bindTexture(gl.TEXTURE_2D, textures.get(g.texture) || white);
        gl.uniform1f(U.uAlphaTest, g.alpha ? 1 : 0);
        gl.drawElementsInstanced(gl.TRIANGLES, g.count, gl.UNSIGNED_INT, g.first * 4, m.count);
        calls++;
      }
    }
    gl.bindVertexArray(null);
    gl.enable(gl.CULL_FACE);
    return calls;
  }

  return {
    draw,
    models,
    manifest,
    missingTextures: missing,
    instanceCount: manifest.instances.length,
    triangleCount: manifest.indexCount / 3,
    drawCalls: models.reduce((n, m) => n + (m.count ? m.groups.length : 0), 0),
  };
}
