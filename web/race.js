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

// skinFor is the head and hair textures a player chose, as `{Head, Hair}`
// urls keyed by shape-name prefix, or null for the model's own. table is
// data/appearance.json (charconv -appearance); face and hair are the 1-up
// numbers the server stores, 0 for the default.
export function skinFor(table, race, female, face, hair) {
  const l = table && table.races && table.races[race] && table.races[race][female ? 'f' : 'm'];
  if (!l || !(face || hair)) return null;
  const out = {};
  if (face && l.faces && l.faces[face - 1]) out.Head = 'data/appearance/' + l.faces[face - 1] + '.png';
  if (hair && l.hair && l.hair[hair - 1]) out.Hair = 'data/appearance/' + l.hair[hair - 1] + '.png';
  return Object.keys(out).length ? out : null;
}

// loadAppearance fetches the table; a missing file leaves everyone on the
// model's own face and hair.
export function loadAppearance(fetchJson) {
  const r = { table: null };
  r.done = fetchJson('data/appearance.json').then((t) => { r.table = t; }).catch(() => {});
  return r;
}
