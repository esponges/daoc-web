# daoc-web

A fan project that rebuilds Dark Age of Camelot's **Midgard** in the browser, as a
multiplayer game. A Go server owns the world; a dependency-free WebGL2 viewer is
its client; and the interface is drawn from your own install's UI data. The aim is
the closest playable version of 1.65. For now it is a private build for a few
friends.

The repository contains **no game data**. Art, models, zones, sounds and layouts
are converted from your own DAoC install into `web/data/` (gitignored), and the
world's population comes from the community
[OpenDAoC database](https://github.com/OpenDAoC/OpenDAoC-Database) (GPL-3.0),
imported at build time into `server/data/` (gitignored).

## What you can do today

- Log in, make a character from the start table (any of Midgard's twelve
  classes: Warrior, Berserker, Savage, Healer, Shaman, Runemaster, Thane, Skald,
  Spiritmaster, Bonedancer, Hunter or Shadowblade, each with its races, male or female) and walk
  Midgard: all 25 outdoor zones, its dungeons and Jordheim, with day and night,
  collision and the neighbouring-zone borders.
- Fight the zone's creatures: swing speed, hit, evade, parry and block,
  styles, specialization, damage types, armour, death, experience and levels,
  loot and release, repair.
- Cast: direct damage, bolts, heals, damage over time, slows, mesmerize, stun,
  life drain, stat and armour buffs, damage add, area spells, and the Skald's
  songs and chants (speed, regeneration, resists, damage add), which pulse on
  the group until turned off. A SPELLS window, spell bar and cast bar; power
  regenerates and is spent.
- Trade with merchants, keep things in the vault, talk in chat channels, group up,
  ride the horse routes.
- See other players and creatures move and fight, a few hundred milliseconds
  behind the server.
- Play through the game's own windows (menu bar, compass, chat, inventory,
  style bar and the rest) when `web/data/ui/` has been converted, with plain
  panels as the fallback.

## How far from a live 1.65 server

`go run ./cmd/parity` reads the community database and counts what Midgard's
data asks for against what the server implements. This is its current reading;
the figures are checked against `docs/parity/parity.json` by `go test ./cmd/parity`.

