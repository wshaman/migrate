package migrate

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"path"
	"time"
)

func txMigration(in string) string {
	return `begin transaction;
` + in + `;
commit transaction;`
}

// RegisterSQL registers a new database migration using supplied raw SQL.
func RegisterSQL(version int64, author, description, queryUp, queryDown string) {
	fnUp := func(db *sql.DB) error {
		query := txMigration(queryUp)
		_, err := db.Exec(query)
		return err
	}
	fnDown := func(db *sql.DB) error {
		query := txMigration(queryDown)
		_, err := db.Exec(query)
		return err
	}
	register(version, author, description, fnUp, fnDown)
}

func Up(db *sql.DB) error {
	return up(db)
}

func Down(db *sql.DB) error {
	return down(db)
}

func CreateFile(shortDescr, packageName, pathTo string) error {
	loc, _ := time.LoadLocation("UTC")
	dt := time.Now().In(loc)
	version := fmt.Sprintf("%d%d%d%d%d", dt.Year(), dt.Month(), dt.Day(), dt.Hour(), dt.Minute())
	fName := fmt.Sprintf("%s_%s.go", version, shortDescr)
	fmt.Printf("creating a file: %s", fName)
	tmpl := fmt.Sprintf(`package %s

import "github.com/wshaman/migrate"

func init() {
	migrate.RegisterSQL(%s, 
		"YOURNAME@HERE", 
		"%s", 
		`+"`"+`UP_SQL_MIGRATION`+"`"+`,
		`+"`"+`DOWN_SQL_MIGRATION`+"`"+`,
	)
}
`, packageName, version, shortDescr)
	return ioutil.WriteFile(path.Join(pathTo, fName), []byte(tmpl), 0622)
}
