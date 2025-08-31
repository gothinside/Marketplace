package roleutils

import (
	"database/sql"
	"fmt"
)

func HasRole(db *sql.DB, UserID int, role string) bool {
	var userID, roleID int
	err := db.QueryRow("SELECT role_id FROM roles WHERE role_name = $1", role).Scan(&roleID)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	fmt.Println(UserID, roleID)
	err = db.QueryRow("SELECT user_id FROM user_role WHERE user_id = $1 and role_id = $2", UserID, roleID).Scan(&userID)
	if err != nil {
		fmt.Println(err.Error())
		return false
	}
	return true
}
