## Purpose

Defines classic tasks: short kill and delivery jobs that Midgard town NPCs hand out, how a player takes, holds, finishes or abandons one, and the experience and money they pay.

## ADDED Requirements

### Requirement: Town NPCs offer tasks
An NPC that offers tasks SHALL give a player who asks for one either a kill task or a delivery task. The task SHALL be refused when the player already holds a task, or has done the most tasks the rule allows at their level. The NPCs that offer tasks SHALL be chosen by the rule recorded with its source. Status: set by the findings.

#### Scenario: Asking for a task
- **WHEN** a level 8 player with no task asks a task-giving NPC in Mularn for one
- **THEN** they receive a kill or delivery task, shown in their task window

#### Scenario: Already holding one
- **WHEN** the same player asks again while holding the task
- **THEN** the NPC refuses and says they already have a task

### Requirement: A kill task names a fitting creature nearby
A kill task SHALL name a creature placed in a zone within the region that cons between blue and orange to the player. The task SHALL be finished by killing one such creature, alone or with a group, and then returning to the NPC that gave it. Status: set by the findings.

#### Scenario: Killing the target
- **WHEN** the player kills a creature of the named kind
- **THEN** the task window marks it done, and the giver accepts it on return

### Requirement: A delivery task carries an item to another NPC
A delivery task SHALL put a task item in the player's pack and name another NPC in the region. Giving the item to that NPC SHALL finish the task. The task item SHALL not be tradeable, sellable or droppable.

#### Scenario: Delivering
- **WHEN** the player gives the task item to the named NPC
- **THEN** the item leaves their pack and the task is finished

### Requirement: Tasks expire and can be abandoned
A task SHALL expire after the time limit the rule gives. A player SHALL be able to abandon a task. Either way, any task item SHALL be removed and nothing paid. Status: set by the findings.

#### Scenario: Abandoning
- **WHEN** a player abandons a delivery task
- **THEN** the task item is removed and they may ask for a new task

### Requirement: A finished task pays experience and money
A finished task SHALL pay experience and money from the player's level by the rule in the rule book, with a worked example per level band. Status: set by the findings.

#### Scenario: Reward at level 10
- **WHEN** a level 10 player finishes a kill task
- **THEN** they gain the experience and money the rule book's worked example gives for level 10

### Requirement: Tasks survive a reconnect
A held task, its progress and its time limit SHALL be stored with the character and SHALL still be there after logging out and in.

#### Scenario: Reconnect
- **WHEN** a player with a half-done kill task logs out and back in before it expires
- **THEN** the task window shows the same task and progress
