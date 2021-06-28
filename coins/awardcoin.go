package coins

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"sync"

	_ "github.com/mattn/go-sqlite3"
)

var MyDB *sql.DB
var Mutex sync.Mutex

type User struct {
	Roll  int     `json:"roll"`
	Coins float64 `json:"coins"`
}

func SetError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-type", "json")
	json.NewEncoder(w).Encode("Internal server error")
	log.Print(err)
}

func AwardCoin(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {

		//Check if JWT is set and valid
		ok, claims := VerifyJwt(w, r)
		if !ok {
			return
		}

		//Parse credentials
		var user User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-type", "json")
			json.NewEncoder(w).Encode("Error while parsing credentials")
			log.Print(err)
			return
		}

		//Check the authenticity of User as an admin
		if !claims.IsAdmin {
			w.WriteHeader(http.StatusUnauthorized)
			w.Header().Set("Content-type", "json")
			json.NewEncoder(w).Encode("Status Unauthorized")
			return
		}

		if user.Coins < 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-type", "json")
			json.NewEncoder(w).Encode("Invalid input. Coins should be positive")
			return
		}

		//Check if roll exists in database
		if !CheckRoll(user.Roll, w) {
			return
		}

		//Check if roll is not of GenSec or AH
		rows, err := MyDB.Query("SELECT isAdmin FROM User WHERE Roll = User.Roll")
		if err != nil {
			SetError(w, err)
			return
		}
		var check bool
		for rows.Next() {
			rows.Scan(&check)
		}
		if check {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-type", "json")
			json.NewEncoder(w).Encode("GenSec or AH can not be awarded coins.")
			return
		}

		Mutex.Lock()
		tx, err := MyDB.Begin()
		if err != nil {
			SetError(w, err)
			tx.Rollback()
			Mutex.Unlock()
			return
		}
		res, err := tx.Exec("Update User set Coins = Coins + ? where Roll = ? AND Coins + ? <1000", user.Coins, user.Roll, user.Coins)

		if err != nil {
			SetError(w, err)
			json.NewEncoder(w).Encode("Transaction aborted. Amount not updated.")
			tx.Rollback()
		} else if check, err := res.RowsAffected(); check == 1 && err == nil {
			_, err = tx.Exec("Insert into RewardHistory values(?,?,datetime('now','localtime'))", user.Roll, user.Coins)
			if err != nil {
				SetError(w, err)
				json.NewEncoder(w).Encode("Transaction aborted. Amount not updated.")
				tx.Rollback()
			} else {
				w.Header().Set("Content-type", "json")
				json.NewEncoder(w).Encode("Transaction successful. Amount updated.")
				tx.Commit()
			}
		} else {
			w.Header().Set("Content-type", "json")
			json.NewEncoder(w).Encode("Final amount more than upper bound. Transaction aborted. Amount not updated.")
			tx.Rollback()
		}
		Mutex.Unlock()
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}
