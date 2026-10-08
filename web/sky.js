// The sky: the region's skydome, as zoneconv read it from the game's sky
// file (sky_midgard.dat for the Vale of Mularn).
//
// One full-screen pass, drawn before anything else and behind everything.
// Each pixel's view ray picks its colour:
//
// - The canopy. The file gives the zenith and four rings down to the
//   horizon, on the east side and on the west; this spaces the rings evenly
//   in elevation and blends east to west by how far the ray points east.
// - Two cloud layers, projected onto a flat ceiling so they bunch towards the
//   horizon as real cloud decks do, each tiled and scrolled as the file says.
//   The file also gives their tint and opacity at each ring, down to none at
//   the horizon, which is what keeps the horizon clear.
// - The sun, at the direction the scene is lit from, its disc and glow drawn
//   over the clouds and added to the sky, in the file's colours for the sun
//   at its height.

const SKY_VS = `#version 300 es
precision highp float;
out vec2 vNdc;
void main() {
  // One triangle that covers the screen.
  vec2 p = vec2(float((gl_VertexID << 1) & 2), float(gl_VertexID & 2)) * 2.0 - 1.0;
  vNdc = p;
  gl_Position = vec4(p, 0.0, 1.0);
}
`;

const SKY_FS = `#version 300 es
precision highp float;
in vec2 vNdc;
uniform vec3 uF, uS, uU;   // camera forward, right and up
uniform float uK, uAspect; // tan(fov/2), width/height
uniform vec3 uEast[5];     // zenith, then four rings to the horizon
uniform vec3 uWest[5];
uniform vec4 uCloudEast[5];
uniform vec4 uCloudWest[5];
uniform sampler2D uCloud0, uCloud1;
uniform vec2 uCloudTile;   // repeats across the sky, per layer
uniform vec2 uCloudShift0, uCloudShift1;
uniform float uClouds;     // how many layers there are
uniform vec3 uSunDir;
uniform sampler2D uSunTex;
uniform vec4 uSunShape, uSunGlow; // RGBA 0..1
uniform vec2 uSunSize;     // half-size of disc and glow, as tangents
uniform float uSun;
out vec4 outColor;

const float HALF_PI = 1.5707963;

// ring picks a colour by elevation from a zenith-to-horizon list.
vec3 ring3(vec3 c[5], float s) {
  int i = int(floor(s));
  if (i >= 4) return c[4];
  return mix(c[i], c[i + 1], fract(s));
}
vec4 ring4(vec4 c[5], float s) {
  int i = int(floor(s));
  if (i >= 4) return c[4];
  return mix(c[i], c[i + 1], fract(s));
}

void main() {
  vec3 d = normalize(uF + uS * vNdc.x * uK * uAspect + uU * vNdc.y * uK);
  // 0 at the zenith, 4 at the horizon; below it, the horizon's colour.
  float el = asin(clamp(d.y, 0.0, 1.0));
  float s = (1.0 - el / HALF_PI) * 4.0;
  // East is +x in the zone.
  float east = 0.5 + 0.5 * d.x;
  vec3 col = mix(ring3(uWest, s), ring3(uEast, s), east);

  if (d.y > 0.0) {
    // A flat ceiling: texture coordinates spread out towards the horizon.
    vec2 p = d.xz / (d.y + 0.08);
    vec4 tint = mix(ring4(uCloudWest, s), ring4(uCloudEast, s), east);
    if (uClouds > 0.5) {
      vec4 c = texture(uCloud0, p * uCloudTile.x * 0.12 + uCloudShift0);
      col = mix(col, c.rgb * tint.rgb, c.a * tint.a);
    }
    if (uClouds > 1.5) {
      vec4 c = texture(uCloud1, p * uCloudTile.y * 0.12 + uCloudShift1);
      col = mix(col, c.rgb * tint.rgb, c.a * tint.a);
    }
  }

  if (uSun > 0.5) {
    // The disc lies across the sun's direction, in its own small plane.
    float c = dot(d, uSunDir);
    if (c > 0.0) {
      vec3 sx = normalize(cross(uSunDir, vec3(0.0, 1.0, 0.0)));
      vec3 sy = cross(sx, uSunDir);
      vec2 q = vec2(dot(d, sx), dot(d, sy)) / c;
      vec2 g = q / uSunSize.y * 0.5 + 0.5;
      vec2 h = q / uSunSize.x * 0.5 + 0.5;
      if (all(greaterThan(g, vec2(0.0))) && all(lessThan(g, vec2(1.0))))
        col += texture(uSunTex, g).rgb * uSunGlow.rgb * uSunGlow.a;
      if (all(greaterThan(h, vec2(0.0))) && all(lessThan(h, vec2(1.0))))
        col += texture(uSunTex, h).rgb * uSunShape.rgb * uSunShape.a;
    }
  }
  outColor = vec4(col, 1.0);
}
`;

