package role

import (
	"database/sql"
	"fmt"
)

type RoleRepo struct {
	Db *sql.DB
}

type RoleRepoI interface {
	AddRoleForUser(id int, RoleID int) error
}

func (RP *RoleRepo) AddRoleForUser(id int, RoleID int) error {
	_, err := RP.Db.Exec("INSERT INTO user_role VALUES($1, $2)", id, RoleID)
	if err != nil {
		return fmt.Errorf("failed to add role")
	}
	return nil
}

func CreateRoleRepo(db *sql.DB) *RoleRepo {
	return &RoleRepo{Db: db}
}
