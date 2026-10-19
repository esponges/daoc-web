# rules-melee Specification

## Purpose
Defines a melee attack in classic 1.65 terms, from the swing timer to the damage dealt: defences, misses, hit location, weapon skill against armour, variance, caps, critical hits and styles.

## Requirements

### Requirement: A swing resolves in a fixed order
Each swing SHALL be resolved in this order:
1. the defender's evade;
2. their parry;
3. their block;
4. the attacker's miss chance;
5. if it hits, a hit location and the damage.

The first roll that succeeds decides the outcome. Each roll SHALL be a uniform number in [0, 1) compared against its chance. The rules SHALL take their rolls from the caller, so a scenario can fix them. Status: reference.

#### Scenario: A parried swing
- **WHEN** the evade roll fails and the parry roll succeeds
- **THEN** the swing is parried, and block and miss are not rolled

### Requirement: Evade
A player SHALL evade only with the Evade ability (level `e`) and only blows from within 60° of their facing. Their chance in tenths of a percent SHALL be ⌊(900 + quickness + dexterity) × e × 5 / 100⌋. With `n` creatures attacking them, it SHALL be divided by 1 + (n − 1) / 3. A creature SHALL evade blows from the front at its template's evade percent. Status: documented (5% a level, modified by quickness and dexterity), reference for the exact form.

#### Scenario: Evade I
- **WHEN** a Warrior with Evade 1, quickness 35 and dexterity 60 is struck from the front by one creature
- **THEN** their evade chance is 4.9%

#### Scenario: From behind
- **WHEN** the same Warrior is struck from behind
- **THEN** they cannot evade it

### Requirement: Parry
A player SHALL parry only while holding a melee weapon, with their Parry spec, and only blows from within 60° of their facing. Their chance in tenths of a percent SHALL be trunc((2 × dexterity − 100) / 4) + (parry spec − 1) × 5 + 50. It SHALL be divided by the number of attackers and halved against a two-handed weapon. A creature SHALL parry blows from the front at its template's parry percent. Status: documented.

#### Scenario: Parry 20
- **WHEN** a Warrior with dexterity 60 and parry 20 is attacked by one creature with a one-handed weapon
- **THEN** their parry chance is 15%

#### Scenario: Two attackers
- **WHEN** two creatures attack the same Warrior
- **THEN** their parry chance against each is 7.5%

#### Scenario: Untrained
- **WHEN** a Warrior with dexterity 45 has parry 1
- **THEN** their parry chance is 4.8%

### Requirement: Block
A player SHALL block only with the Shield ability, a shield in the left hand, a one-handed weapon or none in the right, and only blows from within 60° of their facing. Their chance in tenths of a percent SHALL be trunc((2 × dexterity − 100) / 4) + (shields spec − 1) × 5 + 50, scaled by the shield's quality and condition. A creature SHALL block at its template's block percent. Status: documented.

#### Scenario: Shields 1
- **WHEN** a Warrior with dexterity 60, shields 1 and a 100% quality, 100% condition shield is attacked from the front
- **THEN** their block chance is 5.5%

#### Scenario: Worn shield
- **WHEN** that shield is 90% quality and 80% condition
- **THEN** the block chance is 3.96%

### Requirement: Miss chance
The base miss chance SHALL be 18%. It SHALL be lowered by 4/3 percentage points for each level the attacker has above the defender, and raised the same for each level below. The result SHALL be clamped to 0–100%. Status: documented (the 18% base), reference for the level term.

#### Scenario: Even fight
- **WHEN** a level 10 Warrior swings at a level 10 creature
- **THEN** the miss chance is 18%

#### Scenario: Higher creature
- **WHEN** the creature is level 13
- **THEN** the miss chance is 22%

### Requirement: Bonuses move the miss chance
The miss chance SHALL be lowered by the attacking style's to-hit bonus and raised by the defence bonus of the defender's last styled attack. It SHALL then be scaled by (100 + the defender's armour bonus − the attacker's weapon bonus) / 100. Each item bonus SHALL be capped by level: 0 below level 15, then 5 more for every 5 levels, up to 35 from level 45. Status: reference.

#### Scenario: A style with a to-hit bonus
- **WHEN** the Warrior uses a style with a to-hit bonus of 5 against a level 10 creature
- **THEN** the miss chance is 13%

#### Scenario: An enchanted weapon
- **WHEN** a level 20 Warrior with a weapon bonus of 10% swings at a level 20 creature
- **THEN** the miss chance is 18% × 0.9 = 16.2%

### Requirement: Hit location on a player
A hit on a player SHALL land on the location the style names, if it names one. Otherwise it SHALL roll: torso 40%, legs 25%, arms 15%, head 10%, hands 5%, feet 5%. The armour worn there SHALL be the armour that counts. Status: reference (a public test measured 44/22/16/12/3/3, and the difference is recorded).

#### Scenario: Bare legs
- **WHEN** a hit lands on the legs of a player wearing no leggings
- **THEN** that hit is resolved against no item armour and no absorption

