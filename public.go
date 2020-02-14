package migrate

import (
	"database/sql"
	"fmt"
	"io/ioutil"
	"path"
	"time"
)

// RegisterSQL registers a new database migration using supplied raw SQL.
func RegisterSQL(version int64, author, description, queryUp, queryDown string) {
	register(version, author, description, queryUp, queryDown)
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
	version := fmt.Sprintf("%d%02d%02d%02d%02d%02d", dt.Year(), dt.Month(), dt.Day(), dt.Hour(), dt.Minute(), dt.Second())
	fName := fmt.Sprintf("%s_%s.go", version, shortDescr)
	fmt.Printf("creating a file: %s\n", fName)
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
