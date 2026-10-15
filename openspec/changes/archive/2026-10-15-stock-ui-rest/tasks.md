## 1. Spike

- [x] 1.1 For each remaining piece (info panel, combat log, loot, death, labels, zone picker, menu bar, compass, clock, fonts) record in design.md whether the install has a layout or only art, with the window id or file; verify by a table in design.md and a throwaway read of the `Release`, `menu_bar`, `Compass` and `Clock` layouts

## 2. Messages in chat

- [x] 2.1 Route the viewer's log and error lines into the stock chat window with a colour per kind, and add a kind filter for its tabs; verify a fight shows both sides' lines in the chat window and none in the old log
- [x] 2.2 Keep the built-in log when the converted data is absent; verify with `web/data/ui/` removed

## 3. Loot and release

- [x] 3.1 Build the loot window on stock frames with item pictures by kind, click to take and Take all, closing on walking away; verify by killing a creature and taking its loot in the browser
  - Verified with a sample corpse in the live page (stock frame, icons by kind, row click and Take all sent the right slots in the right order) and `web/loot.test.mjs`; a real kill and loot is still to be seen end to end.
- [x] 3.2 Show the stock Release window on death with a working release; verify by dying and releasing in the browser
  - Added the `release` message (additive, `docs/protocol.md`; server test `TestPlayerCanReleaseEarly`) so the button does something: the server otherwise stands a dead character up only after 5 s. Verified the stock window in the live page (timer, RELEASE and RELEASE BIND clicks, city and house greyed, hides on revive); a real death with the window up is still to be seen.

## 4. Chrome

- [x] 4.1 Hide the developer panel and key help by default behind a key; verify a new session shows none and the key toggles it
- [x] 4.2 Add a compass control to the renderer and build the Compass and Clock windows; verify the compass follows the camera and the clock shows a time
- [x] 4.3 Build the menu bar with the inventory button working and unbuilt windows disabled; verify the button opens the inventory
- [x] 4.4 Draw name labels and the zone picker in stock fonts and con colours, and load the stock font on the login page; verify in the browser and that 24 labels keep the frame rate usable
  - Names, the zone picker and the login page use the stock TrueType font (the bitmap sheets were not needed: 24 labels held 28-35 fps). Gated on `body.stock`, so the built-in look returns without the converted data.

## 5. Wrap-up

- [x] 5.1 Document the additions in the README; verify the commands run as written
- [x] 5.2 Run all Go and JS tests, and the viewer with `web/data/ui/` removed; verify they pass
