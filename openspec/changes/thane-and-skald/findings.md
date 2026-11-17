# Findings

Read 2026-11-16 from DOLSharp (`Dawn-of-Light/DOLSharp`) and OpenDAoC-Core, shallow clones in the session scratchpad, not in this repo. Formulas are restated in our own words from their shape; no code or identifiers are carried over. File names are for finding them again (under `GameServer/`).

## 1.1 What the community servers do

### Class numbers — `playerclasses/midgard/ClassThane.cs`, `ClassSkald.cs` (OpenDAoC-Core)

- DOLSharp keeps class numbers in its database, not in code, and the community dump has no class table, so OpenDAoC-Core is the only source for them.
- **Thane:** hybrid on the Viking base. Primary strength, secondary piety, tertiary constitution; the power pool comes from piety. Base hit points 720, weapon skill base 360, spec points per level 2.0 (multiplier 20). Races Dwarf, Norseman, Troll.
- **Skald:** hybrid on the Viking base. **Primary charisma, secondary strength, tertiary constitution**; the pool comes from charisma. Base hit points 760, weapon skill base 380, spec points per level 1.5 (multiplier 15). Races Dwarf, Kobold, Norseman, Troll.
- This corrects the proposal's first draft: it had strength first for the Skald and a multiplier of 15 for the Thane. The data's start table agrees with the races above.
- Status: provisional (one source, no 1.65 account).

### Area spells — `spells/DirectDamageSpellHandler.cs`, `spells/SpellHandler.cs`

- **Who is hit.** An enemy spell with a radius hits every enemy that may be attacked inside the radius: living creatures and, where it is allowed, players. The circle is centred on the caster when the spell's range is 0 (Thunder Shout) and on the target otherwise (Mjollnir). A range-0 spell needs no target. Both projects agree.
- **Line of sight.** DOLSharp checks line of sight to each victim before damage for a range-0 area spell. This server has no line-of-sight rule yet (README), so none is applied.
- **Falloff: the two disagree.** DOLSharp has none; each victim takes the same roll wherever it stands. OpenDAoC scales the damage of an area spell with a radius by one minus the distance over the radius, so it falls to nothing at the edge. We follow DOLSharp and record the gap: it is the older design, and no 1.65 account of a falloff was found (see 1.2).
- Status: provisional.

### Pulsing songs and chants — `effects/PulsingSpellEffect.cs`, `spells/SpellHandler.cs`, `ECS-Components/EffectListComponent.cs`

- **Units.** The `Frequency` column is tenths of a second (OpenDAoC multiplies it by 100 milliseconds), so the Skald's value of 50 is a pulse every 5 seconds. The Herald lists the live Song of Travel as a pulse every 6 seconds with a duration of 8; this data differs and the data is followed.
- **Toggle.** Casting a pulsing spell that is already on turns it off ("You cancel your effect"). Both projects.
- **Each pulse** re-casts the spell on its targets. A caster who is dead has it cancelled; a caster who is held (mesmerized, stunned) skips that pulse and keeps the song. If the caster lacks the power for the pulse the song is cancelled; the Skald's songs and chants have no power cost (`PulsePower` 0 in the data).
- **How many at once: the two disagree.** OpenDAoC ends every other pulsing spell when a new one starts (the one exception is a focus damage-shield group that does not apply to Midgard). DOLSharp cancels only a running pulse of the same spell type, so different types could run together, limited by concentration. The Herald's Bard page and the Allakhazam song entries both say only one song can be sung at a time. **We follow OpenDAoC and the sources: one at a time; casting another song ends the first.**
- **Effect duration.** The data gives a duration of 5 seconds against a 5-second pulse, so an effect that lasts exactly its duration would flicker each pulse. We keep the effect for one pulse interval plus one second, so it lapses about six seconds after the caster stops or the member leaves range.
- Status: reference for the toggle, the pulse and the one-at-a-time rule; provisional for the grace second (ours).

### Speed enhancement — `spells/SpeedEnhancementSpellHandler.cs` (both projects, the same rule)

