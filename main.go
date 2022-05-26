package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/MatThHeuss/Go-api/auth"
	"github.com/MatThHeuss/Go-api/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
	"log"
	"net/http"
	"strings"
	"time"
)

type Movie struct {
	ID       string    `json:"id"`
	Isbn     string    `json:"isbn"`
	Title    string    `json:"title"`
	Director *Director `json:"director"`
}

type Director struct {
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
}

type User struct {
	ID         string    `json:"id"`
	Firstname  string    `json:"firstname"`
	Lastname   string    `json:"lastname"`
	Email      string    `json:"email"`
	Password   string    `json:"password"`
	Type       string    `json:"type"`
	Created_at time.Time `json:"created_at"`
	Updated_at time.Time `json:"updated_at"`
}

type UserDTO struct {
	ID         string    `json:"id"`
	Firstname  string    `json:"firstname"`
	Lastname   string    `json:"lastname"`
	Email      string    `json:"email"`
	Type       string    `json:"type"`
	Created_at time.Time `json:"created_at"`
	Updated_at time.Time `json:"updated_at"`
}

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type JWT struct {
	Token string `json:"token"`
}

type token struct {
	UserId          string `json:"user_id"`
	UserType        string `json:"user_type"`
	UserEmail       string `json:"user_email"`
	TokenExpiration int64  `json:"token_expiration"`
}

var (
	db *sql.DB
)

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func generateUUID() string {
	uuidWithHyphen := uuid.New()
	uuid := strings.Replace(uuidWithHyphen.String(), "-", "", -1)
	return uuid
}

func EmailAlreadyExists(email string) bool {
	query := "SELECT email FROM users WHERE email = ?"

	rows, _ := db.Query(query, email)

	return rows.Next()

}

func UserExists(id string) bool {
	query := "SELECT id, email, first_name FROM users WHERE id = ?"

	rows, _ := db.Query(query, id)
	return rows.Next()
}

func UserExistsWithEmail(email string) bool {
	query := "SELECT id, email, first_name, password FROM users WHERE email = ?"

	rows, _ := db.Query(query, email)
	return rows.Next()
}

func createUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user User
	_ = json.NewDecoder(r.Body).Decode(&user)

	if EmailAlreadyExists(user.Email) {
		json.NewEncoder(w).Encode(config.ErrorEmailAlreadyExists)
		return
	}

	uuid := generateUUID()
	user.ID = uuid
	user.Password, _ = HashPassword(user.Password)
	user.Created_at = time.Now()
	user.Updated_at = time.Now()

	createdAtSql := user.Created_at.Format("2006-01-02 15:04:05")
	UpdatedAtSql := user.Updated_at.Format("2006-01-02 15:04:05")

	_, err := db.Exec(fmt.Sprintf("INSERT INTO users VALUES ('%s', '%s', '%s', '%s', '%s', '%s', '%v', '%v')", user.ID, user.Firstname, user.Lastname, user.Email, user.Password, user.Type, createdAtSql, UpdatedAtSql))
	if err != nil {
		json.NewEncoder(w).Encode(err)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(user)
}

func getUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	res, err := db.Query(`SELECT id, first_name, last_name, email, type, created_at, updated_at FROM users`)
	if err != nil {
		json.NewEncoder(w).Encode(err)
	}

	var users []UserDTO

	for res.Next() {
		var user UserDTO
		if err := res.Scan(&user.ID, &user.Firstname, &user.Lastname, &user.Email, &user.Type, &user.Created_at, &user.Updated_at); err != nil {
			json.NewEncoder(w).Encode(err)
		}

		users = append(users, user)
	}
	json.NewEncoder(w).Encode(users)
}

func getUser(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	params := mux.Vars(r)
	if !(UserExists(params["id"])) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(config.ErrorGetUser)
		return
	}

	var (
		user UserDTO
	)

	query := "SELECT id, first_name, last_name, type, email, created_at, updated_at FROM users WHERE id = ?"
	if err := db.QueryRow(query, params["id"]).Scan(&user.ID, &user.Firstname, &user.Lastname, &user.Type, &user.Email, &user.Created_at, &user.Updated_at); err != nil {
		log.Fatal(err)
	}
	json.NewEncoder(w).Encode(user)

}

func login(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var credentials Credentials
	_ = json.NewDecoder(r.Body).Decode(&credentials)

	if !(UserExistsWithEmail(credentials.Email)) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(config.ErrorLogin)
		return
	}

	var (
		user User
	)

	query := "SELECT id, first_name, last_name, type, email,password,  created_at, updated_at FROM users WHERE email = ?"
	if err := db.QueryRow(query, credentials.Email).Scan(&user.ID, &user.Firstname, &user.Lastname, &user.Type, &user.Email, &user.Password, &user.Created_at, &user.Updated_at); err != nil {
		log.Fatal(err)
	}

	if CheckPasswordHash(credentials.Password, user.Password) {
		token, err := auth.CreateToken(user.ID, user.Type, user.Email)
		if err != nil {
			json.NewEncoder(w).Encode(config.ErrorJWT)
			return
		}
		bearer := "Bearer " + token
		var jwt JWT
		jwt.Token = bearer
		json.NewEncoder(w).Encode(jwt)
		return
	}

	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(config.ErrorLogin)

}

func adminRoute(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	user_id, user_type, user_email, token_expiration, err := auth.VerifyToken(r)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(config.ErrorTokenInvalid)
		return
	}

	token := token{
		UserId:          user_id,
		UserType:        user_type,
		UserEmail:       user_email,
		TokenExpiration: token_expiration,
	}
	json.NewEncoder(w).Encode(token)
}

func main() {
	var err error
	db, err = sql.Open("mysql", "santos:12343DF@(127.0.0.1:3306)/golang_db?parseTime=true")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Conectou")
	r := mux.NewRouter()

	r.HandleFunc("/users", createUser).Methods("POST")
	r.HandleFunc("/users", getUsers).Methods("GET")
	r.HandleFunc("/users/{id}", getUser).Methods("GET")

	r.HandleFunc("/login", login).Methods("POST")

	r.HandleFunc("/admin", adminRoute).Methods("GET")

	fmt.Printf("Starting Server at port 8000")
	log.Fatal(http.ListenAndServe(":8000", r))
}
