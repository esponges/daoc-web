// The task window: the task you hold, what to do next, the time left and how
// many you have done at this level.
//
// The server sends `task` whenever it changes; the window shows it, counts the
// time down, and sends abandon-task. With the converted interface it is the
// stock Quest Journal (its tree of entries, the quest title strip and the
// REMOVE button); without, a plain panel.

import { frame, buildWindow } from './uiwin.js';

// taskView turns the task message into what to draw; null when there is none
// to show. The window shows even with no task, saying so.
export function taskView(m, nowS, receivedS) {
  if (!m) return null;
  const used = `${m.n || 0} of ${m.max || 0} tasks done at this level`;
  if (!m.on) return { title: 'Task', lines: ['You have no task.', 'Ask a town guard or a trainer.'], foot: used, hasTask: false };
  const left = Math.max(0, (m.left || 0) - Math.max(0, nowS - receivedS));
  return {
    title: m.nm || 'Task',
    lines: [m.tx || '', m.gv ? 'Given by ' + m.gv : ''].filter(Boolean),
    foot: timeLeft(left) + ' left, ' + used,
    hasTask: true,
    done: !!m.dn,
  };
}

// timeLeft is seconds as "1h 59m" or "12m 5s".
export function timeLeft(s) {
  s = Math.max(0, Math.floor(s));
  const h = Math.floor(s / 3600), m = Math.floor((s % 3600) / 60);
  if (h > 0) return h + 'h ' + String(m).padStart(2, '0') + 'm';
  return m + 'm ' + String(s % 60).padStart(2, '0') + 's';
}

// journalGroups is the journal tree's content: the task as one entry, its
// description, who gave it, the time left and how many are done opened under it;
// with no task, a note saying so. open says whether the entry is unfolded.
export function journalGroups(v, open = true) {
  if (!v || !v.hasTask) return [{ title: 'Tasks', open, rows: v ? v.lines.map((t) => ({ text: t })) : [] }];
  return [{ title: v.title + (v.done ? '  (done)' : ''), open, rows: [...v.lines, v.foot].map((t) => ({ text: t })) }];
}

// createTaskWindow builds the window; update(m) takes a task message, toggle()
// opens or closes it. onAbandon() sends abandon-task.
export function createTaskWindow(doc, data, { onAbandon, now = () => performance.now() / 1000, mover } = {}) {
  if (data && data.window('new_quest_journal')) return createStockJournal(doc, data, { onAbandon, now, mover });
  const root = doc.createElement('div');
  root.className = data ? '' : 'panel task';
  const W = 300;
  let msg = null, at = 0, open = false;
  const render = () => {
    root.replaceChildren();
    if (!open) { root.style.display = 'none'; return; }
    root.style.display = data ? 'block' : '';
    if (!data) root.classList.add('open');
    const v = taskView(msg, now(), at) || taskView({ on: false }, 0, 0);
    const body = doc.createElement('div');
    const line = (t, cls) => { const d = doc.createElement('div'); if (cls) d.className = cls; d.textContent = t; return d; };
    body.append(line(v.title, 'task-title'));
    for (const l of v.lines) body.append(line(l, 'task-text'));
    body.append(line(v.foot, 'task-foot'));
    if (v.hasTask) {
      const b = doc.createElement('button');
      b.textContent = 'Abandon';
      b.addEventListener('click', () => onAbandon?.());
      body.append(b);
    }
    if (data) {
      const H = 110 + 16 * v.lines.length;
      Object.assign(root.style, { position: 'fixed', width: W + 'px', height: H + 'px' });
      Object.assign(body.style, { position: 'absolute', left: '12px', top: '10px', width: W - 24 + 'px', font: '12px sans-serif', color: '#fff', textShadow: '0 1px 1px #000' });
      root.append(frame(doc, data, 'dlg_sm_title_noresize', W, H));
    }
    root.append(body);
  };
  render();
  return {
    el: root,
    get isOpen() { return open; },
    update(m) { msg = m; at = now(); render(); },
    toggle() { open = !open; render(); },
    // tick redraws the countdown while the window is open.
    tick() { if (open) render(); },
  };
}

// createStockJournal builds the stock Quest Journal as the task window.
function createStockJournal(doc, data, { onAbandon, now, mover }) {
  let msg = null, at = 0, open = false, folded = false;
  const win = buildWindow(doc, data, 'new_quest_journal', {
    captions: true, labels: true,
    onClose: () => api.toggle(),
    onClick: (id, ev) => { if (ev === 'RemoveQuest') onAbandon?.(); },
  });
  Object.assign(win.el.style, { position: 'fixed', left: '200px', top: '110px', zIndex: 14, display: 'none' });
  mover?.makeMovable(win.el, 'journal');
  const tree = win.tree('1001');
  if (tree) tree.onToggle((g, isOpen) => { folded = !isOpen; });
  const remove = win.button('RemoveQuest');
  const render = () => {
    win.el.style.display = open ? 'block' : 'none';
    if (!open) return;
    const v = taskView(msg, now(), at) || taskView({ on: false }, 0, 0);
    if (tree) tree.setGroups(journalGroups(v, !folded));
    win.setText('1005', v.hasTask ? v.title : 'No Quest Selected');
    if (remove) remove.disable(!v.hasTask);
  };
  const api = {
    el: win.el,
    get isOpen() { return open; },
    update(m) { msg = m; at = now(); render(); },
    toggle() { open = !open; if (open) mover?.restore('journal'); render(); },
    // tick redraws the countdown while the window is open.
    tick() { if (open) render(); },
  };
  return api;
}
