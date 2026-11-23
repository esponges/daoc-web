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
- [x] 3.2 Name plate guild line. Emblems are not painted on cloaks and shields: which part of those figures takes the texture is not worked out (`docs/formats.md`), so the emblem number is shown in the guild window instead.
- [x] 3.3 Checked in the browser with two characters on a scratch server (`/gm guild`, an invitation, the window, guild chat, the name plate). A formation by a group of 8, officer and alliance chat are covered by server tests, not played. The window is a plain panel, so there is no stock frame to fall back from.

## 4. Census and docs

- [x] 4.1 `cmd/parity` marks GuildRegistrar and EmblemNPC implemented. README and CLAUDE.md.
