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
	Options  	 []map[string]string `json:"options,omitempty"`
	ReadOnly  	 bool `json:"readonly,omitempty"`
	SortOrder    int32 `json:"readonly,omitempty"`
}

func (u SystemConfig) TableName() string {
	return "System_Config"
}

func (u SystemConfig) KeyValMap() map[string]SystemConfig {
	return map[string]SystemConfig{
		 "APP_ENV": {
			Key:  "APP_ENV",
			Val:  os.Getenv("APP_ENV"),
			Name: "VNetwork Client Environment",
			Type: "string",
			ReadOnly: false,
			SortOrder: 0,
		},
		"APP_NAME": {
			Key:  "APP_NAME",
			Val:  os.Getenv("APP_NAME"),
			Name: "VNetwork Client Name",
			Type: "string",
			ReadOnly: true,
			SortOrder: 1,
		},
		"APP_TOKEN_EXPIRE": {
			Key:  "APP_TOKEN_EXPIRE",
			Val:  os.Getenv("APP_TOKEN_EXPIRE"),
			Name: "VNetwork Client Token Expiration",
			Type: "number",
			ReadOnly: false,
			SortOrder: 2,
		},
		"LOBBY_MAX_PLAYERS": {
			Key:  "LOBBY_MAX_PLAYERS",
			Val:  os.Getenv("LOBBY_MAX_PLAYERS"),
			Name: "Maximum Lobby Players",
			Type: "number",
			ReadOnly: false,
			SortOrder: 3,
		},
		"LOBBY_MAX_PERSIST": {
			Key:  "LOBBY_MAX_PERSIST",
			Val:  os.Getenv("LOBBY_MAX_PERSIST"),
			Name: "Lobby Max Persist (minutes)",
			Type: "number",
			ReadOnly: false,
			SortOrder: 4,
		},

		// "TEST_TEXT": {
		// 	Key:  "TEST_TEXT",
		// 	Val:  "Hello World",
		// 	Name: "TEST_TEXT",
		// 	Type: "string/number",
		// 	ReadOnly: false,
		// },
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