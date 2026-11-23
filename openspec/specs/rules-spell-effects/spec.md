# rules-spell-effects Specification

## Purpose
Defines, for the spell types creatures cast in 1.65 Midgard, whether a spell lands, what it does to its target, how resists reduce it, and how long it lasts, using each spell's imported values.

## Requirements

### Requirement: A hostile spell lands by a hit roll
A hostile spell SHALL land with chance 85% + (caster level − target level) / 2 %, clamped to 50%–99%. A spell that does not land SHALL do nothing, and both sides SHALL be told the target resisted it. Status: reference, until a 1.65 source is recorded.

#### Scenario: Even levels
- **WHEN** a level 10 creature casts at a level 10 Warrior
- **THEN** the chance to land is 85%

#### Scenario: A higher caster
- **WHEN** a level 16 creature casts at a level 10 Warrior
- **THEN** the chance is 88%

#### Scenario: Far below
- **WHEN** a level 1 creature casts at a level 90 target
- **THEN** the chance is clamped to 50%

### Requirement: Resists reduce spell damage
Spell damage SHALL be multiplied by (1 − resist / 100), where `resist` is the target's resist to the spell's damage type from race and items, capped at 70%. Status: reference for the cap's value, until a 1.65 source is recorded.

#### Scenario: A resisted fraction
- **WHEN** a spell does 100 before resists to a target with 10% resist to its type
- **THEN** it does 90

### Requirement: Direct damage and bolts vary
A creature's direct-damage or bolt spell SHALL do its `Damage` × a variance uniform between 0.75 and 1.25, then the resist, rounded down; a spell of the no-variance type SHALL use a variance of 1. A player's direct-damage or bolt spell from their book SHALL do its `Damage` × (casting stat + 200) / 275, × a roll uniform between a low end and a high end of 1.25, then the resist, then at most three times the spell's `Damage`, rounded down. The low end is 0.25 + (the caster's spec in the spell's line − 1) / the target's level. A spec above the target's level adds 0.005 per level over, and a caster above the target's level adds 0.02 per level of difference, to both ends; the high end is at least 0.25 and the low end between 0 and the high end. A no-variance spell from a player SHALL do its `Damage` scaled by neither. Status: reference (DOLSharp and OpenDAoC-Core; Mythic's Grab Bag answers confirm that a spec at or above the spell's level gives full effect and that damage starts at +25 percent, with no figures).

#### Scenario: Middle of the range
- **WHEN** a creature's 40-damage direct-damage spell rolls a variance of 1.0 against 10% resist
- **THEN** it does 36

#### Scenario: No variance
- **WHEN** a 40-damage no-variance spell lands on a target with no resist
- **THEN** it does exactly 40

#### Scenario: A fully specced caster
- **WHEN** a level 20 caster with 100 in their casting stat and a spec of 20 casts a 40-damage spell at a level 20 target, rolling the middle of the range
- **THEN** the range is 1.2 to 1.25 and the spell does 53

#### Scenario: A low spec
- **WHEN** the same caster has a spec of 5 and rolls the worst result
- **THEN** the low end is 0.45 and the spell does 19

#### Scenario: The cap
- **WHEN** a level 50 caster with 300 in their casting stat and a spec of 50 casts a 10-damage spell at a level 1 target
- **THEN** the range is far above the cap and the spell does 30

### Requirement: A bolt can be blocked
A bolt that lands SHALL be tested against the target's block chance, as a melee blow from the front would be. A blocked bolt SHALL do no damage. Status: reference.

#### Scenario: Blocked bolt
- **WHEN** a bolt reaches a Warrior with a shield facing the caster and the block roll succeeds
- **THEN** the bolt does no damage and both sides see it blocked

### Requirement: Damage over time ticks
A damage-over-time spell SHALL hit every `Frequency` tenths of a second for `Duration` seconds. A creature's spell SHALL hit for its `Damage` after the resist, with no variance. A player's spell from their book SHALL hit, on each tick, for its `Damage` × (casting stat + 200) / 275 × a roll uniform between a low end and 1.13, then the resist, then at most three times the spell's `Damage`, rounded down. The low end is 0.13 + (the caster's spec in the spell's line − 1) / the target's level, between 0 and 1.13. There is no level-difference term. The casting stat and spec are those of the caster when the spell landed. A second cast of the same spell on the same target SHALL restart its duration and take the new cast's stat and spec, not stack. Status: reference (DOLSharp states the range; OpenDAoC-Core uses the general spell variance for it, and the DOLSharp range is followed).

