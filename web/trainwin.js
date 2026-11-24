// The trainer window: the stock Train window filled from a trainer's offer.
//
// Its upper list shows the specializations that can be trained: name and
// level in the first column, the cost in points of the next level in the
// second. The label above it gives the points left. Clicking a row selects it;
// TRAIN, or a double-click on the row, asks the server for one level in it.
// INFO says what the chosen spec's next levels give. When TRAIN or INFO cannot
// act, onMessage says why. The server decides.

import { buildWindow } from './uiwin.js';
import { infoLines } from './sheet.js';

// trainRows are the list rows of a trainer view: the cost in its own column,
// and a spec that cannot be trained now dimmed with the reason as its tooltip.
export function trainRows(view) {
  return view.rows.map((r) => ({ key: r.key, text: r.text, right: String(r.cost), on: r.can, title: r.hint, view: r }));
}

// pointsText is the label over the list; it fits the template's small box.
export const pointsText = (n) => n + (n === 1 ? ' point' : ' points');

// trainAction is what pressing TRAIN on a row does: {train: key} or {say: text}.
export function trainAction(row) {
  if (!row) return { say: 'Choose a specialization to train first.' };
  if (!row.on) return { say: row.view.reason || 'You cannot train that now.' };
  return { train: row.key };
}

// createStockTrainer builds the window, hidden. show(view) fills and opens it;
// onTrain(key) asks for a level; onMessage(text) is a line for the log.
export function createStockTrainer(doc, data, { onTrain, onMessage, onClose, mover } = {}) {
  const win = buildWindow(doc, data, 'train', {
    captions: true, labels: true,
    onClose: () => api.hide(),
    onClick: (id, ev, label) => {
      if (ev === 'Train' || label === 'TRAIN') train();
      else if (ev === 'TrainInfo' || label === 'INFO') info();
    },
  });
  Object.assign(win.el.style, { position: 'fixed', left: '50%', top: '90px', marginLeft: '-150px', zIndex: 15, display: 'none' });
  mover?.makeMovable(win.el, 'trainer');
  const list = win.list('1002');
  let rows = [], selected = -1, open = false, title = '';
  const act = (row) => {
    const a = trainAction(row);
    if (a.train) onTrain?.(a.train); else onMessage?.(a.say);
  };
  const train = () => act(rows[selected]);
  const info = () => { for (const l of infoLines(rows[selected] && rows[selected].view)) onMessage?.(l); };
  if (list) {
    list.onSelect((i) => { selected = i; });
    list.onActivate((i) => { selected = i; act(rows[i]); });
  }
  const api = {
    el: win.el,
    get isOpen() { return open; },
    get selected() { return selected; },
    // show fills the window; a refill keeps the chosen row when it is still there.
    show(view) {
      const keep = rows[selected] ? rows[selected].key : null;
      rows = trainRows(view);
      title = view.title;
      selected = keep === null ? -1 : rows.findIndex((r) => r.key === keep);
      win.setText('train_skill_points', pointsText(view.points));
      if (list) { list.setRows(rows); list.select(selected); }
      open = true;
      win.el.style.display = 'block';
      mover?.restore('trainer');
    },
    hide() {
      if (!open) return;
      open = false;
      win.el.style.display = 'none';
      onClose?.();
    },
    train,
    info,
    get title() { return title; },
  };
  return api;
}

// pairLeft places the trainer's greeting (the 384-wide talk window) and the
// 300-wide Train window side by side, centred together on a screen that wide,
// and from the left edge on a narrower one.
export function pairLeft(screenW, talkW = 384, trainW = 300, gap = 8) {
  const talk = Math.max(8, Math.round((screenW - talkW - gap - trainW) / 2));
  return { talk, train: talk + talkW + gap };
}
