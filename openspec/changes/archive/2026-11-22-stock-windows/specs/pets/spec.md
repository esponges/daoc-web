# Spec Delta

## MODIFIED Requirements

### Requirement: The pet window shows the pet and its commands
The owner SHALL see a pet window with the pet's name, level, health, stance and command buttons. When `web/data/ui/` is present it SHALL be the stock `MiniPet` window: the pet's name in its title, health in its bar, stance and commands in its checkboxes and buttons. Otherwise it SHALL be a plain panel. The window SHALL close when the pet is released.

#### Scenario: Pet health
- **WHEN** the pet takes damage
- **THEN** the pet window's health bar falls at the same moment

#### Scenario: Stance
- **WHEN** the owner ticks the passive checkbox in the stock pet window
- **THEN** the passive command is sent and, once the server confirms, only that stance checkbox is ticked

#### Scenario: Fallback
- **WHEN** `web/data/ui/` is absent
- **THEN** a plain pet panel offers the same commands
