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
	var predictions map[string]int

	target := new(target.Target)
	sources := new(sources.Sources)
	predictions = make(map[string]int)

	// Initialise thah brubot things
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

	// Initialize target and get fixutres
	target.Init(globalConfig, targetConfig)

	if err = target.Authenticate(); err != nil {
		helpers.Logger.Fatal("A failure occurred authenticating to endpoint: ", err)
	}

	// Get fixtures from roundID
	err = target.Fixtures(roundID)
	if err != nil {
		helpers.Logger.Fatal("Failure extracting fixtures from endpoint: ", err)
	}

	// Initialize sources and retrieve predictions
	sources.Init(globalConfig, sourcesConfig)

	// Retrieve predictions from all sources
	err = sources.Predictions(roundID)
	if err != nil {
		helpers.Logger.Fatal("A failure occurred retrieving predictions from source(s): ", err)
	}

	// Update db with most recent predictions
	err = sources.Update(db)
	if err != nil {
		helpers.Logger.Fatal("A failure occurred updating source(s) ", err)
	}

	// Generate weighted predictions for all sources
	predictions, err = sources.Generate(roundID)
	if err != nil {
		helpers.Logger.Error("A failure occurred generating predictions: ", err)
	}

	// Submit generated predictions to target
	err = target.Predictions(predictions)
	if err != nil {
		helpers.Logger.Fatal("A failure occurred submitting predictions: ", err)
	}

}
