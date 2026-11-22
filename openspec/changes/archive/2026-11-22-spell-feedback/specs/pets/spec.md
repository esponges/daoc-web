# Spec Delta

## ADDED Requirements

### Requirement: The pet's effects show in its window
The server SHALL send the owner the pet's lasting effects, each with its spell name, kind and time left, whenever they change. The pet window SHALL show them as icons with their time left, in the six effect slots of the stock pet window or a row of the plain panel, with the spell name on hover.

#### Scenario: Buffing the pet
- **WHEN** the owner casts a haste buff on their pet
- **THEN** the pet window shows the haste icon with its time counting down, and it goes when the buff ends

#### Scenario: More than six
- **WHEN** the pet holds more than six effects
- **THEN** the six with the least time left show
