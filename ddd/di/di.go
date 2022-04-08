package di

import (
	"database/sql"
	"os"

	au "example.com/application/user"
	"example.com/infrastructure/mysql"
	pu "example.com/presentation/user"
)

var (
	db *sql.DB
)

func Setup() error {
	var err error
	db, err = sql.Open("mysql", os.Getenv("MYSQL_DSN"))
	if err != nil {
		return err
	}

	// migration

	return nil
}

func TearDown() {
	db.Close()
}

func InitUser() pu.UserHandler {
	r := mysql.NewUserRepository(db)
	s := au.NewUserService(r)
	return pu.NewUserHandler(s)
}
