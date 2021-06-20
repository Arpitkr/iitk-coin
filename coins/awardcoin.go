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

type Users struct {
	Roll  int `json:"roll"`
	Coins int `json:"coins"`
}

func SetError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Header().Set("Content-type", "json")
	json.NewEncoder(w).Encode("Internal server error")
	log.Print(err)
}

func AwardCoin(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var user Users
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-type", "json")
			json.NewEncoder(w).Encode("Error while parsing credentials")
			log.Print(err)
			return
		}

		if user.Coins < 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-type", "json")
			json.NewEncoder(w).Encode("Invalid input. Coins should be positive")
			return
		}

		//Check if roll exists in database
		if !CheckRoll(MyDB, user.Roll, w) {
			return
		}

		Mutex.Lock()
		tx, err := MyDB.Begin()
		if err != nil {
			SetError(w, err)
			Mutex.Unlock()
			return
		}
		res, err := tx.Exec("Update User set Coins = Coins + ? where Roll = ? AND Coins + ? <10000", user.Coins, user.Roll, user.Coins)

		if err != nil {
			SetError(w, err)
			json.NewEncoder(w).Encode("Transaction aborted. Amount not updated.")
			tx.Rollback()
		} else if check, err := res.RowsAffected(); check == 1 && err == nil {
			json.NewEncoder(w).Encode("Transaction successful. Amount updated.")
			tx.Commit()
		} else {
			json.NewEncoder(w).Encode("Final amount more than upper bound. Transaction aborted. Amount not updated.")
			tx.Rollback()
		}
		Mutex.Unlock()
	} else {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode("Quit")
	}
}
