# Spec Delta

## Purpose

Defines a character's eight stats, hit points and endurance the way classic 1.65 computes them: from race, class, creation points, level and the items worn.

## ADDED Requirements

### Requirement: Races give the starting stats
A character's stats before creation points SHALL be their race's base stats, in the order strength, constitution, dexterity, quickness, intelligence, piety, empathy, charisma, as the scenario below lists for the Midgard races of 1.65. Status: documented (the classic race tables), and matching the reference server.

#### Scenario: The Midgard races
- **WHEN** a character of each Midgard race is created before any points are spent
- **THEN** their stats are:

  | race | STR | CON | DEX | QUI | INT | PIE | EMP | CHA |
  | --- | --- | --- | --- | --- | --- | --- | --- | --- |
  | Norseman | 70 | 70 | 50 | 50 | 60 | 60 | 60 | 60 |
  | Troll | 100 | 70 | 35 | 35 | 60 | 60 | 60 | 60 |
  | Dwarf | 60 | 80 | 50 | 50 | 60 | 60 | 60 | 60 |
  | Kobold | 50 | 50 | 70 | 70 | 60 | 60 | 60 | 60 |
  | Valkyn | 55 | 45 | 65 | 75 | 60 | 60 | 60 | 60 |

### Requirement: Classes name a primary, secondary and tertiary stat and base numbers
Each class SHALL define:
- its primary, secondary and tertiary stats;
- its base hit points;
- its weapon skill base;
- its specialisation multiplier;
- the races that may take it.

For the Warrior:
- primary strength, secondary constitution, tertiary dexterity;
- base hit points 880;
- weapon skill base 460;
- multiplier 20, meaning 2.0 points per level;
- races Norseman, Troll, Dwarf, Kobold and Valkyn.

Status: reference.

#### Scenario: Warrior definition
- **WHEN** the server looks up class 22
- **THEN** it is the Warrior with the numbers above, and a Kobold may be one

### Requirement: Creation spends exactly 30 points
A new character SHALL distribute exactly 30 points over their stats. Raising a stat by `n` above its race base SHALL cost `n`, plus 1 more for each point above 10, plus 1 more again for each point above 15. A build that spends more or less than 30, or lowers a stat below its base, SHALL be refused. Status: documented.

#### Scenario: Cost of a stat
- **WHEN** a player raises strength by 18
- **THEN** it costs 18 + 8 + 3 = 29 points

#### Scenario: Default build
- **WHEN** a Warrior is created without a build, as the current creation page does
- **THEN** strength, constitution and dexterity are each raised by 10, which spends exactly 30

#### Scenario: Overspent build
- **WHEN** a request spends 31 points
- **THEN** creation is refused with a reason and no character is made

### Requirement: Levels from 6 raise the class stats
On reaching each level from 6 to 50, a character SHALL gain:
- 1 in their primary stat;
- 1 in their secondary stat when (level − 6) is a multiple of 2;
- 1 in their tertiary stat when (level − 6) is a multiple of 3.

Levels 1 to 5 give no stats. Status: documented.

#### Scenario: Troll Warrior at 50
- **WHEN** a Troll Warrior with the default build reaches level 50
- **THEN** their strength is 155, constitution 103, dexterity 60 and quickness 35

#### Scenario: Level 7
- **WHEN** a Warrior goes from level 6 to level 7
- **THEN** only strength rises

### Requirement: Items add to stats up to a cap
Stat bonuses on worn items SHALL add to the matching stat. The total item bonus to any one stat SHALL be capped at 1.5 × level, rounded down. A stat SHALL never be used above 350 or below 1. Status: documented.

#### Scenario: Over the cap
- **WHEN** a level 10 Warrior wears items giving +20 strength in total
- **THEN** only +15 counts

### Requirement: Hit points come from class, level and constitution
A character's maximum hit points SHALL be computed as follows:
1. Let `c` be constitution minus 50; if `c` is negative, double it.
2. Let `h` be the class's base hit points × level.
3. Maximum hit points = 20 + ⌊h / 50⌋ + trunc(h × c / 10000).

The constitution used includes item bonuses. Item hit point bonuses SHALL then add up to 4 × level. Status: documented for the item cap, reference for the formula.

#### Scenario: Level 1 Troll Warrior
- **WHEN** a Troll Warrior with the default build (constitution 80) is level 1
- **THEN** they have 39 hit points

#### Scenario: Level 50 Troll Warrior
- **WHEN** the same Warrior reaches level 50 with constitution 103 and no items
- **THEN** they have 1,133 hit points

#### Scenario: Low constitution
- **WHEN** a level 50 Warrior's constitution is 45
- **THEN** `c` is −10 and they have 856 hit points

#### Scenario: Item hit points
- **WHEN** a level 10 Warrior wears items giving +60 hit points
- **THEN** 40 of them count

### Requirement: Endurance is a pool of 100
A character's maximum endurance SHALL be 100. Endurance SHALL return 6 points a second out of combat and 2 a second in combat. Status: provisional for the regeneration rates.

#### Scenario: After a style chain
- **WHEN** a Warrior with 40 endurance stops fighting
- **THEN** they are back to 100 within ten seconds

### Requirement: Health returns slowly out of combat
Out of combat, a living character SHALL regain 2.5 + 0.25 × level hit points every 3 seconds, rounded down and at least 1. In combat they SHALL regain none. Status: provisional for the tick and the in-combat rule, reference for the amount.

#### Scenario: Level 20 resting
- **WHEN** a level 20 Warrior below full health is out of combat for 30 seconds
- **THEN** they regain 70 hit points

### Requirement: The character sheet is sent to the player
The server SHALL send a player their current stats with and without items, maximum hit points, maximum endurance, weapon skill and armour factor for the held weapon and worn armour, spec levels and unspent points. It SHALL send these on joining and whenever any of them changes.

#### Scenario: Equipping a strength item
- **WHEN** a player puts on a helm with +4 strength
- **THEN** they receive an updated sheet whose strength is 4 higher and whose weapon skill reflects it
