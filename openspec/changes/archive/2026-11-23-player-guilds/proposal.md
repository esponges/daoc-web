# Proposal: player-guilds

## Why

Guilds were the social backbone of 1.65 and there are none here. Midgard's data places two Guild Registrars and two Emblem NPCs, which the census lists as not implemented. The stock layout already has the `Guild` window and its commands: withdraw, promote, demote, note, members sort, show offline and send.

## What Changes

- **Forming a guild:**
  - a group leader with a full group of 8 asks a Guild Registrar to form a guild and pays the fee;
  - every member is asked to agree, and all must;
  - the name must be unique and pass the name rules.
- **Membership:**
  - `/gc invite`, `/gc remove` and `/gc leave`;
  - an invite is answered through the shared accept/decline offers;
  - a character is in at most one guild.
- **Ranks:**
  - ranks 0 (Guildmaster) to 9;
  - per-rank permissions: speak on guild chat, hear and speak on officer chat, invite, promote, remove, emblem, and the message of the day;
  - `/gc promote`, `/gc demote` and `/gc edit`.
- **Chat:** the Guild channel (`/gu`) and the Officer channel (`/o`), with the guild's message of the day at login.
- **Emblem:** the Guildmaster buys an emblem at an Emblem NPC. Cloaks and shields worn by members show it, and other players see it.
- **Alliances:**
  - `/gc ally` forms an alliance between guilds;
  - the alliance has its own channel (`/as`), and only the leading guild's Guildmaster can disband it.
- **Interface:**
  - the stock Guild window: members, ranks, notes, sort, show offline;
  - the guild name under the character's name in the world;
  - the guild shown in `/who`.
- **Left out:** guild houses and guild banks, which came with housing after 1.65.

## Capabilities

### New Capabilities
- `player-guilds`: forming, membership, ranks and permissions, guild, officer and alliance chat, the message of the day, emblems and the Guild window.

### Modified Capabilities
- (none)

## Impact

- **Data:** `players.db` gains guild, member, rank and alliance tables (a migration).
- **Server:** new `world/guilds.go`; the Guild Registrar and Emblem NPC as services; chat routing for three new channels.
- **Protocol (additive):** guild state, membership updates, and a guild name and emblem on entities.
- **Client:**
  - the stock Guild window with a plain fallback;
  - name plates;
  - emblem textures on cloaks and shields. The install's emblem sheets are converted by `uiconv` or `charconv` (task 1.1 finds them).
