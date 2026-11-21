// The character windows the menu bar names: attributes, skills and styles,
// built from the game's own layouts (StatsAttributes and Stats_Spec_Abil) and
// filled from what the server has told the character about themselves. A field
// the server has not sent stays empty: nothing here is invented.

import { buildWindow } from './uiwin.js';
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

// --- rows for the stock controls (pure) ---

// SPELL_ICON are the pictures of the spell sheet used for a spell by what it
// does: the imported icon numbers do not match this install's sheets (see
// CLAUDE.md), so a spell is drawn by its kind, as styles are by weapon line.
export const SPELL_ICON = { damage: 76, heal: 11, buff: 20, debuff: 70, mez: 5, stun: 6, pet: 1, other: 0 };

// spellKind sorts a spell by the server's type name (and who it aims at) into
// the kinds above.
export function spellKind(tp = '', tg = '') {
  if (/Mesmerize/i.test(tp)) return 'mez';
  if (/Stun/i.test(tp)) return 'stun';
  if (/Summon|Charm|Pet/i.test(tp) || /^Pet$/i.test(tg)) return 'pet';
  if (/Heal|Regen|Cure|Resurrect|Resur|Lifetransfer|SpreadHeal/i.test(tp)) return 'heal';
  if (/Damage|Bolt|Lifedrain|Drain|Taunt|Proc/i.test(tp)) return 'damage';
  if (/Debuff|Slow|Snare|Disease|Nearsight|Amnesia|Confusion|Fatigue|Root/i.test(tp)) return 'debuff';
  if (/Buff|Shield|Ward|Armour|Armor|Haste|Speed|Absorb|Ablative|Bladeturn|Resist|Regeneration|Strength|Constitution|Dexterity|Acuity|Wraith|Siphon/i.test(tp)) return 'buff';
  return 'other';
}

// spellGroups is the spells tree's content: a group for each spec line with its
// spells, each a row {id, text, icon, on, drag, title}. open says which groups
// are unfolded, by line name; a line not listed is open. A spell the server
// cannot cast yet is greyed.
export function spellGroups(book, song = 0, open = {}) {
  return ((book && book.lines) || []).map((line) => ({
    title: line.nm,
    open: open[line.nm] !== false,
    rows: line.sp.map((s) => {
      const cast = s.ct ? s.ct + 's' : 'instant';
      return {
        id: s.id,
        text: s.nm + '  ' + s.lv + ', ' + s.pw + ' pw, ' + cast + (s.ok ? '' : ' (not yet)') + (song && s.id === song ? '  (playing)' : ''),
        icon: SPELL_ICON[spellKind(s.tp, s.tg)],
        on: !!s.ok,
        drag: s.ok ? 'spell:' + s.id : undefined,
        title: s.ok ? 'Double-click to cast, drag to the spell bar' : 'You cannot cast this yet',
      };
    }),
  }));
}

// styleGroups is the styles tree's content: a group for each weapon line with
// the styles the character can use now (each carries its id to drag to the
// style bar), then one group of those they have learned but cannot use.
export function styleGroups(slots, also, iconOf = () => 0, open = {}) {
  const v = stylesView(slots, also);
  const groups = [];
  for (const r of v.usable) {
    if (r.header) { groups.push({ title: r.header, open: open[r.header] !== false, rows: [] }); continue; }
    groups[groups.length - 1].rows.push({ id: r.id, text: r.text, icon: iconOf(r.spec), drag: 'style:' + r.id, title: 'Drag onto the style bar' });
  }
  if (v.locked.length) {
    const t = 'Not usable now';
    groups.push({ title: t, open: open[t] !== false, rows: v.locked.map((r) => ({ id: r.id, text: r.text, icon: iconOf(r.spec), on: false })) });
  }
  return groups;
}

// --- the windows ---

