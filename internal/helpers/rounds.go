package helpers

import (
	"database/sql"
	"time"
)

// GetCurrentRound retrieves the current round number being played based on execution date
func GetCurrentRound(db *sql.DB) (int, error) {

	var roundID int

	err := db.QueryRow("SELECT find_round_id_by_date($1);", time.Now()).Scan(&roundID)

	if err != nil {
		return 0, err
	}

	return roundID, nil

}
