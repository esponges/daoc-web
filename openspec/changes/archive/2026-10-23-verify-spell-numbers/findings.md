# Findings

Read 2026-10-22 from DOLSharp (`Dawn-of-Light/DOLSharp`, commit 776619f) and OpenDAoC-Core, cloned into the session scratchpad and not into this repo. Formulas below are stated in our own words from their shape; no code or identifiers are carried over. File names are for finding them again.

## 1.1 What the community servers do

### Spell damage (`spell.damage`) — DOL `spells/SpellHandler.cs`, OpenDAoC `spells/SpellHandler.cs`

- **Base** is the spell's `Damage` times `(casting stat + 200) / 275`. The casting stat is the class's mana stat (Intelligence, Piety, Acuity per class), as modified by items and buffs. A caster with 100 in the stat does 1.09 times the base; with 200, 1.45 times.
- **Variance is a range whose low end depends on spec.** The low end starts at 0.25 plus the caster's spec level in that line (minus 1) divided by the target's level; the high end is 1.25. A spec level equal to the target's level gives a low end of about 1.25, so a fully specced caster does not vary. Over-spec adds a little to both ends. A level-difference correction then shifts both ends, and the ends are clamped to a minimum. (DOL: 0.5% per level of over-spec and a per-level factor from the spell; OpenDAoC: a single level-difference term that is 10% a level at level 0 and 2% from level 50.)
- **Cap** is three times the spell's base damage, scaled by effectiveness.
- Resists apply as a percent off, then the cap, then crits. Both servers apply a hit-chance adjustment; ours has a separate hit roll.
- **This server today** uses a flat 0.75 to 1.25 with no stat and no spec level, so it differs in shape and in level (no `(stat+200)/275`).

### Heals — DOL `spells/HealSpellHandler.cs`

- Heals are drawn from a range: high end is 1.25 times `Value`; low end is `0.25 + (spec level − 1) / spell level`, capped at 1.25 (and at least 1). Players only; creatures always use the 1.25 end. A diseased target gets half. There is no casting-stat factor on heals in this file.
- **This server today** restores exactly `Value`. It is the mean of the community range for a caster whose spec equals the spell's level.

### Power pool (`power.pool`) — DOL `gameobjects/GamePlayer.cs`

- Pool is the larger of 5 and `level × 5 + (casting stat − 50)`. A level 50 caster with 100 in the stat has 300. The camelotherald article the code cites is about Vampiirs only.
- **This server today** is `10 + level × (5 + stat/50)`. With a stat of 100 that is 17 at level 1 (DOL: 55) and 360 at level 50 (DOL: 300), so it is too small early and too large late, and the stat enters per level instead of as a flat addition.

### Power regeneration (`power.regen`) — DOL `propertycalc/PowerRegenerationRateCalculator.cs`, OpenDAoC says it follows DOL

- A tick every 3 seconds returns `5 + level / 2.75` power, plus buffs; at least 1. The file's header says patch 1.87 made combat ticks twice as effective and twice as frequent, so pre-1.87 (our target) combat regen was lower; the file does not say by how much.
- **This server today** returns a fiftieth of the pool out of combat and a two-hundredth in it.

### Health regeneration (`character.health-regen`) — DOL `propertycalc/HealthRegenerationRateCalculator.cs`

- A tick every 3 seconds returns `10 + level × 0.2` below level 26 and `level × 0.6` from 26; the same 1.87 note applies; none while diseased. OpenDAoC copies the same numbers and says so.

### Power cost — DOL `spells/SpellHandler.cs` (`PowerCost`)

- A spell's `Power` that is **negative** is a percentage of the caster's pool; one that is **positive** is a fixed amount. DOL then multiplies by 1.2 and reduces the cost by a focus bonus.
- The imported table has 907 spells with positive power, 164 negative and 620 zero. At level 1 the class spells in the table cost 1 to 3, which only makes sense as fixed amounts.
- **This server today reads it the other way round** (`internal/server/world/pcasting.go`, `powerCost`): negative as a fixed amount, positive as a percent of the pool. At level 1 that costs 1 power (a fiftieth of the pool, minimum 1) where DOL would charge 2 to 3, and at level 50 a 20-power spell costs 20 percent of the pool where DOL charges 20 points (about 7 percent of a 300 pool). This is a bug in the wiring, not a formula choice. Task 2.2 corrects it.

