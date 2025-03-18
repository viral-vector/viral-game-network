package dbtype

import (
	"github.com/surrealdb/surrealdb.go/pkg/models"
)

type Application struct {
	ID      *models.RecordID `json:"id,omitempty"`
	Name    string `json:"name,omitempty" form:"name,label:Name,type:string,required:true"`
	Guid    string `json:"guid,omitempty" form:"guid,label:Guid,type:string,required:true"`
	Group   string `json:"group,omitempty" form:"group,label:Group,type:string,required:false"`
	Date_Created string `json:"date_created,omitempty"`
	Date_Updated string `json:"date_updated,omitempty"`
	Image string `json:"image,omitempty" form:"image,label:Image,type:string,required:true"`
	Version string `json:"version,omitempty" form:"version,label:Version,type:string,required:false"`
	Port string `json:"port,omitempty" form:"port,label:Port,type:number,required:false"`
	Command string `json:"command,omitempty" form:"command,label:Command,type:string,required:false"`
	Lobby_Max_Players string `json:"lobby_max_players,omitempty" form:"lobby_max_players,label:Lobby Players,type:number,required:true"`
	Lobby_Max_Persist string `json:"lobby_max_persist,omitempty" form:"lobby_max_persist,label:Lobby Persist (minutes),type:number,required:true"`
}

func (u Application) TableName() string {
	return "Application"
}