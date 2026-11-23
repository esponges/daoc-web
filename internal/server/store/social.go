package store

// A character's friends and ignores, kept as rows of the other character's id.
const (
	SocialFriend = "friend"
	SocialIgnore = "ignore"
)

// Social lists the characters a character has marked with a kind.
func (s *Store) Social(charID int64, kind string) ([]int64, error) {
	rows, err := s.db.Query(`SELECT other_id FROM social WHERE char_id=? AND kind=? ORDER BY other_id`, charID, kind)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

// SetSocial marks or unmarks another character.
func (s *Store) SetSocial(charID, otherID int64, kind string, on bool) error {
	if on {
		_, err := s.db.Exec(`INSERT OR IGNORE INTO social(char_id, other_id, kind) VALUES(?,?,?)`, charID, otherID, kind)
		return err
	}
	_, err := s.db.Exec(`DELETE FROM social WHERE char_id=? AND other_id=? AND kind=?`, charID, otherID, kind)
	return err
}

// CharID finds a character by name, without regard to case, and returns their
// id and the name as stored; 0 when there is none.
func (s *Store) CharID(name string) (int64, string) {
	var id int64
	var stored string
	if err := s.db.QueryRow(`SELECT id,name FROM character WHERE name=? COLLATE NOCASE`, name).Scan(&id, &stored); err != nil {
		return 0, ""
	}
	return id, stored
}

// CharName is a character's name by id, "" when there is none.
func (s *Store) CharName(id int64) string {
	var name string
	s.db.QueryRow(`SELECT name FROM character WHERE id=?`, id).Scan(&name)
	return name
}