### Buff caps (`spell.buff`) — DOL `propertycalc/StatCalculator.cs`

- A player's stat from spell buffs is capped at `level × 1.25` (our `BuffStatCap` matches); specialisation buffs have a separate cap of `level × 1.5 × 1.25`. No source for the armour-factor cap (`BuffAFCap`) was read; it stays provisional.

### Creature heal decision (`cast.heal`)

- Not a formula either server states; it is this project's choice of when a creature heals. It stays provisional and is not part of the data check.

## 1.2 Published 1.65 accounts (2026-10-22, standard searches)

- **Damage and heal spec dependence: qualitative confirmation, no numbers.** Mythic's Camelot Grab Bag Q&A (reposted at gamebanshee.com/kpxwh and gamebanshee.com/kpxd6) says spec points above a spell's level give a direct-damage spell its full effect with damage starting at +25 percent, and that heal variance is wider when the spec is below the spell's level and narrows as it approaches it. That agrees with the community servers' shape (the low end of the range rises with spec until it meets the high end) but gives no figures. The rules stay at reference for the numbers and cite the Q&A for the shape.
- **Casting-stat factor `(stat + 200) / 275`: no published 1.65 source found.** Query: spell damage formula, casting stat, spec level, variance. Reference only.
- **Power pool and regeneration: no formula found.** Query: power pool formula, casting stat, regeneration per tick before 1.87. The only hit is the 1.66h patch note that items raising a casting stat's cap also raise the pool, which confirms the link to the casting stat and nothing more. Both stay reference (a community server's numbers), with the 1.87 note recorded: pre-1.87 combat regeneration was lower than the community servers' by an unknown amount.
- **Health regeneration, buff caps, armour-factor cap: not searched separately**; the first two follow the community servers (reference); the armour-factor cap stays provisional.

## What the check and the game showed (2026-10-22)

- **Percent heals healed nothing.** 14 Healer and Seer heals (for example Gift of the Heavens, value −40) carry a negative value, which both community servers read as that percent of the caster's maximum health; the server clamped it to zero. Found by the data check; fixed (`rules.PercentHeal`). A Healer at 720 health casting a −40 heal now restores 288.
- **Spell power was read the wrong way round.** Positive power in the data is a fixed cost and negative a percent of the pool; the server did the reverse, so every class spell (907 positive-power spells) cost a percent of the pool. Fixed (`rules.SpellPowerCost`).
- **Pool and regeneration** now follow the community servers: the pool is five a level plus the casting stat above 50, and a tick every 3 seconds returns 5 + level / 2.75. Seen in the game: a level 20 Dwarf Healer with 85 piety showed a pool of 135 (100 + 35).
- **Player damage and heals** now scale with the casting stat and the spec in the spell's line. Not covered: damage over time, damage add, buffs, creature spells (kept as they were; their sources were not read for those), the focus bonus on power cost, and the level-difference term the two servers disagree on (we follow DOL's 2 percent a level).
- **Seen in the game (real server on a scratch port, driven from the browser, not the rendered UI):** a Dwarf Healer at level 20 casting Minor Mending (value 6, spell level 1) 11 times at spec 1 restored 3, 4, 2, 4, 3, 6, 7, 7, 3, 4, 1 (the range for spec 1 is 1 to 7, where the old flat rule always gave 6); after training Mending to 20, five casts restored 7 each (range 7 to 7). Each cast cost 1 power (the fixed cost from the data). No damage spell was cast in the game: a Healer has none, and the Runemaster path is covered by `TestSpellsScaleWithSpec`.
- **Limit of the data check:** `TestSpellAmountsAgreeWithTheTable` runs the rules' functions over every imported spell of an implemented type (1,556 spells); it does not run the world's casting path, which the world tests cover.
- **Still provisional:** `spell.buff` (armour factor cap), `cast.heal` (the creature's choice to heal), power regeneration in combat (the 1.87 note), and everything about damage over time and damage add scaling.
