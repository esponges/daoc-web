# rules-experience Specification

## Purpose
Defines how much experience levels take, what creatures are worth, how kills award experience, and how creatures con to a player, as in classic 1.65.

## Requirements

### Requirement: The experience table
A character SHALL reach each level when their total experience reaches that level's threshold, as listed in the scenario below. Experience SHALL stop accruing at level 50. Status: documented.

#### Scenario: The thresholds
- **WHEN** a character's total experience reaches a threshold
- **THEN** they are at least that level:

| level | threshold |
| --- | --- |
| 2 | 50 |
| 3 | 250 |
| 4 | 850 |
| 5 | 2,300 |
| 6 | 6,350 |
| 7 | 15,950 |
| 8 | 37,950 |
| 9 | 88,950 |
| 10 | 203,950 |
| 11 | 459,950 |
| 12 | 839,950 |
| 13 | 1,399,950 |
| 14 | 2,199,950 |
| 15 | 3,399,950 |
| 16 | 5,199,950 |
| 17 | 7,899,950 |
| 18 | 11,799,950 |
| 19 | 17,499,950 |
| 20 | 25,899,950 |
| 21 | 38,199,950 |
| 22 | 54,699,950 |
| 23 | 76,999,950 |
| 24 | 106,999,950 |
| 25 | 146,999,950 |
| 26 | 199,999,950 |
| 27 | 269,999,950 |
| 28 | 359,999,950 |
| 29 | 479,999,950 |
| 30 | 639,999,950 |
| 31 | 849,999,950 |
| 32 | 1,119,999,950 |
| 33 | 1,469,999,950 |
| 34 | 1,929,999,950 |
| 35 | 2,529,999,950 |
| 36 | 3,319,999,950 |
| 37 | 4,299,999,950 |
| 38 | 5,499,999,950 |
| 39 | 6,899,999,950 |
| 40 | 8,599,999,950 |
| 41 | 12,899,999,950 |
| 42 | 20,699,999,950 |
| 43 | 29,999,999,950 |
| 44 | 40,799,999,950 |
| 45 | 53,999,999,950 |
| 46 | 69,599,999,950 |
| 47 | 88,499,999,950 |
| 48 | 110,999,999,950 |
| 49 | 137,999,999,950 |
| 50 | 169,999,999,950 |

#### Scenario: Level 10 to 11
- **WHEN** a level 10 character has 203,950 experience
- **THEN** they need 256,000 more to reach level 11

### Requirement: Creatures are worth a value by level
A creature's experience value SHALL come from the published table. That table doubles each level up to level 10 (5 × 2^level). After that it grows by a factor of 2 every 2 levels up to 20, every 3 levels up to 35, and every 5 levels up to 50. Status: documented. The table departs from that curve at levels 15, 22 and 48, and the table's values are used:
- 28,980 at level 15;
- 231,705 at level 22;
- 31,625,241 at level 48.

#### Scenario: Values
- **WHEN** creatures of level 10, 20, 35 and 50 are killed
- **THEN** they are worth 5,120, 163,840, 5,242,880 and 41,943,040

#### Scenario: The level 22 quirk
- **WHEN** a level 22 creature is killed
- **THEN** it is worth 231,705, as the table says, not the curve's 260,079

### Requirement: Kills award a damage share, capped
Experience for a kill SHALL be shared among the players who damaged the creature, each by their share of the damage it took. Each player's award SHALL be capped at 1.25 × the value of a creature of that player's own level. If the creature cons grey to any player who damaged it, nobody SHALL receive experience. Status: documented (the 1.25 cap) and reference.

#### Scenario: An even solo kill
- **WHEN** a level 10 Warrior alone kills a level 10 creature
- **THEN** they gain 5,120

#### Scenario: Capped
- **WHEN** a level 10 Warrior alone kills a level 12 creature worth 10,240
- **THEN** they gain 6,400

