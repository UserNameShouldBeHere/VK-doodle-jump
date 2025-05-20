package domain

type UserHeader struct {
	VkId   int    `json:"vkid"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type UserRating struct {
	Name  string `json:"name"`
	Score int    `json:"score"`
}

type UserRatingWithPos struct {
	Users      []UserRating `json:"users"`
	CurrentPos int          `json:"current_pos"`
}

// type LeagueTopUsers struct {
// 	League string       `json:"league"`
// 	Users  []UserRating `json:"users"`
// }
