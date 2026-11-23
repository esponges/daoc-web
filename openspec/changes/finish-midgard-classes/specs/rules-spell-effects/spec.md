## ADDED Requirements

### Requirement: Savage self-buffs
The Savage buffs SHALL apply only to the caster, SHALL cost health rather than power (the amount in the spell's data, as a share of maximum health where the data says so), and SHALL be refused when the cost would kill the caster. Each kind SHALL raise its own stat:
- the resistance buffs raise crush, slash or thrust resistance;
- the DPS buff raises weapon damage;
- the evade and parry buffs raise those chances;
- the combat speed buff shortens swing time.

#### Scenario: Health cost
- **WHEN** a Savage casts a parry buff that costs 10 percent of health
- **THEN** they lose that health, gain the parry bonus for the spell's duration, and their power is unchanged

#### Scenario: Not enough health
- **WHEN** a Savage whose health is below the cost casts a Savage buff
- **THEN** it is refused and nothing is spent

#### Scenario: Combat speed
- **WHEN** a Savage with the combat speed buff swings the same weapon
- **THEN** the swing time is shorter by the buff's amount, within the swing speed floor

### Requirement: Savage endurance heal
The Savage endurance heal SHALL restore endurance to the caster by the spell's value, paid in health.

#### Scenario: Endurance restored
- **WHEN** a Savage at low endurance casts the endurance heal
- **THEN** their endurance rises by the spell's amount and their health falls by its cost

### Requirement: Taunt
Taunt SHALL add the spell's value to the target creature's hate for the caster, and SHALL do no damage.

#### Scenario: Pulling hate
- **WHEN** a creature is fighting a Healer and a Savage taunts it enough to lead its hate list
- **THEN** the creature turns to the Savage
