package migrate

import (
	"database/sql"
	"fmt"
	"sort"
)

// TODO: add checksum

type info struct {
	version     int64
	author      string
	description string
	upSQL       string
	downSQL     string
}

var migrations = map[int64]info{}

// register registers a new database migration with associated metadata.
func register(version int64, author, description, upSQL, downSQL string) {
	migrations[version] = info{
		version:     version,
		author:      author,
		description: description,
		upSQL:       upSQL,
		downSQL:     downSQL,
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

func txExec(tx *sql.Tx, query string, v ...interface{}) error {
	if _, err := tx.Exec(query, v...); err != nil {
		if e := tx.Rollback(); e != nil {
			return e
		}
		return err
	}
	return nil
}

func logf(f string, v ...interface{}) {
	fmt.Printf(f, v...)
}

func logln(f string) {
	fmt.Println(f)
}

func sync(db *sql.DB) error {
	return nil
}

// Run sequentially executes registered DB migrations starting at v0.
func up(db *sql.DB) error {
	if err := ensureChangelogTable(db); err != nil {
		return err
	}
	for _, k := range versions(migrations) {
		m := migrations[k]
		logf("Executing migration v%d %s <%s>", m.version, m.description, m.author)
		ex, err := exists(db, m.version)
		if err != nil {
			logln("Fail (Error)")
			return err
		}
		if ex {
			logln("OK (EXISTS)")
		} else {
			tx, err := db.Begin()
			if err != nil {
				return err
			}
			if err := txExec(tx, m.upSQL); err != nil {
				return err
			}
			if err := txExec(tx, `
INSERT INTO db_changelog(version, author, description, rollback, created) VALUES($1, $2, $3, $4, CURRENT_TIMESTAMP);
	`, m.version, m.author, m.description, m.downSQL); err != nil {
				return err
			}
			if err := tx.Commit(); err != nil {
				logln("Fail (Error)")
				return err
			}
			logln("OK")
		}
	}
	return nil
}

func down(db *sql.DB) error {
	if err := ensureChangelogTable(db); err != nil {
		return err
	}

	var m info
	logln("Downgrading migration")
	q := `SELECT version, rollback, author, description FROM db_changelog ORDER BY created DESC LIMIT 1;`
	if err := db.QueryRow(q).Scan(&m.version, &m.downSQL, &m.author, &m.description); err != nil {
		if err == sql.ErrNoRows {
			logln("No migrations found")
			return nil
		}
		return err
	}
	logf("Last applied migration #%d, %s, by: %s\n", m.version, m.description, m.author)
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	if err := txExec(tx, m.downSQL); err != nil {
		logf("Fail (Error)")
		return err
	}
	if err := txExec(tx, `DELETE FROM db_changelog WHERE version=$1`, m.version); err != nil {
		logf("WARN (Migration down, can't update table)")
		return err
	}
	if err := tx.Commit(); err != nil {
		logf("Fail (Error)")
		return err
	}
	logf("OK (Done)")
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
    rollback 	TEXT NOT NULL,
    created     TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
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
