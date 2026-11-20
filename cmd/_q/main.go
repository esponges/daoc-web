package main

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

func main() {
	db, _ := sql.Open("sqlite", "server/data/world.db")
	rows, err := db.Query(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	cols, _ := rows.Columns()
	for rows.Next() {
		v := make([]any, len(cols))
		p := make([]any, len(cols))
		for i := range v {
			p[i] = &v[i]
		}
		rows.Scan(p...)
		for i := range v {
			fmt.Print(v[i], " | ")
		}
		fmt.Println()
	}
}
