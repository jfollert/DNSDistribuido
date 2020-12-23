package config

import (
	"log"
	"encoding/json"
	"io/ioutil"
)

type NodeInfo struct {
	Id   string `json:"id"`
	Ip   string `json:"ip"`
	Port string `json:"port"`
}

type Config struct {
	DNS []NodeInfo `json:"DNS"`
	Broker NodeInfo   `json:"Broker"`
}

func GenConfig(file string) *Config{
    configFile, err := ioutil.ReadFile(file)
    if err != nil {
		log.Fatalf(err.Error())
	}

	conf:= new(Config)
	json.Unmarshal(configFile, &conf)

	return conf
}