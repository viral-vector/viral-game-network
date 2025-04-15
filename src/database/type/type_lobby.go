package dbtype

import (
	"github.com/surrealdb/surrealdb.go/pkg/models"
)

type Lobby struct {
	ID    *models.RecordID `json:"id,omitempty"`
	Name  string `json:"name" form:"name,label:Name,type:string,required:true"`
	Guid  string `json:"guid,omitempty"`
	Date_Created string `json:"date_created"`
	Date_Updated string `json:"date_updated"`
	Private bool `json:"private" form:"private,label:Private,type:bool,required:false"`
	Code string `json:"code" form:"code,label:Code,type:string,required:false"`
	Lobby_Application *Application `json:"lobby_application,omitempty" form:"lobby_application,label:Application,type:choice,required:true"`
	Lobby_Host *User `json:"lobby_host,omitempty"`
	Lobby_Users []*User `json:"lobby_users,omitempty"`
	Lobby_Server *Server `json:"lobby_server,omitempty"` 
}

func (n Lobby) ModelID() string {
	return n.ID.String()
}

func (n Lobby) TableName() string {
	return "Lobby"
}