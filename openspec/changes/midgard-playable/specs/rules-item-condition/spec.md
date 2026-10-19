# Spec Delta

## Purpose

Defines how weapons and armour wear in combat, what wear does to them, and what a blacksmith charges to repair them.

## ADDED Requirements

### Requirement: Items have a condition
Each weapon and armour item a character holds SHALL have a condition from 0 to 100%. New items SHALL be at 100%, or at the item template's condition when it gives one. Condition SHALL be saved with the item. Status: reference.

#### Scenario: A new item
- **WHEN** a Warrior buys a weapon
- **THEN** it is at 100% condition, shown in its tooltip

### Requirement: Combat wears items
Each melee blow a character lands SHALL lower their weapon's condition by 0.05 points. Each blow that lands on a character SHALL lower the condition of the armour at the hit location by 0.05 points. A blocked blow SHALL wear the shield instead. Condition SHALL NOT fall below 0. Status: provisional.

#### Scenario: Two hundred blows
- **WHEN** a weapon at 100% lands 200 blows
- **THEN** it is at 90%

#### Scenario: A block
- **WHEN** a Warrior blocks a blow
- **THEN** their shield loses 0.05 points and their armour does not

### Requirement: Wear lowers effectiveness
A worn weapon's damage per second, and a worn armour piece's armour factor, SHALL be multiplied by its condition. Status: reference.

#### Scenario: A weapon at 90%
- **WHEN** a weapon with 16.5 DPS is at 90%
- **THEN** combat uses 14.85 DPS for it

### Requirement: Players are warned of wear
A character SHALL be told when an item they hold falls below 90%, 80% and 70% condition, once at each threshold.

#### Scenario: The 90% warning
- **WHEN** a weapon falls from 90.02% to 89.97%
- **THEN** its owner is told once that it needs repair

### Requirement: Blacksmiths repair for money
A blacksmith SHALL repair an item a player hands them back to 100%. The price SHALL be the item's price × the condition missing ÷ 100 × 0.5, rounded up to a copper. A player SHALL be asked to accept the price first. A player without enough money SHALL be refused. Status: provisional.

#### Scenario: Repair price
- **WHEN** a player asks a blacksmith to repair a weapon priced at 2 silver (200 copper) that is at 80%
- **THEN** the price is 20 copper, and after they accept they pay it and the weapon is at 100%

#### Scenario: Not enough money
- **WHEN** the player has less than the price
- **THEN** the blacksmith refuses and nothing changes
