package helpers

import (
	"bufio"
	"encoding/csv"
	"errors"
	"io"
	"os"
	"strconv"
	"time"
)

// FindRoundNumber determines the current round number being played based on execution date and source csv.
func FindRoundNumber() (int, error) {

	var roundNumber, roundYear, roundStartMonth, roundStartDay, roundEndMonth, roundEndDay int

	// Open csv containing round dates.
	roundDatesFile, err := os.Open("round_dates.csv")
	if err != nil {
		return 0, err
	}

	// Assume first line in CSV source are headers, read to clear them.
	roundDatesCsv := csv.NewReader(bufio.NewReader(roundDatesFile))
	if _, err := roundDatesCsv.Read(); err != nil {
		return 0, err
	}

	for {
		roundDatesEntry, err := roundDatesCsv.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return 0, err
		}

		// Assign csv fields to expecting columns orders as roundYear,
		// roundStartMonth, roundStartDay, roundEndMonth, roundEndDay, roundName
		if roundYear, err = strconv.Atoi(roundDatesEntry[0]); err != nil {
			return 0, err
		}
		if roundStartMonth, err = strconv.Atoi(roundDatesEntry[1]); err != nil {
			return 0, err
		}
		if roundStartDay, err = strconv.Atoi(roundDatesEntry[2]); err != nil {
			return 0, err
		}
		if roundEndMonth, err = strconv.Atoi(roundDatesEntry[3]); err != nil {
			return 0, err
		}
		if roundEndDay, err = strconv.Atoi(roundDatesEntry[4]); err != nil {
			return 0, err
		}
		if roundNumber, err = strconv.Atoi(roundDatesEntry[5]); err != nil {
			return 0, err
		}

		// Convert round starting and ending year, month and day to Date for comparisson.
		roundStart := time.Date(roundYear, time.Month(roundStartMonth), roundStartDay, int(0), int(0), int(0), int(0), time.UTC)
		roundEnd := time.Date(roundYear, time.Month(roundEndMonth), roundEndDay, int(23), int(59), int(59), int(0), time.UTC)

		// If time at execution is between round start and end times, retrieve roundName.
		if time.Now().After(roundStart) && time.Now().Before(roundEnd) {
			return roundNumber, nil
		}

	}

	// If no round name is found, we are truly boned.
	return 0, errors.New("Failed to find round name")

}
