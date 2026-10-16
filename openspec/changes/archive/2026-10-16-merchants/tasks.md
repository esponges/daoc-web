## 1. Import

- [x] 1.1 Add `classtype` and `items_list` columns to `spawn` and a `merchant_stock` table (list, item, page, slot) in `internal/importer/world.go`
- [x] 1.2 Read `Mob.ClassType` and `Mob.ItemsListTemplateID` into the spawn row; import `merchantitem`; print read and kept counts; fail when the table is missing or empty
- [x] 1.3 Importer test with a small synthetic dump: a merchant mob, a guard, two stock rows, a missing-table case (no real rows)
- [x] 1.4 Print how many merchants were kept versus dropped for lacking a converted model
- [x] 1.5 Re-import `server/data/world.db` and spot-check one well-known merchant's stock against the dump

## 2. Server rules

- [x] 2.1 Load `merchant_stock` into the catalog, ordered by page then slot; tolerate a store without the table (interaction off, log says to re-import)
- [x] 2.2 Add `Spawn.ClassType` / `ItemsList`, `npcKind`, and `NPC.Kind`; set `ik` on the NPC entity in `zone.go`
- [x] 2.3 Add `Player.shop`, an interact-range constant and a helper that checks zone, alive, kind and range
- [x] 2.4 `interact`: validate and answer with `shop` (items via `itemView` plus price); refusals as `err`
- [x] 2.5 `buy`: stock, money, pack space checks; take money and add item in one `store.Tx`; resend `inv`
- [x] 2.6 `sell`: pack slot only, not worn or priced zero; remove item and pay in one `store.Tx`; resend `inv`
- [x] 2.7 Close the shop (`shopclose`) on leaving range, death, zone change, disconnect and merchant death
- [x] 2.8 World tests: stock join with an unknown item, buy success, too poor, full pack, not in stock, walked away, sell success, sell worn, shop closes on distance

## 3. Protocol

- [x] 3.1 Add `TInteract`, `TBuy`, `TSell`, `TShop`, `TShopClose` and their structs and `ik` on `Entity` in `internal/server/protocol/protocol.go`
- [x] 3.2 Route the three requests in `internal/server/gameserver/server.go`
- [x] 3.3 Document messages and fields in `docs/protocol.md`; add round-trip tests and a test that an unknown-message-tolerant client still decodes `ik`

## 4. Client

- [x] 4.1 `web/net.js`: `interact`, `buy`, `sell`; keep the current shop from `shop`/`shopclose`; node tests in `net.test.mjs`
- [x] 4.2 `web/viewer.js`: a right-click with `moved <= 4` picks like left-click; a merchant (`ik`) is targeted and interacted with; a right-drag is unchanged; offline mode sends nothing
- [x] 4.3 New `web/shop.js` with pure row and height helpers and `createStockShop` over the stock frame, built on `web/loot.js`'s pattern; test the helpers in `web/shop.test.mjs`
- [x] 4.4 Plain-panel fallback for when `web/data/ui/` is absent
- [x] 4.5 Selling: a click on a pack row while the shop is open sends `sell`; the window shows the player's money
- [x] 4.6 Chat feedback for `err` replies (too far, cannot afford, pack full)

## 5. Verify

- [x] 5.1 `go test -short ./...` and every `web/*.test.mjs` pass
- [x] 5.2 Restart `gameserver`, find a merchant in a starter zone, right-click it, buy an item, sell one, walk away and see the window close (done in the browser against Myrkwood's weapon stall; no screenshots are kept in the repo, since they show EA's art)
- [x] 5.3 Move `web/data/ui/` aside and confirm the plain fallback works
- [x] 5.4 Settle the interact range and sell ratio from play; update `design.md` if they change
- [x] 5.5 `openspec validate merchants`
