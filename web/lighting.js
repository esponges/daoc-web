// Sunlight and shadows, shared by every shader that draws the world.
//
// The zone files carry no sun: SECTOR.DAT gives a fog colour and nothing
// else, and shademap.pcx is not a lightmap (no sun direction explains it; it
// follows how bright the ground is, snow against grass). So the sun here is
// this project's: one direction, a warm direct light, and an ambient that
// comes from the sky above and the ground below.
//
// Shadows are a single depth map rendered from the sun, square around the
// player, re-centred each frame. Everything that draws the world draws into
// it with its own shader -- terrain, props, characters -- with the light's
// matrix for the camera's, so a tree's cut-out canopy casts the same holes it
// shows. Outside the square there are no cast shadows, and the edge fades so
// the boundary is not a line on the ground.

// Until a sky file says otherwise (setSkyLight), these stand in.
// Direction toward the sun, scene space (x, height, z): south-east of
// overhead, about 50 degrees up, so shadows read without being stretched.
export const SUN_DIR = normalize3([0.45, 0.78, 0.35]);

let SUN_COLOR = [0.78, 0.73, 0.63];
let SKY_COLOR = [0.56, 0.61, 0.70];
let GROUND_COLOR = [0.40, 0.37, 0.32];

// The sky file's light is a colour times an amount for each of ambient and
// direct (sky_midgard.dat: a pale blue ambient at 0.6, white sun at 0.5).
// How the game's renderer scales those is not known, and taken at face value
// they leave sunlit grass at about 0.9 of its texture; EXPOSURE lifts the
// whole scene so it lands near 1.2, keeping the file's balance between the
// two. The ground below gives back a little less than the sky above.
const EXPOSURE = 1.35;
const BOUNCE = 0.7;

// setSkyLight takes the sky file's ambient and direct light, 0..1 RGB.
export function setSkyLight(ambient, sun) {
  SKY_COLOR = ambient.map((v) => v * EXPOSURE);
  GROUND_COLOR = ambient.map((v) => v * EXPOSURE * BOUNCE);
  SUN_COLOR = sun.map((v) => v * EXPOSURE);
}

// The texture unit the shadow map lives on in every program. The terrain
// takes 0-3 for its atlas and ground layers.
export const SHADOW_UNIT = 5;

// GLSL shared by the world shaders. lightAt gives the light reaching a
// surface: ambient from the sky and ground plus the sun, the sun's share
// scaled by the shadow. twoSided lights a face from either side, for leaves
// and thin shells whose normals cannot be trusted to point outward.
export const LIGHT_GLSL = `
uniform vec3 uSunDir;
uniform vec3 uSunColor;
uniform vec3 uSkyColor;
uniform vec3 uGroundColor;
uniform highp sampler2DShadow uShadowMap;
uniform mat4 uLightVP;
uniform float uShadowOn;
uniform vec2 uShadowTexel; // one texel in shadow-map uv, and in world units

float sunShadow(vec3 world, vec3 n) {
  if (uShadowOn < 0.5) return 1.0;
  // Push the sample off the surface along its normal, a texel and a half,
  // so a surface does not shadow itself where depth and texel disagree.
  vec4 lp = uLightVP * vec4(world + n * uShadowTexel.y * 1.5, 1.0);
  vec3 p = lp.xyz * 0.5 + 0.5;
  // Fade out over the outer tenth of the map rather than stop at its edge.
  vec2 e = min(p.xy, 1.0 - p.xy);
  float edge = smoothstep(0.0, 0.1, min(e.x, e.y));
  if (edge <= 0.0 || p.z >= 1.0) return 1.0;
  float s = 0.0;
  for (int y = -1; y <= 1; y++) {
    for (int x = -1; x <= 1; x++) {
      s += texture(uShadowMap, vec3(p.xy + vec2(x, y) * uShadowTexel.x, p.z - 0.0005));
    }
  }
  return mix(1.0, s / 9.0, edge);
}

vec3 lightAt(vec3 n, vec3 world, bool twoSided) {
  float d = dot(n, uSunDir);
  float sun = twoSided ? abs(d) : max(d, 0.0);
  // Only a face toward the sun can be shadowed; one facing away is dark
  // already, and testing it would only add acne.
  if (sun > 0.0) sun *= sunShadow(world, d < 0.0 ? -n : n);
  vec3 amb = mix(uGroundColor, uSkyColor, n.y * 0.5 + 0.5);
  return amb + uSunColor * sun;
}
`;

