## Context

Today the server's only idea of an NPC's role is `Friendly` (`spawn.realm != 0`), so a merchant, a guard and a villager are the same. The importer drops `Mob.ClassType` and `Mob.ItemsListTemplateID`. In the community dump, 1,218 mobs are class `DOL.GS.GameMerchant`, and `merchantitem` maps a list id to item ids with a page and slot. The character already has `Money` (int64, saved with the character), a 40-slot pack, an `inv` message, and `world.addItem`, which saves item and character in one transaction. Right-click is currently camera drag; `viewer.js` already separates a click from a drag with a 4 px threshold.

## Goals / Non-Goals

**Goals:** a right-click on a merchant opens a window of its real stock; buying and selling work and are decided by the server; the foundation (kind, `interact`, range) is reusable for later NPC kinds.

**Non-Goals:** quests, trainers, healers, teleporters, vaults, stable masters; buy-back, repair, quantities, restock or sell-out; auto-walk to the NPC; price changes by realm or faction.

## Decisions

**Kind from class type, as a small map.** `npcKind(classType string) string` returns `"merchant"` for `DOL.GS.GameMerchant` and `""` otherwise. Later kinds are one more map entry plus a handler, not a new mechanism. Alternative considered: flag bits in `Mob.Flags` — rejected, they do not mark merchants. Subclasses with their own behaviour (`Blacksmith`, `GameHealer`) stay `""` until someone builds them, so nobody opens a shop that would be wrong.

**Stock is a join done at load, held per list.** The catalog loads `merchantitem` rows into `map[listID][]StockRow` ordered by page then slot, resolved against `Items` when a shop is built. Unknown items are skipped and counted. Alternative: query SQLite per `interact` — rejected for consistency with how the catalog already caches items and loot.

**The server remembers the open shop.** `Player.shop` holds the NPC id. `buy` and `sell` carry only an item id or pack slot, never a price, and the server re-checks the shop is open, the NPC alive and in range on every request. This keeps the "server decides everything" rule and means a stale client cannot trade with a merchant it walked away from.

**Range is checked on the server's own positions** with one constant (assumed 128 units horizontally plus the merchant's body reach, to match the game's interact distance; tunable in one place). Leaving range is noticed in the per-tick pass that already visits players, which sends `shopclose`. The client does not walk to the NPC; it prints "You are too far away to interact" from the server's `err`.

**Money is copper in an int64; price is `Item.Price`.** The importer already stores `Price` from the table. Selling pays `price / 2`, floor, minimum 1 when price > 0 (assumption: half is the classic ratio; one constant). Sale and payment, and purchase and payment, each go through one `store.Tx` like `addItem`, so a crash cannot leave an item without its money or the reverse.

**Protocol is additive.**

```
client -> server   interact {id}
                   buy      {id: <item id>}            one item at a time
                   sell     {slot}                     pack slot 0-39
server -> client   shop     {npc, nm, items:[{id, nm, k, line, w, d, pr, pg, sl}]}
                   shopclose{}
entity (optional)  ik: "merchant"
```

`shop` items reuse `itemView` (so `k`, `line`, `w`, `d` match `inv`) plus `pr` (price in copper). The client draws pictures by kind and weapon line, never by icon number (see `stock-ui-followups/design.md`).

**Client window.** `web/shop.js` follows `web/loot.js`: pure row/height helpers that tests can import, and a `createStockShop(doc, data, hooks)` that draws a stock frame with rows of picture, name, detail and price. The install has art for a merchant window but no layout for it (the converter writes `skin1` merchant art as PNG), so the window is composed like loot and the inventory are, not from a layout file. Without `web/data/ui/` the same rows go in a plain panel, in the style the built-in panels already use. Selling is a click on a pack row in the inventory window while `shop` is open, which needs one hook where the inventory handles clicks.

**Right-click.** In `mouseup`, a button-2 release with `moved <= 4` runs the same ray pick as left-click. A hit with a merchant `ik` calls `combat.select(hit)` then `net.interact(hit.id)`. `downAt` is currently set only for button 0; it is set for button 2 too, and the drag flag is untouched, so camera behaviour is unchanged.

## Risks / Trade-offs

- **[A merchant spawn's model may not be converted]** The importer keeps only spawns with a converted model, so some merchants will be missing from the world. Mitigation: the census count of merchants kept versus all is printed by the importer; no new conversion work here.
- **[Stock lists may name items the item table lacks]** Mitigation: skipped and counted; an empty shop is valid.
- **[Item prices in the dump may be zero for many items]** Items at zero are shown with a price of zero and are buyable for free; this is data, not a bug, but a review during apply should check how common it is and decide whether to hide them.
- **[`world.db` out of date]** A store without the new columns/tables must not crash the server; interaction is disabled and the log says to re-import.
- **[Windows work on this machine]** Per CLAUDE.md, the server must be restarted after server changes, and UI changes need a look in the browser (`/uidemo.html` plus a real merchant), not only node tests.

## Findings from trying it

Tried in the browser at Myrkwood's weapon stall (Stein, a merchant with 24+ items):

- A right-click from about 300 units away targets the merchant and the server answers "too far away"; after a short walk it opens the shop. Range stays at 128 plus half the merchant's reach: it feels right and needs no change.
- Selling pays half the price (a 30c shield paid 15c); buying took the exact price; "cannot afford" is reported in chat; walking away closes the window. The half ratio stays.
- The stock window scrolls, shows unaffordable prices in red and brings the pack up beside it.
- **Found and fixed:** in the plain fallback the shop panel was centred and grew wide enough to cover the pack's first column, so those cells could not be clicked to sell. The plain panel is now anchored left with a width that leaves the pack clear (`web/game.css`).
- A sale pays for the whole stack (price x count, halved), not one item.
- Dev tip: GM `give` needs an admin account; without one, selling unequipped starter gear is a way to get coins for testing.

## Open Questions

- Whether a shop should stay open on zone-boundary teleports is moot (changing zone closes it).
- Zero-priced items (747 of the 48,448 stock rows) are shown and can be bought free; whether to hide them is left for later.
