# Tasks

## 1. Format work

- [x] 1.1 Find the hair meshes and textures, the face list and the eye textures per race and gender in the install (D1). Write them up in `docs/formats.md`. Check whether 1.65 had size at creation.
- [x] 1.2 `charconv`: convert hair per race and gender, and map NPC hair style and colour. Add a census line for converted hair.

## 2. Rules and server

- [x] 2.1 A source for the stat point costs. Rule-book update to `character.creation`, with tests and a registry entry.
- [x] 2.2 Store migration (D2, D3). Creation validation for appearance and points. Entity appearance fields. Protocol additions and a test.

## 3. Client

- [x] 3.1 `race.js` and `character.js` draw hair and face from the appearance record; helms hide hair.
- [x] 3.2 Creation pickers and stat distribution with a live preview (D4), with the plain page as fallback.
- [x] 3.3 Check in the browser: create characters of each race, and look at guards and townsfolk. Update the README known limitations.

## Outcome

- Done: hair split from the head and drawn with the row's hair skin (NPCs and players); face (1-6) and hair colour (10; trolls 4) chosen at creation, stored, validated, sent as `fc`/`hc` and drawn for you and others; creation page pickers with a texture preview.
- Not done, recorded in the README: hair style (one head mesh per race is converted; the other nine need the `HeadB*` shapes converted and a picker), eye colour (painted into the face skin; nothing to choose), pregame stock creation window (the plain page is used), size.
- The 30 points and their costs were already in the rule book (`rules.RaiseCost`); no 1.65 source for the 1/2/3 tiers was found, so they stay provisional.
