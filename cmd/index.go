package main

import (
	"database/sql"
	"fmt"
	"slices"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
	"github.com/zmaillard/whereami/config"
)

type Node struct {
	Parent   string
	Children []string
}

func (n Node) IsRoot(huc string) bool {
	if n.Parent == huc {
		return true
	} else if _, err := strconv.Atoi(n.Parent); err != nil {
		return true
	}
	return false
}

func (n Node) IsLeaf() bool {
	return len(n.Children) == 0
}

// Determine all of the roots
// For each root - load up descendants
//

func main() {

	cfg, err := config.NewConfigWithVersion("0.0.0")
	if err != nil {
		panic(err)
	}

	db, err := sql.Open("sqlite3", fmt.Sprintf("%s?_journal_mode=WAL", cfg.DbPath))
	if err != nil {
		panic(err)
	}

	/*

		err = createTable(db)
		if err != nil {
			fmt.Println(err)
		}

		err = clearNodes(db)
		if err != nil {
			panic(err)
		}

		nodes, err := buildMapper(db)
		if err != nil {
			panic(err)
		}

		for huc, node := range nodes {
			isRoot := node.IsRoot(huc)

			if !isRoot {
				ans := getAncestors(db, huc, nodes, []string{})
				for i, h := range ans {
					_, err := db.Exec("INSERT INTO hucclosure (parenthuc, childhuc, depth) VALUES (?,?,?)", huc, h, i)
					if err != nil {
						panic(err)
					}
				}
			} else {
				_, err := db.Exec("INSERT INTO hucclosure (parenthuc, childhuc, depth) VALUES (?,?,?)", huc, huc, 0)
				if err != nil {
					panic(err)
				}

			}
		}
		fmt.Println(len(nodes))
	*/
	err = buildIndicies(db)
	if err != nil {
		fmt.Println(err)
	}
}

func createTable(db *sql.DB) error {
	query := `CREATE TABLE "hucclosure"
	(
		parenthuc text    not null
	constraint hucclosure_WBDHU12_huc12_fk
	references WBDHU12 (huc12),
		childhuc  text    not null
	constraint hucclosure_WBDHU12_huc12_fk_2
	references WBDHU12 (huc12),
		depth     integer not null
	)`
	_, err := db.Exec(query)

	return err

}

func getAncestors(db *sql.DB, huc string, nodes map[string]Node, list []string) []string {
	parent, ok := nodes[huc]
	if !ok {
		return list
	}

	if slices.Contains(list, parent.Parent) {
		return list
	}

	if parent.IsRoot(huc) {
		return list
	}

	list = append(list, huc)

	return getAncestors(db, parent.Parent, nodes, list)
}

func buildMapper(db *sql.DB) (map[string]Node, error) {
	mapper := map[string]Node{}
	query := "SELECT huc12, tohuc FROM WBDHU12"
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var huc, tohuc string
		err := rows.Scan(&huc, &tohuc)
		if err != nil {
			return nil, err
		}
		childQuery := "SELECT huc12 FROM WBDHU12 WHERE tohuc = ?"
		childRows, err := db.Query(childQuery, huc)
		if err != nil {
			return nil, err
		}
		defer childRows.Close()
		var children []string
		for childRows.Next() {
			var childHuc string
			err := childRows.Scan(&childHuc)
			if err != nil {
				return nil, err
			}
			children = append(children, childHuc)
		}
		mapper[huc] = Node{Parent: tohuc, Children: children}
	}
	return mapper, nil
}

func clearNodes(db *sql.DB) error {
	// Clear all node tables
	_, err := db.Exec("DELETE FROM hucclosure")
	return err
}

func buildIndicies(db *sql.DB) error {
	query := "CREATE INDEX hucclosure_childhuc_index on hucclosure (childhuc)"
	_, err := db.Exec(query)
	if err != nil {
		return err
	}

	query = "CREATE INDEX hucclosure_parenthuc_index on hucclosure (parenthuc)"
	_, err = db.Exec(query)
	return err
}
