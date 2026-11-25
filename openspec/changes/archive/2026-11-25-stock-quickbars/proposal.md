# Proposal: stock-quickbars

## Why

A player asked on 2026-11-25 why styles sit on their own bar, with Guard, Intercept and Protect as
buttons above it, when in the game they are all on the quickbar, and noted that a character has
several quickbars, each with pages. Checked against the code and the install:

- **Three separate rows, none of them a quickbar.** `web/ui.js` drew the styles as one row in the
  server's bar order (keys 2 to 0, overflowing past ten), the spells as an eight-slot row
  (Shift+1 to Shift+8), and the usable abilities as a row of plain buttons above both.
- **The install has three quickbars.** Every resolution section of `default.ini` places `Quickbar`,
  `Quickbar2` and `Quickbar3`, each with `GroupSize=10`, and the options window has a "Quickbars"
  setting. Each stock quickbar holds anything (styles, spells, abilities) and pages through ten banks
  with the arrows on its side.

## What Changes

- Three quickbars of ten slots, each with ten pages turned by the arrows beside it (the skin's
  `scroll_up` and `scroll_down` buttons) or the mouse wheel. Any slot holds a style, a spell or an
  ability; the same thing may sit on several slots.
- Keys: 1 to 0 for the first bar, Shift+1 to Shift+0 for the second, Alt+1 to Alt+0 for the third.
  1 stays the attack key while the first bar's first slot is empty.
- Styles, spells and abilities are dragged onto a slot from their windows (skills, styles, spells, or
  the plain sheet); a slot is dragged to another to move it; Shift+Right-click clears it.
- What the character newly gains is placed once in the first empty slot of its bar: styles on the
  first (including those held back for the weapon, shaded), spells on the second, abilities on the
  third. A thing taken off is not put back. The old spell bar's slots carry over.
- The layout is the client's, kept in the browser per character, as the stock client keeps it in the
  character's ini file.
- Protocol: `style` takes an optional `id`, to use a style wherever it sits on the server's bar; one
  held back for the weapon is refused with an `err` naming it.

## Impact

`web/quickbars.js` (new), `web/ui.js`, `web/net.js`, `web/charwin.js`, `web/game.css`; the old style
bar (`createStockBar`), spell bars (`createSpellQuickbar`, `createSpellBar`) and ability buttons are
removed. Server: `world/combat.go` (`requestStyleID`), `protocol.Style.ID`, `gameserver` dispatch.
