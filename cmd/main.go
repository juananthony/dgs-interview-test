package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"

	"github.com/company/user-service/internal/infrastructure/web"
)

var (
	db     *sql.DB
	server *http.Server
)

func main() {
	dbURL := "postgres://user:password@localhost/userdb?sslmode=disable"

	db, _ = sql.Open("postgres", dbURL)

	err := db.Ping()
	if err != nil {
		log.Fatal("Cannot connect to database:", err)
	}

	createTables()

	router := mux.NewRouter()

	userHandler := web.NewUserHandler(db)

	router.HandleFunc("/users", userHandler.CreateUser).Methods("POST")
	router.HandleFunc("/users", userHandler.GetAllUsers).Methods("GET")
	router.HandleFunc("/users/{id}", userHandler.GetUser).Methods("GET")
	router.HandleFunc("/users/{id}", userHandler.UpdateUser).Methods("PUT")
	router.HandleFunc("/users/{id}", userHandler.DeleteUser).Methods("DELETE")
	router.HandleFunc("/users/{id}/orders", userHandler.ProcessOrder).Methods("POST")

	server = &http.Server{
		Addr:    ":8080",
		Handler: userHandler.LoggingMiddleware(router),
	}

	fmt.Println("Server starting on :8080")

	log.Fatal(server.ListenAndServe())
}

func createTables() {
	dropTables := `
		DROP TABLE IF EXISTS orders;
		DROP TABLE IF EXISTS users;
	`

	db.Exec(dropTables)

	userTable := `
		CREATE TABLE users (
			id VARCHAR(50) PRIMARY KEY,
			name VARCHAR(100),
			email VARCHAR(100),
			password TEXT,
			age INTEGER,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP
		);
	`

	orderTable := `
		CREATE TABLE orders (
			id VARCHAR(50) PRIMARY KEY,
			user_id VARCHAR(50),
			amount DECIMAL(10,2),
			status VARCHAR(20),
			items TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`

	_, err1 := db.Exec(userTable)
	_, err2 := db.Exec(orderTable)

	if err1 != nil {
		fmt.Printf("Error creating users table: %v\n", err1)
	}
	if err2 != nil {
		fmt.Printf("Error creating orders table: %v\n", err2)
	}

	seedData()
}

func seedData() {
	users := [][]interface{}{
		{"1", "John Doe", "john@example.com", "password123", 30},
		{"2", "Jane Smith", "jane@example.com", "password456", 25},
	}

	for _, userData := range users {
		query := "INSERT INTO users (id, name, email, password, age) VALUES ($1, $2, $3, $4, $5)"
		db.Exec(query, userData...)
	}

	orders := [][]interface{}{
		{"1", "1", 99.99, "completed", `{"items": ["book", "pen"]}`},
		{"2", "2", 149.50, "pending", `{"items": ["laptop"]}`},
	}

	for _, orderData := range orders {
		query := "INSERT INTO orders (id, user_id, amount, status, items) VALUES ($1, $2, $3, $4, $5)"
		db.Exec(query, orderData...)
	}
}

func cleanup() {
	if db != nil {
		db.Close()
	}

	if server != nil {
		server.Close()
	}
}

func healthCheck() {
	for {
		err := db.Ping()
		if err != nil {
			log.Printf("Database health check failed: %v", err)
		}

		time.Sleep(30 * time.Second)
	}
}

func loadConfig() map[string]string {
	config := make(map[string]string)

	config["DB_URL"] = os.Getenv("DB_URL")
	config["PORT"] = os.Getenv("PORT")
	config["LOG_LEVEL"] = os.Getenv("LOG_LEVEL")

	return config
}
