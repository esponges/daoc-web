package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "modernc.org/sqlite"
)

// runQuery prints a query's rows tab-separated, for looking at the store.
func runQuery(path, q string) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer db.Close()
	rows, err := db.Query(q)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	fmt.Println(strings.Join(cols, "\t"))
	for rows.Next() {
		v := make([]any, len(cols))
		p := make([]any, len(cols))
		for i := range v {
			p[i] = &v[i]
		}
		if err := rows.Scan(p...); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		s := make([]string, len(cols))
		for i, x := range v {
			s[i] = fmt.Sprint(x)
		}
		fmt.Println(strings.Join(s, "\t"))
	}
}
