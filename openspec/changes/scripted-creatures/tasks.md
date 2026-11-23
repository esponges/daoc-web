# Tasks

## 1. Research

- [ ] 1.1 For each unimplemented Midgard kind, read its rows in the community data (templates, spawn points, linked adds). Write a one-line behaviour summary per kind in `design.md`.
- [ ] 1.2 Find a public source for each named creature's behaviour (bestiary pages, Herald) and record it. Read the reference server only for understanding.
- [ ] 1.3 Decide whether Enchanter, NameRegistrar, RoyalTreasuryClerk and Recharger are classic 1.65, and write the reasons.

## 2. Behaviour layer

- [ ] 2.1 `world/behaviours.go`: the behaviour table and the brain hooks for presence, adds, flee, call-for-help, timed abilities and cycle. Table tests per spec scenario.
- [ ] 2.2 Fold NightSpawn and DaySpawn into presence, and check the existing night-only creatures still behave.

## 3. Kinds

- [ ] 3.1 Myling, ShiveringPresence, RainbowSprite and OonaUndeadSoldier.
- [ ] 3.2 Named bosses with adds: QueenMajor, Stripe, ZritZrit, Snarls, Oona.
- [ ] 3.3 The remaining named creatures, each sourced or marked behaviour unsourced.
- [ ] 3.4 The Recharger, and whichever of the other services task 1.3 keeps.

## 4. Import and census

- [ ] 4.1 Importer: keep the kind and the parameters it can read. Re-import `world.db`. Add a test that no Midgard placement is dropped without a reason.
- [ ] 4.2 `cmd/parity`: the new statuses (behaviour unsourced, owned by a change). Regenerate the docs and the README table.

## 5. Check

- [ ] 5.1 Convert the models these kinds need, then visit a Myling camp and a boss camp in the browser and fight through an add wave.
