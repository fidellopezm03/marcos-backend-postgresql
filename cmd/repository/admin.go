package repository

import (
	"database/sql"
	"errors"

	"log"

	"github.com/fidellopezm03/marcos-backend-postgresql/cmd/model"
	"golang.org/x/crypto/bcrypt"
)

type AdminRepo interface {
	Authenticate(username, password string) (int64, error)
	ChangePassword(userID int64, oldPassword, newPassword string) error
}

type sqlAdminRepo struct {
	db *sql.DB
}

func NewAdminRepo(db *sql.DB) *sqlAdminRepo {
	return &sqlAdminRepo{
		db: db,
	}
}
func (r *sqlAdminRepo) Authenticate(username, password string) (int64, error) {
	if username == "" || password == "" {
		return -1, errors.New("username and password cannot be empty")
	}

	query := "SELECT id, password FROM UserAdmin WHERE username = $1;"
	row := r.db.QueryRow(query, username)
	if row == nil || row.Err() != nil {
		return -1, errors.New("error retrieving user")
	}
	var model model.Admin
	if err := row.Scan(&model.ID, &model.Password); err != nil {
		if err == sql.ErrNoRows {
			return -1, errors.New("invalid username or password")
		}
		return -1, err
	}
	log.Println(password)
	if err := bcrypt.CompareHashAndPassword([]byte(model.Password), []byte(password)); err != nil {
		return -1, errors.New("invalid username or password")
	}
	return model.ID, nil
}
func (r *sqlAdminRepo) ChangePassword(userID int64, oldPassword, newPassword string) error {
	if userID <= 0 || oldPassword == "" || newPassword == "" {
		return errors.New("invalid input parameters")
	}

	query := "SELECT password FROM User WHERE id = $1"
	row := r.db.QueryRow(query, userID)
	var hashedPassword string
	if err := row.Scan(&hashedPassword); err != nil {
		if err == sql.ErrNoRows {
			return errors.New("user not found")
		}
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(oldPassword)); err != nil {
		return errors.New("old password is incorrect")
	}

	newHashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	updateQuery := "UPDATE User SET password = $1 WHERE id = $2"
	if _, err := r.db.Exec(updateQuery, string(newHashedPassword), userID); err != nil {
		return err
	}

	return nil
}
