// The selection ring drawn on the ground under the current target.
//
// A band of quads around the target, each vertex set on the terrain
// beneath it, so the ring lies over slopes rather than cutting into them. It
// is rebuilt every frame -- 48 segments is a few hundred floats -- and drawn
// translucent, after the opaque scene and without writing depth.

const SEGMENTS = 48;

const RING_VS = `#version 300 es
precision highp float;
layout(location=0) in vec3 aPos;
layout(location=1) in float aEdge;
uniform mat4 uViewProj;
out float vEdge;
void main() {
  vEdge = aEdge;
  gl_Position = uViewProj * vec4(aPos, 1.0);
}
`;

const RING_FS = `#version 300 es
precision highp float;
in float vEdge;
uniform vec3 uColor;
uniform float uAlpha;
out vec4 outColor;
void main() {
  // Brightest along the middle of the band, fading to both edges.
  float a = uAlpha * (1.0 - abs(vEdge * 2.0 - 1.0));
  outColor = vec4(uColor, a);
}
`;

export function createRing(gl, { program, uniforms }) {
  const prog = program(gl, RING_VS, RING_FS, 'ring');
  const U = uniforms(gl, prog, ['uViewProj', 'uColor', 'uAlpha']);
  const n = (SEGMENTS + 1) * 2;
  const pos = new Float32Array(n * 3);
  const edge = new Float32Array(n);
  for (let i = 0; i < n; i++) edge[i] = i % 2;

  const vao = gl.createVertexArray();
  gl.bindVertexArray(vao);
  const pbo = gl.createBuffer();
  gl.bindBuffer(gl.ARRAY_BUFFER, pbo);
  gl.bufferData(gl.ARRAY_BUFFER, pos.byteLength, gl.DYNAMIC_DRAW);
  gl.enableVertexAttribArray(0);
  gl.vertexAttribPointer(0, 3, gl.FLOAT, false, 0, 0);
  const ebo = gl.createBuffer();
  gl.bindBuffer(gl.ARRAY_BUFFER, ebo);
  gl.bufferData(gl.ARRAY_BUFFER, edge, gl.STATIC_DRAW);
  gl.enableVertexAttribArray(1);
  gl.vertexAttribPointer(1, 1, gl.FLOAT, false, 0, 0);
  gl.bindVertexArray(null);

  // draw puts a ring of the given radius around world (x, y). color is RGB
  // 0..1; time drives a slow pulse.
  function draw(viewProj, x, y, radius, groundAt, color, time) {
    // Wide enough to read from a low camera, where it foreshortens.
    const width = Math.max(8, radius * 0.3);
    const r0 = radius - width / 2, r1 = radius + width / 2;
    for (let i = 0; i <= SEGMENTS; i++) {
      const a = (i / SEGMENTS) * Math.PI * 2, c = Math.cos(a), s = Math.sin(a);
      for (let k = 0; k < 2; k++) {
        const r = k ? r1 : r0, wx = x + c * r, wy = y + s * r;
        const o = (i * 2 + k) * 3;
        // A little above the ground, so it does not flicker against it.
        pos[o] = wx; pos[o + 1] = groundAt(wx, wy) + 2; pos[o + 2] = wy;
      }
    }
    gl.useProgram(prog);
    gl.uniformMatrix4fv(U.uViewProj, false, viewProj);
    gl.uniform3fv(U.uColor, color);
    gl.uniform1f(U.uAlpha, 0.75 + 0.2 * Math.sin(time * 4));
    gl.bindVertexArray(vao);
    gl.bindBuffer(gl.ARRAY_BUFFER, pbo);
    gl.bufferSubData(gl.ARRAY_BUFFER, 0, pos);
    gl.enable(gl.BLEND);
    gl.blendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA);
    gl.depthMask(false);
    gl.disable(gl.CULL_FACE);
    gl.drawArrays(gl.TRIANGLE_STRIP, 0, n);
    gl.enable(gl.CULL_FACE);
    gl.depthMask(true);
    gl.disable(gl.BLEND);
    gl.bindVertexArray(null);
  }

  return { draw };
}
