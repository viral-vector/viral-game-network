package dbtype

import (
	"github.com/surrealdb/surrealdb.go/pkg/models"
)

type User struct {
	ID      *models.RecordID `json:"id,omitempty"`
	Name    string `json:"name,omitempty" form:"name,label:Name,type:string,required:true"`
	Guid    string `json:"guid,omitempty"`
	Date_Created string `json:"date_created,omitempty"`
	Date_Updated string `json:"date_updated,omitempty"`
	Date_LastLogin string `json:"date_last_login,omitempty"`
}

func (u User) TableName() string {
	return "User"
}

func (u User) SetGUID(guid string) {
	u.Guid = guid
}