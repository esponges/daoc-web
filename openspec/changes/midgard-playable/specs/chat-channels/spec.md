# Spec Delta

## Purpose

Adds the private and group chat a 1.65 player expects, alongside say and broadcast: a tell to one player and a channel for the group.

## ADDED Requirements

### Requirement: A tell reaches one player
`/send <name> <text>` and `/tell <name> <text>` SHALL deliver the text only to the named player of the sender's realm who is in the game, in any zone. The sender SHALL see their own line as sent. A name SHALL match without regard to case.

#### Scenario: A tell across zones
- **WHEN** Ragnar in Vale of Mularn sends a tell to Sigrid in Myrkwood Forest
- **THEN** Sigrid sees "Ragnar sends, ..." and Ragnar sees "You send, ..." and no one else receives it

#### Scenario: Not in the game
- **WHEN** the named player is not logged in or does not exist
- **THEN** the sender is told that the player is not in the game, and nothing is sent

### Requirement: Group chat reaches the group
`/g <text>` SHALL deliver the text to every member of the sender's group, in any zone, and to no one else. A player not in a group SHALL be told so.

#### Scenario: Group line
- **WHEN** a member of a three-member group types `/g pull now`
- **THEN** all three see the line on the group channel and a player beside them outside the group does not

#### Scenario: No group
- **WHEN** a player who is not in a group types `/g hello`
- **THEN** they are told they are not in a group

### Requirement: Channels are told apart
Each chat line SHALL carry its channel (say, broadcast, tell, group or system), and the client SHALL colour and tab it as the stock chat window does for that channel, with the plain chat as fallback.

#### Scenario: Tabs
- **WHEN** a tell and a group line arrive
- **THEN** they show in their channel colours, and the stock chat window's Group tab shows the group line

### Requirement: Chat limits apply to every channel
The existing per-player rate limit and line length limit SHALL apply to tells and group lines as to say.

#### Scenario: Flooding tells
- **WHEN** a player sends tells faster than the rate limit
- **THEN** they are told they are talking too fast and the extra lines are dropped
