// Shared scaffolding for the headless tests: serve web/ to the modules'
// relative fetches, and a WebGL2 context that accepts everything and draws
// nothing, so the real loading and posing code runs with no browser.

import { readFile } from 'node:fs/promises';
import { fileURLToPath } from 'node:url';
import { dirname, join } from 'node:path';

export const here = dirname(fileURLToPath(import.meta.url));

globalThis.fetch = async (url) => {
  try {
    const buf = await readFile(join(here, url));
    return {
      ok: true, status: 200,
      json: async () => JSON.parse(buf.toString('utf8')),
      arrayBuffer: async () => buf.buffer.slice(buf.byteOffset, buf.byteOffset + buf.byteLength),
    };
  } catch {
    return { ok: false, status: 404 };
  }
};

export const gl = new Proxy({}, {
  get: (_, k) => (typeof k === 'string' && /^[A-Z_0-9]+$/.test(k) ? 0 : () => ({})),
});

export const helpers = {
  program: () => ({}),
  uniforms: (_gl, _p, names) => Object.fromEntries(names.map((n) => [n, {}])),
  loadImage: async () => ({}),
};

let failures = 0;
export function check(name, ok, detail) {
  console.log(`  ${ok ? 'ok  ' : 'FAIL'}  ${name}${detail ? '   ' + detail : ''}`);
  if (!ok) failures++;
}

export function finish() {
  console.log(failures ? `\n${failures} check(s) failed` : '\nall checks passed');
  process.exitCode = failures ? 1 : 0;
}
