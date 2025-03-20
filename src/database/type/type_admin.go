package dbtype

import (
	"github.com/surrealdb/surrealdb.go/pkg/models"
)

type Admin struct {
	ID *models.RecordID `json:"id,omitempty"`
	Name string `json:"name,omitempty" form:"name,label:Name,type:string,required:true"`
	Password string `json:"password,omitempty" form:"password,label:Password,type:string,required-create:true"`
	Email string `json:"email,omitempty" form:"email,label:Email,type:string,required:true"`
	Phone string `json:"phone,omitempty" form:"phone,label:Phone,type:string,required:true"`
	Date_Created string `json:"date_created,omitempty"`
	Date_Updated string `json:"date_updated,omitempty"`
	Date_LastLogin string `json:"date_last_login,omitempty"`
}

func (u Admin) TableName() string {
	return "Admin"
}