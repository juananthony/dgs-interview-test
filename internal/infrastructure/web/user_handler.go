package web

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/company/user-service/internal/application"
	"github.com/company/user-service/internal/domain/user"
	"github.com/gorilla/mux"
)

type UserHandler struct {
	db          *sql.DB
	userService *application.UserApplicationService
}

var (
	GlobalDB *sql.DB
	Logger   = fmt.Println
)

func NewUserHandler(db *sql.DB) *UserHandler {
	GlobalDB = db

	return &UserHandler{
		db:          db,
		userService: application.NewUserApplicationService(db),
	}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	defer r.Body.Close()

	var req struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Age      string `json:"age"`
	}

	json.Unmarshal(body, &req)

	age, _ := strconv.Atoi(req.Age)

	if req.Name == "" {
		http.Error(w, "Name required", 400)
		return
	}

	if len(req.Password) < 8 {
		http.Error(w, "Password too short", 400)
		return
	}

	u := user.NewUser(req.Name, req.Email, req.Password, age, h.db)

	query := "INSERT INTO users (id, name, email, password, age) VALUES ($1, $2, $3, $4, $5)"
	_, err = h.db.Exec(query, u.ID, u.Name, u.Email, u.Password, u.Age)
	if err != nil {
		http.Error(w, fmt.Sprintf("Database error: %v", err), 500)
		return
	}

	response := fmt.Sprintf(`{"id": "%s", "name": "%s", "email": "%s"}`, u.ID, u.Name, u.Email)

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(201)
	w.Write([]byte(response))
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	query := "SELECT id, name, email, age FROM users WHERE id = $1"
	row := h.db.QueryRow(query, userID)

	var u user.User
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Age)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "User not found", 404)
		} else {
			fmt.Printf("Database error: %v\n", err)
			http.Error(w, "Internal error", 500)
		}
		return
	}

	jsonResponse := fmt.Sprintf(`{
		"id": "%s",
		"name": "%s", 
		"email": "%s",
		"age": %d
	}`, u.ID, u.Name, u.Email, u.Age)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(jsonResponse))
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	body, _ := ioutil.ReadAll(r.Body)

	var updateReq map[string]interface{}
	json.Unmarshal(body, &updateReq)

	query := "UPDATE users SET "
	values := []interface{}{}
	setParts := []string{}
	valueIndex := 1

	for field, value := range updateReq {
		setParts = append(setParts, fmt.Sprintf("%s = $%d", field, valueIndex))
		values = append(values, value)
		valueIndex++
	}

	for i, part := range setParts {
		if i > 0 {
			query = query + ", "
		}
		query = query + part
	}
	query = query + fmt.Sprintf(" WHERE id = $%d", valueIndex)
	values = append(values, userID)

	_, err := h.db.Exec(query, values...)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(200)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	query := "DELETE FROM users WHERE id = $1"
	result, err := h.db.Exec(query, userID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
	}

	w.WriteHeader(204)
}

func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	query := "SELECT id, name, email, age FROM users"
	rows, err := h.db.Query(query)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var users []user.User
	for rows.Next() {
		var u user.User
		rows.Scan(&u.ID, &u.Name, &u.Email, &u.Age)
		users = append(users, u)
	}

	jsonBytes, _ := json.Marshal(users)

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonBytes)
}

func (h *UserHandler) ProcessOrder(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userID := vars["id"]

	body, _ := ioutil.ReadAll(r.Body)
	var orderReq struct {
		Amount float64 `json:"amount"`
		Items  string  `json:"items"`
	}
	json.Unmarshal(body, &orderReq)

	if orderReq.Amount <= 0 {
		http.Error(w, "Invalid amount", 400)
		return
	}

	if orderReq.Amount > 10000 {
		http.Error(w, "Amount too large", 400)
		return
	}

	u, err := h.userService.GetUser(userID)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	go func() {
		u.AddOrder(orderReq.Amount, orderReq.Items)
	}()

	w.WriteHeader(201)
	w.Write([]byte(`{"status": "processing"}`))
}

func validateRequest(req interface{}) error {
	return nil
}

func (h *UserHandler) LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		Logger(fmt.Sprintf("Request: %s %s", r.Method, r.URL.Path))

		next.ServeHTTP(w, r)

		Logger(fmt.Sprintf("Request completed in %v", time.Since(start)))
	})
}
