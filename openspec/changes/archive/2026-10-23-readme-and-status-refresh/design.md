# Design

## Context

README.md is 1800 lines: converter pipeline first, "Multiplayer" at line 1303, three parity snapshots at 1568. `docs/parity/parity.json` already holds the current numbers; CLAUDE.md says to read the README for detail, so the file-format sections must stay findable. See proposal.md.

## Goals / Non-Goals

- Goals: a README a friend can read in minutes; no loss of the format write-ups; numbers that cannot drift unnoticed.
- Non-Goals: implementing features; changing the census; rewriting the format prose.

## Decisions

- **Move, don't rewrite, the format sections.** They go verbatim to `docs/formats.md` with their headings, so existing `README.md#...` anchors in code comments and specs are updated by grep. Alternative: split by topic into many files; rejected as churn with no reader benefit.
- **Status table is hand-written from `parity.json`, plus a test.** A small Go test (or extension of `go test ./cmd/parity`) compares the README's figures with the JSON, so the README cannot go stale silently. Alternative: generate the section; rejected because the table has prose around it.
- **Archive first.** `openspec archive` syncs the two finished changes into `openspec/specs/`; the README then links to final specs.
- **findings.md, not a second change.** Ranking the path to level 50 needs reading the items, trainer and class data; it is recorded here and a follow-up is proposed separately.

## Risks / Trade-offs

- [Broken links to README anchors] → grep for `README.md#` and `](#` across the repo and fix each.
- [Archive sync conflicts with existing specs] → run `openspec validate` before and after.
- [README test is brittle] → compare only the implemented/total figures.
