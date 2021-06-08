package auth

import (
	"database/sql"
	"encoding/json"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	Name     string `json:"name"`
	Roll     int    `json:"roll"`
	Email    string `json:"email"`
	Password string `json:"passwd"`
	Role     string `json:"role"`
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

func Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		//Open databse present in root directory
		database, err := sql.Open("sqlite3", "../Info.db")
		CheckError(err)

		//Create table if it does not exist
		table, err := database.Prepare("CREATE TABLE IF NOT EXISTS User(Roll INTEGER, Name VARCHAR(25), Email VARCHAR(100), Password VARCHAR(200), Role VARCHAR(15)) ")
		CheckError(err)
		_, err = table.Exec()
		CheckError(err)

		//Reading data from request body. Role field of request body will be nil. Role will be assigned by another end point.
		var user User
		err = json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-type", "json")
			json.NewEncoder(w).Encode(err)
			return
		}

		//Checking if user already exists in database
		data, err := database.Prepare("SELECT * FROM User WHERE Email = ?")
		CheckError(err)
		rows, err := data.Query(user.Email)
		CheckError(err)
		if rows.Next() {
			w.Header().Set("Content-type", "json")
			json.NewEncoder(w).Encode("This email is already in use")
			return
		}

		//GenerateHashedPassword
		passwd, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
		CheckError(err)
		user.Password = string(passwd)

		//Add data to database. Role is assigned NULL at the time of signup.
		query, err := database.Prepare("INSERT INTO User(Roll, Name, Email, Password, Role) VALUES(?,?,?,?,NULL)")
		CheckError(err)
		_, err = query.Exec(user.Roll, user.Name, user.Email, user.Password)
	}
}
