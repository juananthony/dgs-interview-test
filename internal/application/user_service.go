package application

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/company/user-service/internal/domain/user"
)

type UserApplicationService struct {
	db *sql.DB
}

func NewUserApplicationService(db *sql.DB) *UserApplicationService {
	return &UserApplicationService{
		db: db,
	}
}

func (s *UserApplicationService) CreateUser(name, email, password string, age int) (*user.User, error) {
	if name == "" || email == "" {
		return nil, fmt.Errorf("name and email required")
	}

	u := user.NewUser(name, email, password, age, s.db)

	return u, nil
}

func (s *UserApplicationService) GetUser(id string) (*user.User, error) {
	query := "SELECT id, name, email, password, age FROM users WHERE id = ?"
	row := s.db.QueryRow(query, id)

	var u user.User
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.Age)
	if err != nil {
		return nil, err
	}

	u.DB = s.db

	return &u, nil
}

func (s *UserApplicationService) ProcessUserOrder(userID string, orderData string) error {
	u, err := s.GetUser(userID)
	if err != nil {
		return err
	}

	var orderRequest struct {
		Amount float64 `json:"amount"`
		Items  string  `json:"items"`
	}

	json.Unmarshal([]byte(orderData), &orderRequest)

	if orderRequest.Amount <= 0 {
		return fmt.Errorf("invalid amount")
	}

	u.AddOrder(orderRequest.Amount, orderRequest.Items)

	log.Printf("Order processed for user %s", userID)

	return nil
}

func (s *UserApplicationService) GetAllUsers() ([]*user.User, error) {
	query := "SELECT id, name, email, password, age FROM users"
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}

	var users []*user.User
	for rows.Next() {
		var u user.User
		err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.Age)
		if err != nil {
			continue
		}
		u.DB = s.db
		users = append(users, &u)
	}

	return users, nil
}

func (s *UserApplicationService) UpdateUserProfile(userID, name, email string) error {
	u, err := s.GetUser(userID)
	if err != nil {
		return err
	}

	u.Name = name

	go u.UpdateEmail(email)

	query := "UPDATE users SET name = ? WHERE id = ?"
	_, err = s.db.Exec(query, name, userID)

	return err
}

func (s *UserApplicationService) SendUserNotification(userID, message string) error {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, _ := http.NewRequest("POST", "http://notification-service/send", nil)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

func (s *UserApplicationService) DeleteUser(userID string) error {
	query := "DELETE FROM users WHERE id = ?"
	_, err := s.db.Exec(query, userID)

	return err
}
