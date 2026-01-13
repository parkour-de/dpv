package entities

type Census struct {
	Entity
	Year        int         `json:"year"`
	MemberCount int         `json:"memberCount"`
	Members     []MemberRow `json:"members"`
}

type MemberRow struct {
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Gender    string `json:"gender"`
	BirthYear int    `json:"birthYear"`
}
