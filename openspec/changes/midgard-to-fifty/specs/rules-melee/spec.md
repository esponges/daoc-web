## ADDED Requirements

### Requirement: A left-hand weapon swings alongside the main hand
A player who wears a left-hand weapon they are trained for SHALL, on each main-hand swing, also swing the left hand, at the chance and damage the left-hand rule gives from their Left Axe (or Hand to Hand) spec. Each hand SHALL roll its own hit, defences and damage. A swing with both SHALL take the swing time the rule gives for both weapons. Status: set by the findings.

#### Scenario: A Shadowblade with two axes
- **WHEN** a Shadowblade with Left Axe 20 swings with an axe in each hand
- **THEN** the main hand swings, and the left hand swings at the chance the rule gives, each with its own result

#### Scenario: Untrained
- **WHEN** a Warrior holds an axe in the left hand
- **THEN** it cannot be worn in the left-hand slot, and no left-hand swing happens

### Requirement: Stealth openers need stealth
A style whose opening is "from stealth" SHALL be usable only on the first swing that ends the user's stealth. Otherwise it SHALL be refused as any other style whose opening is not met.

#### Scenario: From stealth
- **WHEN** a hidden Shadowblade opens with a stealth-opener style
- **THEN** the style fires, with its growth damage and effects

#### Scenario: Not hidden
- **WHEN** the same Shadowblade, not hidden, tries the style
- **THEN** it is refused because its opening is not met
