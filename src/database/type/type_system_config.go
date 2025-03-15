package dbtype

import (
	"os"
	"github.com/surrealdb/surrealdb.go/pkg/models"
)

type SystemConfig struct {
	ID          *models.RecordID `json:"id,omitempty"`
	Name         string `json:"name,omitempty"`
	Key   		 string `json:"key"`
	Val   		 string `json:"val"`
	Type  		 string `json:"type"` // string, choice, bool
	Version   	 string `json:"version"`
	Date_Created string `json:"date_created"`
	Date_Updated string `json:"date_updated,omitempty"`
	Options  	[]map[string]string `json:"options,omitempty"`
	ReadOnly  	bool `json:"readonly,omitempty"`
}

func (u SystemConfig) TableName() string {
	return "System_Config"
}


func (u SystemConfig) KeyValMap() map[string]SystemConfig {
	return map[string]SystemConfig{
		// "APP_KEY": {
		// 	Key:  "APP_KEY",
		// 	Val:  os.Getenv("APP_KEY"),
		// 	Name: "Application Key",
		// 	Type: "string",
		//	ReadOnly: false,
		// },
		"APP_ENV": {
			Key:  "APP_ENV",
			Val:  os.Getenv("APP_ENV"),
			Name: "Application Environment",
			Type: "string",
			ReadOnly: false,
		},
		"APP_NAME": {
			Key:  "APP_NAME",
			Val:  os.Getenv("APP_NAME"),
			Name: "Application Name",
			Type: "string",
			ReadOnly: true,
		},
		// "APP_PREFORK": {
		// 	Key:  "APP_PREFORK",
		// 	Val:  os.Getenv("APP_PREFORK"),
		// 	Name: "Application Prefork",
		// 	Type: "bool",
		//	ReadOnly: false,
		// },
		"APP_TOKEN_EXPIRE": {
			Key:  "APP_TOKEN_EXPIRE",
			Val:  os.Getenv("APP_TOKEN_EXPIRE"),
			Name: "Token Expiration (minutes)",
			Type: "number",
			ReadOnly: false,
		},
		// "CACHE_ENDPOINT": {
		// 	Key:  "CACHE_ENDPOINT",
		// 	Val:  os.Getenv("CACHE_ENDPOINT"),
		// 	Name: "Cache Endpoint",
		// 	Type: "string",
		//	ReadOnly: false,
		// },
		// "CACHE_USERNAME": {
		// 	Key:  "CACHE_USERNAME",
		// 	Val:  os.Getenv("CACHE_USERNAME"),
		// 	Name: "Cache Username",
		// 	Type: "string",
		//	ReadOnly: false,
		// },
		// "CACHE_PASSWORD": {
		// 	Key:  "CACHE_PASSWORD",
		// 	Val:  os.Getenv("CACHE_PASSWORD"),
		// 	Name: "Cache Password",
		// 	Type: "string",
		//	ReadOnly: false,
		// },
		// "STORE_ENDPOINT": {
		// 	Key:  "STORE_ENDPOINT",
		// 	Val:  os.Getenv("STORE_ENDPOINT"),
		// 	Name: "Store Endpoint",
		// 	Type: "string",
		//	ReadOnly: false,
		// },
		// "STORE_USERNAME": {
		// 	Key:  "STORE_USERNAME",
		// 	Val:  os.Getenv("STORE_USERNAME"),
		// 	Name: "Store Username",
		// 	Type: "string",
		//	ReadOnly: false,
		// },
		// "STORE_PASSWORD": {
		// 	Key:  "STORE_PASSWORD",
		// 	Val:  os.Getenv("STORE_PASSWORD"),
		// 	Name: "Store Password",
		// 	Type: "string",
		//	ReadOnly: false,
		// },
		"LOBBY_MAX_PLAYERS": {
			Key:  "LOBBY_MAX_PLAYERS",
			Val:  os.Getenv("LOBBY_MAX_PLAYERS"),
			Name: "Maximum Lobby Players",
			Type: "number",
			ReadOnly: false,
		},
		"LOBBY_MAX_PERSIST": {
			Key:  "LOBBY_MAX_PERSIST",
			Val:  os.Getenv("LOBBY_MAX_PERSIST"),
			Name: "Lobby Max Persist (minutes)",
			Type: "number",
			ReadOnly: false,
		},
		// "GAME_NAME": {
		// 	Key:  "GAME_NAME",
		// 	Val:  os.Getenv("GAME_NAME"),
		// 	Name: "Game Name",
		// 	Type: "string",
		//	ReadOnly: false,
		// },
		"GAME_PORT": {
			Key:  "GAME_PORT",
			Val:  os.Getenv("GAME_PORT"),
			Name: "Game Port",
			Type: "number",
			ReadOnly: false,
		},
		// "VNET_HOST": {
		// 	Key:  "VNET_HOST",
		// 	Val:  os.Getenv("VNET_HOST"),
		// 	Name: "Virtual Network Host",
		// 	Type: "string",
		//	ReadOnly: false,
		// },
		// "VNET_PORT": {
		// 	Key:  "VNET_PORT",
		// 	Val:  os.Getenv("VNET_PORT"),
		// 	Name: "Virtual Network Port",
		// 	Type: "number",
		//	ReadOnly: false,
		// },
		"VNET_KEY": {
			Key:  "VNET_KEY",
			Val:  os.Getenv("VNET_KEY"),
			Name: "VNetwork Key",
			Type: "string",
			ReadOnly: false,
		},
		"GAME_DOKIMAGE": {
			Key:  "GAME_DOKIMAGE",
			Val:  os.Getenv("GAME_DOKIMAGE"),
			Name: "Game Image",
			Type: "string",
			ReadOnly: false,
		},
		"GAME_COMMANDS": {
			Key:  "GAME_COMMANDS",
			Val:  os.Getenv("GAME_COMMANDS"),
			Name: "Game Commands",
			Type: "string",
			ReadOnly: false,
		},

		// "TEST_BOOL": {
		// 	Key:  "TEST_BOOL",
		// 	Val:  "true",
		// 	Name: "TEST_BOOL",
		// 	Type: "bool",
		// 	ReadOnly: true,
		// },
		// "TEST_CHOICE": {
		// 	Key:  "TEST_CHOICE",
		// 	Val:  "b",
		// 	Name: "TEST_CHOICE",
		// 	Type: "choice",
		// 	ReadOnly: false,
		// 	Options: []map[string]string {
		// 		map[string]string{"value": "a", "label": "Angora"},
		// 		map[string]string{"value": "b", "label": "Bngora"},
		// 		map[string]string{"value": "c", "label": "Cngora"},
		// 	},
		// },
	}
}