package dbtype

type LobbyUser struct {
	ID    string `json:"id,omitempty"`
	User  User `json:"user"`
	Date_Created string `json:"date_created"`
}