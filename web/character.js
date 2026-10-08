// Skinned character rendering for the DAoC viewer.
//
// charconv exports the model in its own authoring space: Z-up, X lateral,
// facing -Y, with a vertex bound to up to four bones. Posing happens here
// rather than in the converter so the skeleton stays animatable.
//
// The skeleton is DAoC's, and its joints use 3ds Max Biped names, which is
// what makes animating it tractable: "Bip01 L Thigh" is unambiguously the left
// thigh, so a gait can be written against the names without needing to parse
// the .kfa animation files.
//
// Each bone's local frame is arbitrary, so a joint is not rotated in its own
// axes. Instead the rotation is specified in the body's frame and conjugated
// into the joint's: R_local = Wt * R_body * W, where W is the joint's bind
// world rotation. That way "swing the thigh forward" means the same thing for
// every joint regardless of how the artist oriented it.

import { Skeleton, xform, Clip } from './skeleton.js';
import { LIGHT_GLSL, LIGHT_UNIFORMS, bindLight } from './lighting.js';

const MAX_BONES = 52; // the largest bone count on any one shape, with headroom

const CHAR_VS = `#version 300 es
precision highp float;
layout(location=0) in vec3 aPos;
layout(location=1) in vec3 aNormal;
layout(location=2) in vec2 aUV;
layout(location=3) in vec4 aBone;
layout(location=4) in vec4 aWeight;
uniform mat4 uViewProj;
uniform mat4 uModel;
uniform vec3 uCamPos;
uniform vec4 uBones[${MAX_BONES * 3}];
out vec3 vNormal;
out vec2 vUV;
out vec3 vWorld;
out float vDist;

// Bone matrices arrive as three rows of four; the fourth row is implicit.
mat4 boneMatrix(int i) {
  vec4 r0 = uBones[i * 3 + 0];
  vec4 r1 = uBones[i * 3 + 1];
  vec4 r2 = uBones[i * 3 + 2];
  return mat4(
    r0.x, r1.x, r2.x, 0.0,
    r0.y, r1.y, r2.y, 0.0,
    r0.z, r1.z, r2.z, 0.0,
    r0.w, r1.w, r2.w, 1.0);
}

void main() {
  vec4 p = vec4(aPos, 1.0);
  vec3 skinned = vec3(0.0);
  vec3 normal = vec3(0.0);
  for (int k = 0; k < 4; k++) {
    float w = aWeight[k];
    if (w <= 0.0) continue;
    mat4 m = boneMatrix(int(aBone[k]));
    skinned += w * (m * p).xyz;
    normal += w * (mat3(m) * aNormal);
  }
  vec4 world = uModel * vec4(skinned, 1.0);
  vNormal = mat3(uModel) * normal;
  vUV = aUV;
  vWorld = world.xyz;
  vDist = length(world.xyz - uCamPos);
  gl_Position = uViewProj * world;
}
`;

const CHAR_FS = `#version 300 es
precision highp float;
in vec3 vNormal;
in vec2 vUV;
in vec3 vWorld;
in float vDist;
uniform sampler2D uTex;
uniform float uHasTex;
uniform vec3 uDiffuse;
uniform vec3 uFogColor;
uniform float uFogStart;
uniform float uFogEnd;
${LIGHT_GLSL}
out vec4 outColor;
void main() {
  vec3 base = uDiffuse;
  if (uHasTex > 0.5) base *= texture(uTex, vUV).rgb;
  vec3 n = normalize(vNormal);
  // Two-sided: the meshes are thin shells and some faces end up away from the
  // light, which otherwise leaves whole limbs unlit.
  vec3 lit = base * lightAt(n, vWorld, true);
  float fog = clamp((vDist - uFogStart) / max(uFogEnd - uFogStart, 1.0), 0.0, 1.0);
  outColor = vec4(mix(lit, uFogColor, fog), 1.0);
}
`;