| | implemented |
| --- | --- |
| Creature spell types | 15 of 15 |
| Class spell types | 73 of 73 (987 of 987 spells) |
| Style features | 13 of 14 (1,382 of 1,383 styles) |
| Abilities | 22 of 30 (88 of 100 grants) |
| Creature kinds | 24 of 110 (29,119 of 29,932 placements; 7 kinds left out as not classic) |
| Quests | 0 of 0 (the 81 XP Item data quests are marked not classic) |
| Keep doors | 0 of 432 |
| Keeps | 0 of 79 |
| Relics | 0 of 6 |
| Battlegrounds | 0 of 5 |
| Levelling gear | 0 gaps in 10 five-level bands (31 to 50 is the project's added stock) |
| Rules | 27 documented, 66 reference, 46 provisional |

So: the solo-and-small-group player-versus-environment loop works, for all twelve of
Midgard's classes (trainers teach them), with pets and charm, bows, stealth,
poisons, the left hand and tasks (kill and delivery tasks from town NPCs, up to level 20).
Not started: quests, crafting, housing,
keeps, relics, battlegrounds and every realm but Midgard. Merchant armour and
weapons run to level 50 with stock this project added above level 30. [Known limitations](#known-limitations) has the rest.

## Where things are

| path | what |
| --- | --- |
| `cmd/gameserver`, `internal/server/*` | the game server: accounts, world, combat, rules, store, wire protocol |
| `web/` | the browser client and its tests |
| `cmd/*conv`, `cmd/dbimport` | converters from your install and the community database |
| `docs/formats.md` | the converters and the file formats worked out from the install |
| `docs/protocol.md` | the wire format; messages are JSON, versioned, only ever added to |
| `docs/census`, `docs/parity` | what parses and converts; how far the server is from classic Midgard |
| `openspec/` | specs (`specs/`) and changes (`changes/`, `changes/archive/`) |

Needs Go 1.26+ and a DAoC install (by default
`C:\Program Files (x86)\Electronic Arts\Dark Age of Camelot`; `-game` changes it).
The first-run recipe is under [Multiplayer](#multiplayer) below.

## Multiplayer

`go run ./cmd/gameserver` runs the game for several players at once: a Go
server that owns the world, and the viewer as its client. Midgard's outdoor
zones are playable, from Vale of Mularn, Myrkwood Forest and Gotar through to
the Shrouded Isles, and so are its dungeons and Jordheim. The design is in
`openspec/changes/archive/2026-10-15-multiplayer-server/`; the wire format is in
[`docs/protocol.md`](docs/protocol.md).

### Running it

```bash
# once: convert the zones, and the creatures that live in them
go run ./cmd/worldconv -only outdoor           # every Midgard zone; long
go run ./cmd/charconv -outfit troll-warrior && go run ./cmd/animconv -char web/data/char/troll-warrior
go run ./cmd/charconv -outfit norseman      && go run ./cmd/animconv -char web/data/char/norseman
# every creature is web/data/char/m<its model number>; convert the ones the zones place:
go run ./cmd/dbimport -all-models -out server/data/all.db    # a scratch import, to list the models
for m in $(go run ./cmd/dbimport -out server/data/all.db -q "select distinct model from spawn" | tail -n +2); do
  go run ./cmd/charconv -model $m -name m$m && go run ./cmd/animconv -char web/data/char/m$m
done

# the world's population, from the community database (see below)
git clone https://github.com/OpenDAoC/OpenDAoC-Database server/data/community
go run ./cmd/dbimport                          # writes server/data/world.db
# how spells look: the effects those spells use (needs the import above; see docs/formats.md)
go run ./cmd/effectconv                       # writes web/data/effects/
# the pets the summon spells make (Spiritmaster, Bonedancer, Hunter) use models too:
for m in $(go run ./cmd/petmodels -missing); do
  go run ./cmd/charconv -model $m -name m$m && go run ./cmd/animconv -char web/data/char/m$m
done

DAOC_ADMIN_PASSWORD=choose-one go run ./cmd/gameserver   # then open localhost:8777
```

`dbimport` keeps only the spawns whose creature is converted, so the model
list comes from an import that keeps them all (`-all-models`, into a scratch
`-out`), and the real import follows the conversions. Without `world.db` the
server still starts, with every converted zone empty. A few models do not
convert (skeletons, composites) and their creatures are left out.

Log in at `/login.html`, make a character (a Troll Warrior starts in Galplen
in Myrkwood Forest, a Norse Warrior in Vasudheim in East Svealand, as the data
and ZAM's starting-zones table agree) and press Play. Two browsers, two
characters: each sees the other walk, fight and talk.

**Settings** come from flags, the environment, or `server/data/server.env`
(`KEY=VALUE` lines; the directory is gitignored). The server refuses to start
with no way to log in.

| variable | |
| --- | --- |
| `DAOC_ADMIN_PASSWORD` or `DAOC_ADMIN_HASH` | the administrator's password, or its bcrypt hash; an administrator may use the `/gm` commands |
| `DAOC_DISCORD_CLIENT_ID`, `DAOC_DISCORD_CLIENT_SECRET` | log in with Discord: create an application at discord.com/developers, add `<DAOC_PUBLIC_URL>/auth/discord/callback` as its redirect, scope `identify` |
| `DAOC_CLOCK_SHIFT` | a testing aid: seconds added to real time for the game clock, to see a dusk or dawn without waiting (game time is real time times 24 modulo a day); zero by default |
| `DAOC_PUBLIC_URL` | where browsers reach the server, `http://localhost:8777` by default; `https://` also marks cookies Secure |
| `DAOC_SESSION_KEY` | signs the session cookie; generated into `server/data/session.key` if unset |
| `DAOC_ADDR` / `-addr` | the address to listen on; `localhost:8777`, or `:8777` for the whole network |

Players, characters and inventories live in `server/data/players.db`, a SQLite
file separate from the imported world so re-importing never touches them.

**Re-importing while the server runs.** A running server holds `world.db`
open, and on Windows the file cannot then be replaced. `dbimport` builds the
new store beside it and, when the old one is in use, leaves it as
`world.db.next` and says so. The server swaps it in when it next starts: it
checks the staged store opens, has the world's tables and has spawns, renames
`world.db` to `world.db.prev` and moves the new one into place. A staged store
that fails the check is left where it is and the reason is logged. A failed
import leaves `world.db` as it was.

```bash
go run ./cmd/dbimport                  # with the server running: writes world.db.next
# stop the server (Ctrl-C), which folds players.db-wal into players.db; then back it up
cp server/data/players.db server/data/players-$(date +%F).db
DAOC_ADMIN_PASSWORD=choose-one go run ./cmd/gameserver   # the log says it swapped
```

`players.db` is in WAL mode, so while the server runs recent saves are in
`players.db-wal`: back it up after stopping the server, not before.

To roll back, stop the server, put the previous store and the backup back,
and start the previous build:

```bash
mv server/data/world.db.prev server/data/world.db
cp server/data/players-YYYY-MM-DD.db server/data/players.db
```

### What the server decides

Everything. The client sends what the player holds down and what they ask for;
the server moves the character, runs every creature, rolls every blow and
keeps every item. The browser predicts its own movement with the same maths
(checked against vectors the Go tests write, to zero difference) and hides the
latency; other players and creatures are drawn about 120 ms in the past,
interpolated.

- **Movement and collision.** 85 units a second walking, 191 running, the
  game's own rates. Trees block as cylinders of the radius `nifs.csv` gives
  and buildings as the models' own collision hulls, which `propconv` now keeps
  in `collision.bin`; indoors the walls are the geometry, and a ledge or void
  is as solid as a wall. Walking off an outdoor zone's edge enters the
  neighbouring zone; walking into an exit (a dungeon mouth, a city gate)
  enters that place.
- **Creatures** are the imported spawns, each placed, levelled and sized by the
  database, wandering within its roam, aggressive where it is, calling its
  packmates, giving up a chase at its tether and walking home to heal.
  A creature hates by damage and fights whoever it hates most, player or not.
- **Guards** (`GameGuard` spawns) watch their aggro range, or 750 units when
  the spawn gives none. They attack an aggressive creature that comes into it,
  or one fighting a player of their realm who is in it, chase it to their
  tether and walk back to their post. A guard's damage counts in the kill's
  experience share, and that share goes to no one.
- **Combat** follows classic 1.65, as written in the rule book (below) and
  computed in `internal/server/rules`. A swing is settled in this order:
  1. the defender's evade, parry and block, from their abilities, specs and
     dexterity (or a creature's template), and only from the front;
  2. an 18% miss chance moved by the level difference and style bonuses;
  3. if it hits, a hit location, then damage: weapon skill against the armour
     worn there, with variance from the spec, a damage cap, critical hits and
     the creature's race resists.

  A Warrior's styles come from the style table for the weapon held, unlocked
  by trained spec. There are anytime styles, openings after a parry, block or
  evade, positional ones from the side or behind, and chains (Crushing Blow,
  then Slam). Each adds growth damage and costs endurance by weapon speed,
  and half if the swing fails. Creatures' hit points, blows and armour come
  from their level.
- **Progression.**
  - Characters have the eight stats of their race plus a 30-point build, and
    gain stats from level 6.
  - Kills give the published experience values, shared by damage and capped at
    1.25 × a same-level kill; grey kills give nothing. A group's damage counts
    as one damager, and its award is split by level among the members alive
    and within 6,000 units, each part capped the same way; the kill is grey
    to the group if it is grey to its highest member in range (provisional).
  - Levels (to 50) give stats, hit points and specialisation points, spent at a
    trainer of your class (or the Viking trainer).
  - Dead creatures drop money and loot from their loot tables (or, failing a
    table, now and then something of their level). Only their killer can take
    it for thirty seconds, into a pack of forty.
  - Items go on and off in the pack panel (`I`). Their stat bonuses count up to
    the item caps.
  - What a Warrior may use comes from the class's abilities by level: studded
    armour until level 10 and chain after, larger shields with level, and one-
    or two-handed swords, axes and hammers.
  - Every item change is saved in one transaction with the character.
- **Death.** From level 6, a death to a creature costs experience at once.
  Releasing costs as much again, into a gravestone where you fell, and costs
  constitution; you may release ten seconds after dying, and are released
  after fifteen minutes. You come back at your bind point whole, with creatures
  sent home. A Healer, Shaman or Spiritmaster's resurrection spell offers to raise
  a dead player of their realm who has not released (fifteen seconds to accept):
  the raised character skips the release stage (no second loss, no constitution
  loss, no gravestone) but carries resurrection sickness for a minute, doing 30
  percent less damage. Target your gravestone and
  `/pray` there to win its experience back; a healer restores constitution
  for money.
- **Binding.** A character is bound at their start, and `/bind` beside a
  bindstone of their realm binds them there. A character saved before binding
  is bound at the bindstone nearest to where they first log in.
- **Groups and chat.** Up to eight players of one realm group: invite a
  targeted player (click them) or `/invite <name>`, and the invited one
  accepts or declines (an invitation lapses in 60 seconds). The leader alone
  invites and `/remove`s; `/disband` leaves; when the leader leaves, the
  first to join leads. A member stays grouped across zones, and leaves on
  logging out. Members see each other's health, endurance and zone in the
  stock group window (GROUP on the menu bar; MINI for the small one), or a
  plain panel without the converted interface. `/send` or `/tell <name>`
  reaches one player in any zone, `/g` the group; tells are pink and group
  lines teal, and the chat window's Group tab shows only the group.

Keys in network mode: `W/S/A/D/Q/E` and the mouse as before, `Enter` to talk
(`!` to shout, `/who`, `/pray`, `/bind`, `/invite`, `/disband`, `/tell`, `/g`), `1` attack, `2`-`9`, `0` the styles on the bar (lit when
usable), `I` the pack, `C` the character sheet (`Shift+C` the free camera), a
click on a corpse to loot it, a right-click on a merchant, trainer or healer. `Space` jumps, `X` (or `/sit`) sits and stands, `Z` (or `/sprint`) turns sprint on and off, `M` opens the Command window, and the classic emotes are slash commands (`/wave`, `/bow`, `/dance`, `/cheer`, `/clap`, `/cry`, `/curtsey`, `/flex`, `/kiss`, `/laugh`, `/point`, `/salute`, `/bang`, `/victory`, `/beckon`, `/blush`). A seated character regains health, power and endurance twice as fast, and stands up to do anything else; a long fall costs health, and a lake is swum, not walked. The numbers are the rule book's `rules-resting`, `rules-falling` and `rules-movement`, all provisional. Administrators: `/gm goto
103`, `/gm tp x y`, `/gm level n`, `/gm give <item id>`.

### Where the world's content comes from

The client install has no spawns, stats, loot, styles or persistence: those
lived on the game's servers. This project takes them from the
[OpenDAoC database](https://github.com/OpenDAoC/OpenDAoC-Database), published
under the GPL-3.0. It describes content owned by Electronic Arts. It is **not
redistributed here**: it is cloned into `server/data/community/`, which is
gitignored like `web/data/`, and read by this project's own importer into
`server/data/world.db`, also ignored. No rows, dumps or code from OpenDAoC or
DOLSharp are committed, so this repository stays MIT, and the server's combat
rules are written here from documented mechanics. Use is private, on your own
machine or network. Check the licence and the EA question again before
hosting it publicly.

What the data spike found, on zone 100: a spawn's region coordinates minus the
zone's `OffsetX`/`OffsetY` times 8192 are local zone units, with no axis flip,
and 2,202 of 2,219 spawns land within 300 units of the derived terrain; the
spawn model numbers are the client's `monsters.csv` ids.

**Importing.** `go run ./cmd/dbimport` takes about 6 seconds, and
`go run ./cmd/dbimport -q "select name, count(*) from spawn where zone=100 group by name"`
looks at the result. It parses the dump's SQL itself, so no MySQL is needed,
and keeps spawns (in zone units), NPC templates, loot, items, styles, exits,
start locations, bind points and each class's starting equipment (a new
character is given exactly what that table lists). A spawn is dropped if it is in no zone, uses
a model with no converted character (`web/data/char/m<N>`), or sits more than
500 units from the derived ground; the counts are printed.

Almost every spawn names an NPC template marked `ReplaceMobValues`, and the
emulator the data is kept for lets such a template overwrite the spawn row.
The importer does the same: the spawn takes the template's name, aggression,
flags, speed, class type and level list, and the server picks one of those
levels each time the creature spawns. Model, size and equipment stay the
row's. Ignoring the mark left thousands of creatures at the wrong level, such
as level 26 aggressive frogs beside the Galplen start that the template makes
level 1 and passive.

### The rule book

The aim is classic 1.65 Midgard as a public server played it. Each mechanic the
server implements has its own capability spec under `openspec/specs/rules-*/`,
for example `rules-melee` or `rules-death`. A spec says what the rule decides,
what goes in, the formula in this project's words, its caps, and worked
examples with concrete numbers. Those examples are the tests: each one is a
table test in `internal/server/rules`, named after it. Each rule carries a
status:

- **documented**: a public source states it;
- **reference**: no public page gives the number, so it matches how the classic
  reference server (OpenDAoC) behaves, which was read only to understand it;
- **provisional**: neither exists, so the value is this project's choice, and
  the spec says so.

The registry in `internal/server/rules` lists every rule with its status and
sources, and the parity census (below) counts them. No code from a GPL server
is copied into the repository, its specs or its tests.

Sources the rules come from:

- [disorder.dk](https://disorder.dk/daoc/), the 1.65 mechanics pages:
  - [chance to hit](https://disorder.dk/daoc/chance-to-hit/)
  - [evade](https://disorder.dk/daoc/evade/)
  - [parry](https://disorder.dk/daoc/parry/)
  - [blocking](https://disorder.dk/daoc/blocking/)
  - [stats](https://disorder.dk/daoc/stats/) and [stat caps](https://disorder.dk/daoc/stats-cap/)
  - [hit point cap](https://disorder.dk/daoc/hit-point-cap/)
  - [armour factor cap](https://disorder.dk/daoc/armor-factor-cap/)
  - [swing speed](https://disorder.dk/daoc/melee-swing-speed-cap/)
  - [hit location](https://disorder.dk/daoc/armor-hit-chance-and-location/)
  - [skill cap](https://disorder.dk/daoc/skill-cap/)
  - [melee damage](https://disorder.dk/daoc/melee-damage-calculation/)
- The [Camelot Herald wiki](https://camelotherald.fandom.com/): melee damage,
  weapon skill, styles.
- Mythic's Friday Grab Bag answers on parry, block and evade, as quoted on
  those pages, and the knowledge-base note that raised the experience clamp to
  1.25 times a same-level kill.
- The published experience and constitution-restore tables
  (silicondragon.com's "XPs" page, now only in archives).
- Mythic's description of con colour bands ("level divided by ten plus one"),
  reposted in a group experience explainer.
- ZAM's [starting-zones table](https://camelot.allakhazam.com/Start_Zones.html)
  (Shrouded Isles and Trials of Atlantis era): the town each Midgard race and
  base class starts in. The imported start rows agree with it.
- The community database's horse routes: a ticket item's id is also the id of a
  path whose points are the ride (spike over the dump: 216 of 223 tickets).
  Nothing is copied from OpenDAoC for it; the stable master's behaviour is our own.
- A 2001 forum thread on the Warrior shield abilities
  ([quartertothree.com](https://www.quartertothree.com/boards/messages/17/1030.shtml)):
  what Guard, Protect, Intercept and Engage do in outline. It gives no ranges,
  durations or shares. The abilities' numbers and order of resolution follow
  OpenDAoC's (Guard 256 units and +5% a level, Intercept 128 units at 50% with
  a 60 s reuse, Protect 10% a level moved to the protector within 1,000 units,
  Engage +65% block for 5 endurance a blow), read to understand them.
- The Camelot Herald class library pages for the Skald and the Bard, and Allakhazam's
  spell entries for the songs ([camelot.allakhazam.com](https://camelot.allakhazam.com/db/spelldb.html?cspell=553)):
  only one song can be sung at a time, a speed song does not work in combat, and
  the Song of Travel values by level. Friday's Grab Bag answer on resist caps
  ([story 736](https://camelot.allakhazam.com/story.html?story=736)): buff and item
  resist caps are separate.
- The Camelot Herald's patch notes (1.37c: a resurrection offer is answered within
  15 seconds; [1.52](https://camelotherald.fandom.com/wiki/Patch_Notes:_Version_1.52):
  concentration returns after resurrection sickness;
  [1.65](https://camelotherald.fandom.com/wiki/Patch_Notes:_Version_1.65): a timer
  shows the time to release) and the Grab Bag
  answer that the concentration pool depends on level and casting stat, not spec.
  The spell texts in the community database are Mythic's own, and say that a
  resurrection spares the release penalty.
- The 1.38 patch notes ([release notes](https://gamebanshee.com/38kh): falling damage on, Safe Fall in the
  Stealth line) and the 1.87 notes as quoted on the
  [Dawn of Light forum](https://dolserver.sourceforge.net/viewtopic.php?p=90942) (standing regeneration
  doubled to match sitting). They give the shape of falling and resting; the numbers are this project's
  and provisional.
- The Camelot Herald class library page for the Savage
  ([camelotherald.com/node/194](https://www.camelotherald.com/node/194): the Savagery line modifies evade,
  parry, attack speed and DPS, and turns hit points into endurance) and the 1.65H test patch notes
  ([gamebanshee.com/3ydp](https://gamebanshee.com/3ydp): Savage buffs take a percent of maximum health). The
  community database gives each Savage spell's cost, value and duration; reading the values as percents and a
  Taunt's hate scale are this project's, provisional.
- [OpenDAoC-Core](https://github.com/OpenDAoC/OpenDAoC-Core) (GPL-3.0), read to
  understand rules that no public page gives a number for. Nothing is copied
  from it.

### Parity census

How far is the server from classic Midgard? `go run ./cmd/parity` reads the
community database and lists, for Midgard (regions 100, 101 and 151), what the
data asks for:

- the spell types creatures and classes use;
- the style features;
- the abilities granted;
- the creature kinds, by class type;
- quests, keep doors, keeps, relics and battlegrounds;
- the Warrior weapons, shields and armour Midgard merchants sell in each
  five-level band, with the gaps marked (the data has none above level 30; the
  importer adds generic gear for 31 to 50 from `internal/importer/addedstock.go`,
  shown with a star, so there are no gaps
  left).

For each, it says whether the server implements it, and it counts the rules by
status. It writes `docs/parity/parity.json` and `docs/parity/README.md`: names
and counts only, so they are committed like the asset census.
`go test ./cmd/parity` (not `-short`, with the dump checked out) fails if
anything that was implemented no longer is, or if a rule's status weakens.
The current reading is the table at the top of this file.

### Classes, spells and power

How a spell looks comes from the client's own tables (`docs/formats.md`, "How a
spell looks"): the server sends each spell's client effect as `fx` on its events,
and with `web/data/effects` converted (`cmd/effectconv`, and `cmd/animconv` for the
casting clips) a caster plays the spell's cast animation with its hand effects, a
bolt flies and the target effect plays on the target. Without that data spells
work and log as they always did. The log also shows your pet's fight, and the pet
window and the row under the clock show lasting effects, helpful ones framed in
gold and harmful ones in red.

The creation page offers every race and class pair the data's start table lists
for all twelve Midgard classes, each with a
choice of male or female. A caster has a power pool (`rules/power.go`;
reference: the community servers' numbers, no public 1.65 account) and a
spellbook built from the imported spell lines of the specs they train
(`line_spell`, with each spell's power a fixed cost or a percentage of the
pool). `cast` asks for a spell; the server checks
that it is known, ready, targeted, in range and paid for, runs the cast time,
stops it on damage or movement, and applies it with the same engine creatures
use. The stock SPELLS window (the `Stats_Spec_Abil` layout), a spell bar
(Shift+1 to Shift+8) and a cast bar draw it; without `web/data/ui/` a plain
panel (P) does. A spell aimed at another needs line of sight at the start and at
landing ("You can't see your target!"): the server walks the segment between the
two eye points through the zone's collision bins and heightmap, for players and
creatures alike (a creature without sight closes in instead of casting). A
lasting buff with a `Concentration` cost holds that much of the caster's pool
while it is on, with no timer, until it is cancelled (the K list), the caster
dies or leaves, the target leaves, or the target stays more than 5,000 units
away for 20 seconds. Not yet: the focus bonus on power cost and scaling of buffs and debuffs by spec (the
sources differ; buffs are flat here).
An area spell hits every creature inside its radius with no falloff, around the
caster when its range is 0 and around the target otherwise. A song or chant is a
toggle that lands on the group and pulses on it every few seconds; only one is
on at a time. A speed song does not take hold in combat. Player
damage, heals, damage over time and damage add follow the casting stat, the
spec and the weapon's swing as the community servers do.

### The game's own interface

Every window is drawn from the game's own templates, art and fonts when they
have been converted: the menu bar, compass, clock and summary; chat (Main, Chat
and Group tabs, a 500-line scrollback, resizable and movable); Attributes,
Skills, Styles and Spells; the quest journal; the merchant, vault and trainer;
the healer's talk (Interact); the pet (MiniPet), Concentration, the cast bar
(Timer) and the quiver; and the spell bar and quickbar, drawn from the skin's
own slot art. Windows move by their title and remember where they were left
(`/resetui` puts them back). What the install has no layout for (the pack, the
loot window, the quickbar) is built on the stock frame and art by hand:

```bash
go run ./cmd/uiconv -skin atlantis    # writes web/data/ui/atlantis/ (not committed)
```

`ui/uimain.xml` in the install lists about eighty window files; each defines
windows, buttons, frames and text areas with sizes, positions and the textures
they use. `uiconv` merges them into one `layout.json`, writes every texture it
names as PNG, copies the TrueType fonts, parses the bitmap fonts and converts
the icon sheets and the classic skin's window art. `web/uiwin.js` builds the
DOM from that; `web/hud.js` and `web/pack.js` are the windows. Without
`web/data/ui/` the plain built-in panels are used, so nothing depends on it. The
developer panel (zone, grid, triangles, key help) is hidden while playing;
the backquote key shows it.

What the install does and does not hold:

- **Bitmap fonts carry their own metrics.** Under each row of glyphs in a font
  sheet is a marker row of opaque segments, one per glyph, whose widths are the
  advances; characters run in Windows-1252 order from `!`. One sheet
  (`brit9`) breaks the rule and is left out.
- **Chat, the summary and experience windows have layout files.** The quick
  bar, style bar and inventory do not: the inventory is drawn over
  `data/skin1/pperdoll.bmp` with slot positions measured from the art, and
  the style bar is built by hand.
- **The loot window has no layout either.** It is built on the stock frame
  with the stock button; the Release window, menu bar, compass and clock are
  built from their layouts. The compass scrolls a 512-pixel strip by the camera's
  heading, with north the zone's -y (checked against the zone offsets and across
  a zone border), and the clock shows the server's game time (real time times
  24, with `DAOC_CLOCK_SHIFT` to move it). What was watched in the browser is in
  `openspec/changes/archive/2026-10-22-midgard-playable/proofs.md`.
- **Icons are drawn by kind, not by number.** The icon sheets are 10 by 10
  grids of 32-pixel icons, one file per hundred numbers. The imported
  database's icon numbers come from a later client: only about a tenth of
  gear fits a sheet here and style numbers fit no single rule. So a style is
  drawn with a picture of its weapon line and an item with a picture of what
  it is.

### Tests

```bash
go test ./internal/server/...    # accounts, the protocol, movement, collision, creatures, combat, loot, a real WebSocket session
go test ./internal/ui            # the interface readers, against the install when it is there
node web/net.test.mjs            # the browser's movement and prediction against the server's vectors
node web/collision.test.mjs      # and its collision
node web/uiwin.test.mjs          # the interface renderer's data handling
node web/pack.test.mjs           # the inventory window's geometry
node web/loot.test.mjs           # the loot rows
node web/chrome.test.mjs         # the clock text
```

`web/uidemo.html` draws a handful of the stock windows and every font for a
look at them (`/uidemo.html` on the running server). `?w=Merchant,mini_pet` draws just
those windows (a WindowId or a template name, with sample rows in their lists) and
`?w=all` every window the game builds from the stock layouts; a control type the
builder cannot draw is named once on the console.

## Known limitations

- **Day and night are a tint, not the sky file's.** The game clock (real time
  times 24) drives a darkening and a blue cast over the converted clear day, and
  night-only creatures appear at night; the sky file's own dawn, dusk, night,
  stars, a moon and storms are not converted, and the sun's direction is chosen
  here rather than following the clock. The point lights in an
  outdoor zone's `lights.csv` are not lit; only cities and dungeons light
  theirs, and those cast no shadows. Cast shadows reach 2400 units from the player;
  beyond that only slope lighting remains.
- **Grass grows through floors.** The density map keeps it off Mularn's
  square, but nothing stops a tuft inside a building or under a fence on
  ground the map marks as meadow; the shapes are a reading of the atlas, not
  the game's code.
- **Some fire pits are cold.** The 13 `x06_campfire` stone rings carry no
  emitter and nothing in the zone's files adds one, so they stay unlit. The
  forge sparks' `NiParticleBomb` impulse is not modelled, and particles are
  not sorted, so overlapping smoke can blend in the wrong order.
- **Two textures are missing from the install.** `BAG.nif` names `mfiga6.dds`
  and `mheada3.dds`, which exist nowhere in the game directory; those three
  props draw white.
- **Most of the game's animation clips are still unused.** Idle, walk, run, death, the
  combat stance, flinch, three attacks, casting, jumping, landing, sitting, swimming and
  the sixteen classic emotes are converted per race; the other ~2000 humanoid clips
  (the later emotes, the other sit poses, ladders, boats) are reachable by name through
  `animconv` but not wired to anything.
- **5 of 4032 `.kfa` files do not parse**, each with a non-finite value in
  its keyframe data. `-list` reports them; none is in the humanoid
  locomotion set.
- **Offline, there is no collision.** The single-player viewer still walks
  through anything but the heightmap. In network mode trees, buildings and
  dungeon walls block; models with neither a hull nor a trunk radius, such as
  small props, do not.
- **Interiors are reached through their exits, and stay thin.** A dungeon
  mouth in an outdoor zone takes a player in, and the way back out. Dungeons
  have no water: the game's pools and underground rivers are not converted.
  Trollheim (94 of 138 exits and lights) and Tuscaran Glacier (41 of 51)
  fit worse than the rest, because their layouts leave parts out; see
  [Cities and dungeons](docs/formats.md#cities-and-dungeons). Tuscaran Glacier and Midgard's
  Darkspire are almost unlit at the start. Nothing falls: a ledge is a wall.
- **Population and numbers are a community database's, not the live game's.**
  Spawns, levels, loot tables and styles are the OpenDAoC database's, which
  models patch 1.65; the hit points, damage, experience and weapon-skill
  numbers are this project's. Creatures and players cast, and styles and
  weapons fire their procs, but there are no quests, crafting,
  realm-versus-realm or housing (tasks only, to level 20). A creature
  type with no converted character is left out, and style icons are not
  converted. The table at the top says how much of each is done.
- **Line of sight is provisional.** Hulls and trunks block, terrain is sampled every
  128 units, and floors and ceilings count only between storeys; eye height and the
  storey rule are this project's choice, to be tuned in play. Melee needs the creature
  to reach you, and a ranged shot does not exist yet.
- **1 of 633 figures does not parse** — see [Parsing the rest of
  figures/](docs/formats.md#parsing-the-rest-of-figures).
- **Some of what is read is not drawn.** Projected and environment textures,
  vertex morphs, alpha, colour, visibility and UV animation, flipbooks, fog
  properties and particle collisions are read exactly so the models holding
  them parse, and those models draw without them;
  [`docs/census/README.md`](docs/census/README.md) counts the files carrying
  each. Point gravity is approximated as it acts at the emitter.
- **Midgard is checked by eye; the rest of the world is measured.** All 25
  Midgard outdoor zones draw their terrain, ground detail, grass and every
  placement, and a test holds them there. The census converts 98% of
  outdoor placements across all 118 zones `worldconv` converts, and the
  viewer opens any of them from its zone picker, without NPCs and starting
  in the middle of the map; but the other realms' zones have only been
  measured, not looked at. Fifty-four zones use an older `fixtures.csv` with
  a heading but no axis-angle, and Z 0 for anything on the ground; Midgard's
  look right, the rest are unchecked. Ground detail agreement is lowest in
  the forests (Myrkwood Forest 0.56), where the game's tiles carry the
  trees' shading. Of the city blocks, 169 of 174 read; the five that do
  not are `_p00` files with a bad header, none of them placed by
  `city.csv`. The other realms' cities and dungeons convert but, like
  their outdoor zones, have only been measured.
- **Faces, not hair.** A humanoid's head wears the Face skin from
  `monsters.csv`; the hair is a separate style number that is not mapped, so
  a head shows only what its face texture paints.
- **Character options are thin.** A character is one of the twelve playable
  classes in the races the start table allows, male or female, with no face,
  hair or name-style choices beyond that. Repair is a plain panel; the stock
  frame has none. The clock has no plain window when the converted interface is
  absent. The journal, map, war map and main-menu buttons are greyed, with a
  tooltip saying why.
- **A long session in one tab can stall the GPU.** Moving between zones
  reloads the page, and the viewer hands its WebGL context back as it goes;
  even so, after many zones in one tab the browser can stop drawing. A fresh
  tab clears it.
- Heightmap edges stop at sample 255 (65280 units), 256 units short of the
  nominal zone edge. Harmless here; matters when stitching zones together.

## Licensing

The converter and viewer are original work, MIT licensed — see `LICENSE`.

The game assets they read are copyright Electronic Arts / Broadsword Online
Games. Converting them for your own use is one thing; redistributing the
converted output is not, so **no game data is in this repository**: `web/data/`
is gitignored, and you generate it by running the converters against your own
install. Without it the viewer has nothing to show and `skeleton.test.mjs` has
nothing to check.

Nothing here touches the live game servers or its network protocol.
