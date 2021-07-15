package coins

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/mattn/go-sqlite3"
)

type transfer struct {
	FromRoll int     `json:"fromroll"`
	ToRoll   int     `json:"toroll"`
	Coins    float64 `json:"coins"`
}

func CheckRoll(roll int, w http.ResponseWriter) bool {
	rows, err := MyDB.Query("Select * from User where Roll = ?", roll)
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

		//Check if JWT is set and valid
		ok, claims := VerifyJwt(w, r)
		if !ok {
			return
		}

		var trans transfer
		err := json.NewDecoder(r.Body).Decode(&trans)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-type", "json")
			json.NewEncoder(w).Encode("Error while parsing credentials")
			log.Print(err)
			return
		}

		if trans.FromRoll != claims.Roll {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-type", "json")
			json.NewEncoder(w).Encode("Status unauthorized")
			return
		}

		if trans.Coins < 0 {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-type", "json")
			json.NewEncoder(w).Encode("Invalid input. Coins should be positive")
			return
		}

		//Check if Roll1 exists in database
		if !CheckRoll(trans.FromRoll, w) {
			return
		}

		//Check if Roll2 exists in database
		if !CheckRoll(trans.ToRoll, w) {
			return
		}

		flag := 0
		s1, s2 := strconv.Itoa(trans.FromRoll), strconv.Itoa(trans.ToRoll)
		if len(s1) != len(s2) || s1[0:2] != s2[0:2] {
			flag = 1
		}

		Mutex.Lock()
		tx, err := MyDB.Begin()
		if err != nil {
			SetError(w, err)
			tx.Rollback()
			Mutex.Unlock()
			return
		}

		//Update Sender's account
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

		//Calculate applicable taxes
		received, tax := 0.0, 0.0
		if flag == 0 {
			tax = (2.0 * trans.Coins) / 100.0
			received = trans.Coins - tax
		} else {
			tax = (33.0 * trans.Coins) / 100.0
			received = trans.Coins - tax
		}

		//Update Receiver's account
		res, err = tx.Exec("UPDATE User SET Coins = Coins + ? WHERE Roll = ? And Coins + ? <1000", received, trans.ToRoll, received)
		if err != nil {
			SetError(w, err)
			json.NewEncoder(w).Encode("Transaction aborted. Amount not updated.")
			tx.Rollback()
			Mutex.Unlock()
			return
		} else if n, _ := res.RowsAffected(); n == 0 {
			w.Header().Set("Content-type", "json")
			json.NewEncoder(w).Encode("Final amount more than upper bound. Transaction aborted. Amount not updated.")
			tx.Rollback()
			Mutex.Unlock()
			return
		} else {
			_, err = tx.Exec("Insert into TransferHistory values(?,?,?,?,?,datetime('now','localtime'))", trans.FromRoll, trans.ToRoll, trans.Coins, received, tax)
			if err != nil {
				SetError(w, err)
				json.NewEncoder(w).Encode("Transaction aborted. Amount not updated.")
				tx.Rollback()
			} else {
				w.Header().Set("Content-type", "json")
				json.NewEncoder(w).Encode("Transaction successful. Amount updated.")
				tx.Commit()
			}
			Mutex.Unlock()
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}
