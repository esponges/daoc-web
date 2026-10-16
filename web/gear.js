// What a creature wears and carries, from the equipment the server sends.
//
// The server describes each worn piece as [slot, model, colour, extension]
// using the community database's numbers. For held items (slots 10 right
// hand, 11 left hand, 12 two-hand, 13 ranged) the install's items.csv says
// which model a number draws; charconv -gear converts those once into
// data/items/<name>/ and writes data/gear.json, the lookup read here. A
// number the lookup lacks, or one whose kind does not fit the slot (a bow in
// a hand slot, a shield in the right hand), is skipped, so the creature
// keeps its plain model. With no gear.json at all nothing is drawn.

export const SLOT = { right: 10, left: 11, twoHand: 12, ranged: 13 };

// KINDS mirror gamedata.Kind* in Go.
const FITS = {
  [SLOT.right]: ['onehand', 'twohand', 'melee'],
  [SLOT.left]: ['shield', 'onehand', 'melee'],
  [SLOT.twoHand]: ['twohand', 'melee'],
  [SLOT.ranged]: ['ranged'],
};

export function kindFitsSlot(kind, slot) {
  return (FITS[slot] || []).includes(kind);
}

// The skeleton socket a piece is held on. Shields have their own; a bow is
// drawn in the left hand.
export function socketFor(slot, kind) {
  if (slot === SLOT.left && kind === 'shield') return 'Bip01 L Shield';
  if (slot === SLOT.left || slot === SLOT.ranged) return 'Bip01 L Held';
  return 'Bip01 R Held';
}

// resolveHeld picks the held pieces of an equipment list that the lookup can
// draw. It returns what to draw and how many held pieces it had to skip.
export function resolveHeld(eq, table) {
  const pieces = [];
  let skipped = 0;
  if (!table || !table.items) return { pieces, skipped };
  for (const [slot, model] of eq || []) {
    if (slot < SLOT.right || slot > SLOT.ranged) continue;
    const it = table.items[model];
    if (!it || !kindFitsSlot(it.kind, slot)) { skipped++; continue; }
    pieces.push({ slot, item: it.item, socket: socketFor(slot, it.kind) });
  }
  return { pieces, skipped };
}

// --- armour ---
//
// The install does not say which art the database's armour numbers mean, so
// armour is dressed by tier. gear.json (charconv -gear) says what each worn
// number is: an armour level from 0 cloth to 4 plate, and a realm, read from
// the imported items that use the number. A humanoid creature's mesh carries
// every tier of each slot (shapes tagged slot and tier); a worn piece picks
// the tier for its level, a texture of that realm and level, and a tint.

// The tier of each slot's shapes that matches a level 0 to 4. Slots have
// different numbers of tiers (arms six, gloves three), so the levels share
// some; the last tier of a slot is its plate.
const TIER = {
  body: [1, 2, 3, 4, 5], lbody: [1, 2, 2, 3, 4], arms: [1, 2, 3, 4, 6],
  legs: [1, 2, 2, 3, 4], gloves: [1, 1, 2, 2, 3], boots: [1, 2, 3, 4, 5],
};

// The shape slots each worn database slot dresses. The torso piece covers
// both the short and the long body.
const PIECE_SLOTS = { 22: ['gloves'], 23: ['boots'], 25: ['body', 'lbody'], 27: ['legs'], 28: ['arms'] };

const REALM_LETTER = { 1: 'b', 2: 'n', 3: 'h' };

// tiersOf lists the tiers a character's shapes have for each slot.
export function tiersOf(shapes) {
  const out = {};
  for (const s of shapes || []) {
    if (!s.slot) continue;
    (out[s.slot] ||= []).push(s.tier);
  }
  for (const k of Object.keys(out)) out[k] = [...new Set(out[k])].sort((a, b) => a - b);
  return out;
}

// pickTier is the wanted tier, or the nearest the character has below it
// (or its lightest if it has none that low).
export function pickTier(avail, want) {
  let best = avail[0];
  for (const t of avail) if (t <= want) best = t;
  return best;
}

// The realm letter a creature defaults to when a worn number names none,
// from the file its model came from ("HFirbolgM.NIF" is Hibernian).
export function defaultRealm(source) {
  const c = (source || '').charAt(0).toUpperCase();
  return c === 'H' ? 'h' : c === 'B' ? 'b' : 'n';
}

