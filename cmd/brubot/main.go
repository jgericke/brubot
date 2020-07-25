package main

import (
	"brubot/config"
	"brubot/internal/helpers"
	"brubot/internal/sources"
	"database/sql"
)

func main() {

	var err error
	//var source sources.Source

	var globalConfig config.GlobalConfig
	var targetConfig config.TargetConfig
	var sourcesConfig config.SourcesConfig
	var db *sql.DB
	var roundID int

	// initialise
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

	helpers.Logger.Debugf("targetConfig: %s", targetConfig.Client.UserAgent)

	roundID, err = helpers.GetCurrentRound(db)

	if err != nil {
		helpers.Logger.Panic("A failure occurred determining roundID: ", err)
	}

	// debug
	helpers.Logger.Debug("roundID: ", roundID)

	sources := new(sources.Sources)
	sources.Init(globalConfig, sourcesConfig)

	err = sources.Predictions(roundID)

	if err != nil {
		helpers.Logger.Fatal("A failure error occurred retrieving predictions from source(s): ", err)
	}

	err = sources.Update(db)
	if err != nil {
		helpers.Logger.Fatal("A failure occurred updating source(s) ", err)
	}

}
