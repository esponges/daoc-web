# friends-and-social Specification

## Purpose
The friends and ignore lists, the anon, away and looking-for-group marks, and assist, follow, face and inspect.

## Requirements

### Requirement: Friends list
Each character SHALL keep a friends list of character names on the server, shown with online status, and SHALL be told when a friend logs in or out.

#### Scenario: Friend comes online
- **WHEN** Bjorn, on Thora's friends list, logs in
- **THEN** Thora reads "Bjorn has entered the game." and the Friends window shows him online

### Requirement: Anonymous, away and looking for group
/anon SHALL hide the player's zone, level and class from /who for everyone but their group. /afk SHALL answer tells with the away message until the player acts. /lfg SHALL mark the player in /who.

#### Scenario: Away reply
- **WHEN** Thora is /afk "back soon" and Bjorn sends her a tell
- **THEN** Bjorn reads "Thora is AFK: back soon"

### Requirement: Ignore
/ignore SHALL hide another character's say, shout, emote and tells from the ignoring player. It SHALL be kept on the server.

#### Scenario: Ignored tell
- **WHEN** Thora ignores Bjorn and Bjorn sends her a tell
- **THEN** Thora sees nothing

### Requirement: Assist, follow, face
/assist SHALL set the player's target to their target's target. /follow SHALL move the player behind their target at up to their run speed until they move or the target is lost. /face SHALL turn the player toward their target.

#### Scenario: Follow
- **WHEN** Thora follows Bjorn and Bjorn walks away
- **THEN** Thora moves after him, keeping close, until she presses a movement key

### Requirement: Inspect
/inspect on a player in range SHALL show their worn items.

#### Scenario: Inspect
- **WHEN** Thora inspects Bjorn
- **THEN** she sees Bjorn's equipped items and their names

### Requirement: Stick
`/stick` (and the Command window's Stick) SHALL keep the player within melee range of their target, player or creature, facing it, until they press a movement key, the target dies or goes beyond range, or they give the command again.

#### Scenario: Sticking to a fleeing creature
- **WHEN** a player sticks to a creature that runs away
- **THEN** the player follows it, stays in melee range and keeps facing it

#### Scenario: Breaking off
- **WHEN** the player presses a movement key while sticking
- **THEN** sticking ends with a message
