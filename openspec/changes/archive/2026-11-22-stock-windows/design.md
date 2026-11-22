# Design: stock-windows

## Context

`web/uiwin.js` turns a converted `WindowTemplate` into DOM. `buildWindow` handles these control types:
- `FullResizeImageDef`, `ChatControlDef`, `ButtonDef`, `ImageAreaDef`, `LabelDef`/`ScalarLabelDef`, `InvisibleButtonDef`, `HorizontalResizeButtonDef`, `CompassControlDef`, `StatusBarDef`.

Everything else falls into `default: continue`. Across the Atlantis layout, the skipped controls include:
- `HorizontalResizeImageDef` ×145, `IconDef` ×150, `StatusIconDef` ×397, `CheckBoxDef` ×166 and `HorizontalResizeImageButtonDef` ×103;
- `ListBoxDef` ×19, `ComboBoxDef` ×27, `HorizontalSliderDef` ×22, `TextAreaDef` ×15, `IconSetDef` ×8 and `DockableIconDef` ×7;
- `TreeControlDef` ×3, `TabsDef` ×2, `DynamicImageDef` ×2, `VerticalResizeImageDef` ×2, `VerticalSliderDef` ×4;
- `EditBoxDef` ×1 and `ClickableEditBoxDef` ×1.

So the windows that need lists (Spells, Skills, Styles, Merchant, Vault, Train, Concentration) were finished in `charwin.js`, `shop.js` and others with absolutely positioned HTML text. Some never used a stock template at all (`pet.js`, `task.js`, `conc.js`, the trainer and healer panels in `ui.js`, the cast bar and quickbar in `spells.js`).

The window templates exist in the converted layout under these ids, all checked in `web/data/ui/atlantis/layout.json`:
- `Merchant` 350×458, `MiniPet` 135×125, `Vault` 300×430, `Quest` (`new_quest_journal`) 512×256;
- `Train` 300×450, `Interact` 384×320, `Concentration` 128×104, `Timer` 300×96;
- `WindowQuiver` 240×100, `stats_spells` 192×416, `stats_spec_abil` 192×416, `stats_combat` 192×416, `Chat` 350×200.

The install's `default.ini` holds the stock quickbars' places (`Quickbar=304,489,1,1,100,0,1`, with a `Quickbar2` and `Quickbar3`, and other values for each resolution); the layout files do not.

`stats_spells` and `stats_combat` have the WindowId `c_windowInvalid`, so they must be looked up by template `Name`, not by WindowId.

The install's `ui/isles` (Shrouded Isles) skin is incomplete. About 100 template references name textures (`page1`–`page3`, `frames_win`, `navigation`) that it doesn't define, so it draws blank. This change stays on the Atlantis skin, which is also the skin of the 1.65 (Trials of Atlantis era) XML interface.

## Goals / Non-Goals

**Goals:**
- The builder draws every control type above from its template (row art, check art, scrollbar pieces, font, colour).
- Window code binds data to controls by adapter name or control id, through a small API on the built window. It no longer positions its own HTML.
- Each window in the proposal's table opens from its own template.
- Windows can be moved, closed and remembered.

**Non-Goals:**
- Windows for systems we don't have: crafting, keeps, realm war maps, the bazaar, master levels, the LFG and community windows, and the options window beyond what we already offer.
- A Shrouded Isles skin.
- Real spell and item icons by number. They stay chosen by line and kind (CLAUDE.md: the imported icon numbers don't match this install's sheets).
- Server and protocol changes.

## Decisions

### 1. Controls are components with a fill API, registered by type

Each new control is a function `(doc, data, def, rect) → { el, ...api }` in `uiwin.js`, or in a new `uictl.js` if `uiwin.js` grows past about 1000 lines. A table maps the `Def` type to the function, replacing the switch. The built window exposes:
- `win.list(idOrAdapter)` returns `{ setRows(rows), onSelect, selected, scrollTo }`;
- `win.tree(id)` returns `{ setGroups([{ title, open, rows }]) }`;
- `win.check(id)` returns `{ set(bool), onChange }`;
- `win.icon(id)`, `win.text(id)`, `win.slider(id)` and `win.combo(id)`.