- **Value.** `Value` is a percent of normal speed (144 runs at 1.44 times). It is stored as a multiplier on maximum speed next to the slows' multipliers; the multipliers multiply (a 1.44 song and a 40% slow give 1.44 × 0.6).
- **One at a time.** Effects of this spell type replace one another; the stronger counts.
- **Combat.** A pulsing or cast-time speed spell does not take hold on a target that is in combat ("the spell has no effect on you"). Once on, it is removed when the holder lands a hit or is hit, or casts an offensive spell. The Allakhazam entry for a speed song says it does not work in combat. Both projects agree.
- **Not on a horse or a keep guard**, and the effect is removed while stealthed. Mounted travel is the horse routes here, and stealth is not in the game, so neither is relevant.
- Status: reference.

### Health regeneration buff — `spells/RegenBuff.cs`, `propertycalc/HealthRegenerationRateCalculator.cs`

- `Value` is extra hit points added to each health tick, on top of the base regeneration; a tick comes every 3 seconds. The base regeneration differs from this server's (DOLSharp: 10 + 0.2 per level below 26, 0.6 per level above, a later patch's figures); ours is the existing provisional rule.
- **In combat.** DOLSharp lets the tick run in combat at a fifth of the rate (patch 1.87 notes in its comments say the in-combat ticks became twice as frequent); this server's rule gives no regeneration in combat. The buff rides on the tick, so under our rule it works out of combat only. No 1.65 account of the song in combat was found.
- Status: provisional.

### Resist buffs — `spells/ResistBuff.cs`, `propertycalc/ResistCalculator.cs`

- A resist buff adds its `Value` in percent to one resistance, whatever the caster's stat or spec (no scaling). Buffs of one damage type are limited to a total of 24% in DOLSharp's resist calculator; the item bonus has its own cap (half the level plus one) that is separate from the buff cap, and the total of racial, item and buff resist is hard-capped at 70.
- The Friday's Grab Bag answer (Allakhazam story 736) confirms buff and item caps are separate, with a 24%-from-buffs example.
- **Our server today** adds item bonuses to the race resist with no item cap, and has the 70% hard cap only at the point damage is taken. The item cap is not part of this change.
- Status: reference for the flat value and the 24% buff cap; the 70 hard cap is already in the rule book.

## 1.2 Searches for a published 1.65 account

All run 2026-11-16 as web searches.

| Question | Query | Result |
| --- | --- | --- |
| Songs, speed in combat | `Dark Age of Camelot Skald songs only one song at a time speed song breaks in combat Midgard` | Found Camelot Herald class pages (Skald, Bard) and Allakhazam spell entries: only one song at a time; Song of Travel 144% at level 13, 204% at 43, pulse 6 seconds, range 2000; a speed song "does not work in combat". The "breaks when hit" wording came from Camelot Unchained, not DAoC, and is not used. |
| Class numbers | `DAoC Thane Skald class hit points weapon skill spec points per level 1.65 Herald` | Empty for numbers: only class descriptions and player builds. The Thane description (hammer, area and buff spells) matches the data. Numbers stay provisional from OpenDAoC-Core. |
| Area falloff | `DAoC PBAoE area spell damage falloff distance from center Thunder Shout` | Empty: only other games. No 1.65 account of a falloff. |
| Call of a thousand storms | `DAoC "Call of a Thousand Storms" Thane realm ability` | Empty: no source for the name. The census (from the community database) counts it as a Thane ability grant, so it stays listed as not implemented; it is not marked not classic. |
| Regeneration song in combat | `DAoC Song of Rest Skald health regeneration in combat out of combat` | Only that the songs raise the regeneration rate and only one can be sung (Allakhazam); nothing on combat. A community post says patch 1.87 made in-combat ticks twice as frequent, which puts a 1.65 rule before that change. |
| Resist caps | `DAoC resist cap buff 24% spell resist buffs total cap 26% level/2+1 item resist cap` | The Friday's Grab Bag answer above: buff and item caps are separate, 24% from buffs as the example. |

## 1.3 Open questions, settled

1. **Class numbers:** as in 1.1, provisional.
2. **Falloff:** none; provisional; the OpenDAoC falloff is recorded.
3. **Song limit:** one song at a time; a new song ends the running one; death cancels; a held caster skips pulses; no power cost in the data.
4. **Speed song:** value is percent of normal; multiplies with slows; does not take hold in combat and ends when the holder enters combat.
5. **Regeneration song:** extra hit points on each tick; out of combat only under our rule.
6. **Call of a thousand storms:** no source; it stays counted as not implemented and is not marked not classic.