#### Scenario: Five ticks
- **WHEN** a creature's spell with Damage 10, Frequency 40 and Duration 20 lands on a target with no resist
- **THEN** the target takes 10 five times, 4 seconds apart

#### Scenario: Recast
- **WHEN** the same spell lands again on that target after two ticks
- **THEN** one effect remains, with 20 seconds to run

#### Scenario: A player's tick at spec
- **WHEN** a player with 100 in their casting stat and a spec of 20 casts a 10-damage spell at a level 20 target and the tick rolls the middle of the range
- **THEN** the range is 1.08 to 1.13 and the tick does 12

#### Scenario: A low spec tick
- **WHEN** the same caster has a spec of 1 and the tick rolls the worst result
- **THEN** the low end is 0.13 and the tick does 1

#### Scenario: The tick cap
- **WHEN** a caster with 600 in their casting stat casts a 5-damage spell and the tick rolls the best result
- **THEN** the tick does at most 15

#### Scenario: A player's tick uses the cast
- **WHEN** a player's damage-over-time spell has landed and the caster's casting stat then falls
- **THEN** the ticks keep the stat the caster had when it landed

### Requirement: Slows lower speed
A speed-decrease spell SHALL multiply the target's movement speed by (1 − `Value` / 100) for `Duration` seconds. The strongest active slow SHALL apply. Status: reference.

#### Scenario: A 40% slow
- **WHEN** a 40% slow lands on a running Warrior
- **THEN** their speed is 60% of normal for its duration, and the client's predicted movement agrees

### Requirement: Mesmerize holds until hurt
A mesmerize spell SHALL stop its target moving, swinging, casting and using styles for `Duration` seconds. Any damage to the target SHALL end it. Status: reference.

#### Scenario: Broken by damage
- **WHEN** a mesmerized player is hit by another creature
- **THEN** the mesmerize ends at once and the player can act

### Requirement: Stun holds through damage
A stun spell SHALL stop its target moving, swinging, casting and using styles for `Duration` seconds. Damage SHALL NOT end it. Status: reference.

#### Scenario: Hit while stunned
- **WHEN** a stunned player is hit
- **THEN** the stun continues to its end

### Requirement: Life drain heals its caster
A life-drain spell SHALL do damage as direct damage does, and SHALL heal its caster by that damage × `LifeDrainReturn` / 100, up to the caster's maximum health. `LifeDrainReturn` is a percent: the data's 117 drain spells use values from 0 to 100. Status: reference.

#### Scenario: Drain
- **WHEN** a drain with Damage 40, return 50, rolls a variance of 1.0 on a target with no resist
- **THEN** the target loses 40 and the caster regains 20

### Requirement: Heals restore health
A creature's heal spell SHALL restore `Value` health to its target, up to its maximum. A player's heal from their book SHALL restore a roll between a least and a most, up to the target's maximum: the most is 1.25 × `Value` and the least is `Value` × (0.25 + (the caster's spec in the spell's line − 1) / the spell's level), at most 1.25, and at least 1. A heal whose value is negative SHALL restore that percent of the caster's maximum health. Status: reference (DOLSharp and OpenDAoC-Core; Mythic's Grab Bag answers confirm that heal variance narrows as spec nears the spell's level, with no figures).

#### Scenario: A creature heals itself
- **WHEN** a creature at 100 of 300 health casts a 120-point heal on itself
- **THEN** it has 220

#### Scenario: A player heal at spec
- **WHEN** a caster with a spec of 20 casts a level 20 heal of 120
- **THEN** it restores between 144 and 150

#### Scenario: A player heal below spec
- **WHEN** a caster with a spec of 5 casts a level 20 heal of 120
- **THEN** it restores between 54 and 150

#### Scenario: A percent heal
- **WHEN** a caster with 720 health casts a heal whose value is −40
- **THEN** it restores 288

### Requirement: Effects show and end
Each lasting effect SHALL be sent to its target with its spell name, icon id and time left, and SHALL be removed when it ends, when its caster's spell is replaced, or when its target dies. Effects SHALL NOT survive a logout. Status: provisional for the logout rule.

