package migrate

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"testing"

	_ "github.com/lib/pq"
)

func getDB() *sql.DB {
	user := envOrDef("DB_USER", "tst")
	dbName := envOrDef("DB_NAME", "tst")
	host := envOrDef("DB_HOST", "127.0.0.1")
	port := envOrDef("DB_PORT", "5432")
	passwd := envOrDef("DB_PASSWD", "tst")
	dataSource := fmt.Sprintf("dbname=%s host=%s port=%s user=%s password=%s sslmode=%s",
		dbName, host, port, user, passwd, "disable")
	db, err := sql.Open("postgres", dataSource)
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func Test_checkTbl(t *testing.T) {
	db := getDB()
	ensureChangelogTable(db)
}

func Test_versions(t *testing.T) {
	in := map[int64]migration{
		3: {version: 3},
		2: {version: 2},
		7: {version: 7},
		5: {version: 5},
	}
	out := []int64{2, 3, 5, 7}
	if !reflect.DeepEqual(out, versions(in)) {
		t.Fail()
	}
}
