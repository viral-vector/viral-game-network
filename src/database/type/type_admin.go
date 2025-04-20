package dbtype

import (
	"github.com/surrealdb/surrealdb.go/pkg/models"
)

type Admin struct {
	ID *models.RecordID `json:"id,omitempty" form:"id,label:ID,type:text,readonly:true"`
	Name string `json:"name,omitempty" form:"name,label:Name,type:text,required:true"`
	Password string `json:"password,omitempty" form:"password,label:Password,type:text,required-create:true"`
	Email string `json:"email,omitempty" form:"email,label:Email,type:text,required:true"`
	Phone string `json:"phone,omitempty" form:"phone,label:Phone,type:text,required:true"`
	Date_Created string `json:"date_created,omitempty"`
	Date_Updated string `json:"date_updated,omitempty"`
	Date_LastLogin string `json:"date_last_login,omitempty"`
}

func (n Admin) ModelID() string {
	if n.ID == nil {
		return ""
	}
	return n.ID.String()
}

func (n Admin) TableName() string {
	return "Admin"
}