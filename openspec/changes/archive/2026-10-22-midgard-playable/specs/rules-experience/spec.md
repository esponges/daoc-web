# Spec Delta

## ADDED Requirements

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
