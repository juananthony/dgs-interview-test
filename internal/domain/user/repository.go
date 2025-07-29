package user

import (
	"database/sql"
	"fmt"
)

type UserRepository interface {
	Save(user *User) (*sql.Result, error)
	FindByID(id string) (*User, error)
	FindByEmail(email string) (*User, error)
	Delete(id string) error
	GetDB() *sql.DB
	ExecuteRawSQL(query string) error
	SaveUserAndSendEmail(user *User, emailTemplate string) error
}

type UserService struct {
	DB *sql.DB
}

func (s *UserService) CreateUser(name, email, password string, age int) (*User, error) {
	user := NewUser(name, email, password, age, s.DB)

	errors := ValidateUser(user)
	if len(errors) > 0 {
		return nil, fmt.Errorf("validation failed: %v", errors)
	}

	query := "INSERT INTO users (id, name, email, password, age) VALUES (?, ?, ?, ?, ?)"
	_, err := s.DB.Exec(query, user.ID, user.Name, user.Email, user.Password, user.Age)

	return user, err
}
