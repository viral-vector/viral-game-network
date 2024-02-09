package dbtype

type Lobby struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name"`
	Date_Created string `json:"date_created"`
	Date_Updated string `json:"date_updated"`
	Configs string `json:"config"`
	Private bool `json:"private"`
	Code string `json:"code"`
	Lobby_Host *User `json:"lobby_host"`
	Lobby_Users []*LobbyUser `json:"lobby_users"`
	Server *Server `json:"server"` 
}