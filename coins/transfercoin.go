package coins

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	_ "github.com/mattn/go-sqlite3"
)

type transfer struct {
	FromRoll int `json:"roll1"`
	ToRoll   int `json:"roll2"`
	Coins    int `json:"coins"`
}

func CheckRoll(db *sql.DB, roll int, w http.ResponseWriter) bool {
	rows, err := db.Query("Select * from User where Roll = ?", roll)
	if err != nil {
		SetError(w, err)
		return false
	}
	count := 0
	for rows.Next() {
		count++
	}
	if count == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-type", "json")
		json.NewEncoder(w).Encode(fmt.Sprint("Invalid Roll Number ", roll))
		return false
	} else if count > 1 {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-type", "json")
		json.NewEncoder(w).Encode("Internal Server Error")
		log.Print("Multiple instances of roll", roll, "found")
		return false
	}
	return true
}

func TransferCoin(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var trans transfer
		err := json.NewDecoder(r.Body).Decode(&trans)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-type", "json")
			json.NewEncoder(w).Encode("Error while parsing credentials")
			log.Print(err)
			return
		}

		if trans.Coins < 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-type", "json")
			json.NewEncoder(w).Encode("Invalid input. Coins should be positive")
			return
		}

		//Check if Roll1 exists in database
		if !CheckRoll(MyDB, trans.FromRoll, w) {
			return
		}

		//Check if Roll2 exists in database
		if !CheckRoll(MyDB, trans.ToRoll, w) {
			return
		}

		Mutex.Lock()
		tx, err := MyDB.Begin()
		if err != nil {
			SetError(w, err)
			Mutex.Unlock()
			return
		}
		res, err := tx.Exec("UPDATE User SET Coins = Coins - ? WHERE Roll = ? AND COINS - ? >= 0", trans.Coins, trans.FromRoll, trans.Coins)

		//Check if there is an error in updation or balance is less than transfer amount
		if err != nil {
			SetError(w, err)
			json.NewEncoder(w).Encode("Transaction aborted. Amount not updated.")
			tx.Rollback()
			Mutex.Unlock()
			return
		} else if n, _ := res.RowsAffected(); n == 0 {
			w.Header().Set("Content-type", "json")
			json.NewEncoder(w).Encode("Insufficient balance. Transaction aborted. Amount not updated.")
			tx.Rollback()
			Mutex.Unlock()
			return
		}

		res, err = tx.Exec("UPDATE User SET Coins = Coins + ? WHERE Roll = ?", trans.Coins, trans.ToRoll)
		if err != nil {
			SetError(w, err)
			json.NewEncoder(w).Encode("Transaction aborted. Amount not updated.")
			tx.Rollback()
			Mutex.Unlock()
			return
		}
		w.Header().Set("Content-type", "json")
		json.NewEncoder(w).Encode("Transaction successfull. Amount updated.")
		tx.Commit()
		Mutex.Unlock()
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}
