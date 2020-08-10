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
	//var margins map[string]int

	target := new(target.Target)
	sources := new(sources.Sources)
	//margins = make(map[string]int)

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

	err = sources.Weights(previousRoundID, db)
	if err != nil {
		helpers.Logger.Fatal("Failure calculating target weights from previous rounds results: ", err)
	}

	helpers.Logger.Debugf("%s", sourcesConfig.Sources[0].Name)
}
