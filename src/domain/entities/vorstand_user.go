package entities

// VorstandUser represents a minimal user for Vorstand display
type VorstandUser struct {
	Key                      string `json:"_key"`
	Firstname                string `json:"firstname"`
	Lastname                 string `json:"lastname"`
	Email                    string `json:"email,omitempty"`
	AuthorizedRepresentative bool   `json:"authorizedRepresentative"`
	Function                 string `json:"function,omitempty"`
}
