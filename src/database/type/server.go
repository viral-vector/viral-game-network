package dbtype

type Server struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name"`
	Guid 	string `json:"guid"`
	Date_Created string `json:"date_created"`
	Date_Updated string `json:"date_updated"`
	Address string `json:"address"`
	Configs string `json:"config"`
	Lobby *Lobby `json:"lobby"`
}