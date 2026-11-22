# Spec Delta

## ADDED Requirements

### Requirement: Spells keep their client effect
The importer SHALL store each spell's client effect number from the spell dump, and the server SHALL send it with the cast, spell, effect and resist events of that spell. A world store imported before this column SHALL load, with every spell's client effect read as zero.

#### Scenario: Minor Dusk Dart
- **WHEN** the spell dump gives Minor Dusk Dart client effect 2501 and a player casts it
- **THEN** the stored spell carries 2501 and the cast and spell events carry it too

#### Scenario: An older world store
- **WHEN** the server starts on a world store without the column
- **THEN** it runs, and its spell events carry no client effect
