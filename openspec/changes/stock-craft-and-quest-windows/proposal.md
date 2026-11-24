# Proposal: stock-craft-and-quest-windows

## Why

`crafting` and `classic-quests` use plain panels. The stock interface in the install has the windows for them (`Craft`, `MiniCraft`, `Tradeskills`, and the Quest Journal), and the project's rule is that anything the stock layout has is drawn from it, with the plain panel as the fallback.

## What Changes

- The stock Craft and Tradeskills windows show the crafting message's skills, recipes and materials.
- The stock Quest Journal shows the quests message, and the offer and reward choice use its dialogs.
- The plain panels remain when `web/data/ui/` is absent.

## Capabilities

### Modified Capabilities
- `crafting`: shown in the stock windows.

## Impact

`web/uiwin.js`, `web/ui.js`, `web/craft.js`, `web/quest.js`; no server change.
