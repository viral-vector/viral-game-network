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
			Name: "VNetwork Environment",
			Type: "string",
			ReadOnly: true,
			SortOrder: 0,
		},
		"VNET_NAME": {
			Key:  "VNET_NAME",
			Val:  "",
			Name: "VNetwork Name",
			Type: "string",
			ReadOnly: false,
			SortOrder: 1,
		},
		"APP_TOKEN_EXPIRE": {
			Key:  "APP_TOKEN_EXPIRE",
			Val:  "",
			Name: "VNetwork Token Life",
			Type: "number",
			ReadOnly: false,
			SortOrder: 2,
		},
	}
}