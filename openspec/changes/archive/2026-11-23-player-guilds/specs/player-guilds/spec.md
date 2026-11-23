## ADDED Requirements

### Requirement: Forming a guild
A guild SHALL be formed at a Guild Registrar by the leader of a group of 8 players of the same realm, none of them already in a guild, for the registrar's fee. Every member must agree within 30 seconds. The name SHALL be unique and pass the character name rules. The leader SHALL become Guildmaster (rank 0) and the others rank 9.

#### Scenario: Formed
- **WHEN** a leader of 8 unguilded players asks to form "Hammers of Odin", pays, and all agree
- **THEN** the guild exists, the leader is its Guildmaster and the others are members

#### Scenario: Short group
- **WHEN** a leader of 5 players asks to form a guild
- **THEN** it is refused, naming the 8 players needed

### Requirement: Membership
A member with invite permission SHALL invite an unguilded player of the realm, who accepts or declines. A member SHALL be able to leave. A member with remove permission SHALL be able to remove a member of lower rank.

#### Scenario: Invite accepted
- **WHEN** an officer invites Bjorn and he accepts
- **THEN** Bjorn is a rank 9 member and the guild reads "Bjorn has joined the guild."

### Requirement: Ranks and permissions
Each of ranks 0 to 9 SHALL carry permissions:
- guild chat speak and hear;
- officer chat speak and hear;
- invite, promote, remove;
- emblem;
- message of the day.

Only rank 0 SHALL edit permissions. A member SHALL only promote others up to one rank below their own, and demote only members ranked below them.

#### Scenario: Promotion limit
- **WHEN** a rank 3 member with promote permission tries to promote someone to rank 2
- **THEN** it is refused

### Requirement: Guild channels
Guild chat SHALL reach online members whose rank can hear it. Officer chat SHALL reach ranks with officer-hear. Alliance chat SHALL reach every guild in the alliance. The message of the day SHALL be shown at login.

#### Scenario: Officer chat
- **WHEN** an officer speaks on /o
- **THEN** only members whose rank hears officer chat see the line

### Requirement: Emblem
The Guildmaster (or a rank with emblem permission) SHALL choose an emblem number from 1 to 84 at an Emblem NPC, for a fee. The number SHALL be shown in the guild window. Painting it on members' cloaks and shields is not yet done, because the install's emblem textures (`insignia/*_emblem_NNN.dds`) are not yet mapped onto those figures.

#### Scenario: Emblem chosen
- **WHEN** a Guildmaster near an Emblem NPC chooses emblem 12
- **THEN** the fee is taken and every member's guild window shows emblem 12

### Requirement: Alliances
A Guildmaster SHALL be able to invite another guild's Guildmaster into an alliance. The leading guild's Guildmaster SHALL be able to remove a guild or disband the alliance.

#### Scenario: Alliance chat
- **WHEN** two guilds are allied and a member of one speaks on /as
- **THEN** online members of both guilds see it

### Requirement: Guild window and display
The Guild window SHALL list members with rank, level, class, zone (unless anonymous), note and online state. The guild name SHALL appear under the member's name in the world and in /who.

#### Scenario: Name plate
- **WHEN** Thora of "Hammers of Odin" is seen by another player
- **THEN** her name plate shows "<Hammers of Odin>" under her name
