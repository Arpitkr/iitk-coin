package main

import (
	"database/sql"
	_ "fmt"

	_ "github.com/mattn/go-sqlite3"
)

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

func AddData(m map[int]string, data *sql.DB) {
	table, err := data.Prepare("INSERT INTO User (ROLL, NAME) VALUES(?,?)")
	CheckError(err)
	for roll, name := range m {
		_, err = table.Exec(roll, name)
		CheckError(err)
	}
}

func main() {
	data, err := sql.Open("sqlite3", "Finance.db")
	CheckError(err)

	table, err := data.Prepare("CREATE TABLE IF NOT EXISTS User(ROLL INTEGER, NAME VARCHAR(60))")
	CheckError(err)

	_, err = table.Exec()
	CheckError(err)

	m := map[int]string{200190: "Arpit", 200521: "Kunal", 200145: "Aryan", 190461: "Varun"}
	AddData(m, data)

}
