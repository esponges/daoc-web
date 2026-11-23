# rules-abilities Specification

## Purpose
Defines the Warrior's 1.65 defensive abilities: Guard, Protect, Intercept and Engage. It covers when each is gained, how it is used, and what it does in a fight against creatures. The numbers follow OpenDAoC's, read to understand them (nothing copied); a 2001 forum thread agrees in outline.

## Requirements

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
Guard, Protect and Intercept SHALL be used on a targeted member of the user's group, within 2,000 units when used. Using one again on the same member SHALL end it. A user SHALL have each of them on at most one member at a time. Engage SHALL be used on a targeted enemy. Status: reference (OpenDAoC).

#### Scenario: Not grouped
- **WHEN** a Warrior uses Guard on a player who is not in their group
- **THEN** the server refuses: the target must be in their group

### Requirement: Guard blocks for a member
A blow at a guarded member SHALL be tried against the member's own evade, parry and block first. If it is not stopped, and the guarding Warrior holds a shield, stands within 256 units of the member and has the attacker in front of them, the blow SHALL be tested against the Warrior's block chance plus 5% for each Guard level. A success SHALL block the blow for the member, and both SHALL be told. Status: reference.

#### Scenario: Guarded blow
- **WHEN** a creature's blow gets past a guarded Skald's defences, the guarding Warrior stands 150 units away with a 30% block chance and Guard II, and the roll is 0.35
- **THEN** the guard chance is 40%, the Warrior blocks the blow and the Skald takes nothing

#### Scenario: Too far
- **WHEN** the Warrior is 400 units from the Skald
- **THEN** Guard does not act on that blow

### Requirement: Intercept takes blows for a member
While a Warrior has Intercept armed on a member, a melee blow at that member within 128 units of the Warrior, when the Warrior can act, SHALL be taken by the Warrior instead on a 50% roll. The blow SHALL land on the Warrior as a hit, with no evade, parry, block or miss, and its damage SHALL be reduced by the Warrior's armour. Intercept SHALL then end; when the roll fails, or the Warrior is out of range, it SHALL stay armed. It SHALL be usable again 60 seconds after use. Status: reference.

#### Scenario: Intercepted
- **WHEN** a creature's blow lands at a member the Warrior intercepts for, with the Warrior 100 units away and a roll of 0.49
- **THEN** the Warrior takes the blow and Intercept ends

#### Scenario: Intercept fails its roll
- **WHEN** the roll is 0.5
- **THEN** the member takes the blow as usual and Intercept stays armed

### Requirement: Protect moves the creature's anger
When a creature's hate for a protected member rises, a share SHALL move from the member to each Warrior protecting them within 1,000 units who can act: 10%, 20% or 30% by Protect level. Status: reference.

#### Scenario: Protect II
- **WHEN** a protected member earns 100 hate with a creature, and the Warrior has Protect II
- **THEN** the creature's hate for the member rises by 80 and for the Warrior by 20

### Requirement: Engage raises block against one enemy
A Warrior with a shield SHALL engage an enemy that no enemy has attacked in the last 5 seconds. Against blows from that enemy from the front, their block chance SHALL rise by 65 percentage points, and each blow it applies to SHALL cost 5 endurance. While an enemy has attacked the creature in the last 5 seconds, the raise SHALL not apply. Status: reference.

#### Scenario: Engaged
- **WHEN** a Warrior with a shield engages a creature nobody has attacked
- **THEN** that creature's blows from the front meet a block chance 65 points higher, and each such blow costs 5 endurance

#### Scenario: Already in combat
- **WHEN** a group member attacked the creature 2 seconds ago
- **THEN** Engage is refused

### Requirement: Engage ends
An engaged Warrior SHALL not swing. Engage SHALL end when the Warrior cannot pay its endurance, when they target another enemy, or when they start to attack. Status: reference.

#### Scenario: Attacking ends it
- **WHEN** an engaged Warrior starts to attack
- **THEN** Engage ends

### Requirement: Berserk
Berserk SHALL be a timed stance of the Berserker that raises melee damage and lowers the user's own evade, parry and block for its duration, with a reuse timer. Its amounts per ability level SHALL be in the rule book with sources and a status. No source found gives them (a later developer post says Berserk made a Berserker's blows critical, which this project has not adopted), so they are provisional.

#### Scenario: Berserk raises damage
- **WHEN** a Berserker with Berserk active hits the same target with the same weapon as without it
- **THEN** the base damage is higher by the rule book's factor, and the Berserker's evade, parry and block chances are lower by the rule book's share

#### Scenario: Reuse
- **WHEN** a Berserker uses Berserk again before its reuse timer ends
- **THEN** it is refused with the remaining time

### Requirement: Charge
Charge SHALL raise the user's run speed for a short time even in combat, with a reuse timer. Slows and snares SHALL not apply while it lasts.

#### Scenario: Charge in combat
- **WHEN** a Savage in combat uses Charge
- **THEN** their speed rises for its duration and a snare cast on them during it has no effect

### Requirement: QuickCast
QuickCast SHALL make the user's next spell cast with no cast time, with a reuse timer. It SHALL end when that spell goes off or after its window.

#### Scenario: Instant next spell
- **WHEN** a Runemaster uses QuickCast and then casts a bolt
- **THEN** the bolt fires without a cast bar, and the next spell after it has its normal cast time

### Requirement: Enhanced Evade is left out, with its reason
Enhanced Evade SHALL stay unimplemented while no source says what it did or that 1.65 had it. The parity census SHALL list it with that reason and the date searched.

#### Scenario: Census
- **WHEN** `cmd/parity` runs
- **THEN** Enhanced Evade is listed as a gap with its reason

### Requirement: Safe Fall is counted
Safe Fall SHALL be reported by the parity census as implemented, because the falling rules already reduce damage by Stealth specialization.

#### Scenario: Census Safe Fall
- **WHEN** `cmd/parity` runs
- **THEN** Safe Fall is listed as implemented

### Requirement: Danger Sense
Danger Sense SHALL warn its holder when a hidden enemy-realm player comes within its range. It is inert until a second realm can be in a zone with the player.

#### Scenario: Warning
- **WHEN** a hidden enemy player comes within range of a Shadowblade with Danger Sense
- **THEN** the Shadowblade reads "You sense danger nearby!" once per approach

### Requirement: Distraction
Distraction SHALL be aimed at a ground target. Idle creatures in its radius SHALL turn to face the point.

#### Scenario: Creatures look away
- **WHEN** a Shadowblade uses Distraction on ground near two idle creatures
- **THEN** both turn to face the ground target

### Requirement: Registered and tested
Each rule in this change SHALL have a table test per numbered scenario and a registry entry with its sources and status, so `cmd/parity` counts it.

#### Scenario: Parity census
- **WHEN** `cmd/parity` runs after this change
- **THEN** abilities read 27 of 30, and each remaining gap (Climb Walls, Enhanced Evade, Call of a thousand storms) has a reason