#### Scenario: Shared
- **WHEN** two level 10 Warriors kill a level 10 creature, one dealing 75% of the damage
- **THEN** they gain 3,840 and 1,280

#### Scenario: A grey kill
- **WHEN** a level 20 Warrior helps kill a level 9 creature
- **THEN** no one gains experience from it

### Requirement: Con colours
A creature's con colour to a player of level `L` SHALL depend on its level `t` and a band width `w` = min(5, ⌊L / 10⌋ + 1), with bands as the scenario below lists. Status: documented for level 40 and above (Mythic's description of the bands). Below 40 it is provisional: the reference server uses a narrower table at low levels, and the difference is recorded.

#### Scenario: The bands
- **WHEN** a player of level `L` cons a creature of level `t`
- **THEN** its colour is:

| colour | levels |
| --- | --- |
| grey | t ≤ L − 3w |
| green | up to L − 2w |
| blue | up to L − w |
| yellow | up to L |
| orange | up to L + w |
| red | up to L + 2w |
| purple | above L + 2w |

#### Scenario: Level 40
- **WHEN** a level 40 player cons creatures of levels 25, 30, 35, 40 and 45
- **THEN** they are grey, green, blue, yellow and orange

#### Scenario: Level 50
- **WHEN** a level 50 player cons a level 61 creature
- **THEN** it is purple

### Requirement: Clients see the con colour
Each creature entity in a snapshot SHALL carry its con colour to the receiving player, as a number from −3 (grey) to 3 (purple). A client that does not know the field SHALL ignore it.

#### Scenario: Target window
- **WHEN** a level 10 player targets a level 12 creature
- **THEN** its entity carries con 1 (orange, since `w` is 2) and the client colours its name to match

### Requirement: Existing characters keep their level
A character saved before this change SHALL keep their level. Their experience SHALL be moved to the new table at the same fraction of the way through that level.

#### Scenario: Halfway through 10
- **WHEN** a level 10 character saved halfway to 11 under the old curve logs in
- **THEN** they are level 10 with 331,950 experience

### Requirement: A group shares a kill as one damager
For a kill's damage share, all damage dealt by members of one group SHALL count as the group's. The group's award SHALL be the creature's value × the group's share. Solo players and other groups SHALL keep their own shares, as in "Kills award a damage share, capped". Status: provisional.

#### Scenario: A group and a solo player
- **WHEN** two grouped level 10 Warriors deal 75% of a level 10 creature's damage and a solo level 10 Warrior deals 25%
- **THEN** the group's award is 3,840, split between its members, and the solo Warrior gains 1,280

### Requirement: A group's award is split by level among members in range
A group's award SHALL be split among the members who are alive and within 6,000 units of the creature when it dies, in proportion to their levels, rounded down. Each member's part SHALL then be capped at 1.25 × the value of a creature of their own level. Members out of range or dead SHALL get nothing and SHALL NOT count in the split. There SHALL be no group bonus until a 1.65 source for one is recorded. Status: provisional.

#### Scenario: Two even members
- **WHEN** a group of two level 10 Warriors, both in range, kills a level 10 creature alone
- **THEN** each gains 2,560

#### Scenario: Split by level, then capped
- **WHEN** a group of a level 10 and a level 5 character, both in range, kills a level 10 creature worth 5,120
- **THEN** the level 10 gains 3,413, and the level 5's share of 1,706 is capped at 200

#### Scenario: A member far away
- **WHEN** a third member is 9,000 units from the kill
- **THEN** they gain nothing and the other two split the award as if the group had two members

### Requirement: A group kill is grey by its highest member in range
A group's kill SHALL give the group nothing if the creature cons grey to the highest-level member in range. Status: provisional.

#### Scenario: A high-level helper
- **WHEN** a level 20 and a level 10 member, both in range, kill a level 10 creature
- **THEN** the creature is grey to the level 20 and neither member gains experience
