package repository

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/fidellopezm03/marcos-backend-postgresql/cmd/model"
	"golang.org/x/crypto/bcrypt"
)

type Content struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	Idimg       int64  `json:"idImg"`
}
type AdminRepo interface {
	Authenticate(username, password string) (string, error)
	ChangePassword(userID int64, oldPassword, newPassword string) error
	FindImg(id string) (string, error)
	SaveImg(id int64, path string) (string, error)
	CreateImg(idContent int64, path, name string) (int64, error)
	SaveContent(id int64, title, description string) error
	CreateContent(title, description, location string) (int64, error)
	GetAllcontent() (map[string][]Content, error)
	SaveInfo(key, value string) error
	GetAllinfo() (map[string]string, error)
}

const UploadDir = "/uploads"

var aesKey = []byte("aseskeyFromAES")

type sqlAdminRepo struct {
	db *sql.DB
}

func encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	chipertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(chipertext), nil
}

func decrypt(codestr string) (string, error) {
	data, err := hex.DecodeString(codestr)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonceSice := gcm.NonceSize()
	if len(data) < nonceSice {
		return "", fmt.Errorf("data short")
	}
	nonce, chipertext := data[:nonceSice], data[nonceSice:]
	plaintext, err := gcm.Open(nil, nonce, chipertext, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func NewAdminRepo(db *sql.DB) *sqlAdminRepo {
	return &sqlAdminRepo{
		db: db,
	}
}

func (r *sqlAdminRepo) Authenticate(username, password string) (string, error) {
	if username == "" || password == "" {
		return "", errors.New("username and password cannot be empty")
	}

	query := "SELECT username, password FROM UserAdmin WHERE username = $1;"
	row := r.db.QueryRow(query, username)
	if row == nil || row.Err() != nil {
		return "", errors.New("error retrieving user")
	}
	var model model.Admin
	if err := row.Scan(&model.Username, &model.Password); err != nil {
		if err == sql.ErrNoRows {
			return "", errors.New("invalid username or password")
		}
		return "", err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(model.Password), []byte(password)); err != nil {
		return "", errors.New("invalid username or password")
	}
	return model.Username, nil
}

func (r *sqlAdminRepo) SaveContent(id int64, title, description string) error {
	if id < 1 {
		return errors.New("invalid id")
	}

	query := "UPDATE contents SET title = $1, description = $2 WHERE id = $3;"
	res, err := r.db.Exec(query, title, description, id)
	if err != nil {
		return fmt.Errorf("error in query: %v", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error in query: %v", err)
	}
	if n == 0 {
		return errors.New("content not exist")
	}

	return nil
}
func (r *sqlAdminRepo) CreateContent(title, description, location string) (int64, error) {
	location = strings.ToLower(location)
	if location != "consejo" && location != "negocio" {
		return -1, errors.New("error not valid location")
	}

	query := "INSERT INTO contents (title, description,location) VALUES ($1, $2,$3) RETURNING id"

	var idReturning int64
	if err := r.db.QueryRow(query, title, description, location).Scan(&idReturning); err != nil {
		return -1, fmt.Errorf("error in query: %v", err)
	}
	return idReturning, nil

}

func (r *sqlAdminRepo) ChangePassword(userID int64, oldPassword, newPassword string) error {
	if userID <= 0 || oldPassword == "" || newPassword == "" {
		return errors.New("invalid input parameters")
	}

	query := "SELECT password FROM UserAdmin WHERE id = $1;"
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

	updateQuery := "UPDATE UserAdmin SET password = $1 WHERE id = $2;"
	if _, err := r.db.Exec(updateQuery, string(newHashedPassword), userID); err != nil {
		return err
	}

	return nil
}

func (r *sqlAdminRepo) FindImg(id string) (string, error) {
	if len(id) < 1 {
		return "", errors.New("error in id")
	}
	var code string
	query := "SELECT file_path FROM files WHERE id = $1;"
	err := r.db.QueryRow(query, id).Scan(&code)
	if err != nil {
		return "", errors.New("img not found")
	}
	path, err := decrypt(code)
	if err != nil {
		return "", err
	}
	return path, nil
}

func (r *sqlAdminRepo) SaveImg(id int64, path string) (string, error) {
	encPath, err := encrypt(path)
	if err != nil {
		return "", err
	}
	var cypathOldImg string
	query := "SELECT file_path FROM files WHERE id = $1;"
	if err = r.db.QueryRow(query, id).Scan(&cypathOldImg); err != nil {
		return "", errors.New("img not exist")
	}
	pathOldImg, err := decrypt(cypathOldImg)
	if err != nil {
		return "", err
	}

	query = "UPDATE files SET file_path = $1 WHERE id = $2;"
	_, err = r.db.Exec(query, encPath, id)
	if err != nil {
		return "", fmt.Errorf("error in update query: %v", err)
	}

	return pathOldImg, nil
}
func (r *sqlAdminRepo) CreateImg(idContent int64, path, name string) (int64, error) {
	query := "INSERT INTO files (name, file_path) VALUES ($1,$2) RETURNING id;"
	var id int64
	if err := r.db.QueryRow(query, name, path).Scan(&id); err != nil {
		return -1, fmt.Errorf("error insert img :%v", err)
	}
	return id, nil
}
func (r *sqlAdminRepo) SaveInfo(key, value string) error {
	if len(key) < 2 || len(value) == 0 {
		return errors.New("error params not valid")
	}
	query := "UPDATE files SET name = $1 WHERE key = $2;"
	res, err := r.db.Exec(query, value, key)
	if err != nil {
		return fmt.Errorf("error update info: %v", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error cheking update: %v", err)
	}
	if n == 0 {
		return fmt.Errorf("not exist info: %v", err)
	}
	return nil

}
func (r *sqlAdminRepo) GetAllinfo() (map[string]string, error) {

	query := "SELECT key,value FROM info;"
	rows, err := r.db.Query(query)

	if err != nil {
		return nil, fmt.Errorf("error in query: %v", err)
	}
	defer rows.Close()
	var info struct {
		Key   string `db:"key"`
		Value string `db:"value"`
	}
	Info := make(map[string]string)
	for rows.Next() {
		if err := rows.Scan(&info.Key, &info.Value); err != nil {
			log.Println(err)
			continue
		}
		Info[info.Key] = info.Value
	}
	return Info, nil

}
func (r *sqlAdminRepo) GetAllcontent() (map[string][]Content, error) {
	query := "SELECT c.title, c.description, c.location, f.id as idImg FROM contents c LEFT JOIN file f ON c.id = f.content_id;"
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("error geting content: %v", err)
	}
	defer rows.Close()
	ContentMap := make(map[string][]Content)
	var content struct {
		Title       string        `db:"title"`
		Description string        `db:"description"`
		Location    string        `db:"location"`
		Idimg       sql.NullInt64 `db:"idImg"`
	}
	var idImg int64
	for rows.Next() {
		if err := rows.Scan(&content.Title, &content.Description, &content.Location, &content.Idimg); err != nil {
			log.Println(err)
			continue
		}
		idImg = -1
		if ok := content.Idimg.Valid; ok {
			idImg = content.Idimg.Int64
		}
		ContentMap[content.Location] = append(ContentMap[content.Location], Content{
			Title:       content.Title,
			Description: content.Description,
			Idimg:       idImg,
		})

	}
	return ContentMap, nil
}
