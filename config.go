package main

import (
	"errors"
	"github.com/BurntSushi/toml"
	"os"
	"path"
)

const configSubdirName = "rcon"
const configFileName = "config.toml"

const defaultFileContent = `# RCON Config
# You can add servers, removing the # indicating comments, as below
# [someservername]
# hostname = 172.0.0.1
# port = 27015
# password = somepassword

`

type configMap map[string]server

type server struct {
	Host     string `toml:"hostname"` // referred to as hostname in config file for backwards compatibility
	Port     int    `toml:"port"`
	Password string `toml:"password"`
}

func readConfig() (configMap, error) {
	// If config directory doesn't exist, make it
	configDirPath, err := os.UserConfigDir()
	if err != nil {
		return nil, err
	}
	configSubdirPath := path.Join(configDirPath, configSubdirName)
	err = os.Mkdir(configSubdirPath, 0755)
	if err != nil {
		if !errors.Is(err, os.ErrExist) {
			return nil, err
		}
		// if it already exists no need to do anything
	}
	configFilePath := path.Join(configSubdirPath, configFileName)
	configDataBytes, err := os.ReadFile(configFilePath)
	var configData string
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			err := os.WriteFile(configFilePath, []byte(defaultFileContent), 0644)
			if err != nil {
				return nil, err
			}
			configData = defaultFileContent
		} else {
			return nil, err
		}
	} else {
		configData = string(configDataBytes)
	}

	// Decode config info
	var config configMap
	_, err = toml.Decode(configData, &config)
	if err != nil {
		return nil, err
	}
	return config, nil
}
