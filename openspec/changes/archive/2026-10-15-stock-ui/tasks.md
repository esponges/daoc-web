## 1. Spike

- [x] 1.1 Find where the inventory/paperdoll, quick bar and style bar are defined (search `ui/`, `data/skin1/`, `icons/`, and strings in the client binary); verify by recording, per window, the file or "not in the install" in design.md
- [x] 1.2 Parse `uimain.xml` and the chat window with a throwaway reader; verify it prints the chat window's templates, sizes and textures

## 2. Converter

- [x] 2.1 Add `internal/ui` XML reader resolving includes and templates; verify unit tests on inline samples and a run over every `atlantis` file with unknown elements reported
- [x] 2.2 Add TGA and BMP readers beside `internal/dds`; verify tests decode a known atlas to the right size and a corner pixel
- [x] 2.3 Add bitmap-font glyph tables; verify a sample string lays out with the expected widths
- [x] 2.4 Add `cmd/uiconv` writing JSON, PNG and an icon-sheet table to `web/data/ui/`; verify the report counts layouts, textures and missing files and `git status` shows nothing new
  - Done: `cmd/uiconv` writes layout JSON, PNG textures, TrueType fonts, bitmap glyph tables and the 19 icon sheets (verified: report counts, `git status` clean). Open: icon numbers in the imported data do not map onto the install's sheets (see design.md, "Icon numbers"); left unticked until the item and style lookup is settled in 4.2/4.3.

## 3. Renderer and chat

- [x] 3.1 Add `web/uiwin.js` that builds a window from its JSON (frames, buttons with three states, image areas, text); verify a node test over a small layout and the chat window drawn in the browser
- [x] 3.2 Move chat to the stock chat window with channel colours; verify two tabs exchange a line in the new window

## 4. Bars, quick bar, inventory

- [x] 4.1 Target and own health, endurance and mana bars from snapshots; verify the bars follow a fight in the browser
- [x] 4.2 Style and quick bar with stock icons, opener hint and recast shade, rebuilt on weapon change; verify a style is used by click and key
  - Done with the weapon line, not per-style icons: the imported icon numbers do not map onto this install's sheets (design.md, "Icon numbers"); the bar message gained `spec` (additive). Verified in the browser: key 2 and a click both reach the server.
- [x] 4.3 Inventory and gear with item icons, drag to equip and refusal reasons; verify equip, unequip and a refused item
- [x] 4.4 Fall back to the built-in UI when `web/data/ui/` is absent; verify the viewer works with the folder removed

## 5. Wrap-up

- [x] 5.1 Document the UI conversion in the README; verify the commands run as written
- [x] 5.2 Run all Go and JS tests; verify they pass
