`## Context
`
`The client install holds the UI as data: `ui/uimain.xml` includes about 75 window files per skin (`atlantis/`, `isles/`), each defining `WindowTemplate`, `ButtonTemplate` and `ImageAreaTemplate` elements with sizes, positions and texture names. Textures are DDS and TGA atlases; `spell_icons.xml`, `skill_icons.xml`, `item_icons.xml`, `weapon_icons.xml` and `combat_icons.xml` map icons to regions; fonts are bitmap sheets in `fonts/` and `ui/fonts` plus four TTFs; the classic art is in `data/skin1/*.bmp`. The repo reads DDS (`internal/dds`) but not TGA, BMP fonts or this XML. The current client UI is `web/ui.js` (205 lines) and `web/game.css` (54), driven by the multiplayer protocol, which does not change.
`
`## Goals / Non-Goals
`
`**Goals:** the real windows and art for chat, target and health bars, style and quick bar, inventory and gear; layout read from data wherever the install has it.
`**Non-Goals:** every window (maps, keeps, crafting, quests); the original client's scripting or drag-and-drop internals; mobile layouts; changing the protocol.
`
`## Decisions
`
`- **Convert at build time.** `cmd/uiconv` writes JSON and PNG into `web/data/ui/`, as the other converters do, so the browser never parses XML, TGA or DDS. The renderer takes plain JSON.
`- **DOM plus canvas sprites, not a new GL layer.** Windows are absolutely positioned elements using CSS `background-position` into the atlas PNGs. This keeps text, focus and drag-and-drop to the browser and is enough for a few windows.
`- **Skin: `atlantis` first.** It is the newest skin and its files are the most complete; the skin name is a converter flag, so `isles` or `skin1` is a re-run.
`- **Fonts: bitmap sheets converted to glyph tables, TrueType where a layout names one.** Most window text (`arial9`, `arial11`, buttons, titles) is bitmap; chat text is TrueType. A sheet's metrics are in the sheet itself: under each row of glyphs is a marker row of opaque segments, one per glyph, whose widths are the advances, and the characters run in Windows-1252 order from 0x21 (no space glyph; a space is a fraction of the row height). `brit9` is the one sheet that does not follow this layout (101 segments, a 29-pixel row) and is replaced by its sibling `brit9s`.
`- **Staged, each stage replacing its part of `ui.js`.** Chat, then bars, then the style and quick bar, then inventory and gear. The old UI stays as the fallback until the last stage lands.
`- **Where no layout exists, build by hand from stock art.** The spike has not found an inventory, paperdoll or style-bar XML; stage 0 settles this and records it here. A hand-built window is a re-skin, and the spec says so.
`
`## Risks / Trade-offs
`
`- The layout XML may not describe behaviour (tabs, scroll, drag), so the HUD code supplies it → keep the renderer generic and put behaviour in per-window modules.
`- Some layouts depend on engine features (nine-slice frames, tiling, tinting) that need to be worked out from the data → start with chat, the simplest, and grow the renderer from what it needs.
`- Art is EA-owned → converted locally, gitignored, never committed.
`
`## Open Questions
`
`- Are the inventory/paperdoll and style-bar layouts in the install at all? Settled by task 1.
`- Atlantis or the classic skin for the look the user wants? Default is atlantis; say if you prefer another.
`
`## Spike findings (task 1.1)
`
`Where each stage's window is defined in the install:
`
`| Window | Layout | Art |
`| --- | --- | --- |
`| Chat | `ui/atlantis/chat_window.xml` (`WindowId` Chat) | atlantis atlases, `skin1/chat*.bmp` for the classic look |
`| Target | `float_target_window.xml`, `miniinfo_window.xml` | atlantis atlases; `skin1/{red,green,yellow,blue}{bg,full}.bmp` bars |
`| Group, pet, timers, compass, clock | `minigroup_window.xml`, `minipet_window.xml`, `timer_window.xml`, `compass.xml`, `clock_window.xml` | same |
`| Menu bar | `menu_bar_window.xml` | atlantis |
`| Quick bar and style bar | **not in the install** (no `WindowId` for either; the client binary holds no readable name) | `skin1/command.bmp` (132x199), `skin1/combstylesheet.bmp` (192x448), `icons/cbt_*.bmp` |
`| Inventory and paper doll | **not in the install** | `skin1/pperdoll.bmp`, `charsheet.bmp` (192x448 each), `sacks.bmp` (bag tabs), `trade.bmp` |
`| Vault, merchant | `vault_window.xml`, `merchant_window.xml` | `skin1/vault.bmp`, `merchant.bmp` |
`
`So chat, target, group and the menu bar are data-driven. The quick bar, style bar, inventory and paper doll have art but no layout; they are built by hand from the stock sheets, positioned by measuring the art. Those two windows are a re-skin, not a reuse of the original layout.
`
`Icons are not described by the `*_icons.xml` files. Those only name canvases (`itm1_s`, `combat1_b`, ...) pointing at `icons/*.bmp` files the install does not hold under those names. What the install holds is `icons/{itm,wpn,spl,skl,cbt}_N.bmp`, each 320x320 and 24-bit: a 10x10 grid of 32-pixel icons, with `N` the first icon number on the sheet. Icon `n` is therefore on sheet `floor(n/100)*100`, cell `n mod 100` (column `n mod 10`, row `(n mod 100)/10`). This is to be confirmed against a real item icon in task 2.4.
`
`## Icon numbers (task 2.4, partly settled)
`
`Converted: the 19 sheets under `icons/` (`itm`, `wpn`, `spl`, `skl`, `cbt`), 10x10 grids of 32-pixel icons, one file per hundred numbers (`uiconv` writes them as PNG and lists them in `layout.json`). Checked by eye: `wpn_900` is weapons, `cbt_500` combat styles, `spl_*` spells.
`
`**Not settled: the imported data's icon numbers do not line up with this install's sheets.**
`- Gear: the \`item\` table has no icon column; its \`model\` is the only candidate. Of 17,616 weapon and armour rows, 1,522 (9%) fall in a range the install has a sheet for; the rest run up to about 4,800, past the last sheet (about 1,899). The imported database is from a later client era than this install.
`- Styles: \`style.icon\` is 56 to 399 for all 1,199 rows, which is in the spell-sheet range, not the combat sheets (500 to 899). Whether it indexes \`cbt\` with an offset or \`spl\` directly was not confirmed by eye on the first few styles.
`
`Consequence for 4.2 and 4.3: item and style icons need a lookup that can say "no picture", with a generic frame and the name as a fallback, and the style mapping needs checking against a few known styles before it is trusted. This does not affect chat, the bars or the layout work.

Checked for styles (task 4.2): `Two Fists` has icon 105 and the first fist in the combat sheets is at 636; `Bloodletter` has icon 60 and the blood drop is at 505. No single offset fits both, so style icons cannot be placed from `style.icon` with what is in the install. The server's own bar message does not carry an icon either.
