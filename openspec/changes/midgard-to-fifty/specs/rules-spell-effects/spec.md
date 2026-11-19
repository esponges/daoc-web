## ADDED Requirements

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
- disease halves healing received and stops health regeneration, and slows its holder;
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

### Requirement: Regeneration buffs, cures, spread heal and life transfer
These buffs and heals SHALL do what their type names, by values recorded in the rule book:
- an endurance or power regeneration buff adds its value to each endurance or power tick;
- a cure ends every effect of its kind on its target;
- a spread heal heals the most hurt group members, sharing out its amount;
- life transfer moves the caster's health to the target.

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
- **WHEN** a Spiritmaster transfers 50 to a group member
- **THEN** the caster loses the amount the rule gives and the member gains 50

### Requirement: Damage spells that also debuff or slow
A direct damage spell with a sub-spell SHALL deal its damage and then apply the sub-spell when the damage is not resisted. A damage-and-slow spell SHALL do the same with its slow. Status: reference.

#### Scenario: Damage with debuff
- **WHEN** a Spiritmaster's damage-with-debuff spell hits a creature that does not resist
- **THEN** the creature takes the damage and then holds the debuff
