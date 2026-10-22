# Spec Delta

## Purpose

Lets players of the same realm form a group: invite, accept, leave, disband and lead, with members seeing each other's state. Group experience is in `rules-experience`.

## ADDED Requirements

### Requirement: Players invite and accept
A player SHALL be able to invite another player of their realm, by target or by name with `/invite <name>`. The invited player SHALL be asked, and SHALL join only by accepting. An invitation SHALL lapse after 60 seconds or when declined.

#### Scenario: Forming a group
- **WHEN** Ragnar invites Sigrid and she accepts
- **THEN** both are in one group with Ragnar as leader, and both see the other in their group window

#### Scenario: Declining
- **WHEN** Sigrid declines or does not answer within 60 seconds
- **THEN** no group is formed and Ragnar is told she declined

#### Scenario: Another realm
- **WHEN** a player tries to invite a player of another realm
- **THEN** the server refuses with a message and nothing is sent to the other player

### Requirement: Only the leader invites, and a group holds eight
Only the leader SHALL invite into an existing group. A group SHALL hold at most 8 members. A player SHALL be in at most one group, and SHALL NOT be invited while in another.

#### Scenario: A full group
- **WHEN** the leader of an 8-member group invites a ninth player
- **THEN** the server refuses: the group is full

#### Scenario: Already grouped
- **WHEN** a player in another group is invited
- **THEN** the inviter is told the player is already in a group

### Requirement: Leaving, removing and disbanding
A member SHALL be able to leave with `/disband`. The leader SHALL be able to remove a member. When the leader leaves, the longest-standing member SHALL lead. A group left with one member SHALL end. A member who logs out SHALL leave the group.

#### Scenario: The leader leaves
- **WHEN** the leader of a three-member group leaves
- **THEN** the member who joined first becomes leader and both remaining members are told

#### Scenario: Down to one
- **WHEN** one of two members leaves
- **THEN** the group ends and the other is told

### Requirement: Members see each other
Each member SHALL receive, for every other member, their name, level, class, health, endurance, zone, and whether they are within 6,000 units of them, updated at least once a second while it changes. No player outside the group SHALL receive these.

#### Scenario: A member is hurt
- **WHEN** a member takes damage
- **THEN** the other members' group windows show their health lower within a second

#### Scenario: A member in another zone
- **WHEN** a member walks into another zone
- **THEN** they stay in the group, and the others see them as out of range with the zone named

### Requirement: The group window
The client SHALL show the group in the stock group window (`StatsGroup`, and `MiniGroup` when minimised), with the invitation drawn on the stock dialog frame with stock buttons. (`new_group`, first named for the invitation, is a compact list of members' health bars, not a dialog.) Without converted UI data it SHALL show a plain group panel with the same information.

#### Scenario: Stock window
- **WHEN** a player in a group presses GROUP on the menu bar
- **THEN** the stock group window lists each member with name, level and a health bar

#### Scenario: No converted UI
- **WHEN** `web/data/ui/` is absent
- **THEN** a plain panel lists the members and the invitation is a plain prompt
