package dbtype

type SystemEvent struct {
	ID           string `json:"id,omitempty"`
	Date_Created string `json:"date_created"`
	Message      string `json:"message"`
	Severity     string `json:"severity"`
}
