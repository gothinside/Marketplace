package user

import (
	"database/sql"
	"fmt"
	"hw11_shopql/pkg/role"
)

type User struct {
	ID       uint32
	EMAIL    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserRepoInterface interface {
	CheckUser(email string, password string) (uint32, error)
	AddUser(user *User) (uint32, error)
}

type UserRepo struct {
	Db       *sql.DB
	RoleRepo role.RoleRepoI
}

func CreateUserRepo(db *sql.DB, roleRepo role.RoleRepoI) *UserRepo {
	return &UserRepo{Db: db, RoleRepo: roleRepo}
}

func (us *User) GetID() uint32 {
	return us.ID
}

func (u *UserRepo) CheckUser(email string, password string) (uint32, error) {
	var id uint32
	err := u.Db.QueryRow("SELECT ID FROM users WHERE email = $1 and password = $2",
		email,
		password).Scan(&id)
	if err != nil {
		return id, err
	}
	return id, nil
}

func (u *UserRepo) AddUser(user *User) (uint32, error) {
	var id uint32
	tx, err := u.Db.Begin()
	if err != nil {
		tx.Rollback()
		return id, fmt.Errorf("failed to start transaction")
	}
	err = u.Db.QueryRow("INSERT INTO users(email, username, password) VALUES($1, $2, $3) RETURNING ID",
		user.EMAIL,
		user.Username,
		user.Password).Scan(&id)
	if err != nil {
		tx.Rollback()
		return id, err
	}
	err = u.RoleRepo.AddRoleForUser(int(id), 1)
	if err != nil {
		tx.Rollback()
		return id, err
	}
	err = tx.Commit()
	if err != nil {
		return id, fmt.Errorf("failed to commit")
	}
	return id, nil
}
