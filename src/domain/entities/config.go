package entities

type Config struct {
	Entity
	Sequences map[string]int `json:"sequences"`
}
