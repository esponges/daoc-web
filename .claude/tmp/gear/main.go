// Scratch: what do the region-101 lists with level 31+ gear sell?
package main

import (
	"fmt"

	"daocweb/internal/importer"
)

func main() {
	dir := "server/data/community/opendaoc-db-core/"
	type it struct{ lvl, obj, realm int }
	items := map[string]it{}
	importer.ReadTable(dir+"itemtemplate.sql", func(r *importer.Row) error {
		items[r.Str("Id_nb")] = it{r.Int("Level"), r.Int("Object_Type"), r.Int("Realm")}
		return nil
	})
	lists101 := map[string]string{}
	importer.ReadTable(dir+"mob.sql", func(r *importer.Row) error {
		if l := r.Str("ItemsListTemplateID"); l != "" && r.Int("Region") == 101 {
			lists101[l] = r.Str("ClassType")
		}
		return nil
	})
	hist := map[string]int{}
	importer.ReadTable(dir+"merchantitem.sql", func(r *importer.Row) error {
		l := r.Str("ItemListID")
		if _, ok := lists101[l]; !ok {
			return nil
		}
		x := items[r.Str("ItemTemplateID")]
		if x.lvl > 30 && (x.obj == 11 || x.obj == 12 || x.obj == 13 || x.obj == 35) {
			hist[fmt.Sprintf("%s (%s) obj %d realm %d level %d", l, lists101[l], x.obj, x.realm, x.lvl)]++
		}
		return nil
	})
	for k, v := range hist {
		fmt.Println(v, k)
	}
}
