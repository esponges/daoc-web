// The in-game windows, built from the game's own interface data (uiwin.js).
// Each takes the converted UIData and returns plain controls: what to show
// goes in, what the player did comes out through callbacks. The server
// decides everything; these only display and ask.

import { buildWindow, frame, fontFamily, child, num, str, pt, text, icon, styleIconNumber, hrButton } from './uiwin.js';

// Colours of chat lines by kind.
const LINE_COLOR = {
  '': '#ffffff', sys: '#9fd0ff', all: '#ffd76a', err: '#ff9a8a',
  // Game messages, as the combat log used to colour them.
  info: '#cfd8e3', hit: '#ffd76a', 'hit-in': '#ff8a7a', miss: '#9aa8b6', 'miss-in': '#9aa8b6',
  kill: '#8fe39a', death: '#ff5a5a', block: '#8fc7ff', xp: '#c9a6ff', level: '#ffe066',
  spell: '#c9a8ff', 'spell-in': '#ff9ad0', heal: '#8fe39a',
  // Tells and the group channel.
  tell: '#ff8cff', group: '#7fe0e0',
};
// Lines people write; everything else is a game message. The Chat tab shows
// only the first kind, Main shows both, and Group only the group channel.
const SPEECH_KINDS = new Set(['', 'all', 'tell', 'group']);
const TAB_KINDS = { main: null, chat: SPEECH_KINDS, group: new Set(['group']) };

// createStockChat builds the stock chat window: lines in its text area and an
// entry bar that shows while typing. size is the window's width and height.
//   onSend(text)   the player pressed Enter on a non-empty line
//   onToggle(open) typing started or stopped
export function createStockChat(doc, data, { onSend, onToggle, width = 400, height = 150, maxLines = 60 } = {}) {
  const win = buildWindow(doc, data, 'Chat');
  win.resize(width, height);
  const wt = data.windows.get('Chat');
  const area = child(wt, 'ChatControlDef');
  const entry = (wt.c || []).find((c) => c.t === 'FullResizeImageDef' && str(c, 'ControlId') === '1002');
  const entryH = entry ? num(entry, 'Height', 16) : 16;
  const TAB_H = 18; // the stock tab button height; the text area starts below the tabs
  const ap = { x: (pt(area, 'Position') || { x: 5 }).x, y: TAB_H + 2 };
  const ct = data.tpl('ChatControlTemplate', str(area, 'TemplateName'));
  const off = pt(area, 'TextOffset') || { x: 9, y: 8 };
  const fontName = (ct && str(ct, 'FontName')) || 'chat_small';
  const font = data.fonts.get(fontName);

  // The text area fills the window above the entry bar.
  const areaW = width - ap.x * 2, areaH = height - ap.y - entryH - 4;
  const bgTpl = (ct && str(ct, 'BackgroundTemplate')) || '';
  const bg = frame(doc, data, bgTpl, areaW, areaH);
  bg.style.left = ap.x + 'px';
  bg.style.top = ap.y + 'px';
  bg.style.pointerEvents = 'none';
  const old = win.ctl['1000'];
  if (old) old.replaceWith(bg);
  else win.el.appendChild(bg);

  const lines = doc.createElement('div');
  Object.assign(lines.style, {
    position: 'absolute',
    left: off.x + 'px',
    top: off.y + 'px',
    width: areaW - off.x * 2 + 'px',
    height: areaH - off.y * 2 + 'px',
    overflow: 'hidden',
    display: 'flex',
    flexDirection: 'column',
    justifyContent: 'flex-end',
    font: font ? `${font.height + 1}px/${font.height + 3}px "${fontFamily(fontName)}", sans-serif` : '11px/13px sans-serif',
    textShadow: '0 1px 1px #000',
    pointerEvents: 'none',
  });
  bg.appendChild(lines);

  // The entry bar, with a real input over it.
  const bar = doc.createElement('div');
  Object.assign(bar.style, { position: 'absolute', left: '0px', top: height - entryH - 2 + 'px', width: width + 'px', height: entryH + 'px', display: 'none' });
  const oldEntry = win.ctl['1002'];
  if (oldEntry) oldEntry.remove();
  const entryFrame = frame(doc, data, entry ? str(entry, 'TemplateName') : 'chat_entry', width, entryH);
  bar.appendChild(entryFrame);
  const input = doc.createElement('input');
  input.type = 'text';
  input.maxLength = 200;
  input.autocomplete = 'off';
  Object.assign(input.style, {
    position: 'absolute', left: '13px', top: '0px', width: width - 26 + 'px', height: entryH + 'px', boxSizing: 'border-box',
    padding: '0', margin: '0', border: '0', outline: '0', background: 'transparent', color: '#fff',
    font: font ? `${font.height + 2}px "${fontFamily(fontName)}", sans-serif` : '12px sans-serif',
  });
  bar.appendChild(input);
  win.el.appendChild(bar);

  let typing = false;
  const open = (prefill = '') => {
    typing = true;
    bar.style.display = 'block';
    input.value = prefill;
    input.focus();
    onToggle?.(true);
  };
  const close = () => {
    typing = false;
    bar.style.display = 'none';
    input.value = '';
    input.blur();
    onToggle?.(false);
  };
  input.addEventListener('keydown', (e) => {
    e.stopPropagation();
    if (e.key === 'Enter') {
      const t = input.value.trim();
      if (t) onSend?.(t);
      close();
    } else if (e.key === 'Escape') close();
  });
  input.addEventListener('keyup', (e) => e.stopPropagation());

  // Every line is kept; the tab picks which are shown.
  const buffer = [];
  let tab = 'main';
  const shown = (e) => !TAB_KINDS[tab] || TAB_KINDS[tab].has(e.kind);
  const lineEl = (e) => {
    const l = doc.createElement('div');
    l.textContent = e.text;
    l.style.color = LINE_COLOR[e.kind] || LINE_COLOR[''];
    return l;
  };
  const say = (text, kind = '') => {
    const e = { text, kind };
    buffer.push(e);
    if (buffer.length > maxLines * 4) buffer.shift();
    if (!shown(e)) return;
    lines.appendChild(lineEl(e));
    while (lines.childElementCount > maxLines) lines.firstChild.remove();
  };
  const redraw = () => {
    lines.textContent = '';
    for (const e of buffer.filter(shown).slice(-maxLines)) lines.appendChild(lineEl(e));
  };

  // Tabs along the top, drawn with the stock tab button.
  const tabButtons = [];
  [['main', 'Main'], ['chat', 'Chat'], ['group', 'Group']].forEach(([id, label], i) => {
    const b = hrButton(doc, data, 'tab_button', label, 44, () => {
      tab = id;
      tabButtons.forEach((t) => t.select(t === b));
      redraw();
    });
    b.select(id === tab);
    Object.assign(b.el.style, { left: ap.x + 2 + i * 46 + 'px', top: '0px', pointerEvents: 'auto' });
    win.el.appendChild(b.el);
    tabButtons.push(b);
  });

  win.el.style.pointerEvents = 'none';
  input.style.pointerEvents = 'auto';
  return { el: win.el, say, open, close, get typing() { return typing; }, lines, input };
}

