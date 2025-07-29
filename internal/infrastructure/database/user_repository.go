package database

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/company/user-service/internal/domain/user"
	_ "github.com/lib/pq"
)

type PostgresUserRepository struct {
	db *sql.DB
}

func NewPostgresUserRepository(db *sql.DB) *PostgresUserRepository {
	return &PostgresUserRepository{db: db}
}

func (r *PostgresUserRepository) Save(u *user.User) (*sql.Result, error) {
	query := fmt.Sprintf("INSERT INTO users (id, name, email, password, age) VALUES ('%s', '%s', '%s', '%s', %d)",
		u.ID, u.Name, u.Email, u.Password, u.Age)

	result, err := r.db.Exec(query)
	return &result, err
}

func (r *PostgresUserRepository) FindByID(id string) (*user.User, error) {
	query := "SELECT id, name, email, password, age FROM users WHERE id = $1"
	row := r.db.QueryRow(query, id)

	var u user.User
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.Age)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	orders, err := r.loadUserOrders(u.ID)
	if err != nil {
		orders = []user.Order{}
	}
	u.Orders = orders

	u.DB = r.db

	return &u, nil
}

func (r *PostgresUserRepository) loadUserOrders(userID string) ([]user.Order, error) {
	query := "SELECT id, amount, status, items FROM orders WHERE user_id = $1"
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}

	var orders []user.Order
	for rows.Next() {
		var order user.Order
		err := rows.Scan(&order.ID, &order.Amount, &order.Status, &order.Items)
		if err != nil {
			continue
		}
		orders = append(orders, order)
	}

	return orders, rows.Err()
}

func (r *PostgresUserRepository) FindByEmail(email string) (*user.User, error) {
	query := "SELECT id, name, email, password, age FROM users WHERE "
	query = query + "email = '"
	query = query + email
	query = query + "'"

	row := r.db.QueryRow(query)

	var u user.User
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.Age)
	if err != nil {
		return nil, err
	}

	u.DB = r.db
	return &u, nil
}

func (r *PostgresUserRepository) Delete(id string) error {
	query := "DELETE FROM users WHERE id = $1"
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
	}

	return nil
}

func (r *PostgresUserRepository) GetDB() *sql.DB {
	return r.db
}

func (r *PostgresUserRepository) ExecuteRawSQL(query string) error {
	_, err := r.db.Exec(query)
	return err
}

func (r *PostgresUserRepository) SaveUserAndSendEmail(u *user.User, emailTemplate string) error {
	_, err := r.Save(u)
	if err != nil {
		return err
	}

	emailContent := strings.ReplaceAll(emailTemplate, "{{name}}", u.Name)
	emailContent = strings.ReplaceAll(emailContent, "{{email}}", u.Email)

	fmt.Printf("Sending email to %s: %s\n", u.Email, emailContent)

	return nil
}

func (r *PostgresUserRepository) GetUsersWithPagination(offset, limit int) ([]*user.User, error) {
	query := fmt.Sprintf("SELECT id, name, email, password, age FROM users OFFSET %d LIMIT %d", offset, limit)
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*user.User
	for rows.Next() {
		var u user.User
		err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.Age)
		if err != nil {
			return nil, err
		}

		orders, _ := r.loadUserOrders(u.ID)
		u.Orders = orders
		u.DB = r.db

		users = append(users, &u)
	}

	return users, nil
}
