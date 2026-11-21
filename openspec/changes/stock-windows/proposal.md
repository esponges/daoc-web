# Proposal: stock-windows

## Why

Opening most windows in play shows our own panels, not the game's. The menu bar, summary, compass, clock, group and release windows are drawn from the converted stock layout. Everything else isn't, including the windows players use most. Spells, Skills and Styles borrow one stock frame and fill it with plain HTML text. Merchant, pet, vault, task, trainer, healer, concentration, cast bar and quickbar are our own panels. Chat is a fixed box of about seven lines. The cause is in the window builder (`web/uiwin.js`). It draws 11 of the 29 control types the stock layouts use. List boxes, tree controls, checkboxes, icons, tabs, sliders, text areas, combo boxes and the `dialog_inset` strips (145 uses) are all skipped, so each window was finished by hand. In play the split is easy to see: INV. opens a window that looks stock because the pack draws itself on the skin's paper-doll art, but ATTRIB., SKILLS, STYLES, SPELLS and JOURNAL open frames and fills of ours (the first four are one borrowed stock frame placed side by side; JOURNAL is a plain panel), and the spell bar is a row of plain buttons although the stock client has a quickbar of its own. The goal is the 1.65 game in the browser, and the interface is what a player looks at the whole time.

## What Changes

- The window builder draws every control type the Atlantis layouts use:
  - list boxes with scrollbars;
  - tree controls with expandable lines and small icons;
  - checkboxes and tabs;
  - icons and status icons;
  - horizontal and vertical resize images and image buttons;
  - text areas;
  - sliders and combo boxes.
  Each draws in the template's own art and font, and the window code fills it with data by adapter name or control id.
- Each game window is built from its own stock template, not from a borrowed frame:

  | Window | Stock template |
  | --- | --- |
  | Attributes | `StatsAttributes` (stat strips, resist and info list boxes) |
  | Spells | `stats_spells` (spell tree with icons) |
  | Skills | `stats_spec_abil`, filled through its list boxes |
  | Styles | `stats_combat`, a tree of weapon lines with their styles |
  | Combat styles list | `stats_combat` |
  | Merchant | `Merchant` (list box, pages, quantity, coin fields, buy/sell buttons) |
  | Pet | `MiniPet` (title, health bar, stance and command checkboxes, buttons) |
  | Vault | `Vault` |
  | Task, opened by JOURNAL | `Quest` journal (`new_quest_journal`) |
  | Trainer | `Train` |
  | Healer and NPC talk | `Interact` |
  | Concentration | `Concentration` |
  | Cast bar | `Timer` |
  | Quiver | `WindowQuiver` |

- **Chat** behaves like the stock window:
  - its tabs filter channels, with Main showing everything and combat lines going to Main;
  - it scrolls back through earlier lines with a scrollbar;
  - it can be moved and resized, and it starts at a usable size.
- **Quickbar:** the client draws it itself, with no XML layout, but the install's `default.ini` gives its place (`Quickbar`, `Quickbar2`, `Quickbar3`, one position each per resolution). It is rebuilt from the skin's own art: ten slots, a bar number, slots shaded while recharging, and icons chosen by line and kind.
- **Windows can be moved and closed:**
  - stock windows move by their title bars and close with their close buttons, as the templates allow;
  - each window's position is remembered in the browser between sessions.
- **Fallback:** without `web/data/ui/` the built-in panels stay as they are.
- No protocol or server change. Every window shows data the client already receives.

## Capabilities

### New Capabilities
- None.

### Modified Capabilities
- `stock-ui-windows`: the builder draws every stock control type. Each game window is built from its own template, as listed. Client-drawn windows such as the quickbar are built from the skin's art. Windows move, close and remember where they were put.
- `game-hud`: the chat window gets stock tabs, scrollback and resizing; the style bar and quickbar get the stock quickbar look.
- `merchants`: the merchant window uses the `Merchant` template's own list, pages, quantity and coin fields.
- `player-spellcasting`: the SPELLS window is the `stats_spells` tree with icons, and the cast bar is the stock `Timer` window.
- `character-sheet`: the Skills and Styles windows fill the stock list boxes rather than drawing text over the frame.
- `pets`: the pet window is the stock `MiniPet` window.
- `rules-concentration`: the concentration list is the stock `Concentration` window.

## Impact

- **Client:** `web/uiwin.js` gets the new controls. The following move onto their stock templates, keeping their built-in fallbacks: `web/charwin.js`, `web/hud.js` (chat), `web/shop.js`, `web/pet.js`, `web/vault.js`, `web/task.js`, `web/conc.js`, `web/spells.js` (cast bar, quickbar), `web/ui.js` (trainer, healer and window management), `web/pack.js` (quiver). There are new node tests for each control and window.
- **Converter:** `cmd/uiconv` may need to keep fields it now drops (for example tree and list templates' row art) and to export the sub-images the quickbar needs.
- **Server and protocol:** no change.
- **Data:** nothing converted is committed. Icons stay chosen by line and kind, because the imported icon numbers don't match the install's sheets.