// createStockSummary builds the stock summary window: the character's hits,
// endurance and power bars, and the target's name and hits. update takes the
// server's vitals for the character (hp, maxHp, end, maxEnd, level, dead) and
// the target as the viewer holds it (name, level, hp as a fraction, dead), or
// null.
export function createStockSummary(doc, data, { name = '' } = {}) {
  const win = buildWindow(doc, data, 'new_summary');
  win.setText('summary_title', name);
  const frac = (v, max) => (max > 0 ? Math.max(0, Math.min(1, v / max)) : 0);
  return {
    el: win.el,
    update(me, target, targetColor) {
      if (me) {
        win.setBar('summary_player_hits', frac(me.hp, me.maxHp));
        win.setText('summary_player_hits', me.dead ? '' : Math.ceil(me.hp));
        win.setBar('summary_player_end', frac(me.end, me.maxEnd));
        win.setText('summary_player_end', Math.ceil(me.end));
        win.setBar('summary_player_power', 0);
        win.setText('stats_level', me.level || '');
      }
      win.setText('summary_target', target ? target.name + (target.dead ? ' (dead)' : '') : '', target ? targetColor : undefined);
      win.setBar('summary_target_hits', target && !target.dead ? target.hp : 0);
      win.setText('summary_target_hits', target && !target.dead ? Math.round(target.hp * 100) : '');
    },
  };
}

// createStockExp builds the stock experience window: ten coarse bubbles for
// each tenth of the level, and a fine bar for the progress within the tenth
// being filled. update takes the server's vitals (xp, xpNext, level).
export function createStockExp(doc, data) {
  const win = buildWindow(doc, data, 'float_level_exp');
  return {
    el: win.el,
    update(me) {
      const f = me && me.xpNext > 0 ? Math.max(0, Math.min(1, me.xp / me.xpNext)) : 0;
      win.setBar('summary_exp_coarse', Math.floor(f * 10) / 10);
      win.setBar('summary_exp_fine', (f * 10) % 1);
      win.setText('stats_level', me && me.level ? me.level : '');
    },
  };
}

