package target

import (
	"brubot/internal/helpers"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/gocolly/colly/v2"
	"github.com/lib/pq"
)

// Results returns the completed fixture results for a specified round
func (t *Target) Results(previousRoundID int, db *sql.DB) error {

	// roundID *should* typically be currentRound - 1 for retrieving
	// the previous rounds fixture results
	t.PreviousRound.id = previousRoundID
	if err := t.getResults(); err != nil {
		return err
	}

	if err := t.updateResults(db); err != nil {
		return err
	}

	return nil

}

// getResults uses a pre-authenticated client to retrieve fixture results from a specified round
func (t *Target) getResults() error {

	var err error
	var margin int
	var winner string

	t.Client.collector.OnHTML(fmt.Sprintf(t.Client.parser.results["attr_onhtml"], t.PreviousRound.id), func(e *colly.HTMLElement) {

		e.ForEach(t.Client.parser.results["attr_fixture"], func(_ int, cl *colly.HTMLElement) {

			// Split leftTeam and rightTeam based into array using a known delimeter for
			// team name differentiation. We are assuming that there will always be 2 elements
			// in the array returned from the split, 0 being leftTeam and 1 being rightTeam.
			leftTeam := helpers.CleanName(strings.Split(
				cl.Attr(t.Client.parser.results["attr_t_teams"]),
				t.Client.parser.results["attr_t_teams_delimiter"],
			)[0])
			rightTeam := helpers.CleanName(strings.Split(
				cl.Attr(t.Client.parser.results["attr_t_teams"]),
				t.Client.parser.results["attr_t_teams_delimiter"],
			)[1])
			// If the results parser returns a draw then set margin to 0 and winner to draw
			if strings.EqualFold(cl.ChildText(t.Client.parser.results["attr_t_results"]), t.Client.parser.results["attr_t_draw"]) {
				margin = 0
				winner = "draw"
			} else {
				// Split the winner and margin based on a known delimeter for winner team name
				// and margin
				winner = helpers.CleanName(strings.Split(
					cl.ChildText(t.Client.parser.results["attr_t_results"]),
					t.Client.parser.results["attr_t_winner_delimiter"],
				)[0])
				// Sets and converts margin from string to int
				if marginResult, marginErr := strconv.Atoi(
					strings.Split(
						cl.ChildText(t.Client.parser.results["attr_t_results"]),
						t.Client.parser.results["attr_t_winner_delimiter"])[1],
				); marginErr != nil {
					err = marginErr
				} else {
					margin = marginResult
				}
			}
			// Append extracted result to previousRounds Results
			t.PreviousRound.Results = append(t.PreviousRound.Results, result{
				leftTeam:  leftTeam,
				rightTeam: rightTeam,
				winner:    winner,
				margin:    margin,
			})

			helpers.Logger.Debugf("Result has been retrieved, leftTeam: %s, rightTeam: %s, winner: %s, margin: %d",
				leftTeam,
				rightTeam,
				winner,
				margin,
			)

		})

	})

	// If the login attribute is detected in the response body, authentication has somehow failed
	t.Client.collector.OnHTML(t.Client.parser.login["attr_login"], func(e *colly.HTMLElement) {
		err = errors.New("An error occurred during results retrieval, client is not authenticated")
		return
	})

	// Client error has occurred attempting .Visit
	t.Client.collector.OnError(func(r *colly.Response, resError error) {
		helpers.Logger.Errorf("An error occurred results retrieval, client response %+v URL %s error %s", r, r.Request.URL, resError)
		err = fmt.Errorf("An error occurred during results retrieval, client response %+v URL %s error %s", r, r.Request.URL, resError)
		return
	})

	// Client request to the targets results endpoint based on the results roundID
	t.Client.collector.Visit(fmt.Sprintf(t.Client.config.urls["results"], t.PreviousRound.id, t.PreviousRound.id))

	return err

}

// updateResults writes retrieved results for a previous round of fixtures to backend
func (t *Target) updateResults(db *sql.DB) error {

	// backend updates could become their own abstraction as I only use CopyIn
	// and do some level of duplicate checking to prevent duplicate prediction/results updates

	// Temporary ID for duplicate RESULTS checking
	var tmpID int
	// Create an empty context for results update
	sqlCtx := context.Background()
	// Create transaction
	sqlTxn, err := db.BeginTx(sqlCtx, nil)
	if err != nil {
		return err
	}
	// prepare results update with COPY FROM (table, fields[..])
	sqlStmt, err := sqlTxn.Prepare(pq.CopyIn("results", "round_id", "leftteam", "rightteam", "winner", "margin"))
	if err != nil {
		return err
	}

	// looking *very* familiar...will abstract when it makes sense.
	helpers.Logger.Debug("Results update is emminent, hold tight...")

	for idx := range t.PreviousRound.Results {
		// Same "Ugly Check" as source prediction update
		sqlPrdExists := db.QueryRowContext(sqlCtx,
			"SELECT id FROM results WHERE round_id=$1"+
				"AND leftteam=$2 AND rightteam=$3"+
				"AND winner=$4 AND margin=$5",
			t.PreviousRound.id,
			t.PreviousRound.Results[idx].leftTeam,
			t.PreviousRound.Results[idx].rightTeam,
			t.PreviousRound.Results[idx].winner,
			t.PreviousRound.Results[idx].margin).Scan(&tmpID)

		switch {
		case sqlPrdExists == sql.ErrNoRows:
			// ErrNoRows means we are good to go, execute CopyIn
			// with PreviousRound id and results
			_, err = sqlStmt.Exec(
				t.PreviousRound.id,
				t.PreviousRound.Results[idx].leftTeam,
				t.PreviousRound.Results[idx].rightTeam,
				t.PreviousRound.Results[idx].winner,
				t.PreviousRound.Results[idx].margin,
			)
			if err != nil {
				return err
			}
		case sqlPrdExists != nil:
			// Error occurred during query
			return sqlPrdExists
		default:
			helpers.Logger.Debugf("Results update omitted as record already exists with ID: %d", tmpID)
		}

	}
	err = sqlStmt.Close()
	if err != nil {
		return err
	}
	err = sqlTxn.Commit()
	if err != nil {
		return err
	}

	helpers.Logger.Debug("Results update completed sans incidents")

	return nil

}