Rows are plain data: `{ text, color, icon, title, drag }`. The control draws them in its template's font and row art.

- *Alternative:* keep per-window HTML and just restyle it with CSS to look stock. That's rejected, because it's how we got here. It drifts from the layout and has to be redone per window.
- *Alternative:* render windows to a canvas. That's rejected: it loses text selection, tooltips, drag and drop and accessibility, and DOM sprites already reproduce the art exactly.

### 2. Templates are found by WindowId, else by Name

`UIData.windows` keys by WindowId and silently drops `c_windowInvalid` duplicates. A second map keyed by template `Name` makes `buildWindow(doc, data, 'stats_spells')` work. Existing callers keep using WindowIds.

### 3. Unknown controls are reported, not dropped

`default:` logs `uiwin: no drawing for <Type>` once per type. That way a window that looks empty in the browser says why on the console. This is the spec's "unknown control" scenario.

### 4. Quickbar from skin pieces, checked against period screenshots

The quickbar has no XML; the 1.65 client drew it itself. We build it from sub-rectangles of the Atlantis textures: slot frame, background strip, the bar number arrows from `smallarrows`, and the existing icon choice by line and kind. The rectangles are found by viewing the converted PNGs. They are compared side by side with period screenshots of the Trials of Atlantis interface, used as reference only and never committed or stored in the repo. The rectangles go into one table in the client, so correcting one is a one-line change.

- *Alternative:* draw the bar with the `Command` window's 22 buttons. That's rejected: Command is a different window (a page of commands, not slots).

### 5. Chat: buffer, scroll, resize, move

The chat window keeps a ring of 500 lines and draws the visible ones for the active tab. A stock vertical scrollbar, or the mouse wheel, moves the view. "Stick to bottom" holds while the view is at the end. The size starts at 450×230 and is resizable from its corner using the template's resize image buttons. Tab filters stay as in `hud.js` (`TAB_KINDS`); combat, spell and system kinds belong to Main.

### 6. Moving and remembering windows

`buildWindow` adds a drag handle over the title area (`TitleHeight` × `TitleWidth`) when `MoveButton` is true, and wires the close button when `CloseButton` is true. Positions are stored in `localStorage` under `ui.pos.<WindowId|Name>`, wrapped in try/catch. They're clamped into the viewport on open and on resize. Windows that are fixed parts of the HUD (summary, menu bar, compass) stay where they are placed.

### 7. Buying a quantity is repeated buys

The `buy` message has no count, and the protocol is additive-only. The Merchant window's quantity sends that many `buy` messages, and the server decides each one as today. That's no change for the server, and a refusal stops the rest.

### 8. Fallback untouched

Every window keeps its built-in panel path when `stock` is null. The new controls are only reached through `buildWindow`, which only runs with stock data.

## Risks / Trade-offs

- **Some template fields aren't converted:** list, tree and scrollbar templates may reference art fields the converter drops (it reports 100+ unknown elements). → The task for the controls includes checking `uiconv`'s unknown list and extending it where a control needs the field. A re-run of `uiconv` is then required on every install. The README says so.
- **The quickbar look is our reconstruction:** there's no layout for it. → Rectangles in one table, compared against period screenshots, and the result is called out to the user for a look.
- **Breaking current windows:** moving windows onto templates can break flows that work today, such as drag from styles to the bar, selling by click, or cancelling a concentration entry. → Node tests per window for the data binding, and a browser check of each flow before the task is ticked.
- **Remembered positions after a layout change:** a window could be pushed off-screen. → It's clamped on open, and typing `/resetui` in chat clears the remembered positions. The client handles this itself and doesn't send it to the server. The MENU button has no window yet, so it can't carry the reset.

## Migration Plan

This is client-only. Players re-run nothing unless `uiconv` gains fields. In that case the README's setup line `go run ./cmd/uiconv -skin atlantis` is re-run, and the client keeps working on an old conversion, because missing fields are drawn plainer, not failed. Rollback is a revert of the client files.

## Open Questions

- Exact source rectangles for the quickbar frame and slot art. This is settled by looking at the converted textures during the quickbar task; the approach doesn't change.
