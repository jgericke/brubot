package main

import (
	"brubot/config"
	"brubot/internal/target"
	"fmt"

	"log"
)

func main() {

	bruConfig := new(config.Parameters)
	globalConfig := new(config.GlobalConfig)
	targetConfig := new(config.TargetConfig)
	targetEndpoint := new(target.Endpoint)

	if err := bruConfig.Init(); err != nil {
		log.Fatal("Failure reading target config from config.yaml %w", err)
	}
	if err := bruConfig.ParseConfig(globalConfig, targetConfig); err != nil {
		log.Fatal("Unable to parse target config %w", err)
	}

	targetEndpoint.Init(*globalConfig, *targetConfig, 9)
	err := targetEndpoint.Fixtures()

	if err != nil {
		log.Fatal("Failure extracting fixtures from target: ", err)
	}

	for fixture, value := range targetEndpoint.Round.Fixtures {
		fmt.Println(fixture, value)
	}
	/*
		source, err := sources.SourceRugbyVision()
		if err != nil {
			log.Fatal("Failure retrieving from source: ", err)
		} else {
			log.Printf("round.RoundFixtures[0]: %s", source.SourceRound.RoundFixtures[0].TeamnOne)
		}

		log.Printf("%+v\n", round)

			for k, v := range round.RoundFixtures {
				fmt.Println(k, v)
			}
	*/
}
