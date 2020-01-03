package migrate

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"sort"
)

// TODO: add checksum

type info struct {
	version     int64
	author      string
	description string
	fnUp        func(db *sql.DB) error
	fnDown      func(db *sql.DB) error
}

var migrations = map[int64]info{}

// register registers a new database migration with associated metadata.
func register(version int64, author string, description string, fnUp, fnDown func(db *sql.DB) error) {
	migrations[version] = info{
		version:     version,
		author:      author,
		description: description,
		fnUp:        fnUp,
		fnDown:      fnDown,
	}
}
func versions(m map[int64]info) []int64 {
	v := make([]int64, 0, len(m))
	for k := range m {
		v = append(v, k)
	}
	sort.SliceStable(v, func(i, j int) bool { return v[i] < v[j] })
	return v
}

// Run sequentially executes registered DB migrations starting at v0.
func up(db *sql.DB) error {
	if err := ensureChangelogTable(db); err != nil {
		return err
	}
	for _, k := range versions(migrations) {
		m := migrations[k]
		log.Printf("Executing migration v%d %s <%s>", m.version, m.description, m.author)
		ex, err := exists(db, m.version)
		if err != nil {
			log.Print("Fail (Error)")
			return err
		}
		if ex {
			log.Print("OK (EXISTS)")
		} else {
			err := m.fnUp(db)
			if err != nil {
				return err
			}
			_, err = db.Exec(`
				INSERT INTO
					db_changelog(version, author, description, created)
							VALUES($1, $2, $3, CURRENT_TIMESTAMP)
				`, m.version, m.author, m.description)
			if err != nil {
				return err
			}
			log.Print("OK")
		}
	}
	return nil
}

func down(db *sql.DB) error {
	if err := ensureChangelogTable(db); err != nil {
		return err
	}
	var id int64
	log.Println("Downgrading migration")
	q := `SELECT id FROM db_changelog ORDER BY created DESC LIMIT 1;`
	if err := db.QueryRow(q).Scan(&id); err != nil {
		if err == sql.ErrNoRows {
			log.Println("No migrations found")
			return nil
		}
		return err
	}
	fmt.Printf("Last applied migration %d", id)
	m, ok := migrations[id]
	if !ok {
		fmt.Println(" can't be found")
		return errors.New("no migration found")
	}
	fmt.Printf("By: %s. %s\n", m.author, m.description)
	if err := m.fnDown(db); err != nil {
		log.Print("Fail (Error)")
		return err
	}
	if _, err := db.Exec(`DELETE FROM db_changelog WHERE id=$1`, id); err != nil {
		log.Print("WARN (Migration down, can't update table)")
		return err
	}
	log.Print("OK (Done)")
	return nil
}

func ensureChangelogTable(db *sql.DB) error {
	exists := false
	if err := db.QueryRow(`
			SELECT EXISTS (
				SELECT 1
				FROM   information_schema.tables 
				WHERE  table_name = 'db_changelog'
			)`).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return nil
	}
	query := `CREATE TABLE db_changelog
(
    id          BIGSERIAL PRIMARY KEY,
    version     BIGINT    NOT NULL,
    author      TEXT      NOT NULL,
    description TEXT      NOT NULL,
    created     TIMESTAMP NOT NULL
);
`
	if _, err := db.Exec(query); err != nil {
		return err
	}
	return nil
}

func exists(db *sql.DB, version int64) (bool, error) {
	exists := false
	err := db.QueryRow("SELECT 1 FROM db_changelog WHERE version=$1", version).Scan(&exists)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return exists, err
}
