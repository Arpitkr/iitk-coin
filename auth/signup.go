package auth

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"

	"github.com/Arpitkr/iitk-coin/coins"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var MyDB *sql.DB

type User struct {
	Name          string  `json:"name"`
	Roll          int     `json:"roll"`
	Email         string  `json:"email"`
	Password      string  `json:"passwd"`
	IsAdmin       bool    `json:"isAdmin"`
	IsCoordinator bool    `json:"isCoordinator"`
	Coins         float64 `json:"coins"`
}

func SetError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-type", "json")
	json.NewEncoder(w).Encode("Internal server error")
	log.Print(err)
}

func Signup(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		//Create table if it does not exist
		table, err := MyDB.Prepare("CREATE TABLE IF NOT EXISTS User(Roll INTEGER PRIMARY KEY, Name VARCHAR(25), Email VARCHAR(100), Password VARCHAR(200), isAdmin bool, isCoordinator bool, Coins Float DEFAULT 0.0) ")
		if err != nil {
			SetError(w, err)
			return
		}

		_, err = table.Exec()
		if err != nil {
			SetError(w, err)
			return
		}

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
		data, err := MyDB.Prepare("SELECT * FROM User WHERE Email = ?")
		if err != nil {
			SetError(w, err)
			return
		}

		rows, err := data.Query(user.Email)
		if err != nil {
			SetError(w, err)
			return
		}

		if rows.Next() {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-type", "json")
			json.NewEncoder(w).Encode("This email is already in use")
			return
		}

		//GenerateHashedPassword
		passwd, err := bcrypt.GenerateFromPassword([]byte(user.Password), 10)
		if err != nil {
			SetError(w, err)
			return
		}

		user.Password = string(passwd)

		//Add data to database. Role is assigned NULL at the time of signup.
		query, err := MyDB.Prepare("INSERT INTO User(Roll, Name, Email, Password, isAdmin, isCoordinator) VALUES(?,?,?,?,false,false)")
		if err != nil {
			SetError(w, err)
			return
		}

		coins.Mutex.Lock()
		tx, err := MyDB.Begin()
		if err != nil {
			SetError(w, err)
			return
		}

		_, err = tx.Stmt(query).Exec(user.Roll, user.Name, user.Email, user.Password)
		if err != nil {
			SetError(w, err)
			tx.Rollback()
		} else {
			tx.Commit()
		}
		coins.Mutex.Unlock()
	}
}
