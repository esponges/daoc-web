# Tasks

## 1. Sources and assets

- [x] 1.1 Find the emblem textures in the install and write them up in `docs/formats.md`. (Found: `insignia/{b,h,n}_emblem_001-084.dds`. Which part of a cloak or shield takes them is not worked out, so they are not painted on gear: see 3.2.)
- [x] 1.2 Pin the formation fee, the emblem fee and the member limits from 1.65 sources. (Searched: the Herald gives the shape, no numbers; the fees, the 8-player minimum and the limits are provisional.) Rule-book entries with tests and registry entries.

## 2. Server

- [x] 2.1 Store migration and repository for guild, rank, member and alliance (D1). Tests.
- [x] 2.2 `world/guilds.go`: form (D2), invite, leave, remove, promote, demote, edit permissions, motd. Tests per spec scenario.
- [x] 2.3 Guild, officer and alliance channels; motd at login; guild in `/who` and on entities.
- [x] 2.4 The Guild Registrar and Emblem NPC as services; the emblem purchase.
- [x] 2.5 Protocol additions in `protocol.go`, `docs/protocol.md` and a protocol test.

## 3. Client

- [x] 3.1 The Guild window (the stock layout has no Guild window, so it is a plain panel): members, ranks, notes, sort, show offline, buttons.
- [ ] 3.2 Name plate guild line; emblems on cloaks and shields (D3).
- [ ] 3.3 Check in the browser with 8 test characters (admin tools), then the fallback.

## 4. Census and docs

- [ ] 4.1 `cmd/parity` marks GuildRegistrar and EmblemNPC implemented. README and CLAUDE.md.
