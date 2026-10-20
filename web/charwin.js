// The character windows the menu bar names: attributes, skills and styles,
// built from the game's own layouts (StatsAttributes and Stats_Spec_Abil) and
// filled from what the server has told the character about themselves. A field
// the server has not sent stays empty: nothing here is invented.

import { buildWindow, place, text } from './uiwin.js';
import { STAT_NAMES } from './sheet.js';

// --- what each window shows (pure, so node can test it) ---

// attributesView is the attributes window's content: name, level, weapon skill,
// hit points, armour factor, the eight stats and a few lines of detail. Anything
// the server has not sent is '' (realm points, weapon damage, resists).
export function attributesView(me, st) {
  const has = !!st;
  const level = has ? String(st.lv) + (st.half ? '.5' : '') : '';
  const info = [];
  if (has) {
    info.push([st.race, st.cls].filter(Boolean).join(' '));
    info.push('Endurance ' + st.mend);
    if (st.conlost) info.push('Constitution lost ' + st.conlost);
  }
  return {
    name: me && me.name ? me.name : '',
    level,
    weaponSkill: has ? String(st.ws) : '',
    hitPoints: has ? String(st.mhp) : '',
    armourFactor: has ? String(st.af) : '',
    stats: STAT_NAMES.map((_, i) => (has ? String(st.stats[i]) : '')),
    info: info.filter(Boolean),
  };
}

// The adapter each stat's label is bound to, in STAT_NAMES order.
export const STAT_ADAPTERS = ['stats_strength', 'stats_constitution', 'stats_dexterity', 'stats_quickness',
  'stats_intelligence', 'stats_piety', 'stats_empathy', 'stats_charisma'];

// skillsView is the skills window's content: the points left to spend, the
// specialisations with their levels, and the abilities (those the server does
// not act on are marked).
export function skillsView(st) {
  if (!st) return { points: '', specs: [], abilities: [] };
  return {
    points: 'pts: ' + st.pts,
    specs: (st.specs || []).map((s) => ({ text: s.nm + '  ' + s.lv + (s.bonus ? ' (+' + s.bonus + ')' : '') })),
    abilities: (st.ab || []).map((a) => ({ text: a.nm + (a.lv > 1 ? ' ' + a.lv : '') + (a.on ? '' : ' (not yet)'), on: !!a.on })),
  };
}

// stylesView is the styles window's content: the styles the character can use
// now, by weapon line, each with its opening and endurance cost and its bar
// slot, then the ones they have learned but cannot use with what they hold.
export function stylesView(slots, also) {
  const byLine = new Map();
  (slots || []).forEach((s, i) => {
    const line = s.spec || 'Other';
    if (!byLine.has(line)) byLine.set(line, []);
    byLine.get(line).push({ id: s.id, text: s.nm + '  ' + s.opens + ', ' + s.cost + ' end', slot: i, line, spec: s.spec, draggable: true });
  });
  const usable = [...byLine.keys()].sort().flatMap((line) => [{ header: line }, ...byLine.get(line).sort((a, b) => a.slot - b.slot)]);
  const locked = (also || []).map((s) => ({ id: s.id, text: s.nm + '  (needs ' + (s.spec || 'another weapon') + ')', draggable: false, spec: s.spec }));
  return { usable, locked };
}

// --- the windows ---

const GREY = 'rgb(150,150,150)';
const GOLD = 'rgb(255,215,106)';

// listRows lays rows of text into a list box's rectangle, scrolling when they
// do not fit. A row with draggable carries its style id for the bar.
function listRows(doc, data, win, id, rows) {
  const node = (data.windows.get(win.id).c || []).find((c) => c.t === 'ListBoxDef' && (c.c || []).some((x) => x.t === 'ControlId' && x.x === id));
  if (!node) return;
  const r = place(node, win.width, win.height, win.width, win.height);
  const box = doc.createElement('div');
  Object.assign(box.style, { position: 'absolute', left: r.x + 'px', top: r.y + 'px', width: r.w + 'px', height: r.h + 'px', overflowY: 'auto', overflowX: 'hidden' });
  box.className = 'listrows';
  for (const row of rows) {
    const d = doc.createElement('div');
    Object.assign(d.style, { position: 'relative', height: '13px', margin: '0 2px' });
    const color = row.header ? GOLD : row.draggable === false || row.on === false ? GREY : 'rgb(255,255,255)';
    const t = text(doc, data, 'arial9', row.header || row.text, color);
    t.style.top = '1px';
    t.style.left = row.header ? '0px' : '4px';
    t.style.pointerEvents = 'none';
    d.appendChild(t);
    if (row.draggable) {
      d.draggable = true;
      d.style.cursor = 'grab';
      d.title = 'Drag onto the style bar';
      d.addEventListener('dragstart', (e) => e.dataTransfer?.setData('text/plain', 'style:' + row.id));
    }
    box.appendChild(d);
  }
  win.el.appendChild(box);
}

