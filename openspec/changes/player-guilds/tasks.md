# Tasks

## 1. Sources and assets

- [ ] 1.1 Find the emblem textures and the cloak and shield emblem slots in the install (`nifdump`, `mpakls`). Write them up in `docs/formats.md`.
- [ ] 1.2 Pin the formation fee, the emblem fee and the member limits from 1.65 sources. Add a `guilds` rule-book entry with a test and a registry entry.

## 2. Server

- [ ] 2.1 Store migration and repository for guild, rank, member and alliance (D1). Tests.
- [ ] 2.2 `world/guilds.go`: form (D2), invite, leave, remove, promote, demote, edit permissions, motd. Tests per spec scenario.
- [ ] 2.3 Guild, officer and alliance channels; motd at login; guild in `/who` and on entities.
- [ ] 2.4 The Guild Registrar and Emblem NPC as services; the emblem purchase.
- [ ] 2.5 Protocol additions in `protocol.go`, `docs/protocol.md` and a protocol test.

## 3. Client

- [ ] 3.1 The stock Guild window from its template (members, ranks, notes, sort, show offline, buttons), with a plain fallback.
- [ ] 3.2 Name plate guild line; emblems on cloaks and shields (D3).
- [ ] 3.3 Check in the browser with 8 test characters (admin tools), then the fallback.

## 4. Census and docs

- [ ] 4.1 `cmd/parity` marks GuildRegistrar and EmblemNPC implemented. README and CLAUDE.md.
