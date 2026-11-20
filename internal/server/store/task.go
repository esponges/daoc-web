package store

import "database/sql"

// Task is a character's task: the one they hold (Kind is "kill" or "delivery",
// empty for none), and how many they have finished at their level. Expires is
// a Unix time, so a task keeps its limit while the character is away.
type Task struct {
	CharID     int64
	Kind       string
	Giver      string // the NPC who gave it, who is also told of a kill
	GiverID    int64  // and the spawn it is
	Target     string // a kill task's creature
	Receiver   string // a delivery task's NPC
	ReceiverID int64
	Place      string // and the zone it stands in
	ItemName   string // a delivery task's item
	Killed     bool
	Expires    int64
	Done       int
}

// Active says whether the task is one being held.
func (t *Task) Active() bool { return t != nil && t.Kind != "" }

// Task loads a character's task row, nil when they have never had one.
func (s *Store) Task(charID int64) (*Task, error) {
	var t Task
	var killed int
	err := s.db.QueryRow(`SELECT char_id,kind,giver,giver_id,target,receiver,receiver_id,place,item_name,killed,expires,done FROM task WHERE char_id=?`, charID).
		Scan(&t.CharID, &t.Kind, &t.Giver, &t.GiverID, &t.Target, &t.Receiver, &t.ReceiverID, &t.Place, &t.ItemName, &killed, &t.Expires, &t.Done)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	t.Killed = killed != 0
	return &t, nil
}

// SaveTask writes a character's task row, in one transaction with whatever else
// the caller does.
func (t *Tx) SaveTask(k *Task) error {
	killed := 0
	if k.Killed {
		killed = 1
	}
	_, err := t.tx.Exec(`INSERT INTO task(char_id,kind,giver,giver_id,target,receiver,receiver_id,place,item_name,killed,expires,done)
VALUES(?,?,?,?,?,?,?,?,?,?,?,?) ON CONFLICT(char_id) DO UPDATE SET kind=excluded.kind,giver=excluded.giver,giver_id=excluded.giver_id,
target=excluded.target,receiver=excluded.receiver,receiver_id=excluded.receiver_id,place=excluded.place,item_name=excluded.item_name,
killed=excluded.killed,expires=excluded.expires,done=excluded.done`,
		k.CharID, k.Kind, k.Giver, k.GiverID, k.Target, k.Receiver, k.ReceiverID, k.Place, k.ItemName, killed, k.Expires, k.Done)
	return err
}
