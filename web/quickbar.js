// The quickbar's art and place: its slot frame from the skin, and where
// default.ini puts it. quickbars.js builds the bars from these.
//
// The game's client draws its quickbar itself: no XML layout describes it. Its
// frames are in the skin's frames_win texture and its place is in default.ini
// (Quickbar=329,644 on a 1024 by 768 screen, Quickbar2 beneath it, 42 pixels
// lower). The rectangles below were measured off the converted texture; they
// are one table so that correcting one is a one-line change. Where the
// converted skin is absent the callers keep their plain bars.

import { paint } from './uiwin.js';

// SLOT_ART is one slot's frame in the frames_win texture: the dark square with
// its gold bevel. SLOT is the size it is drawn at, and BAR_W the width of a
// bar of ten, as the stock default places it.
export const SLOT_ART = { tex: 'frames_win', sx: 141, sy: 10, sw: 44, sh: 41 };
export const SLOT = 36;
export const GAP = 0;
export const BAR_W = 366;
export const ROW_H = 42;

// quickbarPlace is where a row of the bar sits on a screen of vw by vh pixels:
// centred, its first row 124 pixels from the bottom (as default.ini places it
// on 1024 by 768) and each further row 42 below the one before.
export function quickbarPlace(vw, vh, row = 0) {
  return { left: Math.round((vw - BAR_W) / 2), top: Math.round(vh - 124 + row * ROW_H) };
}

// slotFrame builds one slot's frame element, size pixels square.
export function slotFrame(doc, data, size = SLOT) {
  const e = doc.createElement('div');
  Object.assign(e.style, { position: 'absolute', left: '0px', top: '0px', width: size + 'px', height: size + 'px' });
  paint(e, data, { ...SLOT_ART, dw: size, dh: size });
  return e;
}
