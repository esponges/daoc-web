package store

import (
	"errors"
)

// Guild is a player guild as it is kept: its name, realm, messages, emblem and
// the alliance it is in (0 for none).
type Guild struct {
	ID       int64
	Name     string
	Realm    int
	Motd     string // for every member
	OfficerM string // for the officers
	Emblem   int    // the emblem's number, 0 for none
	Alliance int64
	Perms    [10]int // what each rank 0 to 9 may do, as bits
	Titles   [10]string
}

// GuildMember is a character's place in a guild.
type GuildMember struct {
	CharID  int64
	GuildID int64
	Rank    int
	Note    string
}

// ErrGuildName is returned when a guild's name is taken.
var ErrGuildName = errors.New("store: that guild name is taken")

const guildSchema = `
CREATE TABLE IF NOT EXISTS guild(
  id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL UNIQUE COLLATE NOCASE, realm INTEGER NOT NULL,
  motd TEXT NOT NULL DEFAULT '', omotd TEXT NOT NULL DEFAULT '', emblem INTEGER NOT NULL DEFAULT 0,
  alliance INTEGER NOT NULL DEFAULT 0, created INTEGER NOT NULL DEFAULT 0);
CREATE TABLE IF NOT EXISTS guild_rank(
  guild_id INTEGER NOT NULL REFERENCES guild(id) ON DELETE CASCADE, rank INTEGER NOT NULL,
  title TEXT NOT NULL DEFAULT '', perms INTEGER NOT NULL DEFAULT 0, PRIMARY KEY(guild_id, rank));
CREATE TABLE IF NOT EXISTS guild_member(
  char_id INTEGER PRIMARY KEY REFERENCES character(id) ON DELETE CASCADE,
  guild_id INTEGER NOT NULL REFERENCES guild(id) ON DELETE CASCADE, rank INTEGER NOT NULL DEFAULT 9, note TEXT NOT NULL DEFAULT '');
`

// CreateGuild makes a guild with its ranks and puts the given characters in
// it: the first as rank 0, the rest as rank 9. It fails if the name is taken
// or any character is already in a guild.
func (s *Store) CreateGuild(g *Guild, members []int64) error {
	return s.Tx(func(t *Tx) error {
		res, err := t.tx.Exec(`INSERT INTO guild(name, realm, created) VALUES(?,?,strftime('%s','now'))`, g.Name, g.Realm)
		if err != nil {
			return ErrGuildName
		}
		g.ID, _ = res.LastInsertId()
		for r := 0; r < 10; r++ {
			if _, err := t.tx.Exec(`INSERT INTO guild_rank(guild_id, rank, title, perms) VALUES(?,?,?,?)`, g.ID, r, g.Titles[r], g.Perms[r]); err != nil {
				return err
			}
		}
		for i, c := range members {
			rank := 9
			if i == 0 {
				rank = 0
			}
			if _, err := t.tx.Exec(`INSERT INTO guild_member(char_id, guild_id, rank) VALUES(?,?,?)`, c, g.ID, rank); err != nil {
				return errors.New("store: a character is already in a guild")
			}
		}
		return nil
	})
}

// Guilds loads every guild with its ranks.
func (s *Store) Guilds() ([]*Guild, error) {
	rows, err := s.db.Query(`SELECT id,name,realm,motd,omotd,emblem,alliance FROM guild ORDER BY id`)
	if err != nil {
		return nil, err
	}
	var out []*Guild
	for rows.Next() {
		g := &Guild{}
		if err := rows.Scan(&g.ID, &g.Name, &g.Realm, &g.Motd, &g.OfficerM, &g.Emblem, &g.Alliance); err != nil {
			rows.Close()
			return nil, err
		}
		out = append(out, g)
	}
	rows.Close()
	for _, g := range out {
		rr, err := s.db.Query(`SELECT rank,title,perms FROM guild_rank WHERE guild_id=?`, g.ID)
		if err != nil {
			return nil, err
		}
		for rr.Next() {
			var rank, perms int
			var title string
			if err := rr.Scan(&rank, &title, &perms); err == nil && rank >= 0 && rank < 10 {
				g.Titles[rank], g.Perms[rank] = title, perms
			}
		}
		rr.Close()
	}
	return out, nil
}

// GuildMembers loads every membership.
func (s *Store) GuildMembers() ([]GuildMember, error) {
	rows, err := s.db.Query(`SELECT char_id,guild_id,rank,note FROM guild_member`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []GuildMember
	for rows.Next() {
		var m GuildMember
		if err := rows.Scan(&m.CharID, &m.GuildID, &m.Rank, &m.Note); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// SaveGuild writes a guild's messages, emblem, alliance and ranks.
func (s *Store) SaveGuild(g *Guild) error {
	return s.Tx(func(t *Tx) error {
		if _, err := t.tx.Exec(`UPDATE guild SET motd=?, omotd=?, emblem=?, alliance=? WHERE id=?`, g.Motd, g.OfficerM, g.Emblem, g.Alliance, g.ID); err != nil {
			return err
		}
		for r := 0; r < 10; r++ {
			if _, err := t.tx.Exec(`UPDATE guild_rank SET title=?, perms=? WHERE guild_id=? AND rank=?`, g.Titles[r], g.Perms[r], g.ID, r); err != nil {
				return err
			}
		}
		return nil
	})
}

// SetGuildMember puts a character in a guild at a rank, or changes theirs.
func (s *Store) SetGuildMember(m GuildMember) error {
	_, err := s.db.Exec(`INSERT INTO guild_member(char_id, guild_id, rank, note) VALUES(?,?,?,?)
		ON CONFLICT(char_id) DO UPDATE SET guild_id=excluded.guild_id, rank=excluded.rank, note=excluded.note`, m.CharID, m.GuildID, m.Rank, m.Note)
	return err
}

// RemoveGuildMember takes a character out of their guild.
func (s *Store) RemoveGuildMember(charID int64) error {
	_, err := s.db.Exec(`DELETE FROM guild_member WHERE char_id=?`, charID)
	return err
}

// DeleteGuild removes a guild and its members.
func (s *Store) DeleteGuild(id int64) error {
	_, err := s.db.Exec(`DELETE FROM guild WHERE id=?`, id)
	return err
}

