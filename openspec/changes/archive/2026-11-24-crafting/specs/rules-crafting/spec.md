## ADDED Requirements

### Requirement: Success chance
The chance to succeed SHALL depend on the difference between the recipe's level and the crafter's skill: 100 percent at or below the skill, a point less for each point above it, and no attempt from 50 above. Status: provisional (no public source gives the 1.65 curve).

#### Scenario: Orange recipe
- **WHEN** a crafter's skill is 10 below a recipe's level
- **THEN** the success chance is 90 percent, below the 100 percent of an equal-level recipe

#### Scenario: Out of reach
- **WHEN** a recipe is more than 50 above the crafter's skill
- **THEN** it cannot be tried

### Requirement: Quality
A successful craft SHALL yield a quality from 94 to 100 percent: each point above 94 is won with a probability of 50 percent at equal level, a point of percent more for each ten skill above the recipe, within 20 and 80 percent. 100 percent is a masterpiece.

#### Scenario: Distribution
- **WHEN** crafts are simulated
- **THEN** every quality from 94 to 99 occurs, masterpieces are rare, and a more skilled crafter does better

### Requirement: Craft time
Craft time SHALL be 2 seconds plus one for each 50 levels of the recipe and each ten materials, between 3 and 60 seconds.

#### Scenario: Time
- **WHEN** a level 500 recipe with 20 materials is crafted
- **THEN** it takes 14 seconds

### Requirement: Skill gain and the secondary cap
A successful craft SHALL raise the skill by one with a certain chance while the recipe is at or above the skill, falling a point for each point the skill is above it, and none from 100 above. A failed craft SHALL use up half of each material stack, rounded down. A secondary skill SHALL not pass the primary's; a primary stops at 1100.

#### Scenario: Gain
- **WHEN** a crafter of skill 150 makes a level 100 recipe
- **THEN** the chance to gain is 50 percent

#### Scenario: Secondary cap
- **WHEN** a secondary skill equals the primary's
- **THEN** it does not rise

### Requirement: Registered and tested
Each crafting rule SHALL have table tests per scenario and a registry entry with sources and status.

#### Scenario: Parity census
- **WHEN** `cmd/parity` runs
- **THEN** the crafting rules are listed
