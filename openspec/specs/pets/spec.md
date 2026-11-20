# pets Specification

## Purpose
Lets a player own a creature that follows them and fights on command, through summoning (Spiritmaster, Bonedancer, Hunter) or charm (Hunter), with the owner's commands, the pet window, and how a pet takes part in experience and aggro.

## Requirements

### Requirement: A summon spell creates a pet from its template
A summon spell SHALL create one creature from the npc template the spell names, at the level the spell's value gives. It SHALL stand beside the caster and be owned by them. A player SHALL own at most one pet, except a Bonedancer's commander and its minions. Summoning again SHALL release the old pet first. Status: reference.

#### Scenario: A Spiritmaster summons
- **WHEN** a level 10 Spiritmaster casts their summoning spell
- **THEN** a pet of the spell's template appears beside them, at the level the rule gives, shown to every nearby player with the owner's name

#### Scenario: Summoning again
- **WHEN** the Spiritmaster summons while a pet is up
- **THEN** the first pet is released and the new one appears

#### Scenario: Template model not converted
- **WHEN** the template's model has no converted character in `web/data/char/`
- **THEN** the server still creates the pet and the client draws it with the fallback model

### Requirement: The owner commands the pet
The owner SHALL be able to tell the pet to attack their target, follow, stay or release, and to set its stance to passive, defensive or aggressive. A command SHALL work from the pet window and from `/pet <command>` in chat. A command from anyone but the owner SHALL be refused.

#### Scenario: Attack
- **WHEN** the owner targets a creature and gives "attack"
- **THEN** the pet runs to that creature and fights it until it dies or the owner gives another command

#### Scenario: Unreachable target
- **WHEN** the owner orders an attack on a creature across water or a wall the pet cannot cross
- **THEN** after 4 seconds without headway the pet gives it up, returns to following, says so to the owner, and leaves that creature alone for 15 seconds

#### Scenario: Defensive
- **WHEN** a defensive pet's owner is attacked by a creature
- **THEN** the pet attacks that creature

#### Scenario: Passive
- **WHEN** a passive pet's owner is attacked
- **THEN** the pet keeps following and does not fight

### Requirement: A pet follows its owner and leaves with them
A following pet SHALL keep near its owner at the owner's speed and through zone exits within the same region. The pet SHALL be released when its owner dies, logs out or leaves the region, or when the owner is beyond the pet's leash range for longer than the rule allows.

#### Scenario: Owner runs
- **WHEN** the owner runs 2000 units across open ground
- **THEN** the pet follows and stays within its follow distance

#### Scenario: Owner dies
- **WHEN** the owner dies
- **THEN** the pet is released and disappears

### Requirement: A pet's fight is its owner's fight
Damage a pet deals SHALL count as the owner's for creature hate and for experience. A creature that hates a pet SHALL also count the owner as an enemy. A kill made by the pet SHALL give experience to the owner (and the owner's group) as if the owner had made it. Status: reference.

#### Scenario: The pet kills
- **WHEN** a pet alone kills a creature that cons yellow to its owner
- **THEN** the owner gains the experience for a yellow kill, and the corpse's loot belongs to the owner

#### Scenario: Creature turns
- **WHEN** a creature fighting a pet takes damage from the owner's spell
- **THEN** its hate list holds both the pet and the owner

### Requirement: Charm takes an existing creature as a pet
A charm spell SHALL make a creature of an allowed kind, not above the spell's level limit, into the caster's pet. The charm SHALL be able to break by the rule in the rule book, and a broken charm SHALL turn the creature hostile to its former owner. Status: set by the findings.

#### Scenario: Charm an animal
- **WHEN** a Hunter charms a wolf within their charm's level limit
- **THEN** the wolf becomes their pet and stops fighting them

#### Scenario: Too high
- **WHEN** the Hunter tries to charm an animal above the limit
- **THEN** the spell is refused with a message saying so

#### Scenario: Wrong kind
- **WHEN** the Hunter tries to charm a creature that is not an animal
- **THEN** the spell is refused

### Requirement: A Bonedancer commands minions through a commander
A Bonedancer's commander SHALL accept up to three minions summoned while it is up. The minions SHALL follow the commander and take its attack orders. Minion spells SHALL refuse to cast with no commander. Releasing the commander SHALL release its minions. Status: reference.

#### Scenario: Full army
- **WHEN** a Bonedancer with a commander summons three minions
- **THEN** all three follow the commander, and a fourth summon is refused

#### Scenario: Commander dies
- **WHEN** the commander is killed
- **THEN** its minions are released

### Requirement: The pet window shows the pet and its commands
The owner SHALL see a pet window with the pet's name, level, health, stance and command buttons, using the stock interface when `web/data/ui/` is present and a plain panel otherwise. The window SHALL close when the pet is released.

#### Scenario: Pet health
- **WHEN** the pet takes damage
- **THEN** the pet window's health bar falls at the same moment

#### Scenario: Fallback
- **WHEN** `web/data/ui/` is absent
- **THEN** a plain pet panel offers the same commands
