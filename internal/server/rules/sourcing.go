package rules

// Sourcing pass of 2026-11-24 (openspec rules-sourcing-pass): for each rule
// still without a documented 1.65 source, what was searched and what turned
// up, so the search is not repeated. A rule leaves this list when a source is
// found and its status changes. "Searched" is not a claim the rule is wrong.
const searchedOn = "2026-11-24: "

var searched = map[string]string{
	"melee.left-hand":                         "A 2011 test page quoted on a freeshard forum gives the same 0.625 + 0.0034 x spec share; no patch named; Hand to Hand blow chances and left-hand swing speed not found.",
	"archery.damage":                          "Camelot Herald and freeshard pages searched; 1.88I (2007) made bow DPS count, after 1.65; no 1.65 damage formula found.",
	"archery.draw":                            "No draw time or Rapid Fire figure found in Herald, Allakhazam or Grab Bag pages.",
	"archery.range":                           "No bow range figure found.",
	"archery.abilities":                       "A 2003 note gives Penetrating Arrow three ranks at 30, 40 and 50 bow spec with no percentages; Critical Shot is said to need stealth, which ours does not; nothing on Sure Shot, Volley endurance or recasts.",
	"ability.berserk":                         "1.65 and 1.65A to I patch notes and class pages searched; nothing for the Berserker stance numbers.",
	"ability.charge":                          "Same search; nothing for Charge.",
	"ability.quickcast":                       "Same search; nothing for QuickCast.",
	"ability.danger-sense":                    "Same search; nothing for Danger Sense.",
	"ability.distraction":                     "Same search; nothing for Distraction.",
	"savage.buffs":                            "1.65H notes support percent-of-maximum-health costs; a search snippet of the Savage page shows buffs with a cost near 4 to 5 percent; the full table could not be loaded.",
	"savage.endurance-heal":                   "Nothing found for the endurance heal's size.",
	"savage.taunt":                            "Nothing found for Taunt's hate value.",
	"character.endurance":                     "Forum posts date the standing-regeneration change to 2006; no per-second figures found.",
	"character.health-regen":                  "No per-tick figures found; tick timing threads are from 2009.",
	"power.regen":                             "A 2009 Live test gives about 15 power a tick at level 50 (ours 23), a 3 second out-of-combat tick and a 10 second in-combat tick; 1.87 or later, so not adopted.",
	"resting.seated":                          "Sitting is about twice standing before 1.87, as ours; the 1.87 change is the project's source already.",
	"falling.damage":                          "The 1.38 notes give the shape only; no height or percent found.",
	"falling.safe-fall":                       "The 1.38 notes give the shape only.",
	"falling.water":                           "Nothing found on water landings.",
	"movement.jump":                           "Nothing found for jump speed or gravity.",
	"movement.swim":                           "Nothing found for swim speed.",
	"conc.end":                                "Behaviour, not a number; no source found.",
	"condition.wear":                          "A Grab Bag snippet describes durability lost per point of condition, not a rate per blow.",
	"condition.repair-price":                  "No repair price formula found.",
	"condition.warning":                       "No warning thresholds found.",
	"stealth.detection":                       "A Ranger page lists detection range by stealth bonus tier (75 at 10 percent up to 625 at 50), not our level-gap formula; its patch era is unconfirmed.",
	"experience.group-share":                  "A 2002 Herald post relayed by GameBanshee gives colour bands and no group size bonus; the share rule is not stated.",
	"experience.group-split":                  "The same post says an even split within a colour range, not by level; a group challenge rule (a group of 2 needs a yellow monster, 4 orange, 8 red) is not modelled.",
	"experience.group-grey":                   "The same post agrees: a monster grey to the highest member gives nobody experience.",
	"character.class-hunter":                  "No public class table found (base hit points, weapon skill base, spec multiplier).",
	"character.class-shadowblade":             "No public class table found.",
	"character.class-spiritmaster-bonedancer": "No public class table found.",
	"character.class-thane-skald":             "No public class table found.",
	"character.class-midgard":                 "No public class table found.",
	"spell.area":                              "No area falloff or radius rule found.",
	"spell.buff-af":                           "No source gives a buff armour cap of 2 x level; a 2004 proposal and a 2018 Grab Bag give other readings.",
	"spell.haste":                             "One forum formula agrees with ours but floors the swing at 1.5 seconds, not 1.0; unconfirmed.",
	"spell.health-regen":                      "No 1.65 source found.",
	"spell.mez-duration":                      "Grab Bag answers (2019) give an order of reductions with no percentages.",
	"pet.charm":                               "A forum quote of the 1.46 notes says charm has no maximum level, which disagrees with ours; second hand and not adopted.",
	"cast.heal":                               "Creature behaviour, not a published rule.",
	"loot.money":                              "No money drop formula found.",
	"los.visible":                             "No line-of-sight rule found.",
}

// markSearched fills each registered rule's Searched from the pass above.
func markSearched() {
	for i := range registry {
		if note, ok := searched[registry[i].ID]; ok && registry[i].Searched == "" {
			registry[i].Searched = searchedOn + note
		}
	}
}
