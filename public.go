package migrate

import (
	"database/sql"
)

func txMigration(in string) string {
	return `begin transaction;
` + in + `
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

func CreateFile(shortDescr string) {
	//loc, _ := time.LoadLocation("UTC")
	//dt := time.Now().In(loc)
	//fName := fmt.Sprintf("%d%d%d%d%d_%s.go", dt.Year())
	//fmt.Println(fName)
}
