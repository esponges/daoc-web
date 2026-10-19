## 1. Spike

- [ ] 1.1 For each item in the proposal record in design.md what the server sends or holds (attributes, skills, spells, groups), which stock layouts and art exist, and whether the dangerous spawns near Myrkwood come from the data (`aggro_level`, `aggro_range`, levels) or the server's reading of it; verify by a table in design.md and the printed levels, aggro values and distances of the creatures within 1500 units of the spawn point

## 2. Safety near the spawn

- [ ] 2.1 Implement the rule the spike chooses so creatures far above a new character neither aggro it near where it appears nor stand in its way, reported at load with counts and reasons; verify with a server test over the real zone data that a level 1 character stands at the spawn for five minutes unattacked, and that the imported data is unchanged
- [ ] 2.2 Check the rule does not remove the dangerous creatures from view; verify one is still drawn and coloured red in the browser

## 3. Character windows

- [ ] 3.1 Add what the attributes window needs to the server's messages, owner only, additive and documented; verify with a protocol test that another player never receives it
- [ ] 3.2 Build the attributes window from `StatsAttributes` over the character sheet art, filled from that message; verify it opens from ATTRIB. and shows the character's real values
- [ ] 3.3 Build the styles, skills and spells windows from their layouts, filled from the tables the server already holds; verify each opens and lists a Warrior's real entries
- [ ] 3.4 Enable the menu buttons that now have windows and give the rest a tooltip saying why; verify in the browser and that the fallback UI is unaffected

## 4. Proof of the two flows

- [ ] 4.1 Kill a creature in the browser and take its loot through the stock loot window, one item and Take all; verify with screenshots saved with the change
- [ ] 4.2 Die in the browser, see the Release window, release early and wait out the timer once each; verify with screenshots

## 5. Orientation, time, icons

- [ ] 5.1 Compare the compass against a zone exit of known direction and fix its north if wrong; verify the compass shows that direction when the camera faces the exit
- [ ] 5.2 Look for a day length in the install; if one is found show a game time, else keep local time and add the tooltip; verify the choice is recorded in design.md
- [ ] 5.3 Look for a source that maps the imported icon numbers to this install's sheets; verify by trying it on the known styles `Two Fists` and `Bloodletter`, and either adopt it or record that none was found

## 6. Wrap-up

- [ ] 6.1 Document the new windows and the safety rule in the README; verify the commands run as written
- [ ] 6.2 Run all Go and JS tests and the viewer with `web/data/ui/` removed; verify they pass
