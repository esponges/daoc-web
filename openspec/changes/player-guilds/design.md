# Design: player-guilds

## Context

- `players.db` is the player store (`internal/server/store`), and it already migrates on start.
- Chat channels live in the world, and the stock chat config already filters "Guild", "Officer" and "Alliance".
- The Guild Registrar (`DOL.GS.GuildRegistrar`) and Emblem NPC placements are in the import but skipped.

## Decisions

**D1. Guilds are store rows, cached in the world.**
- Tables:
  - `guild(id, name, realm, motd, omotd, emblem, alliance_id, created)`;
  - `guild_rank(guild_id, rank, title, perms)`;
  - `guild_member(char_id, guild_id, rank, note)`;
  - `alliance(id, leader_guild_id, motd)`.
- The world loads guilds with online members, and writes through on change.

**D2. Forming uses the group and the shared offers.** The registrar's "form" request sends a form offer to each member. Once all accept, the guild is created in one transaction.

**D3. Emblems.**
- The emblem is a number packing logo, pattern and colours, as in the client's emblem picker.
- Task 1.1 finds the install's emblem textures and the cloak and shield slots that take them.
- The viewer composes the emblem onto the cloak and shield texture at load.
- *Fallback:* when the textures are missing, the cloak and shield are plain.

**D4. Fees and limits** go in the rule book as a `guilds` entry. Status: documented where the Herald gives them, provisional otherwise.

## Non-Goals

- Guild houses, guild banks and bounty points. These came after 1.65.
