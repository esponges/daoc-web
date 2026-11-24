// The quest window: the journal of your quests and the current step of each,
// the offer an NPC makes you, and the choice of reward a finished quest gives.
// A plain panel drawn from the server's `quests`, `questoffer` and
// `questreward` messages (docs/protocol.md); the pure parts are here for node.

// questView is the journal's content from a `quests` message.
export function questView(m) {
  if (!m) return { rows: [], done: [] };
  return {
    rows: (m.active || []).map((q) => ({
      id: q.id, title: q.nm,
      text: q.text + (q.need > 1 ? ' (' + (q.have || 0) + ' of ' + q.need + ')' : ''),
    })),
    done: m.done || [],
  };
}

// offerView is the prompt for a `questoffer`.
export function offerView(m) {
  return { id: m.id, head: m.nm + ' - ' + m.npc, text: m.text, reward: m.reward ? 'Reward: ' + m.reward : '' };
}

const el = (doc, tag, cls, text) => {
  const e = doc.createElement(tag);
  if (cls) e.className = cls;
  if (text !== undefined) e.textContent = text;
  return e;
};

// createQuests builds the window, hidden.
//   onOp(op, id, n)   a quest request: "accept", "decline", "abandon", "choose" or "journal"
export function createQuests(doc, { onOp, mover } = {}) {
  const root = el(doc, 'div', 'panel quests');
  Object.assign(root.style, { position: 'fixed', left: '50%', top: '110px', marginLeft: '-210px', width: '420px', zIndex: 16, display: 'none' });
  if (mover && mover.makeMovable) mover.makeMovable(root, 'quests');
  const title = el(doc, 'h2', '', 'Quests');
  const offer = el(doc, 'div');
  const list = el(doc, 'div', 'items');
  Object.assign(list.style, { maxHeight: '220px', overflowY: 'auto' });
  const done = el(doc, 'div', 'dim');
  const reward = el(doc, 'div');
  root.append(title, offer, reward, list, done);
  let open = false;
  const show = () => { open = true; root.style.display = 'block'; };
  const btn = (label, fn) => { const b = el(doc, 'button', '', label); b.addEventListener('click', fn); return b; };
  return {
    el: root,
    get isOpen() { return open; },
    set(m) {
      const v = questView(m);
      list.replaceChildren(...(v.rows.length ? v.rows.map((r) => {
        const d = el(doc, 'div', 'row');
        d.append(el(doc, 'b', '', r.title), el(doc, 'div', '', r.text), btn('Abandon', () => onOp && onOp('abandon', r.id)));
        return d;
      }) : [el(doc, 'div', 'dim', 'No quests.')]));
      done.textContent = v.done.length ? 'Done: ' + v.done.join(', ') : '';
    },
    offer(m) {
      const v = offerView(m);
      offer.replaceChildren(el(doc, 'b', '', v.head), el(doc, 'div', '', v.text), el(doc, 'div', 'dim', v.reward),
        btn('Accept', () => { offer.replaceChildren(); onOp && onOp('accept', v.id); }),
        btn('Decline', () => { offer.replaceChildren(); onOp && onOp('decline', v.id); }));
      show();
    },
    reward(m) {
      reward.replaceChildren(el(doc, 'b', '', 'Choose your reward for ' + m.nm),
        ...(m.choose || []).map((c) => btn(c.nm, () => { reward.replaceChildren(); onOp && onOp('choose', m.id, c.n); })));
      show();
    },
    toggle() { open = !open; root.style.display = open ? 'block' : 'none'; if (open) onOp && onOp('journal'); },
  };
}
