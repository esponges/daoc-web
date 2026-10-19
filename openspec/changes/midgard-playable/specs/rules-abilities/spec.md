# Spec Delta

## Purpose

Defines the Warrior's 1.65 defensive abilities: Guard, Protect, Intercept and Engage. It covers when each is gained, how it is used, and what it does in a fight against creatures.

## ADDED Requirements

### Requirement: Abilities are gained from the data's tables
A character SHALL have an ability at the level the imported specialisation-ability table gives it, for their class career or a trained spec. For a Warrior, the data gives:
- Protect I, II and III at career levels 7, 13 and 18;
- Intercept at 12;
- Guard I, II and III at Shields 5, 10 and 15;
- Engage at Shields 7.

A missing ability SHALL be refused when used. Status: documented by the data.

#### Scenario: Guard at Shields 5
- **WHEN** a Warrior trains Shields to 5
- **THEN** Guard I appears among their abilities and on the skills window

#### Scenario: Too early
- **WHEN** a level 6 Warrior uses Protect
- **THEN** the server refuses: they do not have it

### Requirement: Abilities target a group member or an enemy
Guard, Protect and Intercept SHALL be used on a targeted member of the user's group, within 2,000 units when used. Using one again on the same member SHALL end it. A user SHALL have each of them on at most one member at a time. Engage SHALL be used on a targeted enemy. Status: reference.

#### Scenario: Not grouped
- **WHEN** a Warrior uses Guard on a player who is not in their group
- **THEN** the server refuses: the target must be in their group

### Requirement: Guard blocks for a member
While a Warrior guards a member and holds a shield within 256 units of them, a melee blow at the member SHALL first be tested against the Warrior's block chance (`rules-melee`), from any direction. A success SHALL block the blow for the member, and both SHALL be told. Status: provisional for the 256 units.

#### Scenario: Guarded blow
- **WHEN** a creature swings at a guarded Skald, the guarding Warrior stands 150 units away with a 30% block chance, and the block roll is 0.20
- **THEN** the Warrior blocks the blow and the Skald takes nothing

#### Scenario: Too far
- **WHEN** the Warrior is 400 units from the Skald
- **THEN** Guard does not act on that blow

### Requirement: Intercept takes blows for a member
While a Warrior intercepts for a member within 128 units of them, the next melee blow at that member SHALL be taken by the Warrior instead, resolved against the Warrior's defences and armour. Intercept SHALL then end. It SHALL be usable again 60 seconds after use. Status: provisional.

#### Scenario: Intercepted
- **WHEN** a creature's blow lands at a member the Warrior intercepts for, with the Warrior 100 units away
- **THEN** the Warrior takes the blow and Intercept ends

### Requirement: Protect draws the creature's anger
While a Warrior protects a member within 1,000 units of them, each time a creature damages that member, it SHALL add to its hate for the Warrior. The amount SHALL be 10%, 20% or 30% of the damage, by Protect level. Status: provisional.

#### Scenario: Protect II
- **WHEN** a creature deals 100 damage to a member protected by Protect II
- **THEN** the creature's hate for the protecting Warrior rises by 20

### Requirement: Engage raises block against one enemy
A Warrior who engages an enemy and has not been attacked in the last 5 seconds SHALL have a 95% block chance against that enemy's blows from the front. Each blow blocked this way SHALL cost 5 endurance. Engage SHALL end when the Warrior runs out of endurance or targets another enemy. Status: provisional.

#### Scenario: Engaged
- **WHEN** a Warrior with a shield engages a creature it has not been fighting
- **THEN** that creature's blows from the front are blocked 95% of the time, and each block costs 5 endurance

#### Scenario: Already in combat
- **WHEN** the Warrior was hit 2 seconds ago
- **THEN** Engage is refused