export async function createCharacter(gl, base, helpers) {
  const { program, uniforms, loadImage } = helpers;
  const man = await fetch(base + '/char.json').then((r) => {
    if (!r.ok) throw new Error('GET ' + base + '/char.json -> ' + r.status);
    return r.json();
  });
  const blob = await fetch(base + '/mesh.bin').then((r) => {
    if (!r.ok) throw new Error('GET ' + base + '/mesh.bin -> ' + r.status);
    return r.arrayBuffer();
  });

  const stride = man.stride;
  const vertBytes = man.vertexCount * stride;
  const expected = vertBytes + man.indexCount * 4;
  if (blob.byteLength !== expected) {
    throw new Error(
      'mesh.bin is ' + blob.byteLength + ' bytes, manifest implies ' + expected);
  }

  const prog = program(gl, CHAR_VS, CHAR_FS, 'character');
  const U = uniforms(gl, prog, [
    'uViewProj', 'uModel', 'uCamPos', 'uBones', 'uTex', 'uHasTex',
    'uDiffuse', 'uFogColor', 'uFogStart', 'uFogEnd', ...LIGHT_UNIFORMS,
  ]);

  const vao = gl.createVertexArray();
  gl.bindVertexArray(vao);
  const vbo = gl.createBuffer();
  gl.bindBuffer(gl.ARRAY_BUFFER, vbo);
  gl.bufferData(gl.ARRAY_BUFFER, new Uint8Array(blob, 0, vertBytes), gl.STATIC_DRAW);
  const f = gl.FLOAT, ub = gl.UNSIGNED_BYTE;
  //                  loc size type        norm   stride  offset
  gl.enableVertexAttribArray(0); gl.vertexAttribPointer(0, 3, f, false, stride, 0);
  gl.enableVertexAttribArray(1); gl.vertexAttribPointer(1, 3, f, false, stride, 12);
  gl.enableVertexAttribArray(2); gl.vertexAttribPointer(2, 2, f, false, stride, 24);
  // Bone indices stay as raw 0..255 values, so normalized must be false.
  gl.enableVertexAttribArray(3); gl.vertexAttribPointer(3, 4, ub, false, stride, 32);
  gl.enableVertexAttribArray(4); gl.vertexAttribPointer(4, 4, f, false, stride, 36);
  const ibo = gl.createBuffer();
  gl.bindBuffer(gl.ELEMENT_ARRAY_BUFFER, ibo);
  gl.bufferData(gl.ELEMENT_ARRAY_BUFFER, new Uint8Array(blob, vertBytes), gl.STATIC_DRAW);
  gl.bindVertexArray(null);

  // Textures, one per shape, loaded in parallel and shared by name.
  const texCache = new Map();
  await Promise.all(
    [...new Set(man.shapes.map((s) => s.texture).filter(Boolean))].map(async (name) => {
      const img = await loadImage(base + '/tex/' + name);
      const t = gl.createTexture();
      gl.bindTexture(gl.TEXTURE_2D, t);
      gl.texImage2D(gl.TEXTURE_2D, 0, gl.RGBA, gl.RGBA, gl.UNSIGNED_BYTE, img);
      gl.generateMipmap(gl.TEXTURE_2D);
      gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR);
      gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR);
      texCache.set(name, t);
    }));

  // --- skeleton ---
  const bones = man.bones;
  const skel = new Skeleton(bones);

  // --- recorded animation ---
  // The game's own .kfa clips, if animconv has been run. They are optional:
  // without them the procedural gait still drives the figure, which is what
  // the viewer falls back to.
  const clips = {};
  try {
    const idx = await fetch(base + '/anim/index.json').then((r) => (r.ok ? r.json() : null));
    if (idx) {
      await Promise.all(idx.animations.map(async (a) => {
        const j = await fetch(base + '/anim/' + a.file).then((r) => (r.ok ? r.json() : null));
        if (j) clips[a.name] = new Clip(j, skel);
      }));
    }
  } catch (e) {
    console.warn('recorded animation not loaded:', e.message);
  }

  // --- held items ---
  // charconv converts each into equip/<slot>/ as a one-bone character baked
  // into the frame of its HELD marker, so holding one is setting that bone to
  // the socket's world transform and drawing it with this figure's model
  // matrix. A missing item is skipped, not fatal.
  const equip = [];
  for (const e of man.equip || []) {
    try {
      const socket = skel.boneId(e.bone);
      if (socket < 0) throw new Error('no socket ' + e.bone);
      const item = await createCharacter(gl, base + '/equip/' + e.slot, helpers);
      equip.push({ ...e, socket, item });
    } catch (err) {
      console.warn('equipment ' + e.slot + ' not loaded:', err.message);
    }
  }

  // Scratch, reused every frame.
  const tmpA = xform();
  const skinBuf = new Float32Array(MAX_BONES * 3 * 4);

  function draw(ctx) {
    gl.useProgram(prog);
    gl.uniformMatrix4fv(U.uViewProj, false, ctx.viewProj);
    gl.uniformMatrix4fv(U.uModel, false, ctx.model);
    gl.uniform3fv(U.uCamPos, ctx.camPos);
    gl.uniform3fv(U.uFogColor, ctx.fogColor);
    gl.uniform1f(U.uFogStart, ctx.fogStart);
    gl.uniform1f(U.uFogEnd, ctx.fogEnd);
    bindLight(gl, U, ctx.light);
    gl.uniform1i(U.uTex, 0);
    gl.activeTexture(gl.TEXTURE0);
    gl.bindVertexArray(vao);
    // The model-to-scene mapping flips handedness, so face winding is
    // reversed; with under a thousand triangles it is cheapest to just not
    // cull rather than track the flip through every transform.
    gl.disable(gl.CULL_FACE);
    for (const s of man.shapes) {
      if (s.bones.length > MAX_BONES) {
        throw new Error('shape ' + s.name + ' needs ' + s.bones.length + ' bones');
      }
      for (let b = 0; b < s.bones.length; b++) {
        skel.skinMatrix(s.bones[b], s.invBind[b], tmpA);
        skinBuf.set(tmpA, b * 12);
      }
      gl.uniform4fv(U.uBones, skinBuf.subarray(0, s.bones.length * 12));
      gl.uniform3fv(U.uDiffuse, s.diffuse);
      const t = s.texture && texCache.get(s.texture);
      gl.uniform1f(U.uHasTex, t ? 1 : 0);
      if (t) gl.bindTexture(gl.TEXTURE_2D, t);
      gl.drawElements(gl.TRIANGLES, s.count, gl.UNSIGNED_INT, s.first * 4);
    }
    gl.enable(gl.CULL_FACE);
    gl.bindVertexArray(null);

    // Whatever is held follows its socket in the pose just drawn.
    for (const e of equip) {
      e.item.skeleton.world[0].set(skel.world[e.socket]);
      e.item.draw(ctx);
    }
  }

  return {
    manifest: man,
    height: man.height,
    // How much larger than authored the game draws this model, from
    // monsters.csv. 1 for anything converted without a table row.
    scale: man.scale || 1,
    forwardY: man.forwardY || -1,
    triangles: man.indexCount / 3,
    boneCount: bones.length,
    missingJoints: skel.missingJoints,
    pose: (phase, intensity) => skel.poseWalk(phase, intensity),
    // poseClip plays a recorded clip by name; returns false if it is absent,
    // so the caller can fall back to the procedural gait.
    poseClip: (name, time) => {
      const c = clips[name];
      if (!c) return false;
      skel.poseClip(c, time);
      return true;
    },
    // poseBlend cross-fades from one clip to another. A missing outgoing clip
    // is fine -- the fade just starts from nothing -- but the incoming one
    // has to exist, and false means the caller should fall back.
    poseBlend: (from, fromTime, to, toTime, f) => {
      const b = clips[to];
      if (!b) return false;
      skel.poseCross(clips[from] || null, fromTime, b, toTime, clips[from] ? f : 1);
      return true;
    },
    // poseLayered is poseBlend with a one-shot action on top at weight w,
    // over the upper body only when upper is set.
    poseLayered: (from, fromT, to, toT, f, act, actT, w, upper) => {
      const b = clips[to];
      if (!b) return false;
      skel.poseLayered(clips[from] || null, fromT, b, toT, clips[from] ? f : 1,
        clips[act] || null, actT, w, upper ? skel.upperBodyMask() : null);
      return true;
    },
    clipDuration: (name) => (clips[name] ? clips[name].duration : 0),
    clipRate: (name) => (clips[name] ? clips[name].rate : 1),
    // groundSpeed is a clip's own travel speed in model units per second,
    // read from its planted feet; 0 if absent or not a gait.
    groundSpeed: (name) => (clips[name] ? skel.groundSpeed(clips[name]) : 0),
    clips,
    clipNames: Object.keys(clips),
    skeleton: skel,
    equip,
    // What it holds, which decides how it fights.
    armed: equip.some((e) => e.slot === 'right'),
    shield: equip.some((e) => e.slot === 'left'),
    draw,
  };
}

// modelMatrix places the character in the scene.
//
// The model is authored Z-up with X lateral; the scene is Y-up with world X
// and Y on the ground plane. Both the terrain and the character use the same
// mapping (model x -> scene x, model z -> scene y, model y -> scene z) so that
// they agree, at the cost of a handedness flip that draw() absorbs.
// yaw follows the same convention as the viewer's camera: at yaw y the
// character faces (sin y, 0, -cos y), so pointing the orbit camera and the
// character with the same number puts the camera behind them.
export function modelMatrix(x, groundY, z, yaw, scale = 1) {
  const c = Math.cos(yaw) * scale, s = Math.sin(yaw) * scale;
  // Columns are the images of model X, Y and Z. Model -Y is the facing
  // direction, so model Y maps to (-sin, 0, cos) and the figure looks along
  // (sin, 0, -cos) as required.
  return new Float32Array([
    c, 0, s, 0,
    -s, 0, c, 0,
    0, scale, 0, 0,
    x, groundY, z, 1,
  ]);
}
