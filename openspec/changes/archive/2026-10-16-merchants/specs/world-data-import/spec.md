## ADDED Requirements

### Requirement: Class type and stock list are kept with each spawn
The importer SHALL keep each spawn's class type and the id of the stock list it names, and SHALL import the merchant stock table (list id, item id, page, slot). It SHALL print read and kept counts for that table and fail when the table is missing or empty.

#### Scenario: A merchant row
- **WHEN** a synthetic dump has a mob of class type `DOL.GS.GameMerchant` naming list `L1` and two stock rows for `L1`
- **THEN** the spawn row carries both values and the stock table holds two rows

#### Scenario: Missing stock table
- **WHEN** the merchant stock file is absent or has no rows
- **THEN** the import fails and says which table
