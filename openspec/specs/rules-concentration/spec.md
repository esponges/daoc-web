# rules-concentration Specification

## Purpose
Defines the concentration pool that limits a caster's long-lasting buffs, which spells draw on it, how long such a buff holds, and how the caster sees and ends what they hold.

## Requirements

### Requirement: Casters have a concentration pool
Every class with a spell line SHALL have a concentration pool whose size grows with level, from the rule recorded in the rule book with its source. Classes without spells SHALL have no pool. The pool is 4.4 times the level, adjusted by the caster's casting stat: a stat over 50 adds (stat − 50) / 2 percent, one under 50 takes away (50 − stat) percent. Status: reference (DOLSharp and OpenDAoC agree).

#### Scenario: A level 20 Healer
- **WHEN** a level 20 Healer's character sheet is sent
- **THEN** it carries the concentration pool size the rule gives for level 20, and the amount in use

#### Scenario: A Warrior
- **WHEN** a Warrior's character sheet is sent
- **THEN** it carries no concentration pool

### Requirement: A concentration spell holds its cost while it is on
A spell with `Concentration` greater than 0 SHALL be refused when its cost exceeds the caster's free concentration. Once it lands, its cost SHALL be held from the pool until the buff ends. The buff SHALL have no timer. Status: reference.

#### Scenario: Room in the pool
- **WHEN** a Shaman with 20 free concentration casts a 6-concentration strength/constitution buff on a group member
- **THEN** the buff lands with no timer and the Shaman's free concentration is 14

#### Scenario: No room
- **WHEN** the same Shaman with 4 free concentration casts it
- **THEN** the cast is refused with "This spell requires 6 concentration points to cast!" and no power is spent

#### Scenario: Recasting on the same target
- **WHEN** the Shaman casts the same buff again on the same target
- **THEN** the old buff is replaced and its concentration is returned first

### Requirement: A concentration buff ends with range, death or cancel
A concentration buff SHALL end and return its cost when:
- the caster cancels it;
- the caster dies;
- the caster or the target leaves the zone;
- the target is out of the caster's concentration range (5,000 units) for longer than the grace time (20 seconds).

A self-buff SHALL ignore range. The range is 5,000 units and the grace time 20 seconds. Status: reference for the range (OpenDAoC's default), provisional for the grace time.

#### Scenario: The target runs off
- **WHEN** a buffed group member stays beyond the concentration range past the grace time
- **THEN** the buff ends on them and the caster's concentration is returned

#### Scenario: The caster dies
- **WHEN** a Healer holding three concentration buffs dies
- **THEN** all three end and the targets' effects lists drop them

### Requirement: The caster sees and can cancel what they hold
A caster SHALL see a list of their concentration buffs, each with its spell and target, and SHALL be able to cancel any one of them. When `web/data/ui/` is present the list SHALL be the stock `Concentration` window: concentration used and available in its bar and label, and the buffs in its list box. Otherwise it SHALL be a plain panel.

#### Scenario: Cancelling from the list
- **WHEN** a Healer cancels one entry in their concentration list
- **THEN** that buff ends on its target and the freed concentration shows at once

#### Scenario: The stock window
- **WHEN** a Healer holding two concentration buffs opens the concentration window with the converted interface present
- **THEN** the stock Concentration window lists both buffs with their targets and its bar shows the concentration they use

#### Scenario: No converted interface
- **WHEN** `web/data/ui/` is absent
- **THEN** the concentration list draws as a plain panel with the same entries and cancel buttons
