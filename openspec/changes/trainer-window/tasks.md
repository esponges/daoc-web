# Tasks

## 1. Server and protocol

- [x] 1.1 `trainer` gains `say` (greeting on first open only) and per-spec `gives` (`lv`, `nm`, `k`), documented in `docs/protocol.md` with a protocol test.
- [x] 1.2 Greeting text in `openTrainer`, with and without unspent points.
- [x] 1.3 Refusal names the class's and base class's trainers (from `rules.Classes`, the pair `TaughtBy` accepts).
- [x] 1.4 Level-up points line adds "Visit a <class> trainer to spend them."
- [x] 1.5 `gives` from the catalog functions the sheet uses (styles by spec and level, `spec_ability`, spell lines), up to five above the current level.
- [x] 1.6 Server tests: Savage at Savage trainer (greeting, six specs, gives for Hand to Hand), Savage at Healer trainer (refusal text), level 6 line, re-sent offer without `say`.

## 2. Client

- [x] 2.1 `trainerView`: two columns, specific grey reasons, next unlock in the tooltip; node tests.
- [x] 2.2 `trainwin.js`: header "N points", double-click trains, TRAIN/INFO messages, INFO enabled and writing `gives` to the log; node tests.
- [x] 2.3 Second list column in `uictl` if the list does not already draw `Columns` 2.
- [x] 2.4 `ui.js`: greeting in the talk window beside the Train window; plain fallback panel with greeting, columns, messages and INFO.
- [x] 2.5 `trainclose` from the server (tick check like `checkShop`); the client closes the Train and talk windows.

## 3. Check

- [x] 3.1 `go test -short ./...`, every `web/*.test.mjs`, `openspec validate trainer-window`.
- [ ] 3.2 In the browser, as a Savage: right-click a Viking trainer in Mularn and a Savage trainer, train Hand to Hand by TRAIN and by double-click, INFO, the wrong trainer, walking away. Then the fallback with `web/data/ui/` moved aside.
