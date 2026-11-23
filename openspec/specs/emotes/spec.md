# emotes Specification

## Purpose
Players express themselves with the 1.65 client's emotes: a slash command that plays the emote's animation on every nearby client and writes the emote line on the Emote chat channel.

## Requirements

### Requirement: The classic emote set
The server SHALL accept `/<emote>` for each emote in the classic set (`design.md` D6), whose animation the client's `anims.csv` names for player anim sets. An emote outside the set SHALL be answered with "Unknown command", as any other unknown command is.

#### Scenario: A known emote
- **WHEN** a player types `/wave`
- **THEN** the server accepts it as an emote

#### Scenario: A later emote
- **WHEN** a player types an emote the 1.65 client did not have
- **THEN** they are told the command is unknown

### Requirement: Everyone nearby sees the emote
An accepted emote SHALL send an `emote` event with the emote and its target to the players who see the emoting player. Each client SHALL play that emote's clip once on the emoting figure and then return to its posture.

#### Scenario: A bow seen by a friend
- **WHEN** a player types `/bow` with another player nearby
- **THEN** the other player's client plays the bow clip on that figure once

#### Scenario: A clip not converted
- **WHEN** an emote's clip is missing from a figure's conversion
- **THEN** the emote line still appears and the figure stays idle

### Requirement: Emote lines
An emote SHALL write a line on the Emote channel, in second person to the emoter, third person to onlookers within say range, and second person to the target when there is one. Each emote has its own lines with and without a target, written in this project's words.

#### Scenario: Wave at a target
- **WHEN** Thora targets Brann and types `/wave`
- **THEN** Thora reads "You wave at Brann.", Brann reads "Thora waves at you.", and others nearby read "Thora waves at Brann."

#### Scenario: Wave at no one
- **WHEN** Thora has no target and types `/wave`
- **THEN** Thora reads "You wave." and others nearby read "Thora waves."

### Requirement: When an emote is refused
An emote SHALL be refused while the player is dead, in combat, in the air or swimming. A seated player who emotes SHALL stand up first. Emotes SHALL share the chat rate limit.

#### Scenario: In combat
- **WHEN** a player in combat types `/dance`
- **THEN** nothing is played and they are told they cannot do that in combat

#### Scenario: Seated
- **WHEN** a seated player types `/cheer`
- **THEN** they stand up and cheer
