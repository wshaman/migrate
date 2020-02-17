package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/lib/pq"
	"github.com/wshaman/migrate"
)

func help() {
	c := os.Args[0]
	fmt.Printf(`Usage:
	%s command [params]
commands:
	help show this screen and exit
	up run all migrations 
	down rollback 1 last migration
	create creates a new migration file template 
Eg:
%s create add_table_users
`, c, c)
}

func dbConnect() (*sql.DB, error) {
	envOrDef := func(name, def string) (val string) {
		if val = os.Getenv(name); val != "" {
			return val
		}
		return def
	}

	dataSource := fmt.Sprintf("dbname=%s host=%s port=%s user=%s password=%s sslmode=%s",
		envOrDef("DB_NAME", "tst"),
		envOrDef("DB_HOST", "127.0.0.1"),
		envOrDef("DB_PORT", "5432"),
		envOrDef("DB_USER", "tst"),
		envOrDef("DB_PASSWD", "tst"),
		"disable")
	return sql.Open("postgres", dataSource)
}

func main() {
	if len(os.Args) < 2 {
		help()
		os.Exit(0)
	}
	db, err := dbConnect()
	if err != nil {
		//log.Fatal(err)
		fmt.Println(err.Error())
	}
	command := strings.ToLower(os.Args[1])
	switch command {
	case "up":
		err = migrate.Up(db)
	case "down":
		err = migrate.Down(db)
	case "sync":
		err = migrate.Sync(db)
	case "create":
		if len(os.Args) < 3 {
			help()
			os.Exit(1)
		}
		err = migrate.CreateFile(os.Args[2], "main", "./")
	default:
		help()
	}
	if err != nil {
		fmt.Println(err.Error())
	}
}
