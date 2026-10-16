// Which converted base model draws a player of a given race and gender.
//
// charconv -races writes data/races.json: for each database race number the
// converted character (a monsters.csv row, "m137" for Troll Male) of each
// gender. A race the file lacks, a gender it lacks for that race, or no file
// at all gives no base, and the caller keeps drawing the player with the
// archetype's outfit (`char` in the welcome and the snapshot).

// baseFor is the converted character directory for a race and gender, or null.
export function baseFor(table, race, female) {
  const r = table && table.races && table.races[race];
  const key = r && (female ? r.f : r.m);
  return key || null;
}

// loadRaces fetches the table once. fetchJson(url) returns a promise of parsed
// JSON; a missing file is not an error, it leaves every player on the outfit.
export function loadRaces(fetchJson) {
  const r = { table: null, state: 'loading' };
  r.done = fetchJson('data/races.json')
    .then((t) => { r.table = t; r.state = 'ok'; })
    .catch(() => { r.state = 'none'; });
  return r;
}

// candidates lists the directories to try for a player, best first: the race
// and gender base, then the outfit the server named.
export function candidates(table, { race, female, char }) {
  const out = [];
  const base = baseFor(table, race, female);
  if (base) out.push(base);
  if (char && !out.includes(char)) out.push(char);
  return out;
}
