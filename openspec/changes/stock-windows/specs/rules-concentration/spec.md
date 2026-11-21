# Spec Delta

## MODIFIED Requirements

### Requirement: The caster sees and can cancel what they hold
A caster SHALL see a list of their concentration buffs, each with its spell and target, and SHALL be able to cancel any one of them. When `web/data/ui/` is present the list SHALL be the stock `Concentration` window: concentration used and available in its bar and label, and the buffs in its list box. Otherwise it SHALL be a plain panel.

#### Scenario: Cancelling from the list
- **WHEN** a Healer cancels one entry in their concentration list
- **THEN** that buff ends on its target and the freed concentration shows at once

#### Scenario: The stock window
- **WHEN** a Healer holding two concentration buffs opens the concentration window with the converted interface present
- **THEN** the stock Concentration window lists both buffs with their targets and its bar shows the concentration they use

#### Scenario: No converted interface
- **WHEN** `web/data/ui/` is absent
- **THEN** the concentration list draws as a plain panel with the same entries and cancel buttons
