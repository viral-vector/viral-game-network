package dbtype

type Server struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name"`
	Guid 	string `json:"guid"`
	Status  string `json:"status"`
	Date_Created string `json:"date_created"`
	Date_Updated string `json:"date_updated"`
	Address string `json:"address"`
	Port int32 `json:"port"`
	Configs string `json:"config"`
}