### Requirement: Weapon damage and its cap
A player's base damage SHALL be min(weapon DPS, 1.2 + 0.3 × level) × weapon speed in seconds × (1 + (speed in tenths − 20) × 0.003). For a two-handed weapon it SHALL then be multiplied by 1.1 + 0.005 × spec. The damage cap SHALL be 3 × base damage, set before the weapon's quality and condition multiply the base damage. Status: documented for the DPS cap, reference for the rest.

#### Scenario: Over-level weapon
- **WHEN** a level 10 Warrior uses a 4.5 DPS, 3.0 s sword
- **THEN** the DPS counts as 4.2 and the base damage is 12.978, capped at 38.934

### Requirement: Weapon skill against armour
Weapon skill SHALL be level × class base × 0.005 × (1 + strength × 0.005) × variance. Variance SHALL be uniform from `lo` to `lo + 0.5`, where `lo` = 0.75 + 0.5 × clamp((spec − 1) / (target level + 1), 0, 1). The armour term SHALL be max(1, armour factor / (1 − absorption)). Damage SHALL be min(base × weapon skill / armour term × (1 − resist), cap). Damage dealt SHALL be truncated to a whole number. Status: reference.

#### Scenario: Mid roll on an even creature
- **WHEN** a level 10 Troll Warrior with strength 115 and sword 10 hits a level 10 creature with the sword above, at a variance roll of 0.5
- **THEN** variance is 1.409, the armour term is 27.696 and the hit does 23

#### Scenario: The spread
- **WHEN** the variance roll is 0 and then 1
- **THEN** the same hit does 19 and then 28

#### Scenario: Untrained weapon
- **WHEN** the Warrior's sword is 1 and the roll is 0.5
- **THEN** variance is 1.0 and the hit does 16

#### Scenario: Two-handed
- **WHEN** a level 20 Warrior with strength 125 and axe 20 hits a level 20 creature with a 7.2 DPS, 5.0 s two-handed axe at a roll of 0.5
- **THEN** the hit does 82

### Requirement: Critical hits
A player's melee hit SHALL be critical with a 10% chance. A critical hit SHALL add the damage dealt, styled damage included, times a uniform factor from 0.1 to 1.0 against a creature, or 0.1 to 0.5 against a player. Creatures SHALL NOT score critical hits. Status: reference.

#### Scenario: A big crit
- **WHEN** a 23-point hit on a creature is critical with a factor roll of 1.0
- **THEN** 23 more is added and the target takes 46

### Requirement: Styles add growth damage
A style that lands SHALL add style damage = unrounded hit × growth rate × spec × weapon speed / damage cap. It SHALL be capped at growth rate × spec × weapon speed. A style with a growth rate SHALL add at least 1. The total SHALL be truncated once, after the style is added. A style SHALL only fire if its opening is met when the swing lands. Status: documented.

#### Scenario: A styled hit
- **WHEN** the mid-roll hit above (23.919 before truncation) uses a style with growth rate 0.5 and sword 10
- **THEN** the style adds 23.919 × 0.5 × 10 × 3.0 / 38.934 = 9.215, under its cap of 15, and the hit does 33

### Requirement: Styles cost endurance by weapon speed
A style SHALL cost ⌊weapon speed in tenths × style endurance / 40⌋ endurance, plus 1 if the weapon is faster than 4.0 s, and at least 1. A style whose swing misses or is evaded, parried or blocked SHALL still cost half that, and SHALL NOT fire. A style SHALL NOT be queued without the full cost in hand. Status: reference.

#### Scenario: Quick sword
- **WHEN** a style with endurance 10 is used with a 3.0 s sword
- **THEN** it costs 8 endurance

#### Scenario: Slow axe
- **WHEN** a style with endurance 15 is used with a 5.0 s axe
- **THEN** it costs 18 endurance

### Requirement: Swing speed
A player's time between swings SHALL be weapon speed × (1 − (min(quickness, 250) − 50) × 0.002). It SHALL never be less than 1.5 seconds. Status: documented.

#### Scenario: A slow Troll
- **WHEN** a Troll with quickness 35 swings a 4.0 s hammer
- **THEN** they swing every 4.12 seconds

#### Scenario: Floor
- **WHEN** a character with quickness 100 swings a 1.5 s weapon
- **THEN** they swing every 1.5 seconds

### Requirement: Armour factor of a player's location
The armour factor of the location a hit lands on SHALL be computed as follows:
1. Take the worn item's AF, capped at the level for cloth and at 2 × level for other armour.
2. Multiply it by the item's quality and condition.
3. Cap it again at 2 × level.
4. Add 0.4 × level.

Absorption SHALL be the item's absorb percent: 0 cloth, 10 leather, 19 studded, 27 chain. Status: documented for the absorb values, reference for the rest.

#### Scenario: Studded torso
- **WHEN** a level 10 Warrior is hit on a torso of AF 20, absorb 19%, 100% quality and condition
- **THEN** that hit sees armour factor 24 and absorption 19%
