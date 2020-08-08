package main

import (
	"brubot/config"
	"brubot/internal/helpers"
	"brubot/internal/sources"
	"brubot/internal/target"
	"database/sql"
)

func main() {

	var err error

	var globalConfig config.GlobalConfig
	var targetConfig config.TargetConfig
	var sourcesConfig config.SourcesConfig

	var db *sql.DB
	var roundID int
	var previousRoundID int
	var margins map[string]int

	target := new(target.Target)
	sources := new(sources.Sources)
	margins = make(map[string]int)

	// Initialise brubot
	helpers.LoggerInit()
	globalConfig, targetConfig, sourcesConfig, err = helpers.ConfigInit()
	if err != nil {
		helpers.Logger.Panic("A failure occurred initialising config: ", err)
	}

	db, err = helpers.DBInit(globalConfig)
	if err != nil {
		helpers.Logger.Panic("A failure occurred initialising database connection: ", err)
	}

	defer db.Close()

	roundID, err = helpers.GetCurrentRound(db)
	if err != nil {
		helpers.Logger.Panic("A failure occurred determining roundID: ", err)
	}
	previousRoundID = roundID - 1

	// Initialize target and get fixutres
	target.Init(globalConfig, targetConfig)

	if err = target.Authenticate(); err != nil {
		helpers.Logger.Fatal("A failure occurred authenticating to target: ", err)
	}

	// Gets results from previous rounds fixtures and update db
	err = target.Results(previousRoundID, db)
	if err != nil {
		helpers.Logger.Fatal("Failure extracting results from target: ", err)
	}

	// Gets current fixtures for this round
	err = target.Fixtures(roundID)
	if err != nil {
		helpers.Logger.Fatal("Failure extracting fixtures from target: ", err)
	}

	// Initialize sources and retrieve predictions
	sources.Init(globalConfig, sourcesConfig)

	// Retrieve predicted margins for all fixtures in a round, per source
	err = sources.Predictions(roundID, db)
	if err != nil {
		helpers.Logger.Fatal("A failure occurred retrieving predictions from source(s): ", err)
	}

	// Generate weighted margin predictions for all sources
	margins, err = sources.Margins(roundID)
	if err != nil {
		helpers.Logger.Error("A failure occurred generating predictions: ", err)
	}

	// Submit generated margins to target
	err = target.Predictions(margins)
	if err != nil {
		helpers.Logger.Fatal("A failure occurred submitting predictions: ", err)
	}

}