export const LIGHT_UNIFORMS = ['uSunDir', 'uSunColor', 'uSkyColor', 'uGroundColor',
  'uShadowMap', 'uLightVP', 'uShadowOn', 'uShadowTexel'];

// bindLight sets the shared uniforms on the current program. light is the
// object the shadow map's begin() returns, or null for no shadows.
export function bindLight(gl, U, light) {
  gl.uniform3fv(U.uSunDir, SUN_DIR);
  gl.uniform3fv(U.uSunColor, SUN_COLOR);
  gl.uniform3fv(U.uSkyColor, SKY_COLOR);
  gl.uniform3fv(U.uGroundColor, GROUND_COLOR);
  gl.uniform1i(U.uShadowMap, SHADOW_UNIT);
  gl.uniform1f(U.uShadowOn, light && light.on ? 1 : 0);
  if (light) {
    gl.uniformMatrix4fv(U.uLightVP, false, light.viewProj);
    gl.uniform2f(U.uShadowTexel, 1 / light.size, light.texelWorld);
  }
}

// createShadows makes the depth map. size is its side in texels and radius
// how far from the centre it reaches, in world units.
export function createShadows(gl, { size = 2048, radius = 2400 } = {}) {
  const depth = gl.createTexture();
  gl.bindTexture(gl.TEXTURE_2D, depth);
  gl.texStorage2D(gl.TEXTURE_2D, 1, gl.DEPTH_COMPONENT24, size, size);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_COMPARE_MODE, gl.COMPARE_REF_TO_TEXTURE);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_COMPARE_FUNC, gl.LEQUAL);

  // A program that samples the shadow map cannot draw into it, even with
  // the sampling switched off: WebGL refuses the feedback loop. During the
  // depth pass the unit holds this one-texel stand-in instead.
  const dummy = gl.createTexture();
  gl.bindTexture(gl.TEXTURE_2D, dummy);
  gl.texStorage2D(gl.TEXTURE_2D, 1, gl.DEPTH_COMPONENT24, 1, 1);
  gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_COMPARE_MODE, gl.COMPARE_REF_TO_TEXTURE);
  gl.bindTexture(gl.TEXTURE_2D, null);

  const fbo = gl.createFramebuffer();
  gl.bindFramebuffer(gl.FRAMEBUFFER, fbo);
  gl.framebufferTexture2D(gl.FRAMEBUFFER, gl.DEPTH_ATTACHMENT, gl.TEXTURE_2D, depth, 0);
  gl.drawBuffers([gl.NONE]);
  gl.readBuffer(gl.NONE);
  const ok = gl.checkFramebufferStatus(gl.FRAMEBUFFER) === gl.FRAMEBUFFER_COMPLETE;
  gl.bindFramebuffer(gl.FRAMEBUFFER, null);

  // The light looks along -SUN_DIR from far up its beam. Its view is fixed;
  // only the window onto it moves, in whole texels, so a shadow's edge does
  // not crawl as the player walks.
  const far = 30000;
  const view = lookAtM([SUN_DIR[0] * far, SUN_DIR[1] * far, SUN_DIR[2] * far], [0, 0, 0],
    Math.abs(SUN_DIR[1]) > 0.99 ? [0, 0, 1] : [0, 1, 0]);
  const texelWorld = (2 * radius) / size;
  const light = { viewProj: new Float32Array(16), size, texelWorld, on: ok, radius };

  return {
    ok,
    light,
    // begin aims the map at a point and binds it for drawing depth. The
    // caller then draws casters with light.viewProj as their camera.
    begin(center) {
      const cx = view[0] * center[0] + view[4] * center[1] + view[8] * center[2] + view[12];
      const cy = view[1] * center[0] + view[5] * center[1] + view[9] * center[2] + view[13];
      const cz = view[2] * center[0] + view[6] * center[1] + view[10] * center[2] + view[14];
      const sx = Math.round(cx / texelWorld) * texelWorld;
      const sy = Math.round(cy / texelWorld) * texelWorld;
      // Depth reaches well past the square either way, for tall casters --
      // a mountain behind the player -- and deep receivers.
      const proj = orthoM(sx - radius, sx + radius, sy - radius, sy + radius, -cz - 20000, -cz + 20000);
      mul4(proj, view, light.viewProj);
      gl.bindFramebuffer(gl.FRAMEBUFFER, fbo);
      gl.viewport(0, 0, size, size);
      gl.clear(gl.DEPTH_BUFFER_BIT);
      gl.activeTexture(gl.TEXTURE0 + SHADOW_UNIT);
      gl.bindTexture(gl.TEXTURE_2D, dummy);
      gl.activeTexture(gl.TEXTURE0);
      // Slope-scaled offset keeps lit surfaces off their own depth.
      gl.enable(gl.POLYGON_OFFSET_FILL);
      gl.polygonOffset(2, 4);
      return light;
    },
    // end unbinds the map and leaves it on its unit for the main pass.
    end() {
      gl.disable(gl.POLYGON_OFFSET_FILL);
      gl.bindFramebuffer(gl.FRAMEBUFFER, null);
      gl.activeTexture(gl.TEXTURE0 + SHADOW_UNIT);
      gl.bindTexture(gl.TEXTURE_2D, depth);
      gl.activeTexture(gl.TEXTURE0);
    },
  };
}

