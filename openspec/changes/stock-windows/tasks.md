# Tasks

Every window task ends with a browser check against the stock look, using a scratch server on its own port and data copy. The task is ticked only after that check and its node test pass.

## 1. Controls in the window builder

- [x] 1.1 Replace `buildWindow`'s switch with a table from control type to draw function. Log one console line for each control type nobody draws. Add a lookup of templates by `Name` next to the WindowId lookup, so that `stats_spells` and `stats_combat` (`c_windowInvalid`) build. Verify with a `uiwin.test.mjs` case per lookup and one for the unknown-type log.
- [x] 1.2 Draw `HorizontalResizeImageDef`, `VerticalResizeImageDef`, `HorizontalResizeImageButtonDef`, `IconDef`, `StatusIconDef` and `DynamicImageDef` from their templates. Verify with node tests of the cell layout for each, and with `/uidemo.html` showing the Merchant inset strips and coin decals as in the stock art.
- [x] 1.3 List boxes with the vertical scrollbar template: rows `{text, color, icon, title, drag}`, selection, wheel scrolling and `scrollTo`. Run `uiconv` and extend it if the list or scrollbar templates need fields it now drops, then document the re-run in the README. Verify with node tests (overflow shows the scrollbar, the wheel moves the rows, a click selects) and the Vault list in `/uidemo.html`.
- [x] 1.4 Tree controls (`spell_tree_control` and others): headings that expand and collapse with the tree check art, and rows with small icons. Verify with a node test of expand/collapse and the visible rows, and `stats_spells` in `/uidemo.html`.
- [x] 1.5 Checkboxes, tabs, sliders, combo boxes, text areas and edit boxes, with their fill APIs (`check`, `slider`, `combo`, `text`). Verify with node tests per control, and `MiniPet` and `Interact` drawn in `/uidemo.html`.
- [x] 1.6 Moving, closing and remembering: a drag handle over the title area when `MoveButton` is set, the close button when `CloseButton` is set, positions in `localStorage` (try/catch), clamping into the viewport, and `/resetui` handled by the client. Verify with node tests of the clamp and of a store that throws, and in the browser: move the Spells window, reload, and it is where it was left.
- [x] 1.7 Add every window of this change to `/uidemo.html` with sample data. Verify each draws with no "no drawing for" console line.

## 2. Chat

- [ ] 2.1 Stock chat with a 500-line ring, the stock scrollbar and mouse-wheel scrollback, stick-to-bottom, resizing from the corner, moving, and a 450×230 starting size. Main shows combat, spell and system lines; Chat and Group show their own. Verify with `hud.test.mjs` cases for scrollback, stick-to-bottom and tab filtering, and in the browser during a fight: scroll up, the lines stay put, and switching tabs filters.

## 3. Character windows

- [x] 3.1 Rebuild SPELLS on `stats_spells`: the tree with one heading per line and spells with icon by kind, name and level. A double-click casts, and dragging puts the spell on the quickbar. Verify with a `charwin` node test (a Bonedancer's lines and rows) and in the browser: open SPELLS, expand Bone Army, and cast Summon Bone Boss from it.
- [x] 3.2 Rebuild Skills on `stats_spec_abil`'s three list boxes and points label, and Styles on the `stats_combat` tree (the game's own Styles window), keeping style drag to the bar. Verify with node tests (a level 10 Warrior's specs and Guard I, and a long ability list scrolls) and in the browser: drag a style to the bar and use it by key.
- [x] 3.4 Attributes (ATTRIB. and `C`) on `StatsAttributes`: the stat strips, hit points, armour factor and weapon skill labels, the resist and info list boxes, drawn with the stock insets, with only what the server sends filled in. Verify with a `charwin` node test of the rows and in the browser: open ATTRIB. at level 20 and see it match the stock layout (screenshot for the user).
- [ ] 3.3 Trainer on the stock `Train` window: the specs list, points labels and train button, using the existing trainer messages. Verify with a node test of the rows from a trainer message and in the browser: train a spec at a trainer and see the points drop.

## 4. NPC windows

- [ ] 4.1 Merchant on the stock `Merchant` window, following the merchants spec delta. The list box shows the page's items. The page buttons and label turn the pages. The quantity sends that many `buy` messages and stops at the first refusal. The coin fields show the selected price, the title shows the merchant's name, and selling by click on the pack is kept. Verify with `shop` node tests (pages, quantity, coins) and in the browser: buy three packs of arrows and sell an item.
- [ ] 4.2 Vault on the stock `Vault` window, keeping drag between pack and vault. Verify with a node test of the rows and in the browser: store and take an item at a vault keeper.
- [ ] 4.3 Healer and NPC talk on the stock `Interact` window (title and text area). Verify with a node test and in the browser: right-click a healer.
- [x] 4.4 Task window on the stock `Quest` journal (tree of tasks and abandon button), replaying the last task message on open. Verify with a `task.test.mjs` case and in the browser: ask a task, see it listed, and abandon it.

## 5. Combat windows

- [ ] 5.1 Pet window on the stock `MiniPet`: title, health bar, stance checkboxes ticked from the server's stance, and command buttons. Verify with a `pet` node test and in the browser: summon, set passive, attack, and the bar falls as the pet is hit.
- [ ] 5.2 Concentration on the stock `Concentration` window: bar, label and the list with cancel. Verify with a `conc` node test and in the browser with a Healer holding two buffs, cancelling one.
- [ ] 5.3 Cast bar on the stock `Timer` window: spell name and progress, hidden when the spell lands, fails or is interrupted; the bow draw uses the same window. Verify with a `spells` node test of the timing and in the browser with a three-second cast.
- [ ] 5.4 Quiver on the stock `WindowQuiver` (four icons and counts, click to pick the active quiver slot). Verify with a node test and in the browser with a Hunter holding two arrow kinds.
- [x] 5.5 Quickbar from the skin's art. Take its place from `default.ini` (`Quickbar`, `Quickbar2`, `Quickbar3`; pick the line for the nearest resolution) and replace the plain `.spellbar` buttons with it, so the spell bar is the stock one. Find the slot, frame and number source rectangles in the converted Atlantis textures and put them in one table. Compare against period screenshots of the Trials of Atlantis interface, which are not committed. Build ten slots with keys 1–0 for styles and spells, recharge shading, red for unaffordable styles, tooltips and bar paging. Verify with node tests of slot state and in the browser: put a style and a spell on the bar, use both by key, and save a screenshot for the user.

## 6. Fallback and close

- [ ] 6.1 Fallback: move `web/data/ui/` aside and check that every window of this change opens its built-in panel and works. Verify in the browser, then put the folder back.
- [ ] 6.2 Update the specs' scenarios as built, the README's interface section and CLAUDE.md's "Open now". Then run `go test -short ./...`, all `web/*.test.mjs` and `openspec validate stock-windows`. Verify that all of them pass.

## Workflow follow-up

- A second change, `spell-feedback`, follows: pet combat lines in the log, the cast animation, spell visuals on targets, a proper buff row with target and pet effects, and the double cast fix.
- Archive with `/opsx:archive stock-windows` once the user has looked at the windows.
