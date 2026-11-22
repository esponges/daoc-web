// The chat's pure parts: the ring of lines it keeps, when a new line should
// carry the view along, and the sizes the window may take.

// CHAT_MAX is how many lines the chat keeps; older ones fall off the top.
export const CHAT_MAX = 500;

// CHAT_SIZE is the window's starting size; CHAT_MIN and CHAT_LIMIT bound resizing.
export const CHAT_SIZE = { w: 450, h: 230 };
export const CHAT_MIN = { w: 260, h: 110 };

// pushRing adds entry to the end of list, dropping the oldest past max.
export function pushRing(list, entry, max = CHAT_MAX) {
  list.push(entry);
  while (list.length > max) list.shift();
  return list;
}

// atBottom says whether a view scrolled to top (of total pixels, view of them
// showing) is at the end, within slack: a new line then keeps it there, and a
// view that was scrolled up is left alone.
export function atBottom(top, view, total, slack = 3) {
  return total - view - top <= slack;
}

// chatSize keeps a dragged size within the window's bounds and the screen's.
export function chatSize(w, h, vw, vh) {
  return {
    w: Math.max(CHAT_MIN.w, Math.min(Math.max(CHAT_MIN.w, vw - 20), Math.round(w))),
    h: Math.max(CHAT_MIN.h, Math.min(Math.max(CHAT_MIN.h, vh - 20), Math.round(h))),
  };
}
