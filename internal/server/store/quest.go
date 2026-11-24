package store

// Quest is a character's progress in one quest: the step it is on, how far
// into that step (kills, or items held at the last count), and whether it is
// finished. A row exists from accepting a quest on.
type Quest struct {
	CharID int64
	Quest  string // the quest's id from its data file
	Step   int
	Count  int
	Done   bool
}

const questSchema = `
CREATE TABLE IF NOT EXISTS char_quest(
  char_id INTEGER NOT NULL REFERENCES character(id) ON DELETE CASCADE, quest TEXT NOT NULL,
  step INTEGER NOT NULL DEFAULT 0, count INTEGER NOT NULL DEFAULT 0, done INTEGER NOT NULL DEFAULT 0,
  PRIMARY KEY(char_id, quest));
`

// Quests lists a character's quests, active and done.
func (s *Store) Quests(charID int64) ([]Quest, error) {
	rows, err := s.db.Query(`SELECT quest,step,count,done FROM char_quest WHERE char_id=? ORDER BY rowid`, charID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Quest
	for rows.Next() {
		q := Quest{CharID: charID}
		var done int
		if err := rows.Scan(&q.Quest, &q.Step, &q.Count, &done); err != nil {
			return nil, err
		}
		q.Done = done != 0
		out = append(out, q)
	}
	return out, rows.Err()
}

// SaveQuest writes a quest row.
func (t *Tx) SaveQuest(q Quest) error {
	done := 0
	if q.Done {
		done = 1
	}
	_, err := t.tx.Exec(`INSERT INTO char_quest(char_id,quest,step,count,done) VALUES(?,?,?,?,?)
ON CONFLICT(char_id,quest) DO UPDATE SET step=excluded.step,count=excluded.count,done=excluded.done`,
		q.CharID, q.Quest, q.Step, q.Count, done)
	return err
}

// DeleteQuest forgets a quest, as abandoning one does.
func (t *Tx) DeleteQuest(charID int64, quest string) error {
	_, err := t.tx.Exec(`DELETE FROM char_quest WHERE char_id=? AND quest=?`, charID, quest)
	return err
}
