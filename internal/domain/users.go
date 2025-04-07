package domain

type UserRating struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
}

type LeagueTopUsers struct {
	League string       `json:"league"`
	Users  []UserRating `json:"users"`
}
