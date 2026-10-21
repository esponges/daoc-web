# Findings

What was seen while applying this change: browser checks, census results and anything that differs from 1.65. No screenshots of game art are kept in git.

## 2.6 The start, in the browser (2026-10-19)

A scratch server (port 8788) ran the new build on a fresh import with templates honoured. The user's live server was not touched.

- A new Troll Warrior appeared in Galplen, Myrkwood Forest, among Krek (Viking trainer), Hrodny (Bonedancer trainer), the weapon merchants Gestod and Helga, and a Troll Flayer guard.
- They stood there for 5.4 minutes with nothing attacking. The chat log had no combat lines.
- All creatures within 4,000 units of the start are level 1 and passive (frogs, hatchlings, snake-finders). The nearest level 26 creatures, Haglion's minions, are about 27,000 units away.
- Teleported 800 units from a level 26 Haglion's minion: its name was drawn in purple, as were the level 14–16 svartalf hunters and mora riders around it.
- The load report said "starts: all 183 imported starts are safe". The start rule replaces nothing on this data.

## 3.2 The kit, end to end (2026-10-19)

A Troll Warrior created on the same scratch server, from a store with `start_equip`:
- wears the training axe (slot 40), the small training shield (41) and bronze stelskodd jerkin, leggings, sleeves, boots, gloves and helm (42–47), and nothing else;
- is bound at the Galplen start.

## 3.3 Levelling gear

The census's levelling-gear table (`docs/parity/README.md`) shows Midgard merchants in regions 100, 101 and 151 selling every kind of Warrior weapon, shields, studded and chain from level 1 to 30, and **nothing above level 30**: 32 gaps across bands 31–50.

The data has 1,302 Midgard weapons and chain pieces above level 30, but only 15 merchant lists sell any of them: 6 placements in region 249 (Darkness Falls), and 2 in region 101. The region-101 ones carry the `OrbsMidgardArmor` and `OrbsMidgardWeapons` lists, which sell only level 51 items for OpenDAoC's own "orbs" currency. That is not classic, and level 51 falls outside the table's bands. A Warrior above level 30 levels with drops, as the data stands. Crafting is a later change. The gap is reported and not filled.

## 5.2 After the hate refactor, in the browser (2026-10-19)

Two Troll Warriors (level 1, classic kit) on the scratch server, at a level 2 spectral hog south of Galplen.

- **Two clients on one creature:** both targeted a spectral hog and fought side by side. After the first died, the second kept fighting one ("The spectral hog misses you!") until they died too. The chat logs do not show whether it was the same hog. The turn to the next player is covered by `TestCreatureTurnsToTheNextWhenOneDies`.
- **Death:** "You have been killed.", then the character came back at the Galplen bind with full health.
- **Leash:** the browser try was inconclusive: the second character's target was out of reach and no blows were exchanged before it teleported away. The leash is covered by `TestAggressiveWolfAttacksAndLeashes`, which passes unchanged.
- **Nothing else differed** from before the refactor.

**Finding: a level 1 Warrior loses to a level 2 creature.** Each character hit the level 2 hog for 1–2 damage (training axe: 1.2 DPS, 2.5 s, 90% quality) and took 10–11 per hit from it, with 39 hit points. Both died; the hog was barely scratched. Task 5.3 checked it.

## 5.3 Low-level melee (2026-10-19)

The fight above is the rule book working as written, and the rule stands:
- **Every number follows the rule book**, and the rule book follows OpenDAoC-Core with its default settings. Mob damage factors 2.5 and 120, a damage factor of 1, PvE melee 1.0, weapon skill 100%, NPC armour `int((1 + L/200) × L × 12.5) / 5` and the player's extra 0.4 AF per level all match, read from its `AttackComponent`, `GamePlayer`, `GameNPC`, `ArmorFactorCalculator`, `WeaponSkillCalculator` and `ServerProperties`.
- **The hog was orange, not yellow.** At level 1 the con band is one level wide (level / 10 + 1), so a level 2 creature is orange to a level 1 character. The first reading of this finding called it yellow, which was wrong.
- **Per hit, at the middle of the variance:**
  - against a yellow level 1 creature (31 HP): the Warrior hits for 4.05 and takes 4.12;
  - against an orange level 2 creature (46 HP): the Warrior hits for 1.93 and takes 10.78.
- **Forty simulated fights** (`TestLowLevelFights`): the new Warrior in the kit beat the yellow creature 20 times in 20 and lost to the orange one 20 times in 20.
- **The source:** a Camelot Herald wiki Combat page, as a search result summarised it (the page itself returns HTTP 402 here), says the ideal low-level melee target is one level below you and a yellow creature can be a tough fight.

