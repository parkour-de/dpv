package entities

// VorstandUser represents a minimal user for Vorstand display
type VorstandUser struct {
	Key       string `json:"_key"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
}
