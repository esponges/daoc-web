# Spec Delta

## ADDED Requirements

### Requirement: Guards defend their realm
A guard SHALL attack an aggressive creature of another realm that comes within the guard's aggro range, or the imported default of 750 units when the guard has none. It SHALL also attack such a creature that is fighting a player of the guard's realm within that range. It SHALL chase it up to its leash, as other NPCs do, and SHALL then return to its post. Status: reference.

#### Scenario: A creature follows a player into town
- **WHEN** an aggressive wolf chases a Midgard player past a Mularn guard within its range
- **THEN** the guard attacks the wolf, and the wolf turns on whoever it hates most

#### Scenario: Friendly creatures are left alone
- **WHEN** a non-aggressive creature wanders near a guard
- **THEN** the guard does not attack it

### Requirement: Guard kills give no experience to players
A creature killed with damage from a guard SHALL count the guard's damage as an unshared damager. Players SHALL get only their own share of the damage. Status: reference.

#### Scenario: A guard does half
- **WHEN** a guard deals 50% of a level 10 creature's damage and a level 10 Warrior deals 50%
- **THEN** the Warrior gains 2,560 and the other half is given to no one
