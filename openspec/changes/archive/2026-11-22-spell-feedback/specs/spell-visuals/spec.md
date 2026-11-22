# Spec Delta

## Purpose

Shows spells in the world the way the game's own client does: caster animations, caster effects, projectiles and target effects, all converted from the player's install and played when the server's spell events arrive.

## ADDED Requirements

### Requirement: Spell effects are converted from the install
A converter SHALL read the client's spell, effect and effect-model tables from the install. For every client effect the imported spells use, it SHALL write the caster animations, caster effects, projectile and target effect, each effect with its model, attach bone, duration and rate. Nothing converted SHALL be committed.

#### Scenario: Minor Dusk Dart
- **WHEN** the converter runs over an install and the imported spells include Minor Dusk Dart (client effect 2501)
- **THEN** the output holds effect 2501 with caster animations 41, 42 and 43, caster effects 160 and 161, and target effect 256 with its model, bone and duration

### Requirement: Effect models are converted or listed
Effect models SHALL be converted with the existing model and particle readers. An effect with parts the converter cannot draw SHALL be listed by name and part, and its other parts SHALL still be written.

#### Scenario: An effect part it cannot draw
- **WHEN** an effect model holds a node type the readers do not support
- **THEN** the converter lists the effect and the node type, and writes the effect's other parts

### Requirement: A cast shows on its caster
When a cast begins, the caster SHALL play the spell's starting cast animation, then hold its loop until the spell lands, fails or is interrupted, then play its release. The spell's caster effects SHALL play on the bones the table names for the length of the cast. This SHALL hold for the player's own character, other players and creatures. A figure without the clips SHALL play the generic cast clip, or none.

#### Scenario: The player casts
- **WHEN** the player begins a 2.6-second Minor Dusk Dart
- **THEN** their character plays the start animation, holds the loop with the caster effects on its hands, and plays the release when the spell lands

#### Scenario: Interrupted
- **WHEN** a cast is interrupted
- **THEN** the loop and the caster effects stop and the character returns to its stance

### Requirement: A spell shows where it lands
When a spell lands, its projectile, if it has one, SHALL fly from the caster to the target, and the target effect SHALL play on the target on the bone the table names for the effect's duration. A resisted spell SHALL show its projectile but not its target effect. Effects SHALL be drawn only within the drawing distance of the view.

#### Scenario: A bolt
- **WHEN** a bolt spell lands on a creature
- **THEN** the bolt flies from the caster's hand to the creature and the hit effect plays on it

#### Scenario: A heal
- **WHEN** a heal lands on a player
- **THEN** the heal's effect plays on that player for its duration

### Requirement: Arrows fly
When the server reports an arrow shot, an arrow SHALL fly from the archer to the target and be gone on arrival, whether the shot hits or misses.

#### Scenario: A shot
- **WHEN** a Hunter's arrow is loosed at a creature
- **THEN** an arrow flies from the Hunter to the creature

### Requirement: Spells without converted effects still work
When the effect data is absent, or a spell's effect was not converted, the client SHALL play the generic cast clip if the figure has one and SHALL otherwise behave as before, with no error.

#### Scenario: No effect data
- **WHEN** the client starts without the converted effect data
- **THEN** spells cast, land and log as they do today
