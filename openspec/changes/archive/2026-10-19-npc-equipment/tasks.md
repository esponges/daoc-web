## 1. Spike: can model numbers be resolved?

- [x] 1.1 Pick five well-known NPCs (a guard, a smith, a caster, a monster with a weapon, a robed NPC) and list their `npcequipment` rows
- [x] 1.2 Compare those model and extension values against shapes and textures in the install; record in design.md whether numbers resolve, partly resolve, or must use the by-kind fallback
- [x] 1.3 Inspect a humanoid NPC model's shapes for head selection and write the faces/heads finding into design.md
- [x] 1.4 Find how colour values map to tints (check `color.dat`) and record it

## 2. Import

- [x] 2.1 Add an `npc_equip` table and an `equip` column on `spawn` in `internal/importer/world.go`
- [x] 2.2 Import `npcequipment` rows and link spawns from `mob.EquipmentTemplateID`, falling back to `npctemplate`
- [x] 2.3 Print read/kept counts and fail when the table is missing or empty; add an importer test with a small synthetic dump (no real rows)

## 3. Server and protocol

- [x] 3.1 Load and cache equipment per template in `internal/server/world/catalog.go`; attach to `NPC`
- [x] 3.2 Add the `eq` field to `protocol.Entity`, send it in `npcEntity`, document in `docs/protocol.md`
- [x] 3.3 Add protocol round-trip and world tests: geared NPC, ungeared NPC, unknown template

## 4. Converter and client

Tasks 4.1 to 4.4 (the lookup, reading `eq`, drawing gear, the skipped count) moved to the `npc-gear-mapping` change, which builds them and records why the mapping is approximate. Only faces remain here.

- [ ] 4.5 NPC faces and heads: the head is chosen by the creature's `monsters.csv` row, not by equipment. The converted head texture is `hair_*.png`, so the missing face is probably a skin-mapping problem in the converter; not investigated yet

## 5. Verify

- [x] 5.1 `go test -short ./...` and all `web/*.test.mjs` pass
- [ ] 5.2 Look at guards, a smith and a monster in the browser; check with the `web/data/` lookup moved away that NPCs still draw plain
- [x] 5.3 Update README (where equipment comes from, what resolves) and CLAUDE.md sources if a new source is used (README done in npc-gear-mapping; no new source needed)

## Closed

Archived on 2026-10-19 with the unchecked tasks above moved to `midgard-playable` (section "Carried over").