#### Scenario: Effect icons
- **WHEN** a slow lands on a player
- **THEN** their client shows an effect with the spell name and a countdown, gone when the slow ends

### Requirement: A spell rule's status follows its source
A spell rule SHALL NOT be marked above provisional unless its spec names a source for the formula, and SHALL NOT be marked documented unless that source is a published account of 1.65 itself. Where two sources disagree, the spec SHALL say so and name the one followed.

#### Scenario: A reference rule names its source
- **WHEN** a spell rule is marked reference
- **THEN** its spec names the community server or table the formula was understood from, in the project's own words

#### Scenario: No source found
- **WHEN** no source for a formula can be found
- **THEN** the rule stays provisional and `findings.md` records the search

### Requirement: Spell numbers agree with the imported spell table
For each imported spell of an implemented type, the amount the server would deal or restore for a reference caster and target SHALL be derivable from the spell's own `Damage` or `Value`, and a spell whose amount differs from that derivation SHALL be reported.

#### Scenario: A direct damage spell
- **WHEN** the check runs over a 40-damage direct-damage spell at the middle of its range, with no resist
- **THEN** the amount is within the range the rule gives for that spell's base and the spell is not reported

#### Scenario: A heal
- **WHEN** the check runs over a 120-value heal
- **THEN** the amount is within 0.25 to 1.25 times 120, up to the target's maximum

#### Scenario: A mismatch
- **WHEN** a spell's amount falls outside that range
- **THEN** the check names the spell id and both amounts

#### Scenario: No database
- **WHEN** `server/data/community` is absent
- **THEN** the check is skipped and the rest of the tests run

### Requirement: A damage add is a damage per second
A damage-add spell's `Damage` SHALL be a damage per second. On each landed blow it SHALL add the lesser of `Damage` and the cap for the caster's level, × the blow's swing time in seconds × a roll uniform between 0.5 and 1, rounded down, after the target's resist. The cap is (1.2 + 0.3 × level) × 0.7 damage per second. A damage add whose `Damage` is negative SHALL add that percent of the blow's own damage. It does not depend on the caster's stat or spec. The strongest damage add on a player counts. Status: reference (DOLSharp states the formula and the cap; OpenDAoC-Core scales by swing time with the general variance; the DOLSharp roll is followed).

#### Scenario: A slow weapon
- **WHEN** a level 20 caster's 4.5 per second damage add lands on a 3.0 second swing and rolls the middle
- **THEN** it adds 10

#### Scenario: The damage add cap
- **WHEN** a level 3 caster's 9 per second damage add lands on a 3.0 second swing and rolls the best result
- **THEN** it adds 4, the cap 1.47 times 3 seconds

#### Scenario: A percent add
- **WHEN** a damage add of −20 lands on a blow of 50 damage
- **THEN** it adds 10

#### Scenario: The strongest counts
- **WHEN** two damage-add spells are on a player
- **THEN** each blow adds from the stronger one only

### Requirement: Speed enhancement raises speed
A speed-enhancement spell SHALL multiply its holder's movement speed by its `Value` read as a percent (144 runs at 1.44 times). The strongest active one SHALL count, and it SHALL multiply with any slow. It SHALL not take hold on a target that is in combat, and SHALL end when its holder enters combat. The client's predicted movement SHALL agree with the server at the enhanced speed. Status: reference.

#### Scenario: Simple Song of Travel
- **WHEN** a 144 Song of Travel is on a running character with no slow
- **THEN** they run at 1.44 times their normal speed, and the shared movement vector for that factor passes in Go and JS

#### Scenario: The strongest counts
- **WHEN** a player holds a 144 and a 159 speed enhancement
- **THEN** they run at 1.59 times their normal speed

#### Scenario: With a slow
- **WHEN** a character with a 144 song is hit by a 40% slow
- **THEN** their speed is 1.44 × 0.6 of normal

#### Scenario: Combat
- **WHEN** a character with a song on swings at a creature
- **THEN** the song's effect ends on them, and a new pulse does not take hold until they are out of combat

### Requirement: A health regeneration buff adds to health ticks
A health-regeneration spell SHALL add its `Value` hit points to each health tick while it holds, and the strongest one SHALL count. Under the rule that no health returns in combat, the buff adds nothing in combat. Status: provisional.