export async function createSky(gl, base, spec, { program, uniforms, loadImage }) {
  const prog = program(gl, SKY_VS, SKY_FS, 'sky');
  const U = uniforms(gl, prog, ['uF', 'uS', 'uU', 'uK', 'uAspect', 'uEast', 'uWest',
    'uCloudEast', 'uCloudWest', 'uCloud0', 'uCloud1', 'uCloudTile', 'uCloudShift0', 'uCloudShift1',
    'uClouds', 'uSunDir', 'uSunTex', 'uSunShape', 'uSunGlow', 'uSunSize', 'uSun']);
  const vao = gl.createVertexArray();

  const texture = async (path) => {
    const img = await loadImage(base + '/' + path);
    const t = gl.createTexture();
    gl.bindTexture(gl.TEXTURE_2D, t);
    gl.texImage2D(gl.TEXTURE_2D, 0, gl.RGBA, gl.RGBA, gl.UNSIGNED_BYTE, img);
    gl.generateMipmap(gl.TEXTURE_2D);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR_MIPMAP_LINEAR);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT);
    return t;
  };
  const clouds = await Promise.all(spec.clouds.slice(0, 2).map((c) => texture(c.texture)));
  let sunTex = null;
  if (spec.sunDisc) {
    sunTex = await texture(spec.sunDisc.texture);
    gl.bindTexture(gl.TEXTURE_2D, sunTex);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE);
    gl.texParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE);
  }
  gl.bindTexture(gl.TEXTURE_2D, null);

  // Colour lists, zenith first, as flat arrays for the uniform calls.
  const c3 = (v) => v.map((x) => x / 255);
  const list3 = (rings) => new Float32Array([...c3(spec.zenith), ...rings.flatMap(c3)]);
  const list4 = (rings) => new Float32Array([...c3(spec.cloudZenith), ...rings.flatMap(c3)]);
  const east = list3(spec.east), west = list3(spec.west);
  const cEast = list4(spec.cloudEast), cWest = list4(spec.cloudWest);
  // The file's scroll speeds "are multiplied by 0.001 in the code": texture
  // widths a second. The two decks drift the same way at their own speeds.
  const wind = [0.8, 0.6];

  // draw paints the sky. cam is { f, s, u } (forward, right, up), fov the
  // vertical field of view, sunDir the direction toward the sun.
  function draw(cam, fov, aspect, sunDir, time) {
    gl.useProgram(prog);
    gl.disable(gl.DEPTH_TEST);
    gl.depthMask(false);
    gl.uniform3fv(U.uF, cam.f);
    gl.uniform3fv(U.uS, cam.s);
    gl.uniform3fv(U.uU, cam.u);
    gl.uniform1f(U.uK, Math.tan(fov / 2));
    gl.uniform1f(U.uAspect, aspect);
    gl.uniform3fv(U.uEast, east);
    gl.uniform3fv(U.uWest, west);
    gl.uniform4fv(U.uCloudEast, cEast);
    gl.uniform4fv(U.uCloudWest, cWest);
    gl.uniform1f(U.uClouds, clouds.length);
    const tile = spec.clouds.map((c) => c.tile);
    gl.uniform2f(U.uCloudTile, tile[0] || 1, tile[1] || 1);
    const shift = (i) => {
      const v = (spec.clouds[i]?.speed || 0) * 0.001 * time;
      return [(wind[0] * v) % 1, (wind[1] * v) % 1];
    };
    gl.uniform2fv(U.uCloudShift0, shift(0));
    gl.uniform2fv(U.uCloudShift1, shift(1));
    clouds.forEach((t, i) => {
      gl.activeTexture(gl.TEXTURE0 + i);
      gl.bindTexture(gl.TEXTURE_2D, t);
    });
    gl.uniform1i(U.uCloud0, 0);
    gl.uniform1i(U.uCloud1, 1);
    gl.uniform1f(U.uSun, sunTex ? 1 : 0);
    if (sunTex) {
      const s = spec.sunDisc;
      gl.activeTexture(gl.TEXTURE2);
      gl.bindTexture(gl.TEXTURE_2D, sunTex);
      gl.uniform1i(U.uSunTex, 2);
      gl.uniform3fv(U.uSunDir, sunDir);
      gl.uniform4fv(U.uSunShape, c3(s.shape));
      gl.uniform4fv(U.uSunGlow, c3(s.glow));
      gl.uniform2f(U.uSunSize, s.scale / s.distance, s.glowScale / s.distance);
    }
    gl.activeTexture(gl.TEXTURE0);
    gl.bindVertexArray(vao);
    gl.drawArrays(gl.TRIANGLES, 0, 3);
    gl.bindVertexArray(null);
    gl.depthMask(true);
    gl.enable(gl.DEPTH_TEST);
  }

  return { draw, fog: c3(spec.fog), ambient: spec.ambient, sun: spec.sun };
}
