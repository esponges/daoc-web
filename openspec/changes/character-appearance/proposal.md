# Proposal: character-appearance

## Why

A new character picks a class, a race and a gender, and nothing else. The 1.65 creation screen also chose:

- the face, from several per race and gender;
- the hair style and hair colour;
- the eye colour (from 1.54);
- the starting stat points: 30 to distribute, with costs rising above +10 to a stat;
- the size, in later patches; task 1.1 checks whether 1.65 had it.

The README lists "Faces, not hair" and "Character options are thin" as known limitations. A humanoid's hair is a separate style number that is not mapped, so every head shows only what its face texture paints. This hits creatures as well: every humanoid NPC is bald under its face skin.

## What Changes

- **Hair is mapped.** The hair style number becomes a hair mesh or texture for each race and gender, found in the install (`charconv`). Hair colour is a skin tint, which `charconv` already handles for creatures. Both players and humanoid NPCs wear hair.
- **Faces:** the face list per race and gender, from the install.
- **The creation screen** offers face, hair style, hair colour and eye colour, if the install has eyes as a separate texture.
- **Starting stat points:**
  - 30 points to distribute at creation;
  - the costs follow the 1.65 creation rules: a point costs 1 up to +10, 2 up to +15, and 3 above;
  - the server checks the distribution.

  Today every character gets the race and class base with no choice. This is a rule-book change to `character.creation`.
- **The pregame stock windows,** if `uiconv -skin pregame` gives the character creation layout. The plain creation page is the fallback.

## Capabilities

### New Capabilities
- (none)

### Modified Capabilities
- `player-appearance`: face, hair style, hair colour and eye colour at creation, shown to every viewer.
- `race-visuals`: hair meshes and textures per race and gender, also for humanoid creatures.
- `rules-character-stats`: the 30 starting points and their costs.

## Impact

- **Converters:** `cmd/charconv` maps hair. `docs/formats.md` gets a "Hair" section.
- **Data:** `players.db` adds columns for face, hair, hair colour, eyes and stat points (a migration). Existing characters keep their defaults.
- **Server:** creation validation and the stat rules.
- **Protocol (additive):** appearance fields on entities.
- **Client:** the creation page in `login.js` and `character.js`; `race.js` dresses heads.
