package coins

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/dgrijalva/jwt-go"
)

type Claims struct {
	Name          string `json:"name"`
	Roll          int    `json:"roll"`
	IsAdmin       bool   `json:"isAdmin"`
	IsCoordinator bool   `json:"isCoordinator"`
	jwt.StandardClaims
}

var Key string = "SomeSecret"

//This function is used to verify JWT all over the project. It returns true if JWT valid, false if JWT invalid or if JWT not set.
func VerifyJwt(w http.ResponseWriter, r *http.Request) (bool, Claims) {

	claims := Claims{}

	//Check whether JWT is set or not
	if token, ok := r.Header["Authorization"]; !ok || token == nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode("User not logged in")
		return false, claims
	}

	//Check if JWT valid.
	tokenString := r.Header["Authorization"][0]
	_, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(Key), nil
	})
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode("Invalid User")
		log.Println(err)
		return false, claims
	} else {
		return true, claims
	}
}
