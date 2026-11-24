# Proposal: stock-delve-pack-and-stick

## Why

A player listed these on 2026-11-24. Checked against the code and the install:

- **NPCs do not turn to you.** `Zone.Interact` (`world/shop.go`) opens the NPC's window but leaves its
  facing alone. NPCs already turn smoothly (`turnToward`), so it is a missing call, not missing machinery.
- **No delve.** Nothing shows an item's or spell's details on demand; the pack only has a hover title.
  The client binary's own strings give the stock behaviour: "Right-click for more information." on
  spells and abilities, "Shift+Right-click to remove." on quickbar slots, an "Information/Delve" key
  action and `&delve`. Here a plain right-click on a quickbar or spell bar slot clears it, which is
  the stock Shift+Right-click.
- **Gear does not show its con colour.** Items carry a level (`lv`), but nothing colours an item by it.
- **One picture for many items.** `itemIconNumber` (`web/uiwin.js`) knows three weapon lines (Sword,
  Hammer, Axe); spears, staves, bows, thrown weapons and hand-to-hand all draw the hammer, and arrows,
  potions, poisons and everything else draw one generic icon. The imported icon numbers do not match
  the install's sheets (stock-ui-followups design), so the fix is a hand-picked cell per item kind, not
  an offset.
- **The pack is the wrong shape.** The paper doll art (`skin1/pperdoll.bmp`) draws a pack page as one
  column of eight rows: a small icon box (about 16 px, at x 12, rows 17 px apart from y 278) with room
  for the item's name to its right. `bagCell` (`web/pack.js`) lays 32 px icons four across and two
  down over that art, with no names.
- **No stick.** `/face` works; the Command window's Stick says "not built yet".

## What Changes

- An NPC turns to face a player who interacts with it, holds that while talked to, and turns back to
  its spawn heading after a quiet while.
- Right-click on an item (pack, paper doll, shop, loot, vault, quickbar) or a spell or style (spell bar,
  spells and styles windows, quickbar) delves it: its details written to the chat log, the way the stock
  client does. Gear names are drawn in their con colour to the player. Shift+Right-click clears a
  quickbar or spell bar slot; plain right-click no longer does.
- The server sends what a delve needs (`delve` request, `delve` reply): for gear the level, DPS and
  speed or AF, damage type, bonuses, quality and condition; for a spell its level, cost, range, cast
  time, recast, duration and effect. Additive protocol.
- Item icons chosen per kind from the install's sheets: every Midgard weapon line, bows, arrows and
  quivers, shields by size, armour by slot, potions, poisons, gems, crafting materials, money and
  quest items, with a contact sheet so the picks are checked by eye.
- The pack drawn as the stock list: eight rows per page, small icon, name, count.
- `/stick` and the Command window's Stick: keep within melee range of the target, facing it, until a
  movement key, the target dying or leaving range, or the command again.

## Capabilities

### Modified Capabilities

- `npc-interaction`: the NPC turns to the player.
- `stock-ui-windows`: delve, con-coloured gear, Shift+Right-click to remove, icons by kind, the pack list.
- `friends-and-social`: stick.

## Impact

`internal/server/world` (interact, delve, stick), `internal/server/protocol`, `docs/protocol.md`,
`web/pack.js`, `web/uiwin.js`, `web/quickbar.js`, `web/spells.js`, `web/ui.js`, `web/command.js`, tests.
