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
	Version string `json:"version,omitempty" form:"version,label:Version,type:string,required:true"`
	Command string `json:"command,omitempty" form:"command,label:Command,type:string,required:false"`
}

func (u Application) TableName() string {
	return "Application"
}