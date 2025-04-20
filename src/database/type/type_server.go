package dbtype

import (
	"github.com/surrealdb/surrealdb.go/pkg/models"
)

type Server struct {
	ID      *models.RecordID `json:"id,omitempty" form:"id,label:ID,type:text,readonly:true"`
	Name    string `json:"name" form:"name,label:Name,type:text,readonly:true"`
	Guid 	string `json:"guid" form:"guid,label:GUID,type:text,readonly:true"`
	Status  string `json:"status" form:"status,label:Status,type:text,readonly:true"`
	Date_Created string `json:"date_created" form:"date_created,label:Date Created,type:text,readonly:true"`
	Date_Updated string `json:"date_updated" form:"date_updated,label:Date Updated,type:text,readonly:true"`
	Address string `json:"address" form:"address,label:Address,type:text,readonly:true"`
	Port int32 `json:"port" form:"port,label:Port,type:number,readonly:true"`
	Lobby *Lobby `json:"lobby,omitempty" form:"lobby,label:Lobby,type:text,readonly:true"`
}

func (n Server) ModelID() string {
	if n.ID == nil {
		return ""
	}
	return n.ID.String()
}

func (n Server) TableName() string {
	return "Server"
}