// resolveArmour turns a creature's worn armour into a look: per shape slot the
// tier to draw, the texture URL to dress it in, and an optional tint. Pieces
// whose number the lookup does not know are skipped and counted.
export function resolveArmour(eq, table, shapes, { female = false, realm = 'n' } = {}) {
  const look = { tiers: {}, tex: {}, tint: {}, skipped: 0 };
  const armor = table && table.armor;
  if (!armor) return look;
  const have = tiersOf(shapes);
  for (const [slot, model, color] of eq || []) {
    const targets = PIECE_SLOTS[slot];
    if (!targets) continue;
    const piece = armor.pieces[slot + ':' + model];
    const letter = piece && (REALM_LETTER[piece.realm] || realm);
    const set = piece && armor.textures[letter + ':' + piece.level];
    if (!piece || !set) { look.skipped++; continue; }
    const rgb = color > 0 ? armor.tints[color] : null;
    for (const t of targets) {
      if (!have[t]) continue;
      look.tiers[t] = pickTier(have[t], TIER[t][piece.level]);
      const file = t === 'body' ? (female ? set.bodyF : set.body) : set[t];
      look.tex[t] = 'data/armor/' + file + '.png';
      if (rgb && !(rgb[0] > 250 && rgb[1] > 250 && rgb[2] > 250)) look.tint[t] = rgb.map((v) => v / 255);
    }
  }
  return look;
}

// createGear loads the lookup once and each item model once, however many
// creatures carry it. loadItem(name) returns a promise of a drawable
// character (see createCharacter); fetchJson(url) a promise of parsed JSON.
export function createGear({ loadItem, fetchJson }) {
  let table = null;
  let tableState = 'loading';
  fetchJson('data/gear.json').then((t) => { table = t; tableState = 'ok'; }).catch(() => { tableState = 'none'; });

  const items = new Map(); // name -> { state, ch }
  function item(name) {
    let it = items.get(name);
    if (it) return it;
    it = { state: 'loading', ch: null };
    items.set(name, it);
    loadItem(name).then((ch) => { it.ch = ch; it.state = 'ok'; })
      .catch((e) => { it.state = 'fail'; console.warn('item ' + name + ' not loaded:', e.message); });
    return it;
  }

  let skipped = 0;
  let armourSkipped = 0;
  return {
    get ready() { return tableState === 'ok'; },
    get skipped() { return skipped + armourSkipped; },
    // heldFor resolves a creature's equipment once; call it with the same
    // entity every frame and it only recomputes when the equipment changes.
    heldFor(view, eq) {
      if (tableState === 'loading') return [];
      if (view.heldEq !== eq || view.heldTable !== table) {
        const r = resolveHeld(eq, table);
        view.heldEq = eq; view.heldTable = table; view.held = r.pieces;
        skipped += r.skipped;
      }
      return view.held;
    },
    // lookFor resolves a creature's armour once per equipment, like heldFor.
    // ch is the character the creature is drawn with.
    lookFor(view, eq, ch) {
      if (tableState === 'loading' || !table || !table.armor) return null;
      if (view.lookEq !== eq || view.lookTable !== table) {
        const shapes = ch.manifest.shapes;
        const female = shapes.some((s) => s.slot === 'body' && s.base && /f.png$/i.test(s.texture || ''));
        view.look = resolveArmour(eq, table, shapes, { female, realm: defaultRealm(ch.manifest.source) });
        view.lookEq = eq; view.lookTable = table;
        armourSkipped += view.look.skipped;
      }
      return Object.keys(view.look.tiers).length ? view.look : null;
    },
    // draw puts each loaded held item on its socket in the pose just drawn.
    // skel is the posed skeleton of the creature's character.
    draw(view, eq, skel, ctx) {
      for (const p of this.heldFor(view, eq)) {
        const it = item(p.item);
        if (it.state !== 'ok') continue;
        const socket = skel.boneId(p.socket);
        if (socket < 0) continue;
        it.ch.skeleton.world[0].set(skel.world[socket]);
        it.ch.draw(ctx);
      }
    },
  };
}
