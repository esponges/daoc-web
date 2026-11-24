## ADDED Requirements

### Requirement: Starting stat points
Creation SHALL give 30 points to add to the race's base stats. Raising a stat SHALL cost 1 point a step up to +10 over base, 2 a step from +11 to +15, and 3 a step above +15. The server SHALL refuse a distribution that costs more than 30. Status: documented once a 1.65 source confirms the costs; provisional until then.

#### Scenario: Costs
- **WHEN** a player adds +12 to Strength
- **THEN** it costs 10 + 2×2 = 14 points, leaving 16

#### Scenario: Overspent
- **WHEN** a creation request spends 31 points
- **THEN** creation is refused

#### Scenario: Characters before this change
- **WHEN** a character made before this change logs in
- **THEN** its stats are unchanged
