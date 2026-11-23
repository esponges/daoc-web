package rules

import (
	"regexp"
	"strings"
)

// Player guilds. The shape (a Guild Registrar, /gc commands, ranks 0 to 9, a
// guild, officer and alliance channel) follows the Camelot Herald's guild
// guide; the numbers below are this project's choice and provisional, since no
// 1.65 source found gives them (the one fee found, 10 platinum on the official
// server notes, is undated and later).
const (
	GuildFormMembers = 8     // players needed to form a guild
	GuildFormFee     = 10000 // copper, charged to the leader
	GuildEmblems     = 84    // emblem numbers, as many as the install has per realm
	GuildRanks       = 10
	GuildMaxMembers  = 200
)

// What a guild rank may do, as bits.
const (
	PermGuildSpeak   = 1 << iota // speak on guild chat
	PermGuildHear                // hear guild chat
	PermOfficerSpeak             // speak on officer chat
	PermOfficerHear              // hear officer chat
	PermInvite
	PermPromote
	PermRemove
	PermEmblem
	PermMotd
)

// DefaultPerms is what each rank may do in a new guild: the Guildmaster
// everything; ranks 1 to 3 run it (all but the emblem); ranks 4 to 8 may talk
// and invite; rank 9, the newest, talks and listens.
func DefaultPerms(rank int) int {
	switch {
	case rank == 0:
		return PermGuildSpeak | PermGuildHear | PermOfficerSpeak | PermOfficerHear | PermInvite | PermPromote | PermRemove | PermEmblem | PermMotd
	case rank <= 3:
		return PermGuildSpeak | PermGuildHear | PermOfficerSpeak | PermOfficerHear | PermInvite | PermPromote | PermRemove | PermMotd
	case rank <= 8:
		return PermGuildSpeak | PermGuildHear | PermInvite
	}
	return PermGuildSpeak | PermGuildHear
}

// DefaultTitle is a rank's name in a new guild.
func DefaultTitle(rank int) string {
	switch rank {
	case 0:
		return "Guildmaster"
	case 1:
		return "Officer"
	case 9:
		return "Member"
	}
	return "Rank " + string(rune('0'+rank))
}

var guildName = regexp.MustCompile(`^[A-Za-z][A-Za-z' ]{2,29}$`)

// ValidGuildName says whether a name passes: 3 to 30 letters, spaces and
// apostrophes, starting with a letter, with no two spaces together.
func ValidGuildName(name string) bool {
	return guildName.MatchString(name) && !strings.Contains(name, "  ") && !strings.HasSuffix(name, " ")
}

// MayPromote says whether a member of actorRank may move a member of target
// rank to newRank: only members ranked below them, and only up to one rank
// below their own (rank 0 may promote to 1).
func MayPromote(actorRank, targetRank, newRank int) bool {
	return newRank < targetRank && targetRank > actorRank && newRank > actorRank && newRank >= 0
}

// MayDemote says whether a member of actorRank may lower a member of
// targetRank: only one ranked below them, and never past rank 9.
func MayDemote(actorRank, targetRank int) bool { return targetRank > actorRank && targetRank < GuildRanks-1 }

// MayRemove says whether a member of actorRank may remove one of targetRank.
func MayRemove(actorRank, targetRank int) bool { return targetRank > actorRank }

func init() {
	const c = "player-guilds"
	register(
		Rule{ID: "guild.form", Capability: c, Requirement: "Forming a guild", Status: Provisional,
			Sources: []string{"herald-guilds", "project"}, Scenarios: []string{"Formed", "Short group"}},
		Rule{ID: "guild.ranks", Capability: c, Requirement: "Ranks and permissions", Status: Provisional,
			Sources: []string{"herald-guilds", "project"}, Scenarios: []string{"Promotion limit"}},
	)
}
