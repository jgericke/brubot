package helpers

import (
	"brubot/config"
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// ConfigInit invokes reading and parsing of config file parameters
func ConfigInit() (config.GlobalConfig, config.TargetConfig, config.SourcesConfig, error) {

	bruConfig := new(config.Parameters)
	globalConfig := new(config.GlobalConfig)
	targetConfig := new(config.TargetConfig)
	sourcesConfig := new(config.SourcesConfig)

	if err := bruConfig.Init(); err != nil {
		return *globalConfig, *targetConfig, *sourcesConfig, err
	}

	if err := bruConfig.ParseConfig(globalConfig, targetConfig, sourcesConfig); err != nil {
		return *globalConfig, *targetConfig, *sourcesConfig, err
	}

	return *globalConfig, *targetConfig, *sourcesConfig, nil
}

// DBInit initialises database connectivity
func DBInit(globalConfig config.GlobalConfig) (*sql.DB, error) {

	db, err := sql.Open("postgres", fmt.Sprintf(
		"host=%s "+
			"port=%d "+
			"user=%s "+
			"password=%s "+
			"dbname=%s "+
			"sslmode=%s",
		globalConfig.DB.Host,
		globalConfig.DB.Port,
		globalConfig.DB.User,
		globalConfig.DB.Password,
		globalConfig.DB.Name,
		globalConfig.DB.SSLMode),
	)

	if err != nil {
		return nil, err
	}

	err = db.Ping()

	if err != nil {
		return nil, err
	}

	return db, nil

}