On these numbers, losing to an orange creature at level 1 is the classic warning colour doing its job. Two pieces are recorded as rule-book scenarios ("A yellow creature at level 1", "An orange creature at level 1"). Whether 1.65 live was gentler at levels 1–2 than this would need a combat log from the time. None was found, so it stays reference.

## 6.6 and 7.5 Casters, in the browser (2026-10-19)

The scratch server (port 8788, a store imported with spells) and a Troll Warrior raised with `/gm level` in the classic kit.

- **Arachite priests** (level 15, Myrkwood Forest): on approach the log read "The arachite priest begins casting a spell!", then "You are slowed!" and "The arachite priest's Grinding Force hits you for 12 damage!" every few seconds. Two badges appeared under the vitals: a red **D** counting down from 15s (the damage over time) and a blue **S** from 1:03 (Constrict Spirit). When the slow ended the log said "You can move at full speed again." and the badge went.
- **Svartalf sorcerers** (level 15): Dissolving Cloud and Lesser Devolution hit for 27, one Clutching Vines was resisted ("You resist the effect of the svartalf sorcerer's Clutching Vines!"), and in melee "The svartalf sorcerer's spell is interrupted!" appeared as a blow landed on one mid-cast. Once its lasting spells are on the target, a sorcerer melees until they wear off, as the casting rule says.
- **Found and fixed: arachites were invisible.** The client refused to draw models with more than 52 bones on a shape. 16 converted creature models need more (the arachites' m404–m407 need 58–64, the largest 77), so those creatures fought and cast unseen and could not be targeted. The limit is now 80 bones, which still fits the 256 vertex uniform vectors every WebGL2 device has. After a reload the arachite priests and weblashers draw, and Tab targets them.
- A character raised to level 30 or 50 still carries the level 1 kit, so it hits for 6–10 and takes 56–85 a blow from level 15–16 creatures. That is the gear, not the rules.
- Line of sight is not checked for spells (README, Known limitations).

## 9.5 Groups and chat, in the browser (2026-10-19)

Two clients on the scratch server, Partyleader and Groupmate, both Troll Warriors.

- **Stock interface.** Partyleader clicked Groupmate ("You target Groupmate.", the summary showing the name), opened GROUP on the menu bar and pressed INVITE. Groupmate saw the invitation on the stock dialog frame and pressed ACCEPT; "You have joined Partyleader's group." and the stock group window opened on both, listing "*Partyleader" and "Groupmate" with class, level and full bars. With Groupmate fighting svartalfs in Myrkwood, Partyleader's bar for them read 29%, then empty (dead), then full after the release. Groupmate's window showed Partyleader as "in Myrkwood Forest", 6,000 units being exceeded.
- `/g pull now` showed as `[Group] Partyleader: "pull now"` in teal on both, and alone on the Group tab. A tell showed as `You send, "meet at the bridge" to Groupmate.` and `Partyleader sends, "meet at the bridge"` in pink; `/tell Nobody hello` gave "Nobody is not in the game." `/disband` from Groupmate gave "You leave the group." and "Groupmate has left the group. Your group has disbanded.", and both windows closed. DECLINE gave the inviter "Groupmate declines your invitation."
- **Without `web/data/ui/`.** `/invite groupmate` brought up the plain prompt, Accept opened the plain group panel on both ("★ Partyleader 1 Warrior", "Groupmate 20 Warrior", red health bars), and `/g`, a tell and `/disband` read and coloured the same in the plain chat.
- **`new_group` is not an invitation.** The layout's `new_group` window holds eight health bars bound to `group_health0`–`7`, a compact group list. The invitation is drawn on the stock `dlg_sm_title_noresize` frame with `button_large` ACCEPT and DECLINE; the `groups` spec now says so.
- FIND, SET LFG, BENEFITS, ASSIST, FOLLOW, STICK, FACE, CGF, OPTIONS and the window's own WARMAP and MAP buttons are greyed: nothing is behind them.

## 10.1 Sources for the Warrior's abilities (2026-10-19)

One bounded search. The Camelot Herald wiki pages (Guard, Protect, Intercept, Engage) answer HTTP 402 here and Neoseeker's 1.63 beginner's guide 403, so only search summaries of them were seen. One period source was read in full:

- **ZAM/Allakhazam, "Hero class guide" v2.0** ([camelot.allakhazam.com/guides/Hero_guide_v.2.0.html](https://camelot.allakhazam.com/guides/Hero_guide_v.2.0.html)), written before patch 1.52 (it calls 1.52 upcoming in Europe). The Hero's Guard, Protect, Intercept and Engage are the same abilities the Warrior has.

What it says, against the spec as written:

| ability | the guide | the spec | outcome |
| --- | --- | --- | --- |
| Guard | one member, shield worn, within trading distance; blocks blows at them; one guard per member | block chance at the member within 256 units | agrees; 256 stays provisional (no number given); one guard per member added |
| Protect | moves the anger the member draws by **healing, nuking or debuffing** to the protector; does nothing for a member engaged in melee; no numbers | a share of the **damage the member takes** | disagrees: the mechanic follows the guide; the 10/20/30% stay provisional |
| Intercept | stays armed until it triggers, then about **one minute** to recharge; the blow is taken with evade, parry and block not helping; the member close and **faced** | next blow taken, resolved against the Warrior's defences and armour, 60 s | reuse documented (60 s); no defence roll; facing added; 128 units provisional |
| Engage | a monster that is attacking you; **every** attack from it blocked for a little endurance each; you cannot attack it; fails if the group attacked it first | 95% block when not attacked for 5 s, 5 endurance a block | disagrees on the condition; 95% and 5 stay provisional (the guide gives "every" and "a small amount") |

Two community notes agree with the guide: a 2001 Quarter to Three forum post (Intercept takes the hit with no armour absorption, so harder than the member would have; Guard needs you right next to the member) and the Gaheris FAQ, as a search summarised it (Protect I and II move only a small fraction of a healer's anger; the protector's weapon must be out).

## 18.1 Two clients on the 1.65 loop (2026-10-20 and 21)

Two characters on the user's live server, each in its own browser tab as the administrator: `Browsertroll` (Troll Warrior, level 12) and `Secondtroll` (Norse Warrior, level 15). This was not an evening's play; a person at a keyboard has not yet done that. It is what a script-assisted session could reach, in a browser that drew one to four frames a second, so the loop was walked in pieces and each piece is written down as it went.

**What was exercised, and how it went**

- **Group.** `/invite Secondtroll` then `/accept` formed the group in one second; both clients listed both members. After one went to Myrkwood Forest by `/gm goto 105` the list showed the other in another zone, out of range, and a `/g` line reached the member in the other zone as `[Group] Browsertroll: "hello from Mularn"`.
- **Bind.** `/bind` near the binder said "You are bound here. When you die and release, you will return to this place."
- **Dying and praying.** Killed by a pack of young grendelorms (four level 7 creatures hitting for 21 to 37 each, about twelve seconds from 265 hit points), the character was released early in about a second at the bind point, and left alone came back after exactly five seconds. A gravestone and the prayer were not exercised: the character was at level 12 with no experience, so nothing was lost and no stone was made.
- **Loot.** Coin from every level 1 creature (9 to 12 copper); one item, a `Tattered Rawhide Mjuklaedar Sleeves`, from a mud snake, taken through the stock window.
- **Night.** The game clock ran through dusk, night and dawn within the session (a game day is an hour); at night a pack of rabid wolfhounds (level 11) stood in the Vale and the level 7 grendelorms came for the character.
- **Horse routes, the vault and repair** were each watched earlier in this change, in their own sections, and were not repeated here.
- **Not exercised:** casters (no caster was fought; creature spells are in the server but were not looked at), banking and repair in this session, and an evening of levelling.

**What broke or felt unlike 1.65**

1. **Level 1 creatures drop coin only.** The community data has item drops for named creatures only, and the server adds a 22% chance of one random item of the creature's level; an ordinary kill pays 9 to 12 copper and almost never an item. 1.65 had loot generators by creature level and kind, which the imported data does not carry, so a new player here very likely gets less loot than in the real game.
2. **`1` toggles attacking.** Pressing it a second time, easy to do by habit, stops the attack ("You stop attacking."), with nothing on the bar to show whether the fight is on. Whether that is how 1.65 felt was not checked.
3. **Attacking does not walk to the target.** Pressing attack at a creature across the field does nothing until the player walks up and the character faces it; a character standing still turns to the target by itself, slowly at a low frame rate. How 1.65 behaved was not checked.
4. **Night packs in the middle of the Vale.** The night-only wolfhounds and the grendelorm packs stand in open ground; a level 12 character died in twelve seconds. This is what the data places. The clock brings every night spawn out at once at dusk, which has not been compared with 1.65.
5. **Pack-to-pack equip is easy to get wrong.** `equip` takes a source slot and a destination slot in the one numbering, where 0 to 39 are pack slots and 40 to 47 worn positions; sent to a pack slot, it moves the item within the pack. Only a script does this today.
6. **The test browser is the weak part.** At one to four frames a second the page took about ten seconds to turn a character half way round and a hit the player cannot see arrives anyway; the shared screen drawing went stale after a few seconds with two tabs and had to be refreshed by reloading. That is the environment, not the game, but a slow machine would see the same.