#### Scenario: Song of Rest out of combat
- **WHEN** a level 20 character out of combat holds a 10-point Song of Rest
- **THEN** each health tick returns 10 more hit points than it would without the song

#### Scenario: In combat
- **WHEN** the same character is in combat with the song on
- **THEN** no tick returns hit points

### Requirement: Resist buffs raise a resist
A resist buff SHALL add its `Value` in percent to its holder's resist against its damage type: body, spirit, energy, heat, cold or matter, whatever the caster's stat or spec. The strongest buff for a damage type SHALL count, and resist buffs SHALL add at most 24% to one resist. The total SHALL still be capped by the resist cap of 70%. Status: reference.

#### Scenario: A heat chant against a fire bolt
- **WHEN** a player with 5% heat resist from their race holds a 10-point Heat Diminishing Chant and is hit by a heat spell
- **THEN** the spell meets 15% heat resist

#### Scenario: Another damage type is untouched
- **WHEN** the same player is hit by a cold spell
- **THEN** the heat chant adds nothing to the resist it meets

#### Scenario: The buff cap
- **WHEN** a player holds a 30-point body resist buff
- **THEN** it adds 24% to their body resist

### Requirement: Area damage does not fall off
Area damage SHALL not fall off with distance from the centre: each creature inside the radius takes its own full roll. The centre SHALL be the caster when the spell's range is 0, and the target otherwise. Status: provisional (the community servers differ and no 1.65 account was found).

#### Scenario: Centre and edge
- **WHEN** an area spell hits one creature at its centre and one at its edge
- **THEN** each takes a damage roll from the same range, resisted by its own resist

### Requirement: Stat and resist debuffs lower what they name
A stat debuff (strength, dexterity, strength/constitution, dexterity/quickness) SHALL lower its holder's named stats by its value. A resist debuff SHALL lower one resist by its value. A combat-speed debuff SHALL lengthen its holder's swings by its percent. The strongest debuff of each kind SHALL count. A debuff SHALL be resisted or not as any hostile spell is. Status: reference.

#### Scenario: Strength debuff on a creature
- **WHEN** a creature with 60 strength holds a 20-point strength debuff
- **THEN** its melee damage is computed with 40 strength

#### Scenario: Body resist debuff
- **WHEN** a creature with 10% body resist holds a 15% body resist debuff and takes a body spell
- **THEN** the spell meets −5% body resist

### Requirement: Disease, nearsight, confusion and amnesia
The control debuffs SHALL each do one thing, by values recorded in the rule book:
- disease halves healing received, lowers strength by 7.5 percent and slows its holder by 15 percent;
- nearsight cuts its holder's spell range;
- confusion makes a creature switch to a random nearby target;
- amnesia wipes a creature's hate list, and interrupts a cast.

Status: set by the findings.

#### Scenario: Disease
- **WHEN** a diseased player is healed for 200
- **THEN** they gain the amount the rule gives for a diseased target

#### Scenario: Amnesia
- **WHEN** a Healer's amnesia lands on a creature fighting them
- **THEN** the creature forgets its enemies and may turn on the nearest one again

### Requirement: Constitution, dexterity and acuity buffs
A constitution buff SHALL raise its holder's constitution by its value, a dexterity buff their dexterity, and an acuity buff the stat their class casts with (and nothing for a class that does not cast), each limited by the buff cap on stats. With a `Concentration` cost they are concentration buffs. Status: reference.

#### Scenario: Acuity on a Healer
- **WHEN** a level 20 Healer holds a 21-point acuity buff
- **THEN** their piety is raised by 21, within the buff cap for their level (25)

### Requirement: Haste and the damage shield
A haste (combat-speed buff) SHALL shorten its holder's swings by its percent. A damage shield SHALL deal its damage to each attacker who hits the holder in melee. Status: reference.

#### Scenario: Haste
- **WHEN** a player with a 3.0-second swing holds a 10% haste
- **THEN** their swing time is reduced by 10% under the swing-speed rule

#### Scenario: Damage shield
- **WHEN** a creature hits a player holding a 5-point damage shield
- **THEN** the creature takes the shield's damage, adjusted by its resist

