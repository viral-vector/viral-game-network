package dbtype

type LobbyMessage struct {
	ID    string `json:"id,omitempty"`
	User_ID  string `json:"user_id"`
	Date_Created string `json:"date_created"`
	Body string `json:"body"`
}