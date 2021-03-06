package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var Key string = "SomeSecret"

type Claims struct {
	Name          string `json:"name"`
	Roll          int    `json:"roll"`
	IsAdmin       bool   `json:"isAdmin"`
	IsCoordinator bool   `json:"isCoordinator"`
	jwt.StandardClaims
}

type Credentials struct {
	Password string `json:"passwd"`
	Email    string `json:"email"`
}

func checkHashPassword(pass1 string, pass2 string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(pass1), []byte(pass2))
	return err == nil
}

func JWT(claims *Claims) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(Key))
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func Login(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var cred Credentials
		err := json.NewDecoder(r.Body).Decode(&cred)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Header().Set("Content-type", "json")
			json.NewEncoder(w).Encode("Error while parsing credentials")
			return
		}

		query, err := MyDB.Prepare("SELECT * FROM User WHERE Email = ?")
		if err != nil {
			SetError(w, err)
			return
		}

		rows, err := query.Query(cred.Email)
		if err != nil {
			SetError(w, err)
			return
		}

		//Check if User present in database
		if !rows.Next() {
			w.Header().Set("Content-type", "json")
			json.NewEncoder(w).Encode("This email is not registered. Kindly sign up first.")
			return
		}

		//Authenticate Password
		var user User
		err = rows.Scan(&user.Roll, &user.Name, &user.Email, &user.Password, &user.IsAdmin, &user.IsCoordinator, &user.Coins)
		if err != nil {
			SetError(w, err)
			return
		}
		check := checkHashPassword(user.Password, cred.Password)
		rows.Close()

		//Return Invalid password if password not authenticated
		if !check {
			w.WriteHeader(http.StatusUnauthorized)
			w.Header().Set("Content-type", "json")
			json.NewEncoder(w).Encode("Invalid password")
			return
		}

		//Generating claims for Json web token
		claims := &Claims{
			user.Name,
			user.Roll,
			user.IsAdmin,
			user.IsCoordinator,
			jwt.StandardClaims{
				ExpiresAt: time.Now().Add(time.Minute * 15).Unix(),
			},
		}
		tokenString, err := JWT(claims)
		if err != nil {
			SetError(w, err)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tokenString)
	}
}
