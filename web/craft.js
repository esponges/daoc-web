// The crafting window: your tradeskills, what you can make from them, and
// the materials each recipe needs against what you carry. A plain panel drawn
// from the server's `craft` message (docs/protocol.md); the pure parts are
// here for node to test.

// craftView is the window's content from a `craft` message. skill picks the
// recipes of one skill (0 for all); a recipe can be made when every material
// is held and the server gives it a chance above 0.
export function craftView(m, { skill = 0 } = {}) {
  if (!m) return null;
  const rows = (m.recipes || [])
    .filter((r) => !skill || r.sk === skill)
    .map((r) => {
      const mats = (r.m || []).map((x) => ({ id: x.id, name: x.nm, need: x.n, have: x.have, ok: x.have >= x.n }));
      return {
        id: r.id, name: r.nm, level: r.lv, skill: r.sk, chance: r.ok, secs: r.s, mats,
        can: r.ok > 0 && mats.every((x) => x.ok),
        text: r.nm + '  (' + r.lv + ', ' + r.ok + '%)',
      };
    });
  return {
    prof: m.pn || '', offer: m.on || '', offerSkill: m.offer || 0,
    skills: (m.skills || []).map((s) => ({ id: s.id, name: s.nm, value: s.v, cap: s.cap, text: s.nm + ' ' + s.v + ' / ' + s.cap })),
    rows,
  };
}

// matLine is a recipe's materials as one line, each with what is held.
export function matLine(row) {
  return row.mats.map((x) => x.name + ' ' + x.have + '/' + x.need).join(', ');
}

const el = (doc, tag, cls, text) => {
  const e = doc.createElement(tag);
  if (cls) e.className = cls;
  if (text !== undefined) e.textContent = text;
  return e;
};

// createCraft builds the window, hidden.
//   onOp(op, id)   a crafting request: "open", "join", "make" or "cancel"
export function createCraft(doc, { onOp, mover } = {}) {
  const root = el(doc, 'div', 'panel craft');
  Object.assign(root.style, { position: 'fixed', left: '50%', top: '70px', marginLeft: '-230px', width: '460px', zIndex: 15, display: 'none' });
  if (mover && mover.makeMovable) mover.makeMovable(root, 'craft');
  const title = el(doc, 'h2', '', 'Crafting');
  const skills = el(doc, 'div', 'dim');
  const bar = el(doc, 'div');
  const filter = doc.createElement('select');
  bar.append(filter);
  const list = el(doc, 'div', 'items');
  Object.assign(list.style, { maxHeight: '190px', overflowY: 'auto' });
  const detail = el(doc, 'div');
  const status = el(doc, 'div', 'dim');
  const progress = doc.createElement('progress');
  progress.max = 1; progress.value = 0; progress.style.width = '100%'; progress.style.display = 'none';
  const btn = (label, fn) => { const b = el(doc, 'button', '', label); b.addEventListener('click', fn); return b; };
  let selected = '', state = null, open = false, timer = 0;
  const join = btn('Join', () => onOp && onOp('join'));
  const make = btn('Make', () => selected && onOp && onOp('make', selected));
  const cancel = btn('Cancel', () => onOp && onOp('cancel'));
  const actions = el(doc, 'div');
  actions.append(join, make, cancel);
  root.append(title, skills, bar, list, detail, progress, status, actions);

  const draw = () => {
    const v = craftView(state, { skill: Number(filter.value) || 0 });
    if (!v) return;
    title.textContent = v.prof ? 'Crafting: ' + v.prof : 'Crafting';
    skills.textContent = v.prof ? v.skills.map((s) => s.text).join('  -  ') : (v.offer ? 'A master here teaches ' + v.offer + '.' : 'Find a craft master to learn a craft.');
    join.style.display = !v.prof && v.offer ? '' : 'none';
    make.style.display = v.prof ? '' : 'none';
    const keep = filter.value;
    filter.replaceChildren(new Option('All skills', '0'), ...v.skills.map((s) => new Option(s.name, String(s.id))));
    filter.value = keep || '0';
    list.replaceChildren(...v.rows.map((r) => {
      const d = el(doc, 'div', 'row', r.text);
      d.style.cursor = 'pointer';
      d.style.opacity = r.can ? '1' : '.55';
      if (r.id === selected) d.style.background = 'rgba(255,255,255,.12)';
      d.addEventListener('click', () => { selected = r.id; draw(); });
      return d;
    }));
    const row = v.rows.find((r) => r.id === selected);
    detail.textContent = row ? matLine(row) + ' - ' + row.secs + 's' : '';
    make.disabled = !row || !row.can;
  };
  filter.addEventListener('change', draw);
  const stop = () => { clearInterval(timer); progress.style.display = 'none'; };
  return {
    el: root,
    get isOpen() { return open; },
    set(m) { state = m; draw(); },
    // started shows the progress bar for a craft of secs seconds.
    started(secs) {
      stop();
      const t0 = Date.now();
      progress.style.display = '';
      progress.value = 0;
      timer = setInterval(() => { progress.value = Math.min(1, (Date.now() - t0) / (secs * 1000)); }, 100);
    },
    // done ends the progress bar and shows the result line.
    done(m) { stop(); status.textContent = m.text || ''; },
    toggle() { open = !open; root.style.display = open ? 'block' : 'none'; if (open) { onOp && onOp('open'); draw(); } },
    show() { open = true; root.style.display = 'block'; draw(); },
  };
}