// createStockCharWindows builds the attributes, skills and styles windows,
// hidden. update(me, stats, bar, also) refills them; toggle(name) opens or
// closes one ("attributes", "skills", "styles").
export function createStockCharWindows(doc, data) {
  const attrib = buildWindow(doc, data, 'StatsAttributes', { captions: true });
  const skills = buildWindow(doc, data, 'Stats_Spec_Abil', { captions: true });
  const styles = buildWindow(doc, data, 'Stats_Spec_Abil', { captions: true });
  skills.id = styles.id = 'Stats_Spec_Abil';
  // The styles window reuses the layout: its first list holds the styles that
  // can be used now, the second those learned but not usable.
  const wins = { attributes: attrib, skills, styles };
  const root = doc.createElement('div');
  const open = { attributes: false, skills: false, styles: false };
  const place3 = { attributes: 0, skills: 200, styles: 400 };
  for (const [name, w] of Object.entries(wins)) {
    Object.assign(w.el.style, { position: 'fixed', left: '12px', top: '60px', marginLeft: place3[name] + 'px', zIndex: 14, display: 'none' });
    root.appendChild(w.el);
  }
  const cap = (w, adapter, value) => w.setText(adapter, value);
  let last = {};

  function render() {
    const { me, st, bar, also } = last;
    const a = attributesView(me, st);
    cap(attrib, 'stats_name', a.name);
    cap(attrib, 'stats_level', a.level);
    cap(attrib, 'stats_weapon_skill', a.weaponSkill);
    cap(attrib, 'stats_hitpoints', a.hitPoints);
    cap(attrib, 'stats_armor_factor', a.armourFactor);
    STAT_ADAPTERS.forEach((ad, i) => cap(attrib, ad, a.stats[i]));
    // The extra-info list: race and class, endurance.
    for (const old of attrib.el.querySelectorAll('.listrows')) old.remove();
    attrib.id = 'StatsAttributes';
    infoRows(attrib, a.info);

    const s = skillsView(st);
    cap(skills, 'stats_spec_points', s.points);
    for (const old of skills.el.querySelectorAll('.listrows')) old.remove();
    listRows(doc, data, skills, '1000', s.specs);
    listRows(doc, data, skills, '1001', s.abilities);

    const v = stylesView(bar, also);
    cap(styles, 'stats_spec_points', '');
    for (const old of styles.el.querySelectorAll('.listrows')) old.remove();
    listRows(doc, data, styles, '1000', v.usable);
    listRows(doc, data, styles, '1001', v.locked);
  }

  // The attributes window's additional-info list box.
  function infoRows(win, lines) {
    const node = (data.windows.get('StatsAttributes').c || []).find((c) => c.t === 'ListBoxDef');
    if (!node) return;
    const r = place(node, win.width, win.height, win.width, win.height);
    const box = doc.createElement('div');
    box.className = 'listrows';
    Object.assign(box.style, { position: 'absolute', left: r.x + 'px', top: r.y + 'px', width: r.w + 'px', height: r.h + 'px', overflowY: 'auto' });
    lines.forEach((l, i) => {
      const t = text(doc, data, 'arial9', l, 'rgb(255,255,255)');
      t.style.left = '4px';
      t.style.top = 2 + i * 13 + 'px';
      box.appendChild(t);
    });
    win.el.appendChild(box);
  }

  return {
    el: root,
    isOpen: (name) => open[name],
    toggle(name, on = !open[name]) {
      open[name] = on;
      wins[name].el.style.display = on ? 'block' : 'none';
      if (on) render();
    },
    update(me, st, bar, also) {
      last = { me, st, bar, also };
      if (Object.values(open).some(Boolean)) render();
    },
  };
}
