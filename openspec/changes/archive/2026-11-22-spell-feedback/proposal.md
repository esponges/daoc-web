# Proposal: spell-feedback

## Why

Casting and pet fighting give almost no feedback in play.

- **No cast animation:** a caster stands idle through the whole cast. The viewer never plays a cast animation on the player's own character, and no converted figure has a cast clip. The converter learned the clip after the models were converted.
- **No effect on screen:** a spell landing only writes a log line. Nothing is drawn on the caster, nothing flies, and nothing shows on the target. Arrows don't fly either.
- **Pet fights are silent:** the log only prints hits the player deals or takes. In a browser check, a summoned pet took a svartalf guard from full health to dead without one line.
- **Effects are hard to read:** the player's own effects are small letter badges by the clock. A pet's effects aren't shown at all.
- **Double cast:** double-clicking a spell sends the cast twice, so every cast also prints "You are already casting a spell."

The install has what we need to fix this faithfully:
- `gamedata.mpk` holds the client's own spell table, `spells.csv`. For each spell it gives the three caster animations, the caster, projectile and target effects, and the icon.
- The effect table `speffects.csv` gives each effect's NIF, the bone it attaches to, and its duration and rate.
- The NIF list `spnifs.csv` names the effect models.
- `effects/` holds about 1,600 effect NIFs.
- The community spell dump carries each spell's `ClientEffect`, which is the row in `spells.csv`. For example, Minor Dusk Dart is 2501: animations 41/42/43, caster effects 160/161, target effect 256.

## What Changes

- **Import:** the importer keeps each spell's `ClientEffect`. The server sends it on `cast`, `spell`, `effect` and `resist` events as a new field. That's an additive protocol change.
- **Conversion:** a new converter reads `spells.csv`, `speffects.csv` and `spnifs.csv` and converts the effect NIFs the imported spells use into web data. This reuses the NIF reader and the particle emitters `propconv` already supports. Effects whose nodes it can't draw are listed, not guessed.
- **Spells on screen:**
  - a cast plays the spell's caster animations: start, a held loop, then the release;
  - the caster's effects run during the cast;
  - a bolt flies from caster to target;
  - the target effect plays where the spell lands, on the bone the table names, for its duration.
  - This holds for the player, other players and creatures alike.
- **Arrows:** an arrow flies from the archer to the target.
- **Cast animation on the player:** the player's own character plays the cast clips, which aren't played today. The character animations are converted again so every figure has them.
- **Pet log:** the log gets the pet's hits, misses and spells, and the blows the pet takes, in the game's wording ("Your returned commander hits the svartalf guard for 23 damage!"). The client tells its pet apart by the pet message's id, so the server is unchanged for this.
- **Pet effects:** the pet's effects show as icons in the pet window, which the stock layout gives six slots. The `pet` message gains the pet's effects list (additive).
- **Own effects row:** the player's effects show as a row of icons with time left, placed so no other window covers it. Icons are chosen by spell kind, because most icon numbers in `spells.csv` (800–3700) are on sheets this install doesn't ship (it has 0–499 and 1800).
- **Double cast fix:** a double-click or a held key asks for one cast.

## Capabilities

### New Capabilities
- `spell-visuals`: converting the client's spell effects and animations from the install, and playing them on casters, projectiles and targets in the view.

### Modified Capabilities
- `game-messages`: the pet's combat and the blows it takes appear in the log.
- `world-data-import`: spells keep their client effect.
- `player-spellcasting`: one cast is asked for per press, and the player's own character plays the cast animations.
- `pets`: the pet's effects are sent to its owner and shown in the pet window.
- `game-hud`: the player's own effects row.

## Impact

- **Importer:** `internal/importer` adds a `client_effect` column to `spell`, so a re-import is needed. The world store migration follows the existing pattern.
- **Server:** `world` puts the client effect on spell events and the pet's effects on the pet message. `protocol` adds the new fields. `docs/protocol.md` documents them.
- **Converters:** a new `cmd/effectconv`; `cmd/animconv` gets the cast clips by the table's animation ids and re-runs over all figures.
- **Client:** a new `web/spellfx.js` draws effects, bolts and arrows on the existing particle and model paths. `viewer.js` and `remote.js` play cast clips and send events to `spellfx`. `effects.js` gets the pet log lines and the effects row. `spells.js` and `ui.js` get the one-cast guard. `pet.js` shows effect icons.
- **Data:** converted effects go to `web/data/effects/` (gitignored). Nothing converted is committed.
- **Order:** independent of `stock-windows`. The pet effect icons go into whichever pet window exists when this lands.