function normalize3(v) {
  const l = Math.hypot(v[0], v[1], v[2]);
  return [v[0] / l, v[1] / l, v[2] / l];
}

// Column-major 4x4 helpers, the same convention as the viewer's.
function lookAtM(eye, target, up) {
  const z = normalize3([eye[0] - target[0], eye[1] - target[1], eye[2] - target[2]]);
  const x = normalize3([up[1] * z[2] - up[2] * z[1], up[2] * z[0] - up[0] * z[2], up[0] * z[1] - up[1] * z[0]]);
  const y = [z[1] * x[2] - z[2] * x[1], z[2] * x[0] - z[0] * x[2], z[0] * x[1] - z[1] * x[0]];
  return new Float32Array([
    x[0], y[0], z[0], 0,
    x[1], y[1], z[1], 0,
    x[2], y[2], z[2], 0,
    -(x[0] * eye[0] + x[1] * eye[1] + x[2] * eye[2]),
    -(y[0] * eye[0] + y[1] * eye[1] + y[2] * eye[2]),
    -(z[0] * eye[0] + z[1] * eye[1] + z[2] * eye[2]),
    1,
  ]);
}

function orthoM(l, r, b, t, n, f) {
  return new Float32Array([
    2 / (r - l), 0, 0, 0,
    0, 2 / (t - b), 0, 0,
    0, 0, -2 / (f - n), 0,
    -(r + l) / (r - l), -(t + b) / (t - b), -(f + n) / (f - n), 1,
  ]);
}

function mul4(a, b, out) {
  for (let c = 0; c < 4; c++) {
    for (let r = 0; r < 4; r++) {
      let s = 0;
      for (let k = 0; k < 4; k++) s += a[k * 4 + r] * b[c * 4 + k];
      out[c * 4 + r] = s;
    }
  }
  return out;
}

// For the headless test: the light matrix for a centre, without a GL context.
export function lightMatrix(center, { size = 2048, radius = 2400 } = {}) {
  const far = 30000;
  const view = lookAtM([SUN_DIR[0] * far, SUN_DIR[1] * far, SUN_DIR[2] * far], [0, 0, 0], [0, 1, 0]);
  const texelWorld = (2 * radius) / size;
  const cx = view[0] * center[0] + view[4] * center[1] + view[8] * center[2] + view[12];
  const cy = view[1] * center[0] + view[5] * center[1] + view[9] * center[2] + view[13];
  const cz = view[2] * center[0] + view[6] * center[1] + view[10] * center[2] + view[14];
  const sx = Math.round(cx / texelWorld) * texelWorld;
  const sy = Math.round(cy / texelWorld) * texelWorld;
  return mul4(orthoM(sx - radius, sx + radius, sy - radius, sy + radius, -cz - 20000, -cz + 20000), view, new Float32Array(16));
}