// createStockCharWindows builds the attributes, skills, styles and spells
// windows from their own stock templates, hidden. update(me, stats, bar, also,
// book, song) refills them; toggle(name) opens or closes one ("attributes",
// "skills", "styles", "spells").
//   onCast(id)          a spell was double-clicked
//   onAssign(id)        a spell was asked for the spell bar
//   onAction(name)      a window button: "concentration", "pet", "quiver",
//                       "pray", "bind"
//   mover               makes the windows draggable and remembers their places
export function createStockCharWindows(doc, data, { onCast, onAssign, onAction, mover, styleIcon = () => 0 } = {}) {
  const closeOf = (name) => () => api.toggle(name, false);
  const attrib = buildWindow(doc, data, 'StatsAttributes', {
    captions: true, labels: true, onClose: closeOf('attributes'),
    onClick: (id, ev, label) => { if (label === 'PRAY') onAction?.('pray'); else if (label === 'BIND') onAction?.('bind'); },
    disabled: (ev, id, label) => label !== 'PRAY' && label !== 'BIND',
    tooltip: () => 'Not built yet.',
  });
  const skills = buildWindow(doc, data, 'Stats_Spec_Abil', { captions: true, onClose: closeOf('skills') });
  const styles = buildWindow(doc, data, 'stats_combat', {
    captions: true, labels: true, onClose: closeOf('styles'),
    onClick: (id, ev) => { if (ev === 'ToggleQuiver') onAction?.('quiver'); },
  });
  const spells = buildWindow(doc, data, 'stats_spells', {
    captions: true, labels: true, onClose: closeOf('spells'),
    onClick: (id, ev) => {
      if (ev === 'ToggleConcentration') onAction?.('concentration');
      else if (ev === 'ToggleMiniPet') onAction?.('pet');
    },
  });
  const wins = { attributes: attrib, skills, styles, spells };
  const root = doc.createElement('div');
  const open = { attributes: false, skills: false, styles: false, spells: false };
  const place3 = { attributes: 0, skills: 200, styles: 400, spells: 600 };
  for (const [name, w] of Object.entries(wins)) {
    Object.assign(w.el.style, { position: 'fixed', left: '12px', top: '60px', marginLeft: place3[name] + 'px', zIndex: 14, display: 'none' });
    root.appendChild(w.el);
    mover?.makeMovable(w.el, name);
  }

  // Which tree headings the player folded, by name, kept across refills.
  const folded = { spells: {}, styles: {} };
  const watch = (win, id, key) => {
    const tree = win.tree(id);
    if (!tree) return;
    tree.onToggle((g, isOpen, title) => { folded[key][title] = isOpen; });
  };
  watch(spells, '1000', 'spells');
  watch(styles, '1000', 'styles');
  const dbl = spells.tree('1000');
  if (dbl) dbl.onActivate((i, row) => { if (row.id && row.on !== false) onCast?.(row.id); });

  let last = {};
  const fill = (win, adapter, value) => win.setText(adapter, value);

  function render() {
    const { me, st, bar, also, book, song } = last;
    const a = attributesView(me, st);
    fill(attrib, 'stats_name', a.name);
    fill(attrib, 'stats_level', a.level);
    fill(attrib, 'stats_weapon_skill', a.weaponSkill);
    fill(attrib, 'stats_hitpoints', a.hitPoints);
    fill(attrib, 'stats_armor_factor', a.armourFactor);
    STAT_ADAPTERS.forEach((ad, i) => fill(attrib, ad, a.stats[i]));
    // The extra-info list: race and class, endurance.
    attrib.list('1100')?.setRows(a.info.map((t) => ({ text: t })));

    const s = skillsView(st);
    fill(skills, 'stats_spec_points', s.points);
    skills.list('1000')?.setRows(s.specs);
    skills.list('1001')?.setRows(s.abilities);
    skills.list('1002')?.setRows([]);

    styles.tree('1000')?.setGroups(styleGroups(bar, also, styleIcon, folded.styles));

    const groups = spellGroups(book, song, folded.spells);
    spells.tree('1000')?.setGroups(groups.length ? groups : [{ title: 'You have no spells.', open: false, rows: [] }]);
  }

  const api = {
    el: root,
    isOpen: (name) => open[name],
    toggle(name, on = !open[name]) {
      open[name] = on;
      wins[name].el.style.display = on ? 'block' : 'none';
      if (on) { mover?.restore(name); render(); }
    },
    update(me, st, bar, also, book, song = 0) {
      last = { me, st, bar, also, book, song };
      if (Object.values(open).some(Boolean)) render();
    },
  };
  return api;
}

