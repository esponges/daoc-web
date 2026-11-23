## ADDED Requirements

### Requirement: One realm per character
A character SHALL belong to the realm chosen at creation and SHALL never change realm. An account MAY hold characters of different realms.

#### Scenario: Creation
- **WHEN** a player creates a character
- **THEN** they first choose Albion, Midgard or Hibernia (each that is playable), and the races and classes offered follow

### Requirement: Social actions stay within a realm
Tells, say, groups, guilds, trade and duels SHALL only happen between characters of the same realm. Speech from another realm SHALL reach the listener garbled, as in the game.

#### Scenario: A tell across realms
- **WHEN** a Midgard character sends a tell to an Albion character
- **THEN** it is refused as if the name were not online

### Requirement: Enemies are marked
A character of another realm SHALL be shown with the enemy name colour, and SHALL be attackable only where realm-versus-realm combat is allowed.

#### Scenario: Seen in a frontier
- **WHEN** a Midgard player sees an Albion player
- **THEN** the name shows in the enemy colour
