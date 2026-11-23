# Tasks

## 1. Format work

- [ ] 1.1 Find the hair meshes and textures, the face list and the eye textures per race and gender in the install (D1). Write them up in `docs/formats.md`. Check whether 1.65 had size at creation.
- [ ] 1.2 `charconv`: convert hair per race and gender, and map NPC hair style and colour. Add a census line for converted hair.

## 2. Rules and server

- [ ] 2.1 A source for the stat point costs. Rule-book update to `character.creation`, with tests and a registry entry.
- [ ] 2.2 Store migration (D2, D3). Creation validation for appearance and points. Entity appearance fields. Protocol additions and a test.

## 3. Client

- [ ] 3.1 `race.js` and `character.js` draw hair and face from the appearance record; helms hide hair.
- [ ] 3.2 Creation pickers and stat distribution with a live preview (D4), with the plain page as fallback.
- [ ] 3.3 Check in the browser: create characters of each race, and look at guards and townsfolk. Update the README known limitations.
