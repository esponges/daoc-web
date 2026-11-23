## MODIFIED Requirements

### Requirement: Power regeneration
Power SHALL regenerate on a fixed tick of three seconds, by five plus the caster's level divided by 2.75, in combat or out of it, and twice that while seated out of combat (`rules-resting`), and SHALL not regenerate while the caster is dead. Casting SHALL spend the spell's power cost when the cast completes. Status: provisional for combat: the community servers state this rate for the time after patch 1.87, which made combat ticks twice as effective, so 1.65 combat regeneration was lower by an amount no source gives.

#### Scenario: Resting
- **WHEN** a level 50 Runemaster sits for a minute
- **THEN** their power is higher by about 46 times the ticks passed

#### Scenario: Interrupted cast
- **WHEN** a cast is interrupted
- **THEN** no power is spent
