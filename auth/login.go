package auth

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var Key string = "SomeSecret"

type Claims struct {
	Name string `json:"name"`
	Roll int    `json:"roll"`
	Role string `json:"role"`
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

		//Opening database
		database, err := sql.Open("sqlite3", "../Info.db")
		CheckError(err)

		query, err := database.Prepare("SELECT * FROM User WHERE Email = ?")
		CheckError(err)

		rows, err := query.Query(cred.Email)
		CheckError(err)

		//Check if User present in database
		if !rows.Next() {
			w.Header().Set("Content-type", "json")
			json.NewEncoder(w).Encode("This email is not registered. Kindly sign up first.")
			return
		}

		//Authenticate Password
		var user User
		rows.Scan(&user.Roll, &user.Name, &user.Email, &user.Password, &user.Role)
		check := checkHashPassword(user.Password, cred.Password)

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
			user.Role,
			jwt.StandardClaims{
				ExpiresAt: time.Now().Add(time.Minute * 1).Unix(),
			},
		}
		tokenString, err := JWT(claims)
		if err != nil {
			log.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tokenString)
	}
}
