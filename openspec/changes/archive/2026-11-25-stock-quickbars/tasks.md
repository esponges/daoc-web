# Tasks: stock-quickbars

## 1. Server
- [x] 1.1 `style` takes an optional `id`; `requestStyleID` queues it by id and refuses a held-back style by name
- [x] 1.2 Test: `TestStyleByID`; protocol round trip; `docs/protocol.md`

## 2. Client
- [x] 2.1 `web/quickbars.js`: layout model (three bars, ten pages, ten slots), drags, new-thing placement, legacy spell bar carry-over, resolve and shading, keys, placement
- [x] 2.2 The bars drawn from the skin (slot frames, `scroll_up`/`scroll_down`, page number) and as plain slots without it
- [x] 2.3 `web/ui.js` uses them for styles, spells and abilities; the old style bar, spell bars and ability buttons removed
- [x] 2.4 Abilities draggable from the skills window and plain sheet; held-back styles draggable
- [x] 2.5 Tests: `web/quickbars.test.mjs`; conc, charwin, spells, quickbar tests moved off the removed bars
- [x] 2.6 Checked in the browser: bars filled, Alt+1 uses Guard, key 2 sends a style by id, drag between bars, wheel turns a page
