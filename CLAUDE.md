# daoc-web

# Most Importantly
I trust you and your proposed work. Any work that can be verified by you, it should not be
delegated to me for retest. I'll go with your suggestions 9 of 10 times. So unless you really
need to hear about me is something you're not certain at all.

# The GOAL
This is not a POC anymore. We want to have the closest version to 1.65 version playable
via web. For the moment this is just a close friend version. Don't worry about resources
required to scale the application (servers, CDN's, replicas, etc). That's future work

A fan project that rebuilds Dark Age of Camelot's Midgard in the browser. It
started as a set of Go converters that read a local DAoC install and a WebGL2
viewer for the result, and is now a multiplayer game: a Go server that owns the
world, the viewer as its client, and the game's own interface drawn from the
install's UI data. The repo is MIT and contains no game data.

`README.md` is the long reference (file formats worked out from the install,
how each converter works, the server's rules). Start there for detail; this file
is only what you need to work here safely.

## Layout

| path | what |
| --- | --- |
| `cmd/*` | the tools: `zoneconv`, `propconv`, `charconv`, `animconv`, `interiorconv`, `worldconv` (run them over every zone), `uiconv` (the interface), `dbimport` (the world store), `gameserver`, `serve` (static files), `census`, `charshot`, `nifdump`, `mpakls` |
| `internal/*` | readers and converters: `mpak`, `nif`, `dds`, `kfa`, `zone`, `scenery`, `interior`, `ui` (interface XML, TGA, BMP, bitmap fonts); `importer` (SQL dump parser); `server/{auth,store,protocol,world,gameserver}` |
| `web/` | the browser client: `viewer.js` (main), `net.js` (connection and prediction), `ui.js` + `hud.js` + `pack.js` + `loot.js` + `chrome.js` (windows), `uiwin.js` (renders windows from the converted layout), `*.test.mjs` |
| `docs/protocol.md` | the wire format; messages are JSON, versioned, and only ever added to |
| `openspec/` | specs (`specs/`) and changes (`changes/`, `changes/archive/`) |

Gitignored and never committed: `web/data/` (everything converted from the
install, including `web/data/ui/`) and `server/data/` (the community database
checkout, `world.db`, `players.db`, `session.key`, `server.env`).

## Rules that matter

- **No game data in git.** Art, models and the imported database describe
  EA-owned content. Convert from your own install into `web/data/`; never
  commit output or copy it into a spec or test fixture.
- **The community database is GPL-3.0** (OpenDAoC-Database). It is imported at
  build time into `server/data/` by our own importer. Do not commit it and do
  not copy GPL code; the rules in `internal/server/world/rules.go` are our own.
- **The server decides everything.** The client sends what the player holds
  down and what they ask for. Movement is fixed 1/60 s steps shared by Go
  (`world/move.go`) and JS (`web/move.js`), checked against shared vectors.
- **The built-in UI is the fallback.** Anything built from `web/data/ui/` must
  fall back to the plain panels when that folder is absent.
- **A humanoid's `Head` skin in `monsters.csv` is its hair colour**; the face is the
  `Face` column. `charconv -model` uses the Face skin (see the README, "Character
  textures").
- **Icon numbers in the imported data do not match this install's sheets**, so
  styles and items are drawn by weapon line and kind. Do not "fix" this by
  guessing an offset (see `openspec/changes/archive/2026-10-19-stock-ui-followups/design.md`).

## Where things come from

- **Art, models, zones, sounds, interface layouts:** your own DAoC install
  (path under "Running it"). The converters read it directly; the file formats
  were worked out from the install's files and are documented in the README.
  Nothing converted is committed.
- **World population and numbers** (spawns, creature templates, levels, loot,
  items, styles): the [OpenDAoC database](https://github.com/OpenDAoC/OpenDAoC-Database),
  GPL-3.0, models patch 1.65. Cloned into `server/data/community/` and read by
  `cmd/dbimport`. The client install has none of this. It describes EA-owned
  content, so use stays private; re-check the licence and the EA question before
  hosting publicly. README: "Where the world's content comes from".
- **Rules and formulas** (hit points, damage, experience, weapon skill): the
  target is classic 1.65. They are written here in `internal/server/rules`,
  from the rule book (README: "The rule book"). OpenDAoC-Core and DOLSharp
  (GPL) may be read to understand a mechanic, but nothing is copied from them.
- **Sources adopted for the rules:**
  - disorder.dk's 1.65 mechanics pages: hit, evade, parry, block, stats and caps,
    hit point cap, AF cap, swing speed, hit location, skill cap, melee damage;
  - the Camelot Herald wiki: melee damage, weapon skill, styles;
  - Mythic's Grab Bag answers, and the knowledge-base note on the 1.25 experience
    clamp;
  - the published experience and constitution-restore tables (silicondragon.com);
  - Mythic's con colour band description;
  - ZAM's starting-zones table (camelot.allakhazam.com/Start_Zones.html): the
    start town of each race and class;
  - A 2001 forum thread on Guard, Protect, Intercept and Engage
    (quartertothree.com/boards/messages/17/1030.shtml): outline only, no numbers;
  - OpenDAoC-Core, for reference-status rules only.

  If you use another source (a wiki, a format write-up, a community table), add
  it here and to the README.

## Running it

The install is expected at
`C:\Program Files (x86)\Electronic Arts\Dark Age of Camelot` (`-game` changes it).

```bash
go run ./cmd/worldconv -only outdoor            # zones (long); also -only interiors
go run ./cmd/charconv -outfit troll-warrior && go run ./cmd/animconv -char web/data/char/troll-warrior
go run ./cmd/uiconv -skin atlantis              # the stock interface -> web/data/ui/
git clone https://github.com/OpenDAoC/OpenDAoC-Database server/data/community
go run ./cmd/dbimport                           # world population -> server/data/world.db
DAOC_ADMIN_PASSWORD=choose-one go run ./cmd/gameserver    # http://localhost:8777
```

The README's Multiplayer section has the full first-run recipe (every creature
model is converted separately) and the settings table (`DAOC_ADMIN_PASSWORD`,
Discord login, `DAOC_PUBLIC_URL`, `server/data/server.env`). Log in at
`/login.html`, make a character, press Play. `/?zone=105` is the offline viewer;
`?play=<characterId>` is network mode. In play the developer panel is hidden;
the backquote key shows it.

After changing server code, restart the server: a running `gameserver.exe` is
the old build.

## Testing

```bash
go test -short ./...        # -short skips the census over a real install
for t in web/*.test.mjs; do node $t; done
openspec validate <change>
```

Tests that read the install skip when it is absent. For UI work also look at it
in a browser (`/uidemo.html` draws sample stock windows); passing node tests do
not show that a window looks right. Check the fallback by temporarily moving
`web/data/ui/`.

## Working here

- **Specs drive work.** Changes live in `openspec/changes/<name>/` (proposal,
  specs, design, tasks); `/opsx:propose`, `/opsx:apply` and `/opsx:archive`
  run them. Archived changes sync into `openspec/specs/`. Open now:
  `midgard-playable`.
- **Game rules have a rule book.** Every mechanic the server decides has a
  `rules-<mechanic>` spec giving the formula in our words, its sources, a status
  (documented, reference or provisional) and worked examples. Each numbered
  scenario is a table test in `internal/server/rules`, and the rule is
  registered there so `cmd/parity` counts it. A change that adds or alters a
  rule updates its spec, test and registry entry together.
- **Protocol changes are additive.** New messages and fields only; document
  them in `docs/protocol.md` and add a protocol or server test.
- **Go** is not on PATH in this environment: `~/goroot-dist/go/bin`.
- **Shell:** long heredocs and scripts with backslash escapes get mangled here.
  Write files with the editor tools, and in JS replacement strings avoid `$`
  patterns (`String.replace` with a function argument is safe).
- **Browser automation:** a held key cannot be sent with the key action. Move
  the character by dispatching `keydown` and `keyup` events from page script
  with a delay between them. On a world store imported before templates were
  honoured, creatures near the Galplen start can kill a level 1 character in
  two blows (re-import; see `midgard-playable` design 1a).
- Commit only when asked; end commit messages with the attribution lines the
  session gives.
