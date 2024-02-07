package dbtype

type User struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name,omitempty"`
	Guid    string `json:"guid,omitempty"`
	Date_Created string `json:"date_created"`
	Date_Updated string `json:"date_updated"`
	Date_LastLogin string `json:"date_last_login"`
	Configs string `json:"config"`
}