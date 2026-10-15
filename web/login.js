// Log in, pick or make a character, and go play.

// The game's own typeface, when the interface has been converted.
fetch('data/ui/atlantis/fonts/myriadpro-regular.ttf', { method: 'HEAD' })
  .then((r) => { if (r.ok) document.body.classList.add('stock'); })
  .catch(() => {});

const $ = (id) => document.getElementById(id);
const show = (id, on) => $(id).classList.toggle('hidden', !on);
const msg = (t) => { $('msg').textContent = t || ''; };

async function api(path, opts) {
  const r = await fetch(path, { credentials: 'same-origin', ...opts });
  if (!r.ok) throw Object.assign(new Error((await r.text()).trim() || r.statusText), { status: r.status });
  return r.status === 204 ? null : r.json();
}

const post = (path, body) => api(path, { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(body) });

async function render() {
  const cfg = await api('/api/config');
  let me = null;
  try { me = await api('/api/me'); } catch (e) { if (e.status !== 401) throw e; }
  show('login', !me);
  show('chars', !!me);
  if (!me) {
    show('discord-box', cfg.discord);
    show('admin-box', cfg.admin);
    return;
  }
  $('who').textContent = me.name + (me.admin ? ' (admin)' : '');
  const list = $('list');
  list.replaceChildren();
  for (const c of me.characters) {
    const row = document.createElement('div');
    row.className = 'char';
    const info = document.createElement('div');
    info.innerHTML = '<b></b><small></small>';
    info.querySelector('b').textContent = c.name;
    info.querySelector('small').textContent = `${c.label}, level ${c.level} · ${c.zoneName || 'zone ' + c.zone}`;
    const play = document.createElement('a');
    play.className = 'btn primary'; play.textContent = 'Play';
    play.href = `/?play=${c.id}&zone=${c.zone}`;
    const del = document.createElement('button');
    del.type = 'button'; del.textContent = 'Delete';
    del.addEventListener('click', async () => {
      if (!confirm(`Delete ${c.name}? This cannot be undone.`)) return;
      try { await api('/api/characters/' + c.id, { method: 'DELETE' }); render(); } catch (e) { msg(e.message); }
    });
    row.append(info, play, del);
    list.appendChild(row);
  }
  if (!me.characters.length) list.textContent = 'None yet. Make one below.';
  show('create', me.characters.length < me.max);
  const sel = $('arch');
  if (!sel.options.length) for (const a of me.archetypes) sel.add(new Option(a.label, a.key));
}

$('admin-box').addEventListener('submit', async (e) => {
  e.preventDefault();
  msg('');
  try { await post('/auth/admin', { password: $('pw').value }); $('pw').value = ''; render(); } catch (err) { msg(err.message); }
});
$('new').addEventListener('submit', async (e) => {
  e.preventDefault();
  msg('');
  try { await post('/api/characters', { name: $('name').value, arch: $('arch').value }); $('name').value = ''; render(); } catch (err) { msg(err.message); }
});
$('logout').addEventListener('click', async () => { await post('/auth/logout', {}); render(); });

const m = new URLSearchParams(location.search).get('msg');
if (m) msg(m);
render().catch((e) => msg(e.message));
