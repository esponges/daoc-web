# Spec Delta

## ADDED Requirements

### Requirement: Fights at level 1 follow the same rules
A level 1 character's fights SHALL use the same blow, armour and hit point rules as every other level, with no low-level adjustment. At level 1 the con band is one level wide, so a creature of level 2 is orange. Creature armour and damage grow from near nothing at level 1, so one level makes a large difference: a new Warrior wins an even fight with a yellow creature and loses to an orange one. Status: reference (OpenDAoC-Core with its default settings).

#### Scenario: A yellow creature at level 1
- **WHEN** a level 1 Troll Warrior (strength 110, 39 hit points) with the training axe (1.2 DPS, 2.5 seconds, 90% quality) and a bronze stelskodd jerkin (AF 4, 19% absorption, 85% quality) fights a level 1 creature (31 hit points), at the middle of each variance range
- **THEN** the Warrior hits it for 4.05 and it hits the Warrior's torso for 4.12

#### Scenario: An orange creature at level 1
- **WHEN** the same Warrior fights a level 2 creature (46 hit points)
- **THEN** the Warrior hits it for 1.93 and it hits the Warrior's torso for 10.78
