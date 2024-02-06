package dbtype

type Server struct {
	ID      string `json:"id,omitempty"`
	Name    string `json:"name"`
	Address string `json:"address"`
}