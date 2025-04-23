package dbtype

import (
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

func (n SystemConfig) ModelID() string {
	return n.ID.String()
}

func (n SystemConfig) TableName() string {
	return "System_Config"
}

func (n SystemConfig) KeyValMap() map[string]SystemConfig {
	return map[string]SystemConfig{
		"VNET_NAME": {
			Key:  "VNET_NAME", 
			Val:  "",
			Name: "VNetwork Name",
			Type: "text",
			ReadOnly: false,
			SortOrder: 1,
		},
		"VNET_TOKEN_EXPIRE": {
			Key:  "VNET_TOKEN_EXPIRE",
			Val:  "",
			Name: "VNetwork Token Life",
			Type: "number",
			ReadOnly: false,
			SortOrder: 2,
		},
	}
}