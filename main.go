package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/joho/godotenv"
)

// User model struct
type User struct {
	gorm.Model

	Name  string `gorm:"size:64"`
	Email string `gorm:"type:varchar(100);unique"`
}

// UserData struct
type UserData struct {
	Name  string
	Email string
}

// Get all users from DB
func getAllUsers(w http.ResponseWriter, r *http.Request) {
	db := connectToDatabase()
	defer db.Close()

	var users []User

	// Get all users from DB
	db.Find(&users)

	// Encode into json
	json.NewEncoder(w).Encode(users)

}

// Create a new user in DB
func createUser(w http.ResponseWriter, r *http.Request) {

	var userData UserData

	db := connectToDatabase()
	defer db.Close()

	err := json.NewDecoder(r.Body).Decode(&userData)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user := User{Name: userData.Name, Email: userData.Email}

	// Save new user in DB
	db.Save(&user)

	json.NewEncoder(w).Encode(user)

}

// Update existing user data in DB
func updateUser(w http.ResponseWriter, r *http.Request) {

	var userData UserData
	var user User

	db := connectToDatabase()
	defer db.Close()

	// Get user id from URL path
	vars := mux.Vars(r)
	userID := vars["id"]

	db.Where("id = ?", userID).Find(&user)

	if user.ID == 0 {
		http.Error(w, "User with this id does not exist.", http.StatusBadRequest)
		return
	}

	err := json.NewDecoder(r.Body).Decode(&userData)

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update user data in DB
	db.Model(&user).Updates(&userData)

	json.NewEncoder(w).Encode(user)

}

// Delete existing user from DB
func deleteUser(w http.ResponseWriter, r *http.Request) {

	var user User

	db := connectToDatabase()
	defer db.Close()

	// Get user id from URL path
	vars := mux.Vars(r)
	userID := vars["id"]

	db.Where("id = ?", userID).Find(&user)

	if user.ID == 0 {
		http.Error(w, "User with this id does not exist.", http.StatusBadRequest)
		return
	}

	db.Delete(&user)

	json.NewEncoder(w).Encode("User deleted successfully!")
}

// Routing
func handleRequests() {

	// Define new router
	router := mux.NewRouter().StrictSlash(true)

	// Define API handlers
	router.HandleFunc("/users", getAllUsers).Methods("GET")
	router.HandleFunc("/users", createUser).Methods("POST")
	router.HandleFunc("/users/{id:[0-9]+}", updateUser).Methods("PATCH")
	router.HandleFunc("/users/{id:[0-9]+}", deleteUser).Methods("DELETE")

	log.Fatal(http.ListenAndServe("127.0.0.1:8000", router))
}

// Connect to postgres database
func connectToDatabase() *gorm.DB {

	// Load DB config
	dbUser := os.Getenv("DB_USER")
	dbName := os.Getenv("DB_NAME")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort, _ := strconv.Atoi(os.Getenv("DB_PORT"))

	dbURI := fmt.Sprintf("host=%s user=%s dbname=%s password=%s port=%d sslmode=disable", dbHost, dbUser, dbName, dbPassword, dbPort)

	// Connect to DB
	db, err := gorm.Open("postgres", dbURI)

	if err != nil {
		panic(err)
	}

	return db

}

// Initialization
func init() {

	// Load .env file
	err := godotenv.Load()
	if err != nil {
		panic(err)
	}

	db := connectToDatabase()

	// Close DB
	defer db.Close()

	// Database migration
	db.Debug().AutoMigrate(&User{})

}

// Main
func main() {

	handleRequests()

}
