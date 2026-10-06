package dbtype

import (
	"encoding/json"
	"fmt"
	"github.com/surrealdb/surrealdb.go/pkg/models"
	"strings"
)

type Application struct {
	ID                *models.RecordID `json:"id,omitempty" form:"id,label:ID,type:text,readonly:true"`
	Name              string           `json:"name,omitempty" form:"name,label:Name,type:text,required:true"`
	Guid              string           `json:"guid,omitempty" form:"guid,label:Guid,type:text,required:true"`
	Group             string           `json:"group,omitempty" form:"group,label:Group,type:text,required:false"`
	Date_Created      string           `json:"date_created,omitempty"`
	Date_Updated      string           `json:"date_updated,omitempty"`
	Image             string           `json:"image,omitempty" form:"image,label:Image,type:text,required:true"`
	Version           string           `json:"version,omitempty" form:"version,label:Version,type:text,required:false"`
	Port              string           `json:"port,omitempty" form:"port,label:Port,type:number,required:false"`
	Command           string           `json:"command,omitempty" form:"command,label:Command,type:text,required:false"`
	Lobby_Max_Players string           `json:"lobby_max_players,omitempty" form:"lobby_max_players,label:Lobby Players,type:number,required:true"`
	Lobby_Max_Persist string           `json:"lobby_max_persist,omitempty" form:"lobby_max_persist,label:Lobby Persist (minutes),type:number,required:true"`
}

func (n Application) ModelID() string {
	if n.ID == nil {
		return ""
	}
	return n.ID.String()
}

func (n Application) TableName() string {
	return "Application"
}

func (n *Application) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		n.ID = nil
		return nil
	}
	parts := strings.SplitN(string(text), ":", 2)
	if len(parts) != 2 || parts[0] != n.TableName() || parts[1] == "" {
		return fmt.Errorf("invalid application record ID: %q", text)
	}
	n.ID = &models.RecordID{ID: parts[1], Table: n.TableName()}
	return nil
}

// UnmarshalJSON accepts both form record IDs and complete API representations.
func (n *Application) UnmarshalJSON(data []byte) error {
	if len(data) > 0 && data[0] == '"' {
		var id string
		if err := json.Unmarshal(data, &id); err != nil {
			return err
		}
		return n.UnmarshalText([]byte(id))
	}
	type applicationJSON Application
	return json.Unmarshal(data, (*applicationJSON)(n))
}
