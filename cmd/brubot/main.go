package main

import (
	"brubot/config"
	"brubot/internal/helpers"
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
	//var predictions map[string]int

	target := new(target.Target)
	//sources := new(sources.Sources)
	//predictions = make(map[string]int)

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

	// Initialize target and get fixutres
	target.Init(globalConfig, targetConfig)

	if err = target.Authenticate(); err != nil {
		helpers.Logger.Fatal("A failure occurred authenticating to endpoint: ", err)
	}

	// debug
	helpers.Logger.Infof("Source[0] name: %s", sourcesConfig.Sources[0].Name)

	// Get results for previous roundID
	err = target.Results(roundID - 1)
	if err != nil {
		helpers.Logger.Fatal("Failure extracting results from endpoint: ", err)
	}

}
