## 1. NPC facing

- [x] 1.1 `Interact` turns the NPC toward the player (all kinds, also quest and task touches); it holds while talked to and returns to its spawn heading after 15 s quiet. Server test.

## 2. The pack list

- [x] 2.1 Measure the eight rows of `pperdoll.bmp` exactly; `bagCell` returns a row; the cell draws a small icon, the name and the count. Update `pack.test.mjs`.
- [x] 2.2 Drag, click, Shift-click and the sack tabs unchanged; check in the browser, and the fallback pack.

## 3. Delve

- [x] 3.1 `delve` request (pack or worn slot, shop or loot item, spell or style id) and reply; server builds the lines from the catalog; `docs/protocol.md`; protocol and server tests.
- [x] 3.2 Right-click delves in the pack, paper doll, shop, loot, vault, spell bar, quickbar and the spells and styles windows; the reply goes to the chat log with the name in its con colour for gear. Node tests of the pure part.
- [x] 3.3 Shift+Right-click clears a quickbar or spell bar slot; plain right-click no longer does. Tooltips say "Right-click for more information." and "Shift+Right-click to remove."

## 4. Item icons

- [x] 4.1 A contact sheet of the install's item and combat icon sheets (scratch, never committed); pick a cell per kind.
- [x] 4.2 The server sends the item's object type where `line` is not enough (bows, arrows, thrown, staves, spears, potions, poisons, materials); `itemIconNumber` maps kind to the picked cell. Tests; look at the pack and a shop in the browser.

## 5. Stick

- [x] 5.1 `/stick` and the Command window's Stick on the server, sharing Follow's steering; ends on a movement key, the target dying or going beyond range, or the command again. Server test.

## 6. Check

- [ ] 6.1 `go test -short ./...`, every `web/*.test.mjs`, `openspec validate stock-delve-pack-and-stick`; play it through in the browser.
