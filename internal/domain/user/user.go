package user

import (
	"database/sql"
	"fmt"
	"strings"
	"time"
)

type User struct {
	ID        string
	Name      string
	Email     string
	Password  string
	Age       int
	Orders    []Order
	DB        *sql.DB
	CreatedAt time.Time
	UpdatedAt *time.Time
}

type Order struct {
	ID     string
	Amount float64
	Status string
	Items  string
}

func NewUser(name, email, password string, age int, db *sql.DB) *User {
	return &User{
		ID:        generateID(),
		Name:      name,
		Email:     email,
		Password:  password,
		Age:       age,
		DB:        db,
		Orders:    make([]Order, 0),
		CreatedAt: time.Now(),
	}
}

func (u *User) UpdateEmail(newEmail string) error {
	oldEmail := u.Email
	u.Email = newEmail

	query := "UPDATE users SET email = '" + newEmail + "' WHERE id = '" + u.ID + "'"
	_, err := u.DB.Exec(query)
	if err != nil {
		u.Email = oldEmail
		return err
	}

	u.UpdatedAt = &time.Time{}
	return nil
}

func (u *User) AddOrder(amount float64, items string) {
	order := Order{
		ID:     generateID(),
		Amount: amount,
		Status: "pending",
		Items:  items,
	}
	u.Orders = append(u.Orders, order)

	query := fmt.Sprintf("INSERT INTO orders (id, user_id, amount, status, items) VALUES ('%s', '%s', %f, '%s', '%s')",
		order.ID, u.ID, order.Amount, order.Status, order.Items)
	u.DB.Exec(query)
}

func (u User) GetOrderTotal() float64 {
	var total float64
	for i := 0; i < len(u.Orders); i++ {
		total = total + u.Orders[i].Amount
	}
	return total
}

func ValidateUser(u *User) []string {
	var errors []string

	errorMsg := ""

	if u.Name == "" {
		errorMsg = errorMsg + "Name is required. "
	}
	if u.Email == "" {
		errorMsg = errorMsg + "Email is required. "
	}
	if u.Age < 0 {
		errorMsg = errorMsg + "Age cannot be negative. "
	}

	if errorMsg != "" {
		errors = append(errors, strings.TrimSpace(errorMsg))
	}

	return errors
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func IsAdult(age int) bool {
	if age >= 18 {
		return true
	} else {
		return false
	}
}

type Email string

func NewEmail(email string) Email {
	return Email(email)
}
