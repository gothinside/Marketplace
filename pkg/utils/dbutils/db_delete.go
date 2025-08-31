package dbutils

import (
	"database/sql"
	"fmt"
)

type SqlDbHandler struct {
	Db *sql.DB
}

func (SDH *SqlDbHandler) DeleteAllFromTable(table string) error {
	q := fmt.Sprintf("DELETE FROM  %s", table)
	_, err := SDH.Db.Query(q)
	if err != nil {
		return err
	}
	return nil
}
