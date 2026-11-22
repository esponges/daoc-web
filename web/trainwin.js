// The trainer window: the stock Train window filled from a trainer's offer.
//
// Its upper list shows the specializations that can be trained (name, level
// and the cost in points), the label above it the points left. Clicking a row
// selects it; TRAIN asks the server for one level in it. The server decides.

import { buildWindow } from './uiwin.js';

// trainRows are the list rows of a trainer view: the cost sits in the row, and
// a spec that cannot be trained now is dimmed with the reason as its tooltip.
export function trainRows(view) {
  return view.rows.map((r) => ({ key: r.key, text: r.text + '   ' + r.cost + ' pts', on: r.can, title: r.hint }));
}

// pointsText is the label over the list.
export const pointsText = (n) => n + ' points to spend';

// createStockTrainer builds the window, hidden. show(view) fills and opens it;
// onTrain(key) is a click on TRAIN with a row chosen.
export function createStockTrainer(doc, data, { onTrain, onClose, mover } = {}) {
  const win = buildWindow(doc, data, 'train', {
    captions: true, labels: true,
    onClose: () => api.hide(),
    onClick: (id, ev, label) => { if (ev === 'Train' || label === 'TRAIN') train(); },
    disabled: (ev) => ev === 'TrainInfo',
    tooltip: () => 'Not built yet.',
  });
  Object.assign(win.el.style, { position: 'fixed', left: '50%', top: '90px', marginLeft: '-150px', zIndex: 15, display: 'none' });
  mover?.makeMovable(win.el, 'trainer');
  const list = win.list('1002');
  let rows = [], selected = -1, open = false, title = '';
  const train = () => {
    const r = rows[selected];
    if (r && r.on) onTrain?.(r.key);
  };
  if (list) list.onSelect((i) => { selected = i; });
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
    get title() { return title; },
  };
  return api;
}
