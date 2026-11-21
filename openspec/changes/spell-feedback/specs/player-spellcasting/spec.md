# Spec Delta

## ADDED Requirements

### Requirement: One press asks for one cast
A double-click on a spell, or a key held down on a quickbar slot, SHALL send one cast request. A second request SHALL be sent only for a separate press made after the first cast has ended, failed or been interrupted.

#### Scenario: Double-click
- **WHEN** the player double-clicks Minor Dusk Dart in the spells window
- **THEN** one cast begins and no "You are already casting a spell." line appears

#### Scenario: Held key
- **WHEN** the player holds a quickbar key down through a cast
- **THEN** one cast request is sent
