# quests Specification

## Purpose
Quests as data: steps, conditions, rewards, quest drops, the journal, abandoning, and how content is authored and sourced.

## Requirements

### Requirement: Quest offer
An NPC that gives a quest SHALL offer it in its talk text to a player who meets its conditions and has not done it. A player who does not meet them SHALL see no offer.

#### Scenario: Offer
- **WHEN** a level 5 Warrior talks to the Warrior trainer and the epic's first step is open to them
- **THEN** the talk text offers the quest, and accepting adds it to the journal

#### Scenario: Wrong class
- **WHEN** a Healer talks to the same NPC
- **THEN** no offer appears

### Requirement: Steps advance
Each step SHALL advance when its goal is met:
- talking to the named NPC;
- killing the count of the named creature (a group kill counts for members in range);
- holding the named item;
- delivering it;
- reaching the place;
- or using the object.

The journal SHALL show the current step.

#### Scenario: Kill step
- **WHEN** a step asks for 5 of a creature and the player kills the fifth
- **THEN** the step completes and the journal shows the next step

### Requirement: Quest drops
An item that a step needs SHALL drop from its creature only for players on that step, and SHALL not drop otherwise.

#### Scenario: Not on the quest
- **WHEN** a player not on the quest kills the creature
- **THEN** the quest item does not drop

### Requirement: Rewards
Finishing a quest SHALL give its experience (by the experience rules' quest share), money and items. When there is a choice, the player picks one. A full bag SHALL hold the reward until there is room.

#### Scenario: Reward choice
- **WHEN** a player finishes a quest offering a sword or a shield
- **THEN** they pick one and receive it

### Requirement: Abandon and repeat
A player SHALL be able to abandon an active quest, losing its progress. A finished quest SHALL not be offered again unless it is marked repeatable.

#### Scenario: Abandon
- **WHEN** a player abandons a quest on step 3
- **THEN** it leaves the journal and can be taken again from the start

### Requirement: Content is authored and sourced
Each quest SHALL be a data file written by this project, naming its source. Text SHALL not be copied from a source or from a GPL server. A test SHALL check every quest file against the schema and that each NPC, creature and item it names exists in the world data.

#### Scenario: Broken reference
- **WHEN** a quest names a creature the world data does not have
- **THEN** the quest data test fails, naming the quest and the creature