### Requirement: Bladeturn, ablative and absorption
Bladeturn SHALL stop the next melee or ranged hit completely and then end, unless the hit carries Penetrating Arrow. Ablative armour SHALL absorb its percent of each melee hit until its pool is spent. Armour absorption SHALL raise its holder's absorption by its value. Status: reference.

#### Scenario: Bladeturn
- **WHEN** a player holding a bladeturn is hit in melee
- **THEN** the blow does no damage, a message says it was turned, and the bladeturn ends

#### Scenario: Ablative pool
- **WHEN** a player with a 25%, 100-point ablative takes a 120-damage hit
- **THEN** 30 is absorbed and 70 of the pool remains

### Requirement: A mesmerize duration buff shortens mesmerizes on its holder
A mesmerize duration buff of N percent SHALL shorten every mesmerize that lands on its holder by N percent, never to under one second. The Spiritmaster casts it on their pet. Status: provisional.

#### Scenario: Shorter mesmerize
- **WHEN** a pet holding a 50 percent mesmerize duration buff is mesmerized for 10 seconds
- **THEN** the mesmerize lasts 5 seconds, and never less than one

### Requirement: Regeneration buffs, cures, spread heal and life transfer
These buffs and heals SHALL do what their type names, by values recorded in the rule book:
- an endurance or power regeneration buff adds its value to each endurance or power tick;
- a cure ends every effect of its kind on its target;
- a spread heal heals the most hurt group members, sharing out its amount;
- life transfer heals the target and costs the caster half the heal (see the scenario);

Status: reference.

#### Scenario: Power regeneration
- **WHEN** a Healer holding a power regeneration buff has a power tick
- **THEN** it returns the buff's value more than without it

#### Scenario: Cure of a mesmerize
- **WHEN** a mesmerized player is cured of mezz
- **THEN** the mesmerize ends and they can move again; a cure of poison, disease or nearsight finds nothing to end until those effects exist

#### Scenario: Spread heal
- **WHEN** a Healer's 187-point spread heal is cast with two group members missing 300 and 100 hit points
- **THEN** they gain 141 and 46: shares in proportion to what each is missing, never more than that

#### Scenario: Life transfer
- **WHEN** a Spiritmaster casts a 10 percent life transfer (maximum 1,000) on a group member
- **THEN** the member gains twice the percent of the caster's maximum (200), the caster loses half of that (100) and never falls under 1 hit point, and a diseased member gains half

### Requirement: Damage spells that also debuff or slow
A direct damage spell with a sub-spell SHALL deal its damage and then apply the sub-spell when the damage is not resisted. A damage-and-slow spell SHALL do the same with its slow. Status: reference.

#### Scenario: Damage with debuff
- **WHEN** a Spiritmaster's damage-with-debuff spell hits a creature that does not resist
- **THEN** the creature takes the damage and then holds the debuff

### Requirement: Savage self-buffs
The Savage buffs SHALL apply only to the caster, SHALL cost health rather than power (the amount in the spell's data, as a share of maximum health where the data says so), and SHALL be refused when the cost would kill the caster. Each kind SHALL raise its own stat:
- the resistance buffs raise crush, slash or thrust resistance;
- the DPS buff raises weapon damage;
- the evade and parry buffs raise those chances;
- the combat speed buff shortens swing time.

#### Scenario: Health cost
- **WHEN** a Savage casts a parry buff that costs 10 percent of health
- **THEN** they lose that health, gain the parry bonus for the spell's duration, and their power is unchanged

#### Scenario: Not enough health
- **WHEN** a Savage whose health is below the cost casts a Savage buff
- **THEN** it is refused and nothing is spent

#### Scenario: Combat speed
- **WHEN** a Savage with the combat speed buff swings the same weapon
- **THEN** the swing time is shorter by the buff's amount, within the swing speed floor

### Requirement: Savage endurance heal
The Savage endurance heal SHALL restore endurance to the caster by the spell's value, paid in health.

#### Scenario: Endurance restored
- **WHEN** a Savage at low endurance casts the endurance heal
- **THEN** their endurance rises by the spell's amount and their health falls by its cost

### Requirement: Taunt
Taunt SHALL add the spell's value to the target creature's hate for the caster, and SHALL do no damage.

#### Scenario: Pulling hate
- **WHEN** a creature is fighting a Healer and a Savage taunts it enough to lead its hate list
- **THEN** the creature turns to the Savage
