package entities

type User struct {
	Entity
	Key          string   `json:"_key,omitempty"`
	Email        string   `json:"email"`
	PasswordHash string   `json:"password_hash"`
	Name         string   `json:"name"`
	Vorname      string   `json:"vorname"`
	Roles        []string `json:"roles"`
	// Address fields...
}
