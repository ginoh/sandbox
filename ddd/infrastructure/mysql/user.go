package mysql

import (
	"database/sql"

	"example.com/domain/model/user"
)

type UserRepository struct {
	*sql.DB
}

func NewUserRepository(db *sql.DB) user.UserRepository {
	return &UserRepository{db}
}

func (ur *UserRepository) FindByID(id uint32) (*user.User, error) {
	stmt, err := ur.Prepare(`SELECT id, name, created_at, updated_at FROM users WHERE id = ?`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	u := &user.User{}
	err = stmt.QueryRow(id).Scan(&u.ID, &u.Name, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return u, nil
}

func (ur *UserRepository) Create(u *user.User) (*user.User, error) {
	stmt, err := ur.Prepare(`INSERT INTO users(name, created_at, updated_at) VALUES (?, ?, ?)`)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(u.Name, u.CreatedAt, u.UpdatedAt)
	if err != nil {
		return nil, err
	}

	insertId, _ := result.LastInsertId()
	u.ID = uint32(insertId)

	return u, nil
}

func (ur *UserRepository) Update(u *user.User) error {
	stmt, err := ur.Prepare(`UPDATE users SET name = ?, updated_at = ? WHERE id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(u.Name, u.UpdatedAt, u.ID); err != nil {
		return err
	}

	return nil
}

func (ur *UserRepository) Delete(id uint32) error {
	stmt, err := ur.Prepare(`DELETE FROM users WHERE id = ?`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	if _, err := stmt.Exec(id); err != nil {
		return err
	}

	return nil
}
