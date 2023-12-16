package config_parser

import (
	"encoding/json"
	"go.uber.org/zap"
	"os"
	"time"

	. "SimulatorMAS/internal/logger"
)

var Cfg *ConfigFile

func init() {
	configFile, err := os.ReadFile("./config/config.json")
	if err != nil {
		Log.Fatal("Can't find config file", zap.Error(err))
	}

	err = json.Unmarshal(configFile, &Cfg)
	if err != nil {
		Log.Fatal("Can't read config file", zap.Error(err))
	}

	Cfg.Greens.MaxLifeTime *= time.Second
	Cfg.Greens.IntervalLunchTime *= time.Second
	Cfg.Greens.IntervalAddLifeTime *= time.Second

	Cfg.Rodent.MaxLifeTime *= time.Second
	Cfg.Rodent.IntervalLunchTime *= time.Second
	Cfg.Rodent.IntervalAddLifeTime *= time.Second
}

type ConfigFile struct {
	Greens struct {
		Count               int           `json:"count"`
		IntervalAddLifeTime time.Duration `json:"interval_add_life_time"`
		IntervalLunchTime   time.Duration `json:"interval_lunch_time"`
		MaxLifeTime         time.Duration `json:"max_life_time"`
	} `json:"greens"`
	Rodent struct {
		Count               int           `json:"count"`
		IntervalAddLifeTime time.Duration `json:"interval_add_life_time"`
		IntervalLunchTime   time.Duration `json:"interval_lunch_time"`
		MaxLifeTime         time.Duration `json:"max_life_time"`
		NumberOfCorpses     int           `json:"number_of_corpses"`
		MaxHungryToDie      int           `json:"max_hungry_to_die"`
	} `json:"rodent"`
	Broker struct {
		Logging                  bool   `json:"logging"`
		RetryCount               int    `json:"retry_count"`
		BrokerAddress            string `json:"broker_address"`
		QueueCorpsesName         string `json:"queue_corpses_name"`
		QueueGreensReqToBornName string `json:"queue_greens_req_to_born_name"`
		QueueRodentReqToBornName string `json:"queue_rodent_req_to_born_name"`
	}
}