// styleDragId reads the style id out of what a styles window row carries when
// it is dragged ("style:<id>"), or null for anything else.
export function styleDragId(data) {
  const m = /^style:(\d+)$/.exec(data || '');
  return m ? Number(m[1]) : null;
}

// createStockBar builds the style bar: one stock icon per style with its key
// and endurance cost. A style that cannot be used now is shaded, one the
// character cannot pay for is shaded red, one ready to use gets a gold frame.
//   onUse(slot)  the player clicked a style
export function createStockBar(doc, data, { onUse, onDrop, keys = [] } = {}) {
  const root = doc.createElement('div');
  Object.assign(root.style, { display: 'flex', gap: '3px', flexWrap: 'wrap', justifyContent: 'center', maxWidth: '92vw' });
  let slots = [];
  let cells = [];

  const label = (value, font, x, y) => {
    const t = text(doc, data, font, value, 'rgb(255,255,255)');
    t.style.left = x + 'px';
    t.style.top = y + 'px';
    t.style.pointerEvents = 'none';
    t.style.filter = 'drop-shadow(1px 1px 0 #000)';
    return t;
  };

  function set(next) {
    slots = next;
    root.textContent = '';
    cells = slots.map((s, i) => {
      const cell = doc.createElement('div');
      Object.assign(cell.style, { position: 'relative', width: '36px', height: '36px', cursor: 'pointer' });
      cell.title = `${s.nm} (${s.opens}, level ${s.lv}, ${s.cost} endurance)`;
      cell.appendChild(icon(doc, data, styleIconNumber(s.spec), 2, 2));
      const shade = doc.createElement('div');
      Object.assign(shade.style, { position: 'absolute', left: '2px', top: '2px', width: '32px', height: '32px', pointerEvents: 'none' });
      cell.appendChild(shade);
      if (keys[i]) cell.appendChild(label(keys[i], 'arial9', 4, 4));
      cell.appendChild(label(String(s.cost), 'arial9', 20, 24));
      cell.addEventListener('click', () => onUse?.(i));
      // A style dragged from the styles window lands in this slot.
      cell.addEventListener('dragover', (e) => e.preventDefault());
      cell.addEventListener('drop', (e) => {
        e.preventDefault();
        const id = styleDragId(e.dataTransfer?.getData('text/plain'));
        if (id !== null) onDrop?.(i, id);
      });
      root.appendChild(cell);
      return { cell, shade };
    });
  }

  // update takes the server's vitals: which slots may be used (avail) and endurance.
  function update(me) {
    if (!me) return;
    cells.forEach(({ cell, shade }, i) => {
      const ready = (me.avail || []).includes(i);
      const poor = me.end < slots[i].cost;
      shade.style.background = poor ? 'rgba(160,0,0,.55)' : ready ? 'transparent' : 'rgba(0,0,0,.55)';
      cell.style.outline = ready && !poor ? '1px solid #ffd76a' : 'none';
    });
  }

  return { el: root, set, update, get count() { return slots.length; } };
}

// The server stands a dead character up after this many seconds unless they
// release sooner (world.PlayerRespawnS).
export const RESPAWN_SECONDS = 5;

// createStockRelease builds the stock Release window: a timer bar, a line of
// text and the release buttons. RELEASE and RELEASE BIND both stand the
// character up at their bind point, the only place the server returns them
// to; the city and house buttons stay greyed.
//   onRelease()  the player pressed release
// update(dead, now) shows the window while dead and runs the timer from the
// moment of death; now is in seconds.
export function createStockRelease(doc, data, { onRelease } = {}) {
  const win = buildWindow(doc, data, 'Release', {
    onClick: (ev) => {
      if (ev === 'Release' || ev === 'ReleaseToBind') onRelease?.();
    },
    disabled: (ev) => ev === 'ReleaseToCity' || ev === 'ReleaseToHouse',
  });
  win.el.style.display = 'none';
  let since = null;
  return {
    el: win.el,
    get isOpen() {
      return since !== null;
    },
    update(dead, now) {
      if (!dead) {
        if (since !== null) {
          since = null;
          win.el.style.display = 'none';
        }
        return;
      }
      if (since === null) {
        since = now;
        win.el.style.display = 'block';
      }
      const left = Math.max(0, RESPAWN_SECONDS - (now - since));
      win.setBar('release_timer_time', left / RESPAWN_SECONDS);
      win.setText('release_timer_text', left > 0 ? `You will be released in ${Math.ceil(left)} seconds` : 'Releasing...');
    },
  };
}
